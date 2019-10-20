package main

/************************************************************************/
import (
	"fmt"
	"strconv"
	"strings"
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"bytes"
	"log"
	gateway_log "gatewayPackage/gateway_log"
	GwChars "gatewayPackage/gateway_characteristics"  // rename to "GwChars"
	GP "gatewayPackage/gateway_parse"
)
/************************************************************************/
const CMD_RESTART         string = "FF"

const CMD_RENEW_ID        string = "01"
const CMD_RENEW_PARENT    string = "02"
const CMD_RF_TXPOWER      string = "03"
const CMD_RF_TXRETRY      string = "04"
const CMD_RF2_TXPOWER     string = "05"
const CMD_RF2_TXRETRY     string = "06"
const CMD_RF2_CBA         string = "07"
const CMD_RF2_ACTIVE      string = "08"
const CMD_LED_OFF         string = "09"
const CMD_LED_ON          string = "0A"

const CMD_REPORT_FULL     string = "00"
const CMD_REPORT          string = "0B"


/************************************************************************/
func IoTGateway_reboot(c mqtt.Client, topic string) {
	controlResponse(c, topic, "IoTGateway is rebooting")
	GwChars.Reboot()
}

func IoTGateway_poweroff(c mqtt.Client, topic string) {
	controlResponse(c, topic, "IoTGateway is shutting down")
	GwChars.Poweroff()
}

func IoTGateway_checkupdate(c mqtt.Client, topic string) {
	controlResponse(c, topic, "IoTGateway is checking for updates")
	GwChars.CheckUpdate()
}

var IoTGatewayCommand = map[string] func(mqtt.Client, string) {
	"reboot": IoTGateway_reboot,
	
	"poweroff": IoTGateway_poweroff,
	"powerOff": IoTGateway_poweroff,
	
	"checkupdate": IoTGateway_checkupdate,
	"checkUpdate": IoTGateway_checkupdate,
}


/************************************************************************/
var MySensorsCommand = map[string] func(mqtt.Client, string) {
	"map": MySensors_MapVisualization,
}


/************************************************************************/
var parents = make(map[int64][]int64, 256)
var  num_child_f1 int =  0
var last_child_f1 int64 = -1
var level = 0

var tree strings.Builder


/************************************************************************/
func count_child(id int64) (int64) {
	num_child := len(parents[id])
	var count int64 = int64(num_child)
	
	for i := 0; i < num_child; i++ {
		count += count_child( parents[id][i] )
	}
	return count
}

/************************************************************************/
func get_child_json(f int64) {
	level += 1
	num_child := len(parents[f])

	var f1 int64

	for i := 0; i < num_child; i++ {
		f1 = parents[f][i]
		tree.WriteString(" |")

		if level != 1 {
			tree.WriteString("   ")
			for j := 1; j < (level - 1); j++ {				
				tree.WriteString("    ")
			}		
			tree.WriteString("|")
		}
		
		fmt.Fprintf(&tree, "-- %02X-%s", f1, node_data[f1].name)
		
		node_num_childs := count_child(f1)
		if node_num_childs != 0 {
			fmt.Fprintf(&tree, " (%d)", node_num_childs)
		}
		
		tree.WriteString("\n")
		
		get_child_json(f1)
	}

	level -= 1
}


func MySensors_MapVisualization(c mqtt.Client, topic string) {
	total_node := len(node_data)

	var node_seen = make( []int64, total_node)
	
	for id := range node_data {
		node_seen[ node_data[id].counter ] = id
	}
	
	// clear array
	for i := 0; i < 256; i++ {
		parents[int64(i)] = make([]int64, 0)
	}
	
	for i := 0; i < total_node; i++ {
		id := node_seen[i]
		if id != 0 {
			parent := node_data[id].parentId
			parents[parent] = append(parents[parent], id)
		}
	}
	//fmt.Println("Child:", parents)
	num_child_f1 = len(parents[0])
	if num_child_f1 > 0 {
		last_child_f1 = parents[0][num_child_f1 - 1]
	}
	// fmt.Println("Num F1:", num_child_f1)
	// fmt.Println("Last F1:", last_child_f1)
	
	tree.Reset()  // clear string buffer
	
	fmt.Fprintf(&tree, "00-N0 (%d)\n", total_node - 1)
	get_child_json(0)

	lines := strings.Split(tree.String(), "\n")
	var js strings.Builder
	
	js.WriteString("{")
	for i := 0; i < len(lines) - 2; i++ {
		fmt.Fprintf(&js, `"%s" : "",`,  lines[i])
	}
	fmt.Fprintf(&js, `"%s" : ""`,  lines[len(lines) - 2])
	js.WriteString("}")
	
	c.Publish(topic, 0, false, js.String())
}


