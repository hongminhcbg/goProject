del gateway_checkKey
del gateway_checkKey.exe

set GOOS=linux
set GOARCH=arm
set GOARM=7

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'" -o gateway_checkKey main.go

upx.exe gateway_checkKey
"C:\Users\Admin\Desktop\putty+PSCP\pscp.exe" -pw 12345678 "gateway_checkKey" root@192.168.137.201:/root/iot_gateway
rem xcopy gateway_monitor "C:\Users\Admin\Dropbox\ptitopen Team Folder\MinhNH_D15\Autoupdate"  /y
