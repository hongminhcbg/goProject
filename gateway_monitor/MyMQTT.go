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
	GwChars "gateway_characteristics" // rename to "GwChars"
	checkMos "gateway_checkMos"
	"gateway_log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/************************************************************************/
const THINGSBOARD_TOPIC_TELEMETRY string = `v1/devices/me/telemetry`
const THINGSBOARD_TOPIC_REQUEST string = `v1/devices/me/rpc/request/+`
const THINGSBOARD_TOPIC_RESPONSE string = `v1/devices/me/rpc/response/+`

const MOSQUITTO_HOST string = `tcp://localhost:1883`
const DOMITICZ_TOPIC_OUT string = `domoticz/out`
const DOMITICZ_TOPIC_IN string = `domoticz/in`
const CHECKMOS_TOPIC string = `checkMosquitto`

/// mqtt client
var mqtt_thingsboard mqtt.Client
var mqttThingsBoard2 mqtt.Client
var mqtt_mosquitto mqtt.Client

var thingsboard_1_connected bool = true
var thingsboard_2_connected bool = true

const mqtt_disconnect_timeout uint = 5000
const mqtt_connect_timeout time.Duration = 5

/************************************************************************/
var domoticz_rx_count int = 0
var domoticz_rx_msg string = ""
var checkMosValue string

func MosquittoCallBack_DomoticzOut(c mqtt.Client, message mqtt.Message) {
	domoticz_rx_msg = strings.Replace(string(message.Payload()), `"`, ``, -1) // remove all "
	domoticz_rx_count++
}

/************************************************************************/
func MosquittoCallBack_AdapterOut(c mqtt.Client, message mqtt.Message) {
	//	fmt.Println( string(message.Payload()) )
	//	fmt.Println()
	ThingsboardJson.AddObject(string(message.Payload()))
}

/************************************************************************/
func MosquittoCallBack_AdapterResponse(c mqtt.Client, message mqtt.Message) {
	log.Printf("TOPIC: %s\n", message.Topic())
	log.Printf("MSG k: %s\n", message.Payload())
	topic_str := string(message.Topic())
	arr_id := strings.Split(topic_str, "/")
	idRes := arr_id[len(arr_id)-1]
	if Config.Thingsboard_1.Enable == true {
		if useMqttTB1 == true { // mqtt
			mqtt_thingsboard.Publish(string(message.Topic()), 0, false, string(message.Payload()))
		} else {
			ThingsboardResponseHTTP(string(message.Payload()), TB1HttpClient, idRes)
			log.Print("id res = ", idRes, "\n")
		}
	}
	if Config.Thingsboard_2.Enable == true {
		if useMqttTB2 == true {
			mqttThingsBoard2.Publish(string(message.Topic()), 0, false, string(message.Payload()))
		} else {
			ThingsboardResponseHTTP(string(message.Payload()), TB2HttpClient, idRes)
		}
	}
}

/************************************************************************/
func thingsboard_LostConnect_Handler(c mqtt.Client, err error) {
	GwChars.SetLed_Red("1")

	thingsboard_1_connected = false
	//log.Printf("#%s#\n", lastMsg)
	log.Printf("Thingsboard.LostConnect, reason: %v\n", err)
	gateway_log.Thingsboard_add_log("Thingsboard.LostConnect" + err.Error())
}

func thingsboard_OnConnect_Handler(c mqtt.Client) {
	thingsboard_1_connected = true

	log.Println("Thingsboard.OnConnect")
	gateway_log.Thingsboard_add_log("Thingsboard.OnConnect")
	c.Subscribe(THINGSBOARD_TOPIC_REQUEST, 0, MQTTCallBack_ThingsboardRequest)
}

/************************************************************************/
// func thingspeak_LostConnect_Handler(c mqtt.Client, err error) {
// log.Printf("Thingspeak LostConnect_Handler, reason: %v\n", err)
// }

// func thingspeak_OnConnect_Handler(c mqtt.Client) {
// log.Printf("Thingspeak OnConnect_Handler\n")
// }

/************************************************************************/
func amazon_LostConnect_Handler(c mqtt.Client, err error) {
	thingsboard_2_connected = false

	log.Printf("AWS.LostConnect, reason: %v\n", err)
	gateway_log.Thingsboard_add_log("AWS.LostConnect")
}

func amazon_OnConnect_Handler(c mqtt.Client) {
	thingsboard_2_connected = true

	log.Println("AWS.OnConnect")
	gateway_log.Thingsboard_add_log("AWS.OnConnect")
	c.Subscribe(THINGSBOARD_TOPIC_REQUEST, 0, MQTTCallBack_ThingsboardRequest)
}

