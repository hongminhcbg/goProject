package gatewaymqtttb

import (
	//	"fmt"
	"log"
//	"strings"
	"time"

	//	"io/ioutil"
	//	"crypto/tls"
	//	"crypto/x509"
	GwChars "gateway_characteristics" // rename to "GwChars"
	//checkMos "gateway_checkMos"
	"gateway_log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)
/*********************/

const thingsboardTopicTelemetry string = `v1/devices/me/telemetry`
const thingsboardTopicRequest string = `v1/devices/me/rpc/request/+`
const thingsboardTopicResponse string = `v1/devices/me/rpc/response/+`

/// mqtt client
var mqttThingsboard [2]mqtt.Client

// ThingsboardConnected check status of device
var ThingsboardConnected = [2]bool{false, false}

// lostConnectFuncArr array store all callback function when lost connect
var lostConnectFuncArr = [2]func(mqtt.Client, error){thingsboard1LostConnectHandler, thingsboard2LostConnectHandler}

// onConnectFuncArr array store all callback function when lost connect
var onConnectFuncArr = [2]func(mqtt.Client){thingsboard1OnConnectHandler, thingsboard2OnConnectHandler}

// mqttDisconnectTimeout set timeout when connect 
const mqttDisconnectTimeout uint = 5000

const mqttConnectTimeout time.Duration = 5
/************************************************************************/

var domoticzRxCount int = 0
var domoticzRxMsg string = ""
var checkMosValue string
/************************************************************************/

// thingsboardLostConnectHandler process event lost connect
func thingsboard1LostConnectHandler(c mqtt.Client, err error) {
	GwChars.SetLed_Red("1")
	ThingsboardConnected[0] = false
	//log.Printf("#%s#\n", lastMsg)
	log.Printf("Thingsboard1.LostConnect, reason: %v\n", err)
	gateway_log.Thingsboard_add_log("Thingsboard1.LostConnect" + err.Error())
}
/*********************************/

// thingsboardLostConnectHandler process event lost connect
func thingsboard2LostConnectHandler(c mqtt.Client, err error) {
	GwChars.SetLed_Red("1")
	ThingsboardConnected[1] = false
	//log.Printf("#%s#\n", lastMsg)
	log.Printf("Thingsboard1.LostConnect, reason: %v\n", err)
	gateway_log.Thingsboard_add_log("Thingsboard1.LostConnect" + err.Error())
}
/*********************************/

// thingsboard1OnConnectHandler process when on conect
func thingsboard1OnConnectHandler(c mqtt.Client) {
	ThingsboardConnected[0] = true
	log.Println("Thingsboard1.OnConnect")
	gateway_log.Thingsboard_add_log("Thingsboard1.OnConnect")
//	c.Subscribe(thingsboardTopicRequest, 0, MQTTCallBackThingsboardRequest)
}
/**********************************************************/

// thingsboard2OnConnectHandler process when on conect
func thingsboard2OnConnectHandler(c mqtt.Client) {
	ThingsboardConnected[1] = true
	log.Println("Thingsboard2.OnConnect")
	gateway_log.Thingsboard_add_log("Thingsboard2.OnConnect")
//	c.Subscribe(thingsboardTopicRequest, 0, MQTTCallBackThingsboardRequest)
}
/**********************************************************/

// MqttThingsboardReconnect setup mqtt thingsboard1
func MqttThingsboardReconnect(host, tocken string, callBackTB func(mqtt.Client, mqtt.Message), index int) {
	optsThingsboard := mqtt.NewClientOptions()
	optsThingsboard.AddBroker(host)
	optsThingsboard.SetUsername(tocken)
	optsThingsboard.SetConnectionLostHandler(lostConnectFuncArr[index])
	optsThingsboard.SetOnConnectHandler(onConnectFuncArr[index])
	if mqttThingsboard[index] != nil && mqttThingsboard[index].IsConnected() {
		mqttThingsboard[index].Disconnect(mqttDisconnectTimeout)
	}
	mqttThingsboard[index] = mqtt.NewClient(optsThingsboard)
	mqttThingsboard[index].Connect().WaitTimeout(mqttConnectTimeout * time.Second)
	mqttThingsboard[index].Subscribe(thingsboardTopicRequest, 0, callBackTB)
}
/**************************************/

// RespondMgs respond data to host, publish data to v1/devices/me/rpc/response/id (id is int number TB send)
func RespondMgs(id,msg string){
	if ThingsboardConnected[0] == false {
		log.Println("can't connect to host 1")
	} else {
		topic := `v1/devices/me/rpc/response/` + id
		mqttThingsboard[0].Publish(topic, 0, false, msg)
	}

	if ThingsboardConnected[1] == false {
		log.Println("can't connect to host 2")
	} else {
		topic := `v1/devices/me/rpc/response/` + id
		mqttThingsboard[1].Publish(topic, 0, false, msg)
	}
}
/****************************************/

// SendMgs send msg to TB, msg is json string
func SendMgs(msg string){
	if ThingsboardConnected[0] == true {
		mqttThingsboard[0].Publish(thingsboardTopicTelemetry, 0, false, msg)
	}
	if ThingsboardConnected[1] == true {
		mqttThingsboard[1].Publish(thingsboardTopicTelemetry, 0, false, msg)
	}
}

/******************************************/