go-lindaparser
==============
This is a [Go](https://golang.org/) port of [py-lindaparser](https://github.com/srhnsn/py-lindaparser).

Requirements
------------
* Go

Installation and usage
----------------------
1. Run `go get github.com/srhnsn/go-lindaparser/...`
1. Copy `config.example.yaml` to `assets/config.yaml` and edit accordingly.
1. Run `go-bindata -o assets.go -prefix assets/ assets/...` and copy the created `assets.go` into the `calculate_average_grades` and `find_new_exams` directories. On Windows, you can also just run the `build assets.bat` file. You need to repeat this step, if you edit the file in `assets/config.yaml`.
1. Depending on what you want to do, change into the `calculate_average_grades` or `find_new_exams` directory and run `go run main.go assets.go`. If you run `go build` in these directories, you will get single binary files ready to deploy anywhere without any other dependencies.

Average grade calculator
------------------------
This program fetches and prints the current exam results and calculates the average grades. The view is separated by the Grundstudium and the Hauptstudium.

It is possible to fix wrong ECTS values on Linda by modifying the `ects_overrides` setting in the `config.yaml` file. It is an object where its keys are the course IDs and its values the corrected ECTS values.

New exam finder
---------------
On the first run, this program will create an `exams.json` file containing all currently available exam information. If you run the program again, it will print any new exam results it finds on Linda. If it does not find any new exam results, the program will not output anything.

If you set up your environment so it e-mails all program output to yourself (with [cron](https://en.wikipedia.org/wiki/Cron) for example), you can use this to get prompt notifications about new exam results.
