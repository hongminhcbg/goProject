package main

/************************************************************************/
import (
	"bytes"
	"encoding/json"
	"fmt"
	GwChars "gateway_characteristics" // rename to "GwChars"
	"gateway_commit"
	"gateway_log"
	"log"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

/************************************************************************/
func help_cmd() string {
	var help strings.Builder
	help.WriteString(`{`)
	help.WriteString(`"?     " : "show this help",`)
	help.WriteString(`"??    " : "show list variable",`)
	help.WriteString(`"Object" : "iotgateway/mysensors/adapter/monitor/mosquitto/domoticz/node"`)
	help.WriteString(`}`)
	return help.String()
}

/************************************************************************/
func help_config() string {
	json, err := json.MarshalIndent(Config, "", "  ")
	//fmt.Println(string(json))
	if err != nil {
		return `{"status" : "load config failed"}`
	}
	return string(json)
}

/************************************************************************/
func help_config_cmd() string {
	//	fmt.Println(ConfigCmd)
	return ConfigCmd
}

/*********************************************************************/
func help_adapter() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"adapter. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func help_monitor() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"monitor. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func help_mosquitto() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"mosquitto. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func help_domoticz() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"domoticz. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func help_iotgateway() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"iotgateway. " : "[ reboot | poweroff | checkupdate | commit ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func thingsboardResponse(c mqtt.Client, topic string, msg string) {
	//c.Publish(topic, 0, false, `{"` + msg + `" : ""}`)
	c.Publish(topic, 0, false, msg)

}

/************************************************************************/
const MAX_ARGS int = 7

func Parse_Cmd(total_cmd string) []string {
	list := strings.Split(strings.ToLower(total_cmd), ".")

	//var args [MAX_ARGS]string
	var args = make([]string, MAX_ARGS)

	len := len(list)

	if len > MAX_ARGS {
		len = MAX_ARGS
	}

	for i := 0; i < len; i++ {
		args[i] = list[i]
	}

	return args
}

/************************************************************************/
func IoTGateway_reboot(c mqtt.Client, topic string) {
	thingsboardResponse(c, topic, TB_TextToJSON("IoTGateway is rebooting"))
	GwChars.Reboot()
}

func IoTGateway_poweroff(c mqtt.Client, topic string) {
	gateway_log.Thingsboard_add_log("IoTGateway_poweroff : gateway power off ") //LHM add 0223
	thingsboardResponse(c, topic, TB_TextToJSON("IoTGateway is shutting down"))
	GwChars.Poweroff()
}

func IoTGateway_checkupdate(c mqtt.Client, topic string) {
	gateway_log.Thingsboard_add_log("IoTGateway_checkupdate: gateway checkupdate")
	thingsboardResponse(c, topic, TB_TextToJSON("IoTGateway is checking for updates"))
	GwChars.CheckUpdate()
}
func IoTGateway_commit(c mqtt.Client, topic string) {
	str := gateway_commit.Commit()
	//TB_display_text(c, topic, str)
	thingsboardResponse(c, topic, TB_TextToJSON(str))
	gateway_log.Thingsboard_add_log("IoTGateway_commit: commit dir IoTGateway " + str)
}

func TB_display_text(c mqtt.Client, topic string, text string) {
	var str strings.Builder
	arr_string := strings.Split(text, "\n")
	str.WriteString(`{`)
	for _, val := range arr_string {
		str.WriteString(`"":"` + val + `",`)
	}
	str.WriteString(`"":""`)
	temp_str := str.String()
	temp_str = temp_str + "}"
	fmt.Println("1:" + temp_str)
	thingsboardResponse(c, topic, temp_str)
}

func TB_TextToJSON(text string) string {
	var str strings.Builder
	str.WriteString(`{"`) // JSON begin
	for _, line := range strings.Split(text, "\n") {
		str.WriteString(line)     // JSON Value
		str.WriteString(`":"","`) // this Value end, next Key begin
	}
	str.WriteString(`":""}`) // JSON end
	//	fmt.Println("2:" + str.String())
	return str.String()
}

var IoTGatewayCommand = map[string]func(mqtt.Client, string){
	"reboot": IoTGateway_reboot,

	"poweroff": IoTGateway_poweroff,
	//	"powerOff": IoTGateway_poweroff,

	"checkupdate": IoTGateway_checkupdate,
	//	"checkUpdate": IoTGateway_checkupdate,

	"commit": IoTGateway_commit,
}

/************************************************************************/
func MQTTCallBack_ThingsboardRequest(c mqtt.Client, message mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", message.Topic())
	fmt.Printf("MSG: %s\n", message.Payload())
	dec := json.NewDecoder(bytes.NewReader(message.Payload()))
	var json_decode map[string]interface{}
	if err := dec.Decode(&json_decode); err != nil {
		log.Println(err)
		return
	}
	// fmt.Print("\n")
	// for key := range json_decode {
	// fmt.Printf("%15s   ", key)
	// fmt.Println(json_decode[key])
	// }
	if method, ok := json_decode["method"].(string); ok {
		topic := strings.Replace(message.Topic(), "request", "response", 1)
		//methods := strings.ToLower(method)
		args := Parse_Cmd(method)
		gateway_log.Thingsboard_add_log("MQTTCallBack_ThingsboardRequest(): command received [" + method + "]") //LHM add 0223
		GwChars.Sleep_ms(2000)
		switch args[0] {
		case "?":
			thingsboardResponse(c, topic, help_cmd())
		case "??":
			thingsboardResponse(c, topic, help_config())
		case "???":
			thingsboardResponse(c, topic, help_config_cmd())
		case "adapter":
			switch args[1] {
			case "restart":
				thingsboardResponse(c, topic, TB_TextToJSON("Adapter is restarting"))
				GwChars.Monitor_restart_adapter()

			default:
				thingsboardResponse(c, topic, help_adapter())
			}

		case "monitor":
			switch args[1] {
			case "restart":
				thingsboardResponse(c, topic, TB_TextToJSON("Monitor is restarting"))
				GwChars.Monitor_restart_monitor()

			default:
				thingsboardResponse(c, topic, help_monitor())
			}

		case "mosquitto":
			switch args[1] {
			case "restart":
				thingsboardResponse(c, topic, TB_TextToJSON("Mosquitto is restarting"))
				GwChars.Restart_mosquitto()
			default:
				thingsboardResponse(c, topic, help_mosquitto())
			}

		case "domoticz":
			switch args[1] {
			case "restart":
				thingsboardResponse(c, topic, TB_TextToJSON("Domoticz is restarting"))
				GwChars.Restart_domoticz()
			default:
				thingsboardResponse(c, topic, help_domoticz())
			}

		case "iotgateway":
			if fn, ok := IoTGatewayCommand[args[1]]; ok {
				fn(c, topic)

			} else {
				thingsboardResponse(c, topic, help_iotgateway())
			}

		case "node", "mysensors", "nodecmd":
			log.Println("case node, mysensors send data to adapter")
			mqtt_mosquitto.Publish(string(message.Topic()), 0, false, string(message.Payload())) // FW to mosquitto
			//thingsboardResponse(c, topic, TB_TextToJSON("case node, mysensors"))

		default:
			thingsboardResponse(c, topic, TB_TextToJSON("Unknow object"))
		}
	}
} // end func

/************************************************************************/
