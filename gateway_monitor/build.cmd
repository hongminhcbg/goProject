del gateway_monitor
del gateway_monitor.exe

set GOOS=linux
set GOARCH=arm
set GOARM=7

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'"

upx.exe gateway_monitor

xcopy gateway_monitor "C:\Users\Admin\Dropbox\ptitopen Team Folder\MinhNH_D15\Autoupdate"  /y
pause