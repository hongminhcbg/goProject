del gateway_Command
del gateway_Command.exe

set GOOS=linux
set GOARCH=amd64

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'" -o gateway_Command main.go

rem upx.exe gateway_Command
"C:\Users\Admin\Desktop\putty+PSCP\pscp.exe" -pw 1 "gateway_Command" lhm@192.168.44.139:/home/lhm
rem xcopy gateway_monitor "C:\Users\Admin\Dropbox\ptitopen Team Folder\MinhNH_D15\Autoupdate"  /y
