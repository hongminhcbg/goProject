del gateway_adapter
del gateway_adapter.exe

set GOOS=linux
set GOARCH=arm
set GOARM=7

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'"
upx.exe gateway_adapter
"C:\Users\Admin\Desktop\putty+PSCP\pscp.exe" -pw 12345678 gateway_adapter root@192.168.137.201:/root/iot_gateway
pause