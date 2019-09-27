package main

import (
//	GMT "gateway.mqtt.tb"
//	GHT "gateway.http.tb"
	GwChars "gateway_characteristics" // rename to "GwChars"
	"fmt"
//	"strings"
	GTC "gatewayPackage/tbClient" //gateway thingsboard client
)
var tb1 GTC.TbClient
var tb2 GTC.TbClient
/**********************************************/

// processAllCommand process all command of user send form tb
func processAllCommand(idRes, method, idDev string){
	fmt.Println("processAllCommand", idRes, method, idDev)
	if idDev == "TB1" {
		tb1.Respond(idRes, `{"staus":"ok"}`)
	} else if idDev == "TB2" {
		tb2.Respond(idRes, `{"status":"ok"}`)
	}
}
/*********************************************/
func main(){
	//httpClient = GHT.NewClient("TB1")	
	tb1 = GTC.NewHTTPTbClient("http://172.16.0.158:8080", "sWq5EXCQLC8hGVfoXztT", "TB1")
	//tb1 = GTC.NewHTTPTbClient("https://demo.thingsboard.io:80", "Q8ryIjV0hn1DJvNsGUfg", "TB1")
	tb2 = GTC.NewMQTTTbClient("tcp://demo.thingsboard.io:1883", "Q8ryIjV0hn1DJvNsGUfg", "TB2")	
	tb1.Setup(processAllCommand)
	tb2.Setup(processAllCommand)
	fmt.Printf("value of tb1 is %v \n", tb1)
	fmt.Printf("value of tb2 is %v \n", tb2)
	//temp2.Setup("http://192.168.0.108:8080", "sWq5EXCQLC8hGVfoXztT", "TB1", processAllCommand)
	//httpClient.Setup("http://192.168.0.108:8080", "sWq5EXCQLC8hGVfoXztT", "TB1", processAllCommand)	
	//mqttClient.Setup("tcp://demo.thingsboard.io:1883", "Q8ryIjV0hn1DJvNsGUfg", "TB2", processAllCommand)
	for {
		GwChars.Sleep_ms(2000)
	}
}