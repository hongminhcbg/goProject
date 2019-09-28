package tbclient

import(
	"net/http"
	"fmt"
	"time"
	"log"
//	"gateway_log"
	"io/ioutil"
	"encoding/json"
	"bytes"
	GwChars "gateway_characteristics"
	"strings"
	mqtt "github.com/eclipse/paho.mqtt.golang"

)

// TbClient communication with thingsboard
type TbClient interface{
	// Setup setup device and callback function
	Setup(func(string, string, string))

	// Respond data to host, need idRes to create topic or http link
	// respond user command 
	Respond(idRes string, msg string)

	// Post data to host
	Post(string)
}

// HTTPTbClient http, communication with tb by http protocol
type HTTPTbClient struct {
	urlPost 		string
	urlGet 			string
	urlRes 			string
	ConnectStatus 	bool
	idDev 			string
}
/**************************************************/

// Disable use for thingsboard disable
type Disable struct {
	name string
}
/*****************************************************/

// MQTTTbClient mqtt, communication with tb by mqtt protocol
type MQTTTbClient struct {
	Dev 						mqtt.Client
	thingsboardTopicTelemetry 	string
	thingsboardTopicRequest 	string
	thingsboardTopicResponse 	string
	mqttDisconnectTimeout 		uint
	mqttConnectTimeout 			time.Duration
	tbLostConnectHandler		func(mqtt.Client, error)
	tbOnConnectHandler			func(mqtt.Client)
	Connected 					bool
	idDev						string
	host						string
	monitorTocken				string
}
/**************************************************/

// processAllCommand use for mqtt
var processAllCommand func(string, string, string)
/*********************************************/

// callBackTB1 use for mqtt
func callBackTB1(c mqtt.Client, message mqtt.Message){
	// fmt.Printf("TOPIC: %s\n", message.Topic())
	// fmt.Printf("MSG: %s\n", message.Payload())
	dec := json.NewDecoder(bytes.NewReader(message.Payload()))
	var jsonDecode map[string]interface{}
	if err := dec.Decode(&jsonDecode); err != nil {
		log.Println(err)
		return
	}
	if method, ok := jsonDecode["method"].(string); ok {
		topicStr := string(message.Topic())
		arrID := strings.Split(topicStr, "/")
		idRes := arrID[len(arrID)-1]
		processAllCommand(idRes, method, "TB1")
	}
}

func callBackTB2(c mqtt.Client, message mqtt.Message){
	// fmt.Printf("TOPIC: %s\n", message.Topic())
	// fmt.Printf("MSG: %s\n", message.Payload())
	dec := json.NewDecoder(bytes.NewReader(message.Payload()))
	var jsonDecode map[string]interface{}
	if err := dec.Decode(&jsonDecode); err != nil {
		log.Println(err)
		return
	}
	if method, ok := jsonDecode["method"].(string); ok {
		topicStr := string(message.Topic())
		arrID := strings.Split(topicStr, "/")
		idRes := arrID[len(arrID)-1]
		processAllCommand(idRes, method, "TB2")
	}	
}
/******************************************/

// loop and wait user command
func callbackFuncThread(urlGet string, idDev string, CBFunction func(string, string, string)){
	fmt.Println("Begin loop, urlGet = ", urlGet)
	for {
		req, _ := http.NewRequest("GET", urlGet, nil)
		req.Header.Add("Content-Type", "application/json")
		res, _ := http.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(res.Body)
		dec := json.NewDecoder(bytes.NewReader(body))
		var jsonDecode map[string]interface{}
		if err := dec.Decode(&jsonDecode); err != nil {
			fmt.Println("decode error")
		} else if idRes1, ok := jsonDecode["id"].(float64); ok {
			idRes := fmt.Sprintf("%d", int(idRes1)) //float to string
//			fmt.Printf("id receive = %s\n", idRes)
			if method, ok := jsonDecode["method"].(string); ok {
				fmt.Println(method)
				CBFunction(idRes, method, idDev)
			}
		}
		GwChars.Sleep_ms(1000)
	}
}
/**********************************************/

