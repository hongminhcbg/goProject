del gateway_adapter
del gateway_adapter.exe

set GOOS=linux
set GOARCH=arm
set GOARM=7

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'"
upx.exe gateway_adapter
xcopy gateway_adapter "C:\Users\Admin\Dropbox\ptitopen Team Folder\MinhNH_D15\Autoupdate"  /y
pause