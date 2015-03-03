package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/srhnsn/go-lindaparser"
)

const (
	cachedExamsFilename = "exams.json"
)

func findNewExams(currentExams, cachedExams []lindaparser.Exam) []lindaparser.Exam {
	result := make([]lindaparser.Exam, 0)

	if len(cachedExams) == 0 {
		return []lindaparser.Exam{}
	}

Outer:
	for _, currentExam := range currentExams {
		for _, cachedExam := range cachedExams {
			if cachedExam == currentExam {
				continue Outer
			}
		}

		result = append(result, currentExam)
	}

	return result
}

func getCachedExams() ([]lindaparser.Exam, error) {
	b, err := ioutil.ReadFile(cachedExamsFilename)

	if err != nil {
		return []lindaparser.Exam{}, nil
	}

	var result []lindaparser.Exam

	err = json.Unmarshal(b, &result)

	if err == nil {
		return result, nil
	} else {
		return nil, err
	}
}

func main() {
	config := lindaparser.LoadConfig(Asset)

	us := lindaparser.NewUserSession()

	err := us.Login(config.Username, config.Password)

	if err != nil {
		log.Fatalf("Failed to log in: %s", err)
	}

	currentExams, err := us.GetExams()

	if err != nil {
		log.Fatalf("Failed to get current exams: %s", err)
	}

	cachedExams, err := getCachedExams()

	if err != nil {
		log.Fatalf("Failed to get cached exams: %s", err)
	}

	err = saveExams(currentExams)

	if err != nil {
		log.Fatalf("Failed to save current exams: %s", err)
	}

	newExams := findNewExams(currentExams, cachedExams)

	for _, newExam := range newExams {
		fmt.Printf("%s: %s\n", newExam.Semester, newExam.Name)
	}
}

func saveExams(exams []lindaparser.Exam) error {
	b, err := json.MarshalIndent(exams, "", "    ")

	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cachedExamsFilename, b, 0600)

	return err
}