/************************************************************************/
func Parse_NodeAddr(opt string) (string) {
	addr, err := strconv.ParseInt(opt, 16, 64);
	if err == nil && addr >= 0 && addr <= 0xFFFFFFFFFF{
		
		addr_str := opt
		if (len(opt) % 2) == 1 {
			addr_str = "0" + addr_str
		}
		
		if len(addr_str) == 2 {
			addr_str = "00" + addr_str
		}
		
		return addr_str
	}
	
	return ""
}


/************************************************************************/
func controlResponse(c mqtt.Client, topic string, msg string) {
	//string to string json
	c.Publish(topic, 0, false, `{"` + msg + `" : ""}`)
}


/************************************************************************/
var MySensors_TxPower = map[string]string {
	"min" : "0",
	"low" : "1",
	"high": "2",
	"max" : "3",
}

var MySensors_Datarate = map[string]string {
	"1M"  : "0", 
	"1m"  : "0",
	"2M"  : "1",
	"2m"  : "1",
	"250K": "2",
	"250k": "2",
}


/************************************************************************/
func help_iotgateway() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"iotgateway. " : "[ reboot | poweroff | checkupdate ]"`)
	s.WriteString(`}`)
	return s.String()
}


/************************************************************************/
func help_node() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"node.XX.COMMAND  " : "XX = node ID (hex)",`)
	s.WriteString(`".report " : "",`)
	s.WriteString(`".reportFull " : "",`)
    s.WriteString(`"" : "",`)
	s.WriteString(`".restart     " : "",`)
	s.WriteString(`".renewId     " : "clear & request new ID",`)
	s.WriteString(`".renewParent " : "clear & find parent",`)
	s.WriteString(`" " : "",`)
	s.WriteString(`".RF.txPower.[level]         " : "[ min | low | high | max ]",`)
	s.WriteString(`".RF.txRetry.[max].[timeout] " : "[ 0 - 15 ].[ 250 | 500 | 750 | ... | 4000 ]",`)
	s.WriteString(`"  " : "",`)
	s.WriteString(`".RF2.txPower.[level]         " : "[ min | low | high | max ]",`)
	s.WriteString(`".RF2.txRetry.[max].[timeout] " : "[ 0 - 15 ].[ 250 | 500 | 750 | ... | 4000 ]",`)
	s.WriteString(`".RF2.CBA.[channel].[baud].[address] " : "[ 0 - 125 ].[ 250K | 1M | 2M ].2 - 5 hex bytes",`)
	s.WriteString(`".RF2.active " : "active RF2 config",`)
	s.WriteString(`"   " : "",`)
	s.WriteString(`".led.[state] " : "[ on | off ]"`)
	s.WriteString(`}`)
	return s.String()
}


/************************************************************************/
func help_txpower() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`".min  " : "-18dBm",`)
	s.WriteString(`".low  " : "-12dBm",`)
	s.WriteString(`".high " : " -6dBm",`)
	s.WriteString(`".max  " : "  0dBm"`)
	s.WriteString(`}`)
	return s.String()
}


