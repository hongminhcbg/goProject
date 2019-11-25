package tbclientmqtt

import (
	//	"fmt"
	"log"
	"strings"
	"time"

//	"io/ioutil"
	GwChars "gatewayPackage/gateway_characteristics" // rename to "GwChars"
	"bytes"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	tbclient "gatewayPackage/tbClient"
)
/**************************************/

const thingsboardTopicRequest string 	= 		`v1/devices/me/rpc/request/+`
const thingsboardTopicTelemetry string 	= 		`v1/devices/me/telemetry`
const mqttConnectTimeout time.Duration 	= 			5
/**************************/

// Client mqtt
type Client struct {
	Dev 						mqtt.Client
	host						string
	monitorTocken				string
}
/************************************************************************/

// Start create new client
func Start(host string, monitorTocken string, CB func(tbclient.TbClient, string, string, string), idDev string) *Client {
	client := &Client{}
	client.host = host
	client.monitorTocken = 	monitorTocken

	LostConnect := func(cmqtt mqtt.Client, err error){
		GwChars.SetLed_Red("1")
		log.Printf("%s MQTT LostConnect, reason: %v\n", idDev, err)
	}

	onConnect := func(cmqtt mqtt.Client){
		log.Printf("%s MQTT OnConnect ", idDev)
	}

	onMessage := func(cmqtt mqtt.Client, message mqtt.Message){
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
			if params, ok := jsonDecode["params"].(string); ok {
				CB(tbclient.TbClient(client), idRes, method, params)
			} else {
				CB(tbclient.TbClient(client), idRes, method, "")
			}
		}			
	}

	optsThingsboard := mqtt.NewClientOptions()
	optsThingsboard.AddBroker(client.host)
	optsThingsboard.SetUsername(client.monitorTocken)
	optsThingsboard.SetConnectionLostHandler(LostConnect)
	optsThingsboard.SetOnConnectHandler(onConnect)
	client.Dev = mqtt.NewClient(optsThingsboard)
	client.Dev.Connect().WaitTimeout(mqttConnectTimeout * time.Second)
	client.Dev.Subscribe(thingsboardTopicRequest, 0, onMessage)
	return client
}
/************************************************/

// SetupCallback set up callback function for this client    responseHandler
// func (c *Client) SetupCallback(CB func(tbclient.TbClient, string, string), idDev string){
// }
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
	c.Dev.Publish(thingsboardTopicTelemetry, 0, false, msg)
}
/********************************************************************/