set GOOS=linux
set GOARCH=arm
set GOARM=7
go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'"
pause