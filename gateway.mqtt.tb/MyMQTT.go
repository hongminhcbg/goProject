package gatewaymqtttb

import (
	//	"fmt"
	"log"
	"strings"
	"time"

//	"io/ioutil"
	//	"crypto/tls"
	//	"crypto/x509"
	GwChars "gateway_characteristics" // rename to "GwChars"
	//checkMos "gateway_checkMos"
	"gateway_log"
	"bytes"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)
/*********************/

// Client mqtt
type Client struct {
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
/************************************************************************/
var processAllCommand func(string, string, string)
/************************************************************************/

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

// Setup mqtt protocol
func (c *Client) Setup(host string, monitorTocken string, idDev string, callbackFunc func(string, string, string)){
	c.host = host
	c.monitorTocken = monitorTocken
	optsThingsboard := mqtt.NewClientOptions()
	optsThingsboard.AddBroker(host)
	optsThingsboard.SetUsername(monitorTocken)
	optsThingsboard.SetConnectionLostHandler(c.tbLostConnectHandler)
	optsThingsboard.SetOnConnectHandler(c.tbOnConnectHandler)	
	if c.Dev != nil && c.Dev.IsConnected() {
		c.Dev.Disconnect(c.mqttDisconnectTimeout)
	}
	c.Dev = mqtt.NewClient(optsThingsboard)
	c.Dev.Connect().WaitTimeout(c.mqttConnectTimeout * time.Second)
	if c.idDev == "TB1" {
		c.Dev.Subscribe(c.thingsboardTopicRequest, 0, callBackTB1)
	} else {
		c.Dev.Subscribe(c.thingsboardTopicRequest, 0, callBackTB2)
	}
	processAllCommand = callbackFunc	
}
/****************************/

// NewClient create new client
func NewClient(idDev string) Client {
	client := Client{}
	client.thingsboardTopicRequest = 		`v1/devices/me/rpc/request/+`
	client.thingsboardTopicResponse = 		`v1/devices/me/rpc/response/+`
	client.thingsboardTopicTelemetry = 		`v1/devices/me/telemetry`
	client.mqttConnectTimeout = 			5
	client.mqttDisconnectTimeout = 			5000
	client.idDev = 							idDev
	if idDev == "TB1" {
		client.tbLostConnectHandler = func(c mqtt.Client, err error){
			GwChars.SetLed_Red("1")
			client.Connected = false
			//log.Printf("#%s#\n", lastMsg)
			log.Printf("Thingsboard1.LostConnect, reason: %v\n", err)
			gateway_log.Thingsboard_add_log("Thingsboard1.LostConnect" + err.Error())
		}
		client.tbOnConnectHandler = func(c mqtt.Client){
			client.Connected = true
			log.Println("Thingsboard1.OnConnect")
			gateway_log.Thingsboard_add_log("Thingsboard1.OnConnect")			
		}
	} else {
		client.tbLostConnectHandler = func(c mqtt.Client, err error){
			GwChars.SetLed_Red("1")
			client.Connected = false
			//log.Printf("#%s#\n", lastMsg)
			log.Printf("Thingsboard2.LostConnect, reason: %v\n", err)
			gateway_log.Thingsboard_add_log("Thingsboard2.LostConnect" + err.Error())
		}
		client.tbOnConnectHandler = func(c mqtt.Client){
			client.Connected = true
			log.Println("Thingsboard2.OnConnect")
			gateway_log.Thingsboard_add_log("Thingsboard2.OnConnect")			
		}
	}
	return client
}
/****************************************************************/

// Respond respond msg to host
func (c *Client) Respond(idRes, msg string) {
	topic := `v1/devices/me/rpc/response/` + idRes
	//log.Println("topic respond = ", topic)
	c.Dev.Publish(topic, 0, false, msg)
}
/*****************************************************************/

// Post post data to host
func (c *Client) Post(msg string){
	c.Dev.Publish(c.thingsboardTopicTelemetry, 0, false, msg)
}
/********************************************************************/