/************************************************************************/
func mosquitto_LostConnect_Handler(c mqtt.Client, err error) {
	log.Printf("Mosquitto.LostConnect, reason: %v\n", err)
	gateway_log.Thingsboard_add_log("Mosquitto.LostConnect")
}

func mosquitto_OnConnect_Handler(c mqtt.Client) {
	log.Println("Mosquitto.OnConnect")
	gateway_log.Thingsboard_add_log("Mosquitto.OnConnect")

	/*****************************************************/
	//check mosquitto startup publish message and receive this message
	checMostemp := checkMos.GetStatus(c, CHECKMOS_TOPIC, 2000)
	log.Printf("mosquitto startup: %t", checMostemp)
	if !checMostemp {
		//c.Disconnect(1000)
		GwChars.Sleep_ms(1000)
		//c.Connect().WaitTimeout(mqtt_connect_timeout * time.Second)
		log.Println("log mosquitto")
		mqtt_mosquitto_reconnect()
		return
	}
	if Config.Gateway.Debug_domoticz > 0 {
		c.Subscribe(DOMITICZ_TOPIC_OUT, 0, MosquittoCallBack_DomoticzOut) // for Debug domoticz
	}

	c.Subscribe(THINGSBOARD_TOPIC_TELEMETRY, 0, MosquittoCallBack_AdapterOut)     // adapter_data
	c.Subscribe(THINGSBOARD_TOPIC_RESPONSE, 0, MosquittoCallBack_AdapterResponse) // adapter_response
	/*****************************************************/
}

/************************************************************************/
// func AWS_TlsConfig() *tls.Config {
// 	File_rootCA := Config.Thingsboard_2.RootCA
// 	File_cert   := Config.Thingsboard_2.MonitorClientCert
// 	File_key    := Config.Thingsboard_2.MonitorClientKey

// 	// Import trusted certificates from CAfile.pem
// 	// Alternatively, manually add CA certificates to default openssl CA bundle
// 	certpool := x509.NewCertPool()
// 	pemCerts, err := ioutil.ReadFile(File_rootCA)
// 	if err == nil {
// 		certpool.AppendCertsFromPEM(pemCerts)
// 	}

// 	// Import client certificate/key pair
// 	cert, err := tls.LoadX509KeyPair(File_cert, File_key)  // (public, private)
// 	if err != nil {
// 		//panic(err)
// 	}
// 	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
// 	if err != nil {
// 		//panic(err)
// 	}

// 	// Create tls.Config with desired tls properties
// 	return &tls.Config{
// 		RootCAs: certpool,
// 		InsecureSkipVerify: true,
// 		Certificates: []tls.Certificate{cert},
// 	}
// }

/************************************************************************/
func mqtt_thingsboard_reconnect() {
	//thingsboard_add_log("Begin Thingsboard")
	opts_thingsboard := mqtt.NewClientOptions()
	opts_thingsboard.AddBroker(Config.Thingsboard_1.Host)
	opts_thingsboard.SetUsername(Config.Thingsboard_1.MonitorToken)
	opts_thingsboard.SetConnectionLostHandler(thingsboard_LostConnect_Handler)
	opts_thingsboard.SetOnConnectHandler(thingsboard_OnConnect_Handler)
	if mqtt_thingsboard != nil && mqtt_thingsboard.IsConnected() {
		mqtt_thingsboard.Disconnect(mqtt_disconnect_timeout)
	}
	mqtt_thingsboard = mqtt.NewClient(opts_thingsboard)
	mqtt_thingsboard.Connect().WaitTimeout(mqtt_connect_timeout * time.Second)
}

/************************************************************************/
func mqtt_amazon_reconnect() {
	//thingsboard_add_log("Begin AWS")
	opts_amazon := mqtt.NewClientOptions()
	opts_amazon.AddBroker(Config.Thingsboard_2.Host)
	//opts_amazon.SetTLSConfig( AWS_TlsConfig() )
	opts_amazon.SetUsername(Config.Thingsboard_2.MonitorToken)
	opts_amazon.SetConnectionLostHandler(amazon_LostConnect_Handler)
	opts_amazon.SetOnConnectHandler(amazon_OnConnect_Handler)
	if mqttThingsBoard2 != nil && mqttThingsBoard2.IsConnected() {
		mqttThingsBoard2.Disconnect(mqtt_disconnect_timeout)
	}
	mqttThingsBoard2 = mqtt.NewClient(opts_amazon)
	mqttThingsBoard2.Connect().WaitTimeout(mqtt_connect_timeout * time.Second)
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