/************************************************************************/
func help_txretry() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`".delay " : "tx timeout (250-4000us)",`)
	s.WriteString(`".count " : "max retry (0-15)"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
const MAX_ARGS int = 7

func Parse_Cmd(total_cmd string) ([]string) {
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
func Parse_TxRetry(arg_max string, arg_timeout string) (string) {
	var S strings.Builder
	
	max, err := strconv.ParseInt(arg_max, 10, 64)
	if err == nil && max >= 0 && max <= 15 {
		fmt.Fprintf(&S, "%1X", max)
	} else {
		S.WriteString(" ")
	}

	timeout, err := strconv.ParseInt(arg_timeout, 10, 64)
	if err == nil && timeout >= 250 && timeout <= 4000 {		
		fmt.Fprintf(&S, "%1X", (timeout / 250) - 1)
	} else {
		S.WriteString(" ")
	}

	return S.String()
}


/************************************************************************/
func Parse_CBA(arg_channel string, arg_baud string, arg_addr string) (string) {
	var S strings.Builder
	
	ch, err := strconv.ParseInt(arg_channel, 10, 64);	
	if err == nil && ch >= 0 && ch <= 125 {
		fmt.Fprintf(&S, "%02X", ch)
	} else {
		S.WriteString("  ")
	}
	
	if baud, ok := MySensors_Datarate[arg_baud]; ok {
		S.WriteString(baud)
	} else {
		S.WriteString(" ")
	}
	
	S.WriteString( Parse_NodeAddr(arg_addr) )
	
	return S.String()
}


/************************************************************************/
func MySensors_reportFull(c mqtt.Client, topic string, id int64, args []string) {	
	gateway_log.Thingsboard_add_log("GWadapter, MySensors_reportFull()")
	node_data[id].reportFull = ""  // clear flag
	Send_NodeCmd(id, CMD_REPORT_FULL)  // request
	
	// wait timeout 9s
	for i := 0; i < 9; i++ {
		GwChars.Sleep_ms(1000)
		if node_data[id].reportFull != "" {
			c.Publish(topic, 0, false, node_data[id].reportFull)
			return
		}
	}
	
	controlResponse(c, topic, "Node reportFull timeout")
}
/***********************************************/
// clear flag data of node_data[id]
// send commands to node, A callback function will update new data
// this function wait end check until not empty
func MySensors_waitData(c mqtt.Client, topic string, id int64, cmd string) {
	gateway_log.Thingsboard_add_log("GWadapter, MySensors_waitData()")
	node_data[id].reportFull = ""  // clear flag
	node_data[id].report = ""
	Send_NodeCmd(id, cmd)  // send command downto node
	
	// wait timeout 9s
	for i := 0; i < 9; i++ {
		GwChars.Sleep_ms(1000)
		if node_data[id].reportFull != "" {
			c.Publish(topic, 0, false, node_data[id].reportFull)
			return
		}
		if node_data[id].report != "" {
			c.Publish(topic, 0, false, node_data[id].report)
			return
		}
	}
	
	controlResponse(c, topic, "Node wait data timeout")

	
}
/***************************************************/

func MySensors_report(c mqtt.Client, topic string, id int64, args []string) {
	gateway_log.Thingsboard_add_log("GWadapter, MySensors_report()")
	fmt.Printf("here is GWadapter, MySensors_report() id = %d\n", id)
	node_data[id].report = ""     // clear flag
	Send_NodeCmd(id, CMD_REPORT)  // request new data
	
	// wait timeout 9s
	for i := 0; i < 9; i++ {
		GwChars.Sleep_ms(1000)
		fmt.Printf("i = %d\n", i)
		if node_data[id].report != "" {
			//fmt.Printf("node_data = %s\ni=%d\n", node_data[id].report, i)
			c.Publish(topic, 0, false, node_data[id].report)
			return
		}
	}
	
	controlResponse(c, topic, "Node report timeout")
}
func MySensors_report_cmd(c mqtt.Client, topic string, id int64, cmd string) {
	gateway_log.Thingsboard_add_log("GWadapter, MySensors_report()")
	node_data[id].report = ""     // clear flag
	Send_NodeCmd(id, cmd)  // request new data
	
	// wait timeout 9s
	for i := 0; i < 9; i++ {
		GwChars.Sleep_ms(1000)
		if node_data[id].report != "" {
			fmt.Printf("node_data = %s\ni=%d\n", node_data[id].report, i)
			c.Publish(topic, 0, false, node_data[id].report)
			return
		}
	}
	controlResponse(c, topic, "Node report timeout")
}


func MySensors_restart(c mqtt.Client, topic string, id int64, args []string) {
	gateway_log.Thingsboard_add_log("GW_log: GWadapter,MySensors_restart()")
	Send_NodeCmd(id, CMD_RESTART)
	controlResponse(c, topic, "Node restart ")
}

func MySensors_renewId(c mqtt.Client, topic string, id int64, args []string) {
	gateway_log.Thingsboard_add_log("GWadapter,MySensors_renewId()")

	if id != 0 {
		Send_NodeCmd(id, CMD_RENEW_ID)
		controlResponse(c, topic, "Renew node id")
	}
}

func MySensors_renewParent(c mqtt.Client, topic string, id int64, args []string) {
	gateway_log.Thingsboard_add_log("GWadapter,MySensors_renewParent()")
	if id != 0 {
		Send_NodeCmd(id, CMD_RENEW_PARENT)
		controlResponse(c, topic, "Renew parent")
	}
}


/************************************************************************/
func RF_config(c mqtt.Client, topic string, id int64, args []string) {
	gateway_log.Thingsboard_add_log("GWadapter,RF_config()")
	switch args[3] {
		case "txpower", "txPower":  // node.X.txpower.level
			if level, ok := MySensors_TxPower[args[4]]; ok {
				Send_NodeCmd(id, CMD_RF_TXPOWER + level)
				controlResponse(c, topic, "Config new RF txpower")
			} else {
				c.Publish(topic, 0, false, help_txpower())
			}

		case "txretry", "txRetry":
			if opt := Parse_TxRetry(args[4], args[5]); opt != "  " {
				Send_NodeCmd(id, CMD_RF_TXRETRY + opt)
				controlResponse(c, topic, "Config new RF txretry")
			} else {
				c.Publish(topic, 0, false, help_txretry())
			}

		default: c.Publish(topic, 0, false, help_node())
	}
}


/************************************************************************/
func RF2_config(c mqtt.Client, topic string, id int64, args []string) {
	gateway_log.Thingsboard_add_log("GWadapter,RF2_config()")
	switch args[3] {
		case "txpower", "txPower":
			if level, ok := MySensors_TxPower[args[4]]; ok {
				Send_NodeCmd(id, CMD_RF2_TXPOWER + level)
				controlResponse(c, topic, "Config new RF2 txpower")
			} else {
				c.Publish(topic, 0, false, help_txpower())
			}

		case "txretry", "txRetry":
			if opt := Parse_TxRetry(args[4], args[5]); opt != "  " {
				Send_NodeCmd(id, CMD_RF2_TXRETRY + opt)
				controlResponse(c, topic, "Config new RF2 txretry")
			} else {
				c.Publish(topic, 0, false, help_txretry())
			}

		case "cba", "CBA":
				if opt := Parse_CBA(args[4], args[5], args[6]); opt != "   " {
					Send_NodeCmd(id, CMD_RF2_CBA + opt)
					controlResponse(c, topic, "Config new RF2 CBA")
				} else {
					c.Publish(topic, 0, false, help_node())
				}

		case "active":
				Send_NodeCmd(id, CMD_RF2_ACTIVE)
				controlResponse(c, topic, "Config CMD_RF2_ACTIVE ")
		
		default: c.Publish(topic, 0, false, help_node())
	}
}


/************************************************************************/
func Led_control(c mqtt.Client, topic string, id int64, args []string) {
	gateway_log.Thingsboard_add_log("[Led_control]")
	switch args[3] {
		case "on", "ON":
			Send_NodeCmd(id, CMD_LED_ON)
			controlResponse(c, topic, "Turn on LED")

		case "off", "OFF":
			Send_NodeCmd(id, CMD_LED_OFF)
			controlResponse(c, topic, "Turn off LED")
			
		default: c.Publish(topic, 0, false, help_node())
	}
}


/************************************************************************/
var NodeCommand = map[string] func(mqtt.Client, string, int64, []string) {
  "reportFull":  MySensors_reportFull,
  "reportfull":  MySensors_reportFull,
	
  "report":  MySensors_report,
	
	"restart": MySensors_restart,

	"renewid": MySensors_renewId,
//  "renewId": MySensors_renewId,
	
	"renewparent": MySensors_renewParent,
//	"renewParent": MySensors_renewParent,
	
	"rf": RF_config,
//	"RF": RF_config,
	
	"rf2": RF2_config,
//	"RF2": RF2_config,
	
	"led": Led_control,
//	"LED": Led_control,
}


/************************************************************************/
func help_cmd() (string) {
	var help strings.Builder
	help.WriteString(`{`)
	help.WriteString(`"?      " : "show this help",`)
	help.WriteString(`"??     " : "show list variable",`)
	help.WriteString(`"Object " : "iotgateway/mysensors/adapter/monitor/mosquitto/domoticz/node"`)
	help.WriteString(`}`)
	return help.String()
}


/************************************************************************/
func help_config() (string) {
	json, err := json.MarshalIndent(Config, "", "  ")
	if err != nil {
		return `{"":""}`
	}
	//fmt.Println(string(json))
	return string(json)
}


/************************************************************************/
func help_adapter() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"adapter. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}


/************************************************************************/
func help_monitor() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"monitor. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}


/************************************************************************/
func help_mosquitto() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"mosquitto. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}


/************************************************************************/
func help_domoticz() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"domoticz. " : "[ restart ]"`)
	s.WriteString(`}`)
	return s.String()
}