// Setup setup url for device, create thread always wait user command
func (c HTTPTbClient) Setup(callBackFunc func(string, string, string)){
	fmt.Println(c.idDev, " onconnect HTTP")
	go callbackFuncThread(c.urlGet, c.idDev, callBackFunc)
}
/**************************************/

// Post post msg to host
func (c HTTPTbClient) Post(msg string) {
	req, _ := http.NewRequest("POST", c.urlPost, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)	
}
/*************************************/

// Respond to host
func (c HTTPTbClient) Respond(idRes string, msg string){
	THINGSBOARDCURL := c.urlRes + idRes
	req, _ := http.NewRequest("POST", THINGSBOARDCURL, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}
/*******************************************/

// Setup mqtt callback function
func (c MQTTTbClient) Setup(callbackFunc func(string, string, string)){
	processAllCommand = callbackFunc
}
/***************************************/

// Post post data to host
func (c MQTTTbClient) Post(msg string){
	c.Dev.Publish(c.thingsboardTopicTelemetry, 0, false, msg)
}
/**********************************************/

// Respond respond msg to host
func (c MQTTTbClient) Respond(idRes, msg string) {
	topic := `v1/devices/me/rpc/response/` + idRes
	c.Dev.Publish(topic, 0, false, msg)
}
/********************************************************/

// Setup and do nothng
func (c Disable) Setup(callbackFunc func(string, string, string)){

}
/*****************************************************************/

// Post do nothing
func (c Disable) Post(msg string){

}
/********************************************************/

// Respond nothings to host
func (c Disable) Respond(idRes, mgs string){

}
/*****************************************************************/

// NewHTTPTbClient create new
func NewHTTPTbClient(host string, monitorTocken string, idDev string) HTTPTbClient{
	var c HTTPTbClient
	c.urlGet = host + "/api/v1/" + monitorTocken + "/rpc?timeout=2000000"
	c.urlPost = host + "/api/v1/" + monitorTocken + "/telemetry"
	c.urlRes = host + "/api/v1/" + monitorTocken + "/rpc/" // + id
	c.ConnectStatus = false
	c.idDev = idDev
	return c
}
/***********************************************/

// NewMQTTTbClient create new mqtt
func NewMQTTTbClient(host string, monitorTocken string, idDev string) MQTTTbClient {
	client := MQTTTbClient{}
	client.thingsboardTopicRequest = 		`v1/devices/me/rpc/request/+`
	client.thingsboardTopicResponse = 		`v1/devices/me/rpc/response/+`
	client.thingsboardTopicTelemetry = 		`v1/devices/me/telemetry`
	client.mqttConnectTimeout = 			5
	client.mqttDisconnectTimeout = 			5000
	client.idDev = 							idDev
	client.host = 							host
	client.monitorTocken = 					monitorTocken

	client.tbLostConnectHandler = func(c mqtt.Client, err error){
		GwChars.SetLed_Red("1")
		log.Printf("Thingsboard1.LostConnect, reason: %v\n", err)
		log.Println()
	}
	client.tbOnConnectHandler = func(c mqtt.Client){
		log.Println("Thingsboard1.OnConnect")
	}

	optsThingsboard := mqtt.NewClientOptions()
	optsThingsboard.AddBroker(client.host)
	optsThingsboard.SetUsername(client.monitorTocken)
	optsThingsboard.SetConnectionLostHandler(client.tbLostConnectHandler)
	optsThingsboard.SetOnConnectHandler(client.tbOnConnectHandler)
	client.Dev = mqtt.NewClient(optsThingsboard)
	client.Dev.Connect().WaitTimeout(client.mqttConnectTimeout * time.Second)
	if client.idDev == "TB1" {
		client.Dev.Subscribe(client.thingsboardTopicRequest, 0, callBackTB1)
	} else {
		client.Dev.Subscribe(client.thingsboardTopicRequest, 0, callBackTB2)
	}		
	return client
}
/****************************************/

// NewDisableClient create disable
func NewDisableClient(name string) Disable{
	var c = Disable{}
	c.name = name
	return c
}