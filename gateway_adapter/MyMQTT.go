package main

/************************************************************************/
import (
	"log"
	"time"
	"github.com/eclipse/paho.mqtt.golang"
	"gateway_log"
)

/************************************************************************/
const THINGSBOARD_TOPIC_TELEMETRY string = `v1/devices/me/telemetry`
const THINGSBOARD_TOPIC_REQUEST   string = `v1/devices/me/rpc/request/+`
const THINGSBOARD_TOPIC_RESPONSE  string = `v1/devices/me/rpc/response/+`

const MOSQUITTO_HOST        string = `tcp://localhost:1883`
const DOMITICZ_TOPIC_OUT    string = `domoticz/out`
const DOMITICZ_TOPIC_IN     string = `domoticz/in`

/// mqtt client
var  mqtt_mosquitto    mqtt.Client

const mqtt_disconnect_timeout uint       = 5000
const mqtt_connect_timeout time.Duration = 5


/************************************************************************/
func mosquitto_LostConnect_Handler(c mqtt.Client, err error) {
  log.Printf("Mosquitto.LostConnect, reason: %v\n", err)
	gateway_log.Thingsboard_add_log("Mosquitto.LostConnect")
}

func mosquitto_OnConnect_Handler(c mqtt.Client) {
  log.Println("Mosquitto.OnConnect")
	gateway_log.Thingsboard_add_log("Mosquitto.OnConnect")
	
	c.Subscribe(DOMITICZ_TOPIC_OUT, 0, MosquittoCallBack_DomoticzOut)
	c.Subscribe(THINGSBOARD_TOPIC_REQUEST, 0, MosquittoCallBack_ThingsboardRequest)
}


/************************************************************************/
func mqtt_mosquitto_reconnect() {
	//thingsboard_add_log("Begin Mosquitto")
	opts_mosquitto := mqtt.NewClientOptions()
	opts_mosquitto.AddBroker(MOSQUITTO_HOST)
	opts_mosquitto.SetConnectionLostHandler(mosquitto_LostConnect_Handler)
	opts_mosquitto.SetOnConnectHandler(mosquitto_OnConnect_Handler)
	
	if mqtt_mosquitto != nil && mqtt_mosquitto.IsConnected() {
		mqtt_mosquitto.Disconnect(mqtt_disconnect_timeout)
	}
	mqtt_mosquitto = mqtt.NewClient(opts_mosquitto)
	mqtt_mosquitto.Connect().WaitTimeout(mqtt_connect_timeout * time.Second)
}

/************************************************************************/

