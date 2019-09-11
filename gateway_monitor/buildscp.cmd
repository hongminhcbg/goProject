del gateway_monitor
del gateway_monitor.exe

set GOOS=linux
set GOARCH=arm
set GOARM=7

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'"

upx.exe gateway_monitor
"C:\Users\Admin\Desktop\putty+PSCP\pscp.exe" -pw 12345678 "C:\GoProject\src\gateway_monitor\gateway_monitor" root@192.168.137.201:/root/iot_gateway

pause