/************************************************************************/
func help_mysensors() (string) {
	var s strings.Builder
	s.WriteString(`{`)
	s.WriteString(`"mysensors. " : "[ map ]"`)
	s.WriteString(`}`)
	return s.String()
}

/************************************************************************/
func MosquittoCallBack_ThingsboardRequest(c mqtt.Client, message mqtt.Message) {
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
		gateway_log.Thingsboard_add_log("GWlog adapter: string received " + method)
		topic := strings.Replace(message.Topic(), "request", "response", 1)
		
		args  := Parse_Cmd(method)	

		switch args[0] {
			case "node":
					var id int64 = -1
					// try node.NAME
					for id_key := range node_data {
						nodeName := strings.Replace(args[1], "n", "N", 1)
						if nodeName == node_data[id_key].name {
							id = id_key  // get id
							break
						}
					}

					// try node.ID
					if id == -1 {
						if id_hex, err := strconv.ParseInt(args[1], 16, 64); err == nil {
							id = id_hex
						}
					}

					if node_data[id] == nil {  // node ko tồn tại
						//controlResponse(c, topic, "Node ID not exist")
						c.Publish(topic, 0, false, help_node())
					} else {
						fn, ok := NodeCommand[args[2]];
						if id >= 0 && id <= 255 && ok {
							go fn(c, topic, id, args)
						} else {
							
							c.Publish(topic, 0, false, help_node())
							// not match in function or id out of range
						}
					}
					
			case "mysensors":
				gateway_log.Thingsboard_add_log("GWlog adapter: case mysensors ")
				if fn, ok := MySensorsCommand[args[1]]; ok {
					fn(c, topic)
				} else {
					c.Publish(topic, 0, false, help_mysensors())
				}
			case "nodecmd":
				if val, node_id, timeout := GP.FindCommand(method); node_id != -1 {
					fmt.Printf("val = %s, node_id = %d, timeout = %d\n", val, node_id, timeout) // print to debug
					if timeout > 0 {
						//go MySensors_report(c, topic, node_id, args) // call function the same node.0.report
						go MySensors_waitData(c, topic, node_id, val)
					} else {
						Send_NodeCmd(node_id, val)
						controlResponse(c, topic, "done execute command")
					}
				} else {
					//fmt.Println("error: " + val)
					controlResponse(c, topic, "command error test" + val)
				}
			//	controlResponse(c, topic, "test node cmd")
			default: controlResponse(c, topic, "Unknow object")
		}
	}
} // end func


/************************************************************************/
