package main

import (
	"bytes"
	"fmt"
	"gateway_commit"
	"gateway_log"
	"net/http"
	"strings"

	//  	"strconv"
	"encoding/json"
	"io/ioutil"

	// 	"gateway_log"
	GwChars "gateway_characteristics"
)

/************************/
type httpClient struct {
	urlGet  string
	urlPost string
	urlRes  string
	idRes   string
}

/************************/
//global variable
var httpThingBoardCurlPost string //post data
var httpThingBoardCurlGet string  //get data from TB
var httpThingBoardCurlRes string  //respond data when receive command
// TB1HttpClient store informations of http client 1
var TB1HttpClient = httpClient{urlGet: "", urlPost: "", urlRes: "", idRes: ""}
// TB2HttpClient store informations of http client 2
var TB2HttpClient = httpClient{urlGet: "", urlPost: "", urlRes: "", idRes: ""}
// IoTGatewayCommandHTTP map string with function, avoid a lot of switch-case
var IoTGatewayCommandHTTP = map[string]func(httpClient, string){
	"reboot":      IoTGatewayRebootHTTP,
	"poweroff":    IoTGatewayPoweroffHTTP,
	"checkupdate": IoTGatewayCheckupdateHTTP,
	"commit":      IoTGatewayCommitHTTP,
}

//IoTGatewayRebootHTTP reboot pi
func IoTGatewayRebootHTTP(c httpClient, idRes string) {
	ThingsboardResponseHTTP(TBTextToJSON("IoTGateway is rebooting"), c, idRes)
	GwChars.Reboot()
}
// IoTGatewayPoweroffHTTP turn off pi
func IoTGatewayPoweroffHTTP(c httpClient, idRes string) {
	ThingsboardResponseHTTP(TBTextToJSON("IoTGateway is rebooting"), c, idRes)
	gateway_log.Thingsboard_add_log("IoTGateway_poweroff : gateway power off ")
	GwChars.Poweroff()
}
// IoTGatewayCheckupdateHTTP check file in https://www.dropbox.com/sh/dzkpki95m6zgv4r/AACVV6roExWn1sl5oS96Thh5a?dl=0, 
// if have change download new  file 
func IoTGatewayCheckupdateHTTP(c httpClient, idRes string) {
	ThingsboardResponseHTTP(TBTextToJSON("IoTGateway is checking for updates"), c, idRes)
	gateway_log.Thingsboard_add_log("IoTGateway_checkupdate: gateway checkupdate")
	GwChars.CheckUpdate()
}

// IoTGatewayCommitHTTP copy all data /root/iot_gateway to partition 1 of SD card 
func IoTGatewayCommitHTTP(c httpClient, idRes string) {
	str := gateway_commit.Commit()
	ThingsboardResponseHTTP(TBTextToJSON(str), c, idRes)
	gateway_log.Thingsboard_add_log("IoTGateway_commit: commit dir IoTGateway " + str)
}
/*********************/
func processMsgHTTP(method, idRes, body string, c httpClient) {
	args := Parse_Cmd(method)
	switch args[0] {
	case "?":
		ThingsboardResponseHTTP(helpCmd(), c, idRes) //thingsboardResponse(c, topic, help_cmd())
	case "??":
		ThingsboardResponseHTTP(helpConfig(), c, idRes) //thingsboardResponse(c, topic, help_config())

	case "adapter":
		switch args[1] {
		case "restart":
			ThingsboardResponseHTTP(TBTextToJSON("Adapter is restarting"), c, idRes) //thingsboardResponse(c, topic, TBTextToJSON("Adapter is restarting"))
			GwChars.Monitor_restart_adapter()

		default:
			ThingsboardResponseHTTP(helpAdapter(), c, idRes) //thingsboardResponse(c, topic, help_adapter())
		}

	case "monitor":
		switch args[1] {
		case "restart":
			ThingsboardResponseHTTP(TBTextToJSON("Monitor is restarting"), c, idRes) //thingsboardResponse(c, topic, TBTextToJSON("Monitor is restarting"))
			GwChars.Monitor_restart_monitor()

		default:
			ThingsboardResponseHTTP(helpMonitor(), c, idRes) //thingsboardResponse(c, topic, helpMonitor())
		}

	case "mosquitto":
		switch args[1] {
		case "restart":
			ThingsboardResponseHTTP(TBTextToJSON("Mosquitto is restarting"), c, idRes) //thingsboardResponse(c, topic, TBTextToJSON("Mosquitto is restarting"))
			GwChars.Restart_mosquitto()
		default:
			ThingsboardResponseHTTP(helpMosquitto(), c, idRes) //thingsboardResponse(c, topic, helpMosquitto())
		}

	case "domoticz":
		switch args[1] {
		case "restart":
			ThingsboardResponseHTTP(TBTextToJSON("Domoticz is restarting"), c, idRes)
			GwChars.Restart_domoticz()
		default:
			ThingsboardResponseHTTP(helpDomoticz(), c, idRes)
		}

	case "iotgateway":
		if fn, ok := IoTGatewayCommandHTTP[args[1]]; ok {
			fn(c, idRes)
		} else {
			ThingsboardResponseHTTP(helpIotgateway(), c, idRes)
		}
	case "node", "mysensors", "nodecmd":
		mqtt_mosquitto.Publish(`v1/devices/me/rpc/request/`+idRes, 0, false, body) // FW to mosquitto
	default:
		ThingsboardResponseHTTP(TBTextToJSON("Unknow object"), c, idRes) //thingsboardResponse(c, topic, TBTextToJSON("Unknow object"))
	}
}
/**********************************/

