package main
import (
	GHT "gateway.http.tb"
	"log"
)

// HTTPCallBackThingsboardRequest received command from user, call processAllCommand to handle that
func HTTPCallBackThingsboardRequest(idRes, method string){
	log.Println("idRes = ", idRes)
	log.Println("method = ", method)
	processAllCommand(idRes, method, "HTTP")
}
/*****************************/

// setupHTTP setup http protocol
func setupHTTP() {
	tbResFuncMap["HTTP"] = GHT.RespondMsg
	if Config.Thingsboard_1.Enable == true && useMqttTB1 == false {
		GHT.Setup(Config.Thingsboard_1.Host, Config.Thingsboard_1.MonitorToken, HTTPCallBackThingsboardRequest, 0)
		tb1PostFunc = GHT.PostMsg
	}
	if Config.Thingsboard_2.Enable == true && useMqttTB2 == false {
		GHT.Setup(Config.Thingsboard_2.Host, Config.Thingsboard_2.MonitorToken, HTTPCallBackThingsboardRequest, 1)
		if Config.Thingsboard_1.Enable == true && useMqttTB1 == false {
			tb2PostFunc = doNothing1arg
		} else {
			tb2PostFunc = GHT.PostMsg
		}
	}
}
/************************************************************************/
