@echo off

go-bindata -o assets.go -prefix assets/ assets/...

copy assets.go calculate_average_grades >nul
move assets.go find_new_exams >nul

if errorlevel 1 pause
