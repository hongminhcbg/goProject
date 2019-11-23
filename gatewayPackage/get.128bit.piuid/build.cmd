del get_uid_128
del get_uid_128.exe

set GOOS=linux
set GOARCH=arm
set GOARM=7

go build -i -ldflags "-s -w -X 'main.buildTime=%date% %time%'" -o get_uid_128 main.go

upx.exe get_uid_128
"C:\Users\Admin\Desktop\putty+PSCP\pscp.exe" -pw 12345678 "get_uid_128" root@192.168.137.201:/root/iot_gateway