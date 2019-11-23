del createGatewayTocken.exe

"C:\Users\Admin\Desktop\putty+PSCP\pscp.exe" -pw 12345678 "config.json" root@192.168.137.201:/root/iot_gateway
rem go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'" -o createGatewayTocken.exe main.go
