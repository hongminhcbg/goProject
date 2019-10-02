package main

/************************************************************************/
import (
//	"bytes"
	"encoding/json"
	"fmt"
	GwChars "gateway_characteristics" // rename to "GwChars"
	"gateway_commit"
	"gateway_log"
	"log"
	"strings"
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
/**********************************************/

// processAllCommand process all command of user send form tb
func processAllCommand(c interface{}, idRes, method string){
	client := c.(tbClient)
	fmt.Println(idRes, method)
	args := ParseCmd(method)
	gateway_log.Thingsboard_add_log("MQTTCallBack_ThingsboardRequest(): command received [" + method + "]") //LHM add 0223
	GwChars.Sleep_ms(2000)
	switch args[0] {
		case "?":
			client.Respond(idRes, helpCmd())
		case "??":
			client.Respond(idRes, helpConfig())
		case "???":
			client.Respond(idRes, helpConfigCmd())
		case "adapter":
			switch args[1] {
			case "restart":
				client.Respond(idRes, TBTextToJSON("Adapter is restarting"))
				GwChars.Monitor_restart_adapter()

			default:
				client.Respond(idRes, helpAdapter())
			}

		case "monitor":
			switch args[1] {
			case "restart":
				client.Respond(idRes, TBTextToJSON("Monitor is restarting"))
				GwChars.Monitor_restart_monitor()

			default:
				client.Respond(idRes, helpMonitor())
			}

		case "mosquitto":
			switch args[1] {
			case "restart":
				client.Respond(idRes, TBTextToJSON("Mosquitto is restarting"))
				GwChars.Restart_mosquitto()
			default:
				client.Respond(idRes, helpMosquitto())
			}

		case "domoticz":
			switch args[1] {
			case "restart":
				client.Respond(idRes, TBTextToJSON("Domoticz is restarting"))
				GwChars.Restart_domoticz()
			default:
				client.Respond(idRes, helpDomoticz())
			}

		case "iotgateway":
			switch args[1] {
				case "reboot":
					client.Respond(idRes, TBTextToJSON("IoTGateway is rebooting"))
					GwChars.Reboot()
				case "poweroff":
					client.Respond(idRes, TBTextToJSON("IoTGateway is rebooting"))
					GwChars.Reboot()
				case "checkupdate":
					gateway_log.Thingsboard_add_log("IoTGateway_checkupdate: gateway checkupdate")
					client.Respond(idRes, TBTextToJSON("IoTGateway is checking for updates"))
					GwChars.CheckUpdate()
				case "commit":
					str := gateway_commit.Commit()
					client.Respond(idRes, TBTextToJSON(str))
					gateway_log.Thingsboard_add_log("IoTGateway_commit: commit dir IoTGateway " + str)
				default:
					client.Respond(idRes, helpDomoticz())
			}

		case "node", "mysensors", "nodecmd":
			log.Println("case node, mysensors send data to adapter")
			topicMos := `v1/devices/me/rpc/request/` + idRes
			payload := `{"method":"` + method + `"}`
			mqtt_mosquitto.Publish(topicMos, 0, false, payload) // FW to mosquitto
			//client.Respond(idRes, TBTextToJSON("case node, mysensors"))
		default:
			client.Respond(idRes, TBTextToJSON("Unknow object"))
	}
}
/*************************************************************************/