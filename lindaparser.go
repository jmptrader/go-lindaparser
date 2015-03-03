package lindaparser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	fieldUsername   = "asdf"
	fieldPassword   = "fdsa"
	requestTimeout  = 15
	urlExamOverview = "https://linda.hs-heilbronn.de/qisstudent/rds?state=change&type=1&moduleParameter=studyPOSMenu&nextdir=change&next=menu.vm&subdir=applications&xml=menu&purge=y&navigationPosition=functions%2CstudyPOSMenu&breadcrumb=studyPOSMenu&topitem=functions&subitem=studyPOSMenu"
	urlGrades       = "https://linda.hs-heilbronn.de/qisstudent/rds?state=notenspiegelStudent&next=list.vm&nextdir=qispos/notenspiegel/student&createInfos=Y&struct=auswahlBaum&nodeID=auswahlBaum%7Cabschluss%3Aabschl%3D84%2Cstgnr%3D1&expand=0&asi={asi}"
	urlLogin        = "https://linda.hs-heilbronn.de/qisstudent/rds?state=user&type=1&category=auth.login&startpage=portal.vm&breadCrumbSource=portal"
	userAgent       = "Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.111 Safari/537.36"
)

var (
	patternASI        = regexp.MustCompile(`asi=([^"^&]+)`)
	patternGrade      = regexp.MustCompile(`^(\d+,\d+)`)
	patternGrades     = regexp.MustCompile(`(?s)<tr>\s*` + strings.Repeat(`<td[^>]*>(.+?)</td>\s*`, 10) + "</tr>")
	patternIsLoggedIn = regexp.MustCompile("Sie sind angemeldet als:")

	ASIParseError    = errors.New("Could not parse ASI")
	NotLoggedInError = errors.New("Not logged in")
	NoExamsError     = errors.New("No exam results found")
)

type Exam struct {
	CourseType string  `json:"course_type"`
	ECTS       float32 `json:"ects"`
	Grade      float32 `json:"grade"`
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Passed     bool    `json:"passed"`
	Semester   string  `json:"semester"`
}

type ExamByName []Exam

type UserSession struct {
	ECTSOverrides map[int]float32
	asi           string
	client        *http.Client
	loggedIn      bool
	password      string
	username      string
}

func (a ExamByName) Len() int           { return len(a) }
func (a ExamByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a ExamByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (us *UserSession) GetExams() ([]Exam, error) {
	if !us.loggedIn {
		return nil, NotLoggedInError
	}

	if us.asi == "" {
		asi, err := us.getASI()

		if err == nil {
			us.asi = asi
		} else {
			return nil, err
		}
	}

	url := getAsiUrl(us.asi)
	content, err := us.Open(url, nil)

	if err != nil {
		return nil, err
	}

	exams, err := parseExams(content)

	if err != nil {
		return nil, err
	}

	overrideECTS(exams, us.ECTSOverrides)

	return exams, nil
}

func (us *UserSession) Login(username, password string) error {
	data := url.Values{}

	data.Set(fieldUsername, username)
	data.Set(fieldPassword, password)

	content, err := us.Open(urlLogin, data)

	if err != nil {
		return err
	}

	us.loggedIn = patternIsLoggedIn.MatchString(*content)

	if us.loggedIn {
		return nil
	} else {
		return errors.New("Log-in does not appear to be successful (maybe wrong credentials?)")
	}
}

func (us *UserSession) Open(url string, data url.Values) (*string, error) {
	req, err := getRequest(url, data)

	if err != nil {
		return nil, err
	}

	resp, err := us.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Got wrong HTTP status code (%d)", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	bodyStr := string(body)

	return &bodyStr, nil
}

func (us *UserSession) getASI() (string, error) {
	if !us.loggedIn {
		return "", NotLoggedInError
	}

	content, err := us.Open(urlExamOverview, nil)

	if err != nil {
		return "", err
	}

	match := patternASI.FindStringSubmatch(*content)

	if match == nil {
		return "", ASIParseError
	} else {
		return match[1], nil
	}
}

func NewUserSession() *UserSession {
	var us UserSession

	cookieJar, err := cookiejar.New(nil)

	if err != nil {
		log.Panic(err)
	}

	us.client = &http.Client{
		Jar:     cookieJar,
		Timeout: requestTimeout * time.Second,
	}

	return &us
}

func getAsiUrl(asi string) string {
	return strings.Replace(urlGrades, "{asi}", asi, -1)
}

func getRequest(url string, data url.Values) (*http.Request, error) {
	var req *http.Request
	var err error

	if data == nil {
		req, err = http.NewRequest("POST", url, nil)
	} else {
		req, err = http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	if err == nil {
		return req, nil
	} else {
		return nil, err
	}
}

func overrideECTS(exams []Exam, ectsOverrides map[int]float32) {
	for i, exam := range exams {
		if ects, ok := ectsOverrides[exam.ID]; ok {
			exams[i].ECTS = ects
		}
	}
}

func parseExams(content *string) ([]Exam, error) {
	match := patternGrades.FindAllStringSubmatch(*content, -1)

	if match == nil {
		return nil, NoExamsError
	}

	exams := make([]Exam, len(match))

	for i, rawExam := range match {
		trimWhiteSpaces(rawExam)

		exams[i] = Exam{
			CourseType: rawExam[7],
			ECTS:       parseECTS(rawExam[8]),
			Grade:      parseGrade(rawExam[2]),
			ID:         parseID(rawExam[1]),
			Name:       rawExam[10],
			Passed:     parsePassed(rawExam[3]),
			Semester:   rawExam[9],
		}
	}

	return exams, nil
}

func parseECTS(input string) float32 {
	input = strings.Replace(input, ",", ".", -1)
	value, err := strconv.ParseFloat(input, 32)

	if err != nil {
		log.Panicf(`Could not parse ECTS ("%s"): %s`, input, err)
	}

	return float32(value)
}

func parseGrade(input string) float32 {
	match := patternGrade.FindStringSubmatch(input)

	// There are exams without grades.
	if match == nil {
		return 0
	}

	input = match[1]

	input = strings.Replace(input, ",", ".", -1)
	value, err := strconv.ParseFloat(input, 32)

	if err != nil {
		log.Panicf(`Could not parse grade ("%s"): %s`, input, err)
	}

	return float32(value)
}

func parseID(input string) int {
	value, err := strconv.Atoi(input)

	if err != nil {
		log.Panicf(`Could not parse ID ("%s"): %s `, input, err)
	}

	return value
}

func parsePassed(input string) bool {
	return input == "bestanden"
}

func trimWhiteSpaces(items []string) {
	for i, item := range items {
		items[i] = strings.TrimSpace(item)
	}
}