// CallBackThingsboardRequestHTTP received command of user, parse id, method, call processMsgHTTP to handle data
func CallBackThingsboardRequestHTTP(c httpClient) {
	req, _ := http.NewRequest("GET", c.urlGet, nil)
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	//    fmt.Println(string(body))
	//fmt.Printf("type of body is %T\n", body)
	dec := json.NewDecoder(bytes.NewReader(body))

	var jsonDecode map[string]interface{}
	if err := dec.Decode(&jsonDecode); err != nil {
		return
	}
	if idRes1, ok := jsonDecode["id"].(float64); ok {
		idRes := fmt.Sprintf("%d", int(idRes1)) //float to string
		//c.idRes = idRes
		fmt.Printf("id receive = %s\n", idRes)
		if method, ok := jsonDecode["method"].(string); ok {
			fmt.Println(method)
			processMsgHTTP(method, idRes, string(body), c)
		}
	}
}
/************************************************/

// ThingsboardPostHTTP post data to host
func ThingsboardPostHTTP(c httpClient, msg string) {
	req, _ := http.NewRequest("POST", c.urlPost, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}
/***************************************************/

// ThingsboardResponseHTTP respond when user use command
func ThingsboardResponseHTTP(msg string, c httpClient, idRes string) {
	//THINGSBOARD_CURL := Config.Thingsboard_1.Host + "/api/v1/" + Config.Thingsboard_1.MonitorToken + "/rpc/" + id
	THINGSBOARDCURL := c.urlRes + idRes
	req, _ := http.NewRequest("POST", THINGSBOARDCURL, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
	//fmt.Println("End respond data")
}
/********************************************/

// setupCallBackThingsboardRequestHTTP create thread alway wait command of user
func setupCallBackThingsboardRequestHTTP(c httpClient) {
	fmt.Println("Minh ============> setupCallBackThingsboardRequestHTTP", c)
	for {
		CallBackThingsboardRequestHTTP(c)
		GwChars.Sleep_ms(100)
	}
}
/***********************************************/

//httpTB1Setup setup TB1 http client
func httpTB1Setup() {
	TB1HttpClient.urlPost = Config.Thingsboard_1.Host + "/api/v1/" + Config.Thingsboard_1.MonitorToken + "/telemetry"
	TB1HttpClient.urlGet = Config.Thingsboard_1.Host + "/api/v1/" + Config.Thingsboard_1.MonitorToken + "/rpc?timeout=2000000"
	TB1HttpClient.urlRes = Config.Thingsboard_1.Host + "/api/v1/" + Config.Thingsboard_1.MonitorToken + "/rpc/" // + id
	TB1HttpClient.idRes = ""
	// fmt.Printf("httpThingBoardCurlPost = %s\n", TB1HttpClient.urlPost)
	// fmt.Printf("httpThingBoardCurlGet = %s\n", TB1HttpClient.urlGet)
	// fmt.Printf("httpThingBoardCurlRes = %s\n", TB1HttpClient.urlRes)

	go setupCallBackThingsboardRequestHTTP(TB1HttpClient)
}
/**********************************/

// httpTB2Setup setup TB1 http client
func httpTB2Setup() {
	TB2HttpClient.urlPost = Config.Thingsboard_2.Host + "/api/v1/" + Config.Thingsboard_2.MonitorToken + "/telemetry"
	TB2HttpClient.urlGet = Config.Thingsboard_2.Host + "/api/v1/" + Config.Thingsboard_2.MonitorToken + "/rpc?timeout=2000000"
	TB2HttpClient.urlRes = Config.Thingsboard_2.Host + "/api/v1/" + Config.Thingsboard_2.MonitorToken + "/rpc/" // + id
	TB2HttpClient.idRes = ""
	go setupCallBackThingsboardRequestHTTP(TB2HttpClient)
	//	fmt.Println(TB2HttpClient)
}
