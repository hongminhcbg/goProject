del gateway_checkupdate
del gateway_checkupdate.exe

set GOOS=linux
set GOARCH=arm
set GOARM=7

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'"
upx.exe gateway_checkupdate
xcopy gateway_checkupdate "C:\Users\Admin\Dropbox\ptitopen Team Folder\MinhNH_D15\Autoupdate"  /y
pause