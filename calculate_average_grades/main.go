package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/srhnsn/go-lindaparser"
)

const (
	col1Width = 52
	col2Width = 5
	col3Width = 5
)

var (
	rowLength = col1Width + col2Width + col3Width + 2
)

func getFilteredExams(exams []lindaparser.Exam, courseType string) []lindaparser.Exam {
	result := make([]lindaparser.Exam, len(exams))

	for _, exam := range exams {
		if exam.CourseType == courseType {
			result = append(result, exam)
		}
	}

	return result
}

func main() {
	config := lindaparser.LoadConfig(Asset)

	us := lindaparser.NewUserSession()
	us.ECTSOverrides = config.ECTSOverrides

	err := us.Login(config.Username, config.Password)

	if err != nil {
		log.Fatalf("Failed to log in: %s", err)
	}

	exams, err := us.GetExams()

	if err != nil {
		log.Fatalf("Failed to get exams: %s", err)
	}

	sort.Sort(lindaparser.ExamByName(exams))

	examsG := getFilteredExams(exams, "G")
	examsH := getFilteredExams(exams, "H")

	fmt.Print("Grundstudium:\n\n")
	printExams(examsG)

	fmt.Print("\n\nHauptstudium:\n\n")
	printExams(examsH)
}

func printAverage(ectsTotal, gradesTotal float32) {
	average := gradesTotal / ectsTotal

	averageStr := lindaparser.FormatFloat(average, 2)
	ectsTotalStr := lindaparser.FormatFloat(ectsTotal, -1)
	gradesTotalStr := lindaparser.FormatFloat(gradesTotal, -1)

	fmt.Println(strings.Repeat("-", rowLength))

	fmt.Printf("%s %s %s\n",
		lindaparser.JustifyLeft("Summe", col1Width),
		lindaparser.JustifyLeft(ectsTotalStr, col2Width),
		lindaparser.JustifyLeft(gradesTotalStr, col3Width))

	fmt.Printf("%s %s %s\n",
		lindaparser.JustifyLeft("Durchschnitt", col1Width),
		lindaparser.JustifyLeft("", col2Width),
		lindaparser.JustifyLeft(averageStr, col3Width))
}

func printExams(exams []lindaparser.Exam) {
	printHeader()

	var ectsTotal float32
	var gradesTotal float32

	for _, exam := range exams {
		if exam.Grade == 0 || !exam.Passed {
			continue
		}

		ectsTotal += exam.ECTS
		gradesTotal += exam.Grade * exam.ECTS

		fmt.Printf("%s %s %s\n",
			lindaparser.JustifyLeft(exam.Name, col1Width),
			lindaparser.JustifyLeft(lindaparser.FormatFloat(exam.ECTS, 1), col2Width),
			lindaparser.JustifyLeft(lindaparser.FormatFloat(exam.Grade, 1), col3Width))
	}

	printAverage(ectsTotal, gradesTotal)
}

func printHeader() {
	fmt.Printf("%s %s %s\n%s\n",
		lindaparser.JustifyLeft("Veranstaltung", col1Width),
		lindaparser.JustifyLeft("ECTS", col2Width),
		lindaparser.JustifyLeft("Note", col3Width),
		strings.Repeat("-", rowLength))
}
