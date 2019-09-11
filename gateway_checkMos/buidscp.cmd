del gateway_checkMosquitto
del gateway_checkMosquitto.exe

set GOOS=linux
set GOARCH=arm
set GOARM=7

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'"
upx.exe gateway_checkMosquitto
"C:\Users\Admin\Desktop\putty+PSCP\pscp.exe" -pw 12345678 gateway_checkMosquitto root@192.168.137.201:/root/iot_gateway/
pause