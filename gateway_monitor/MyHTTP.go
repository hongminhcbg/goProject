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
var TB1HttpClient = httpClient{urlGet: "", urlPost: "", urlRes: "", idRes: ""}
var TB2HttpClient = httpClient{urlGet: "", urlPost: "", urlRes: "", idRes: ""}
var IoTGatewayCommandHttp = map[string]func(httpClient, string){
	"reboot":      IoTGateway_reboot_http,
	"poweroff":    IoTGateway_poweroff_http,
	"checkupdate": IoTGateway_checkupdate_http,
	"commit":      IoTGateway_commit_http,
}

/***********************/
//iotgateway function
func IoTGateway_reboot_http(c httpClient, idRes string) {
	ThingsboardResponseHTTP(TB_TextToJSON("IoTGateway is rebooting"), c, idRes)
	GwChars.Reboot()
}
func IoTGateway_poweroff_http(c httpClient, idRes string) {
	ThingsboardResponseHTTP(TB_TextToJSON("IoTGateway is rebooting"), c, idRes)
	gateway_log.Thingsboard_add_log("IoTGateway_poweroff : gateway power off ")
	GwChars.Poweroff()
}
func IoTGateway_checkupdate_http(c httpClient, idRes string) {
	ThingsboardResponseHTTP(TB_TextToJSON("IoTGateway is checking for updates"), c, idRes)
	gateway_log.Thingsboard_add_log("IoTGateway_checkupdate: gateway checkupdate")
	GwChars.CheckUpdate()
}
func IoTGateway_commit_http(c httpClient, idRes string) {
	str := gateway_commit.Commit()
	ThingsboardResponseHTTP(TB_TextToJSON(str), c, idRes)
	gateway_log.Thingsboard_add_log("IoTGateway_commit: commit dir IoTGateway " + str)
}

/*********************/
func processMsgHTTP(method, idRes, body string, c httpClient) {
	args := Parse_Cmd(method)
	switch args[0] {
	case "?":
		ThingsboardResponseHTTP(help_cmd(), c, idRes) //thingsboardResponse(c, topic, help_cmd())
	case "??":
		ThingsboardResponseHTTP(help_config(), c, idRes) //thingsboardResponse(c, topic, help_config())

	case "adapter":
		switch args[1] {
		case "restart":
			ThingsboardResponseHTTP(TB_TextToJSON("Adapter is restarting"), c, idRes) //thingsboardResponse(c, topic, TB_TextToJSON("Adapter is restarting"))
			GwChars.Monitor_restart_adapter()

		default:
			ThingsboardResponseHTTP(help_adapter(), c, idRes) //thingsboardResponse(c, topic, help_adapter())
		}

	case "monitor":
		switch args[1] {
		case "restart":
			ThingsboardResponseHTTP(TB_TextToJSON("Monitor is restarting"), c, idRes) //thingsboardResponse(c, topic, TB_TextToJSON("Monitor is restarting"))
			GwChars.Monitor_restart_monitor()

		default:
			ThingsboardResponseHTTP(help_monitor(), c, idRes) //thingsboardResponse(c, topic, help_monitor())
		}

	case "mosquitto":
		switch args[1] {
		case "restart":
			ThingsboardResponseHTTP(TB_TextToJSON("Mosquitto is restarting"), c, idRes) //thingsboardResponse(c, topic, TB_TextToJSON("Mosquitto is restarting"))
			GwChars.Restart_mosquitto()
		default:
			ThingsboardResponseHTTP(help_mosquitto(), c, idRes) //thingsboardResponse(c, topic, help_mosquitto())
		}

	case "domoticz":
		switch args[1] {
		case "restart":
			ThingsboardResponseHTTP(TB_TextToJSON("Domoticz is restarting"), c, idRes)
			GwChars.Restart_domoticz()
		default:
			ThingsboardResponseHTTP(help_domoticz(), c, idRes)
		}

	case "iotgateway":
		if fn, ok := IoTGatewayCommandHttp[args[1]]; ok {
			fn(c, idRes)
		} else {
			ThingsboardResponseHTTP(help_iotgateway(), c, idRes)
		}
	case "node", "mysensors", "nodecmd":
		mqtt_mosquitto.Publish(`v1/devices/me/rpc/request/`+idRes, 0, false, body) // FW to mosquitto
	default:
		ThingsboardResponseHTTP(TB_TextToJSON("Unknow object"), c, idRes) //thingsboardResponse(c, topic, TB_TextToJSON("Unknow object"))
	}
}
func CallBackThingsboardRequestHTTP(c httpClient) {
	req, _ := http.NewRequest("GET", c.urlGet, nil)
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	//    fmt.Println(string(body))
	//fmt.Printf("type of body is %T\n", body)
	dec := json.NewDecoder(bytes.NewReader(body))

	var json_decode map[string]interface{}
	if err := dec.Decode(&json_decode); err != nil {
		return
	}
	if id_respond_1, ok := json_decode["id"].(float64); ok {
		id_respond := fmt.Sprintf("%d", int(id_respond_1)) //float to string
		//c.idRes = id_respond
		fmt.Printf("id receive = %s\n", id_respond)
		if method, ok := json_decode["method"].(string); ok {
			fmt.Println(method)
			processMsgHTTP(method, id_respond, string(body), c)
		}
	}
}

/************************************************/
func ThingsboardPostHTTP(c httpClient, msg string) {
	req, _ := http.NewRequest("POST", c.urlPost, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}

/***************************************************/
func ThingsboardResponseHTTP(msg string, c httpClient, idRes string) {
	//THINGSBOARD_CURL := Config.Thingsboard_1.Host + "/api/v1/" + Config.Thingsboard_1.MonitorToken + "/rpc/" + id
	THINGSBOARD_CURL := c.urlRes + idRes
	req, _ := http.NewRequest("POST", THINGSBOARD_CURL, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
	//fmt.Println("End respond data")
}

/********************************************/
func setupCallBackThingsboardRequestHTTP(c httpClient) {
	for {
		CallBackThingsboardRequestHTTP(c)
		GwChars.Sleep_ms(100)
	}
}
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

func httpTB2Setup() {
	TB2HttpClient.urlPost = Config.Thingsboard_2.Host + "/api/v1/" + Config.Thingsboard_2.MonitorToken + "/telemetry"
	TB2HttpClient.urlGet = Config.Thingsboard_2.Host + "/api/v1/" + Config.Thingsboard_2.MonitorToken + "/rpc?timeout=2000000"
	TB2HttpClient.urlRes = Config.Thingsboard_2.Host + "/api/v1/" + Config.Thingsboard_2.MonitorToken + "/rpc/" // + id
	TB2HttpClient.idRes = ""
	fmt.Println(TB2HttpClient)

	go setupCallBackThingsboardRequestHTTP(TB2HttpClient)
	//	fmt.Println(TB2HttpClient)
}
