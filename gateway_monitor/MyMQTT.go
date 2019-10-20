package main

/************************************************************************/
import (
	//	"fmt"
	"log"
	"strings"
	"time"

	//	"io/ioutil"
	//	"crypto/tls"
	//	"crypto/x509"
	GwChars "gatewayPackage/gateway_characteristics" // rename to "GwChars"
	checkMos "gatewayPackage/gateway_checkMos"
	gateway_log "gatewayPackage/gateway_log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
//	GMT "gateway.mqtt.tb"
)

/************************************************************************/

const thingsboardTopicTelemetry string = `v1/devices/me/telemetry`

const thingsboardTopicRequest string = `v1/devices/me/rpc/request/+`

const thingsboardTopicResponse string = `v1/devices/me/rpc/response/+`

const mosquittoHost string = `tcp://localhost:1883`

const domiticzTopicOut string = `domoticz/out`

const domiticzTopicIn string = `domoticz/in`

const checkmosTopic string = `checkMosquitto`

/// mqtt client
//var mqtt_thingsboard mqtt.Client
//var mqttThingsBoard2 mqtt.Client
var mqtt_mosquitto mqtt.Client

var thingsboard1connected = true
var thingsboard2connected = true

const mqttDisconnectTimeout uint = 5000
const mqttConnectTimeout time.Duration = 5

/************************************************************************/
var domoticz_rx_count int = 0
var domoticz_rx_msg string = ""
var checkMosValue string

// MosquittoCallBackDomoticzOut for debug domoticz
func MosquittoCallBackDomoticzOut(c mqtt.Client, message mqtt.Message) {
	domoticz_rx_msg = strings.Replace(string(message.Payload()), `"`, ``, -1) // remove all "
	domoticz_rx_count++
}
/************************************************************************/

// MosquittoCallBackAdapterOut callback function to debug
func MosquittoCallBackAdapterOut(c mqtt.Client, message mqtt.Message) {
	//	fmt.Println( string(message.Payload()) )
	//	fmt.Println()
	ThingsboardJson.AddObject(string(message.Payload()))
}
/************************************************************************/

// MosquittoCallBackAdapterResponse massage received form adapter
func MosquittoCallBackAdapterResponse(c mqtt.Client, message mqtt.Message) {
	log.Printf("TOPIC adapter -> monitor: %s\n", message.Topic())
	log.Printf("MSG adapter -> monitor: %s\n", message.Payload())
	topicStr := string(message.Topic())
	arrID := strings.Split(topicStr, "/")
	idRes := arrID[len(arrID)-1]
	tb1.Respond(idRes, string(message.Payload()))
	tb2.Respond(idRes, string(message.Payload()))
}
/************************************************************************/

// mosquittoLostConnectHandler thingsboard2 lost connect
func mosquittoLostConnectHandler(c mqtt.Client, err error) {
	log.Printf("Mosquitto.LostConnect, reason: %v\n", err)
	gateway_log.Thingsboard_add_log("Mosquitto.LostConnect")

}
/***************************************/

// mosquittoOnConnectHandler mosquitto on connect function
func mosquittoOnConnectHandler(c mqtt.Client) {
	log.Println("Mosquitto.OnConnect")
	gateway_log.Thingsboard_add_log("Mosquitto.OnConnect")

	/*****************************************************/
	//check mosquitto startup publish message and receive this message
	checMostemp := checkMos.GetStatus(c, checkmosTopic, 2000)
	log.Printf("mosquitto startup: %t", checMostemp)
	if !checMostemp {
		//c.Disconnect(1000)
		GwChars.Sleep_ms(1000)
		//c.Connect().WaitTimeout(mqttConnectTimeout * time.Second)
		log.Println("log mosquitto")
		mqttMosquittoReconnect()
		return
	}
	if Config.Gateway.Debug_domoticz > 0 {
		c.Subscribe(domiticzTopicOut, 0, MosquittoCallBackDomoticzOut) // for Debug domoticz
	}

	c.Subscribe(thingsboardTopicTelemetry, 0, MosquittoCallBackAdapterOut)     // adapter_data `v1/devices/me/telemetry`
	c.Subscribe(thingsboardTopicResponse, 0, MosquittoCallBackAdapterResponse) // adapter_response `v1/devices/me/rpc/response/+`
}
/************************************************************************/

// mqttMosquittoReconnect setup domoticz, adapter comunication with monitor
func mqttMosquittoReconnect() {
	//thingsboard_add_log("Begin Mosquitto")
	optsMosquitto := mqtt.NewClientOptions()
	optsMosquitto.AddBroker(mosquittoHost)
	optsMosquitto.SetConnectionLostHandler(mosquittoLostConnectHandler)
	optsMosquitto.SetOnConnectHandler(mosquittoOnConnectHandler)
	if mqtt_mosquitto != nil && mqtt_mosquitto.IsConnected() {
		mqtt_mosquitto.Disconnect(mqttDisconnectTimeout)
	}
	mqtt_mosquitto = mqtt.NewClient(optsMosquitto)
	mqtt_mosquitto.Connect().WaitTimeout(mqttConnectTimeout * time.Second)
}
/************************************************************************/
