package gatewaymqtttb

import (
	//	"fmt"
	"log"
	"strings"
	"time"

//	"io/ioutil"
	GwChars "gateway_characteristics" // rename to "GwChars"
//	"gateway_log"
	"bytes"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)
/**************************/

// Client mqtt
type Client struct {
	Dev 						mqtt.Client
	thingsboardTopicTelemetry 	string
	thingsboardTopicRequest 	string
	thingsboardTopicResponse 	string
	mqttDisconnectTimeout 		uint
	mqttConnectTimeout 			time.Duration
	idDev						string
	host						string
	monitorTocken				string
}
/************************************************************************/

// NewClient create new client
func NewClient(host string, monitorTocken string, idDev string) *Client {
	client := &Client{}
	client.thingsboardTopicRequest = 		`v1/devices/me/rpc/request/+`
	client.thingsboardTopicResponse = 		`v1/devices/me/rpc/response/+`
	client.thingsboardTopicTelemetry = 		`v1/devices/me/telemetry`
	client.mqttConnectTimeout = 			5
	client.mqttDisconnectTimeout = 			5000
	client.idDev = 							idDev
	client.host = 							host
	client.monitorTocken = 					monitorTocken
	return client

	// client.tbLostConnectHandler = func(c mqtt.Client, err error){
	// 	GwChars.SetLed_Red("1")
	// 	log.Printf("%s MQTT LostConnect, reason: %v\n", client.idDev, err)
	// }
	// client.tbOnConnectHandler = func(c mqtt.Client){
	// 	log.Printf("%s OnConnect MQTT", client.idDev)
	// }
	// client.tbCallBackMqtt = func(c mqtt.Client, message mqtt.Message){
	// 	dec := json.NewDecoder(bytes.NewReader(message.Payload()))
	// 	var jsonDecode map[string]interface{}
	// 	if err := dec.Decode(&jsonDecode); err != nil {
	// 		log.Println(err)
	// 		return
	// 	}
	// 	if method, ok := jsonDecode["method"].(string); ok {
	// 		topicStr := string(message.Topic())
	// 		arrID := strings.Split(topicStr, "/")
	// 		idRes := arrID[len(arrID)-1]
	// 		client.processAllCommand(idRes, method, client.idDev)
	// 	}			
	// }

	// optsThingsboard := mqtt.NewClientOptions()
	// optsThingsboard.AddBroker(client.host)
	// optsThingsboard.SetUsername(client.monitorTocken)
	// optsThingsboard.SetConnectionLostHandler(client.tbLostConnectHandler)
	// optsThingsboard.SetOnConnectHandler(client.tbOnConnectHandler)
	// client.Dev = mqtt.NewClient(optsThingsboard)
	// client.Dev.Connect().WaitTimeout(client.mqttConnectTimeout * time.Second)
	// client.Dev.Subscribe(client.thingsboardTopicRequest, 0, client.tbCallBackMqtt)
	// return client
}
/************************************************/

// SetupCallback set up callback function for this client
func (c *Client) SetupCallback(CB func(interface{}, string, string)){
	LostConnect := func(cmqtt mqtt.Client, err error){
		GwChars.SetLed_Red("1")
		log.Printf("%s MQTT LostConnect, reason: %v\n", c.idDev, err)
	}

	onConnect := func(cmqtt mqtt.Client){
		log.Printf("%s MQTT OnConnect ", c.idDev)
	}

	callBackmqtt := func(cmqtt mqtt.Client, message mqtt.Message){
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
			CB(c, idRes, method)
		}			
	}

	optsThingsboard := mqtt.NewClientOptions()
	optsThingsboard.AddBroker(c.host)
	optsThingsboard.SetUsername(c.monitorTocken)
	optsThingsboard.SetConnectionLostHandler(LostConnect)
	optsThingsboard.SetOnConnectHandler(onConnect)
	c.Dev = mqtt.NewClient(optsThingsboard)
	c.Dev.Connect().WaitTimeout(c.mqttConnectTimeout * time.Second)
	c.Dev.Subscribe(c.thingsboardTopicRequest, 0, callBackmqtt)
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