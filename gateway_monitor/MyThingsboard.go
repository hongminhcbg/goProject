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
func helpCmd() string {
	var help strings.Builder
	help.WriteString(`{`)
	help.WriteString(`"?     " : "show this help",`)
	help.WriteString(`"??    " : "show list variable",`)
	help.WriteString(`"Object" : "iotgateway/mysensors/adapter/monitor/mosquitto/domoticz/node"`)
	help.WriteString(`}`)
	return help.String()
}

/************************************************************************/
func helpConfig() string {
	json, err := json.MarshalIndent(Config, "", "  ")
	//fmt.Println(string(json))
	if err != nil {
		return `{"status" : "load config failed"}`
	}
	return string(json)
}

/************************************************************************/
func helpConfigCmd() string {
	//	fmt.Println(ConfigCmd)
	return ConfigCmd
}

/*********************************************************************/
func helpAdapter() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"adapter. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func helpMonitor() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"monitor. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func helpMosquitto() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"mosquitto. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func helpDomoticz() string {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"domoticz. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func helpIotgateway() string {
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
const maxArgs int = 7

// ParseCmd convert string to array split by "." VD: nguyen.hong.minh => [nguyen, hong, minh]
func ParseCmd(totalCmd string) []string {
	list := strings.Split(strings.ToLower(totalCmd), ".")

	//var args [maxArgs]string
	var args = make([]string, maxArgs)

	len := len(list)

	if len > maxArgs {
		len = maxArgs
	}

	for i := 0; i < len; i++ {
		args[i] = list[i]
	}

	return args
}

/************************************************************************/
// func IoTGateway_reboot(c mqtt.Client, topic string) {
// 	thingsboardResponse(c, topic, TBTextToJSON("IoTGateway is rebooting"))
// 	GwChars.Reboot()
// }

// func IoTGateway_poweroff(c mqtt.Client, topic string) {
// 	gateway_log.Thingsboard_add_log("IoTGateway_poweroff : gateway power off ") //LHM add 0223
// 	thingsboardResponse(c, topic, TBTextToJSON("IoTGateway is shutting down"))
// 	GwChars.Poweroff()
// }

// func IoTGateway_checkupdate(c mqtt.Client, topic string) {
// 	gateway_log.Thingsboard_add_log("IoTGateway_checkupdate: gateway checkupdate")
// 	thingsboardResponse(c, topic, TBTextToJSON("IoTGateway is checking for updates"))
// 	GwChars.CheckUpdate()
// }
// func IoTGateway_commit(c mqtt.Client, topic string) {
// 	str := gateway_commit.Commit()
// 	//TB_display_text(c, topic, str)
// 	thingsboardResponse(c, topic, TBTextToJSON(str))
// 	gateway_log.Thingsboard_add_log("IoTGateway_commit: commit dir IoTGateway " + str)
// }
/********************************************************/

// TBTextToJSON convert text to json string, minh\n123 ====> {"minh":"", "123":"", "":""}
func TBTextToJSON(text string) string {
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

// var IoTGatewayCommand = map[string]func(mqtt.Client, string){
// 	"reboot": IoTGateway_reboot,

// 	"poweroff": IoTGateway_poweroff,
// 	//	"powerOff": IoTGateway_poweroff,

// 	"checkupdate": IoTGateway_checkupdate,
// 	//	"checkUpdate": IoTGateway_checkupdate,

// 	"commit": IoTGateway_commit,
// }
/**********************************************/

// doNothing if both tb1 and tb2 use semilar protocol => tbPostFuncMap["TB2"] = doNothing
// avoid post double time 
func doNothing1arg(msg string){
	log.Println("2")
}
/************************************************************************/

// MQTTCallBackThingsboardRequest process all message from TB
func MQTTCallBackThingsboardRequest(c mqtt.Client, message mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", message.Topic())
	fmt.Printf("MSG: %s\n", message.Payload())
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
		processAllCommand(idRes, method, "MQTT")
	}
} // end func

/************************************************************************/

// processAllCommand process all command of user send form tb
func processAllCommand(idRes, method, protocol string){
	if respondFunction, ok := tbResFuncMap[protocol]; ok{
		//topic := strings.Replace(message.Topic(), "request", "response", 1)
		args := ParseCmd(method)
		gateway_log.Thingsboard_add_log("MQTTCallBack_ThingsboardRequest(): command received [" + method + "]") //LHM add 0223
		GwChars.Sleep_ms(2000)
		switch args[0] {
			case "?":
				respondFunction(idRes, helpCmd())
			case "??":
				respondFunction(idRes, helpConfig())
			case "???":
				respondFunction(idRes, helpConfigCmd())
			case "adapter":
				switch args[1] {
				case "restart":
					respondFunction(idRes, TBTextToJSON("Adapter is restarting"))
					GwChars.Monitor_restart_adapter()

				default:
					respondFunction(idRes, helpAdapter())
				}

			case "monitor":
				switch args[1] {
				case "restart":
					respondFunction(idRes, TBTextToJSON("Monitor is restarting"))
					GwChars.Monitor_restart_monitor()

				default:
					respondFunction(idRes, helpMonitor())
				}

			case "mosquitto":
				switch args[1] {
				case "restart":
					respondFunction(idRes, TBTextToJSON("Mosquitto is restarting"))
					GwChars.Restart_mosquitto()
				default:
					respondFunction(idRes, helpMosquitto())
				}

			case "domoticz":
				switch args[1] {
				case "restart":
					respondFunction(idRes, TBTextToJSON("Domoticz is restarting"))
					GwChars.Restart_domoticz()
				default:
					respondFunction(idRes, helpDomoticz())
				}

			case "iotgateway":
				switch args[1] {
					case "reboot":
						respondFunction(idRes, TBTextToJSON("IoTGateway is rebooting"))
						GwChars.Reboot()
					case "poweroff":
						respondFunction(idRes, TBTextToJSON("IoTGateway is rebooting"))
						GwChars.Reboot()
					case "checkupdate":
						gateway_log.Thingsboard_add_log("IoTGateway_checkupdate: gateway checkupdate")
						respondFunction(idRes, TBTextToJSON("IoTGateway is checking for updates"))
						GwChars.CheckUpdate()
					case "commit":
						str := gateway_commit.Commit()
						respondFunction(idRes, TBTextToJSON(str))
						gateway_log.Thingsboard_add_log("IoTGateway_commit: commit dir IoTGateway " + str)
					default:
						respondFunction(idRes, helpDomoticz())
				}

			case "node", "mysensors", "nodecmd":
				log.Println("case node, mysensors send data to adapter")
				topicMos := `v1/devices/me/rpc/request/` + idRes
				payload := `{"method":"` + method + `"}`
				mqtt_mosquitto.Publish(topicMos, 0, false, payload) // FW to mosquitto
				//respondFunction(idRes, TBTextToJSON("case node, mysensors"))
			default:
				respondFunction(idRes, TBTextToJSON("Unknow object"))
		}
		
	} else {
		gateway_log.Thingsboard_add_log("protocol not Subport")
	}

}
/*************************************************************************/