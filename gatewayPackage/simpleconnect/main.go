package main

import (
	"fmt"
	GHT "gatewayPackage/tbClient.http" //gateway http thingsboard
	tbclient "gatewayPackage/tbClient"
	"time"
	"strings"
	"os/exec"
	b64 "encoding/base64"
)
var tb1 tbclient.TbClient
func excuteLinuxCommand(method string, params []string) string{
	// fmt.Println("[HVM] ===> linux command: ", commandString)
	// arr := strings.Split(commandString, " ")
	cmd := exec.Command(method, params...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return b64.StdEncoding.EncodeToString([]byte("stderr: \n" + string(stdoutStderr))) 
	}
	return b64.StdEncoding.EncodeToString([]byte("stdout: \n" + string(stdoutStderr))) 
}
//TBTextToJSON convert text to json string, minh\n123 ====> {"minh":"", "123":"", "":""}
func TBTextToJSON(text string) string {
	var str strings.Builder
	str.WriteString(`{"`) // JSON begin
	text = strings.Replace(text, `"`, `\"`, -1)
	for _, line := range strings.Split(text, "\n") {
		str.WriteString(line)     // JSON Value
		str.WriteString(`":"","`) // this Value end, next Key begin
	}
	str.WriteString(`":""}`) // JSON end
	fmt.Println("[HVM] ====> log text2jsonstring" + str.String())
	return str.String()
}

func processAllCommand(c tbclient.TbClient, idRes, method, params string){
	fmt.Println("idRes= ", idRes, " method = ", method, "params = ", params)
	args := strings.Split(params, "$#")
	output := excuteLinuxCommand(method, args)
	c.Respond(idRes, TBTextToJSON(output))
}

func main() {
	fmt.Println("Hello World")
	tb1 = GHT.Start("http://203.162.88.116:8080", "mgQ0bMJho7ICQ8Fyhdpm", processAllCommand, "TB1")	
	for {
		tb1.Post(`{"Room 1.Temp":"27", "Room 1.notice":"2"}`)
		time.Sleep(5000 * time.Millisecond)
	}
}
