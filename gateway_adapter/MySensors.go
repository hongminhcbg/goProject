
package main

/************************************************************************/
import (
	"fmt"
	
	"log"
	"strconv"
	"time"
	"strings"
	"github.com/eclipse/paho.mqtt.golang"
	"encoding/json"
	"bytes"
	"net/http"
	"gateway_log"
  GwChars "gateway_characteristics"      // rename to "GwChars"
//GwChars "gateway_characteristics_win"  // rename to "GwChars" (Windows)
)


/************************************************************************/
type NODE_DATA struct {
	counter      int64     // thứ tự xuất hiện của node
	
	uptime       int64
	uid          string
	rstType     string
	pageErase    int64
	keyUsed      int64
	
	name         string
	seen         bool      // đánh dấu đã kết nối với Node
	
	temp_tb_1    float64
	temp_tb_2    float64
	lost_tb      float64
	customIdx    string
	parentId    int64
	
	reportFull   string
	report       string
	
	txPower      string
	baud         string
	ard          int64
	arc          int64
}

var node_data = make( map[int64]*NODE_DATA )
var node_report int64 = 0


/************************************************************************/
func node_data_init(id int64) {
	node_data[id] = &NODE_DATA{}  // new struct
	
	node_data_reset(id)
	
	node_data[id].counter = int64(len(node_data)) - 1
}


func node_data_reset(id int64) {
	node_data[id].name = "Unknow"
	node_data[id].seen = false
	node_data[id].uid  = ""
	node_data[id].temp_tb_1  = 0.0
	node_data[id].temp_tb_2  = 0.0
	node_data[id].lost_tb    = 0.0
	node_data[id].customIdx = ""
	node_data[id].parentId = -1
	node_data[id].reportFull = ""
	node_data[id].report = ""
	node_data[id].txPower = ""
	node_data[id].baud = ""
	node_data[id].ard = 0
	node_data[id].arc = 0
//node_data[id].counter = 0
}


/************************************************************************/
const GATEWAY_REPORT_LEN      int = 10
const GATEWAY_REPORTFULL_LEN  int = 18

const NODE_REPORT_LEN         int = 10
const NODE_REPORTFULL_LEN     int = 20


/************************************************************************/
func check_device(id int64, uid_hex string) {
	if node_data[id].uid != uid_hex {
		
		if exist_id := get_device_id(uid_hex); exist_id != -1 {
			delete(node_data, exist_id)
		}
		
		node_data[id].uid	= uid_hex
		node_data[id].name = get_config_name(uid_hex)
		gateway_log.Thingsboard_add_log(fmt.Sprintf("New id: %d %s %s\n", id, node_data[id].uid, node_data[id].name))
	}	
}


/************************************************************************/
func gateway_process_reportfull(string_msg string, time_report time.Time) {
	uptime_hex      := string_msg[0:4]
	rf24_config_hex := string_msg[4:6]
//rf24_aw_hex     := string_msg[6:8]
	rf24_retr_hex   := string_msg[8:10]
	rf24_ch_hex     := string_msg[10:12]
	rf24_setup_hex  := string_msg[12:14]
	rf24_addr_hex   := string_msg[14:18]
	
	//===============================================//
	// Check uptime minute
	uptime, err := strconv.ParseInt(uptime_hex, 16, 32) 
	if err != nil {
		return
	}
	uptime_str := fmt.Sprintf("%dd%02dh%02dm", uptime / (24 * 60), (uptime / 60) % 24, uptime % 60)

	//===============================================//
	// Check CRC
	rf24_config, err := strconv.ParseInt(rf24_config_hex, 16, 32) 
	if err != nil {
		return
	}
	rf24_crc_str := "NOCRC"
	switch (rf24_config  >> 2) & 0x03 {
		case 0x02: rf24_crc_str = "CRC8"
		case 0x03: rf24_crc_str = "CRC16"
	}
	
	//===============================================//
	// Check RETR
	rf24_retr, err := strconv.ParseInt(rf24_retr_hex, 16, 32) 
	if err != nil {
		return
	}
	rf24_ard := (rf24_retr >> 4) & 0x0F
	rf24_arc := (rf24_retr >> 0) & 0x0F
	
	//===============================================//
	// Check RF_CH
	rf24_ch, err := strconv.ParseInt(rf24_ch_hex, 16, 32) 
	if err != nil {
		return
	}

	//===============================================//
	// Check RF_SETUP
	rf24_setup, err := strconv.ParseInt(rf24_setup_hex, 16, 32) 
	if err != nil {
		return
	}
	
	rf24_pwr_str   := "PA_MIN"
	rf24_pwr_str_2 := "min"
	switch (rf24_setup >> 1) & 0x03 {
		case 0x03: 
			rf24_pwr_str = "PA_MAX"
			rf24_pwr_str_2 = "max"
			
		case 0x02:
			rf24_pwr_str = "PA_HIGH"  
			rf24_pwr_str_2 = "high"
		
		case 0x01:
			rf24_pwr_str = "PA_LOW"   
			rf24_pwr_str_2 = "low"
	}
	
	rf24_datarate_str := "1Mbps"  // case 0x00
	switch rf24_setup & 0x28 {
		case 0x08: rf24_datarate_str = "2Mbps"
		case 0x20: rf24_datarate_str = "250Kbps"
	}
	
	//===============================================//
	s := ThingsboardJson.ObjectBegin("")
	fmt.Fprintf(s, `"N0.%s":"`            , Config.MySensors.S0)
	fmt.Fprintf(s, `%s<br>`               , time_format(time_report))
	fmt.Fprintf(s, `%s<br>`               , uptime_str)
	fmt.Fprintf(s, `%s - %s<br>`          , rf24_crc_str, rf24_pwr_str)
	fmt.Fprintf(s, `%s - CH_%d - 0x%s<br>`, rf24_datarate_str, rf24_ch, rf24_addr_hex)
	fmt.Fprintf(s, `[%dus, %d]",`         , (rf24_ard + 1) * 250, rf24_arc)
	mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, ThingsboardJson.GetObject())
	
	//===============================================//
	var S strings.Builder
	S.WriteString("{")
	fmt.Fprintf(&S, `"Id Name " : "%02X - %s",`    				, 0, node_data[0].name)
  fmt.Fprintf(&S, `"Parent  " : "",`         				    )
	fmt.Fprintf(&S, `"Reptime " : "%s",`           				, time_format_2(time_report))
	fmt.Fprintf(&S, `"Uptime  " : "%s - %s",`      				, uptime_str, "PwrOn")
	fmt.Fprintf(&S, `"UID     " : "%s",`                  , "Arduino")
  fmt.Fprintf(&S, `"FStore  " : "Erase %d - Used %d",`  , 0, 0)
	fmt.Fprintf(&S, `".txPower.%s":"",`                   , rf24_pwr_str_2)
	fmt.Fprintf(&S, `".txRetry.%d.%d":"",`                , (rf24_ard + 1) * 250, rf24_arc)
	fmt.Fprintf(&S, `".CBA.%d.%s.%s":""`                  , rf24_ch, rf24_datarate_str, rf24_addr_hex)
	S.WriteString("}")
	//fmt.Println(S.String())
	node_data[0].reportFull = S.String()   // set flag

}

/**********************************************************************/
func gateway_process_report(string_msg string, time_report time.Time) {
  var id int64 = 0  // gateway_id
  
//log.Println("Report:", id, string_msg)
	parent_hex := string_msg[0:2]
	uptime_hex := string_msg[2:6]
	uid_hex    := string_msg[6:10]
	
	//===============================================//
	// check ParentID
	parent_id, err := strconv.ParseInt(parent_hex, 16, 64) 
	if err != nil {
		return
	}
	node_data[id].parentId = parent_id

	//===============================================//
	// check uptime minute
	uptime, err := strconv.ParseInt(uptime_hex, 16, 64) 
	if err != nil {
		return
	}
	node_data[id].uptime = uptime
	uptime_str := fmt.Sprintf("%dd %02d.%02d", uptime / (24 * 60), (uptime / 60) % 24, uptime % 60)

	//===============================================//
	// check UID:
	check_device(id, uid_hex)
	
	//===============================================//
	// s := ThingsboardJson.ObjectBegin("")
	// fmt.Fprintf(s, `"%s.%s":"`              , node_data[id].name, Config.MySensors.S0)
	// fmt.Fprintf(s, `%s<br>`                 , time_format(time_report))
	// fmt.Fprintf(s, `%s - %s<br>`            , uptime_str, node_data[id].rstType)
	// fmt.Fprintf(s, `@%02X/%02X<br>`         , parent_id, id)
	// fmt.Fprintf(s, `%s - %s - [%d %d]<br>`  , node_data[id].name, node_data[id].uid, node_data[id].pageErase, node_data[id].keyUsed)
	// fmt.Fprintf(s, `%s<br>`                 , "PA_" + strings.ToUpper(node_data[id].txPower))
	// fmt.Fprintf(s, `[%dus, %d]",`           , (node_data[id].ard + 1) * 250, node_data[id].arc)
	// mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, ThingsboardJson.GetObject())
  
	//===============================================//
	var S strings.Builder
	S.WriteString("{")
	fmt.Fprintf(&S, `"UpTime " : "%s",`   , uptime_str)
	fmt.Fprintf(&S, `"Parent " : "%02X"`  , parent_id)
	S.WriteString("}")
	// fmt.Println(S.String())
	node_data[id].report = S.String()   // set flag
}

/************************************************************************/
func node_process_reportFull(id int64, string_msg string, time_report time.Time) {
	//0 06 0 B F9 64 2 000000E1E0
//log.Println("ReportFull:", id, string_msg)
	numpage_hex    := string_msg[0:1]
	keyused_hex    := string_msg[1:3]
	rst_type       := string_msg[3:4]
	rf24_setup_hex := string_msg[4:5]
	rf24_retr_hex  := string_msg[5:7]
	rf24_ch_hex    := string_msg[7:9]
  rf24_aw_hex    := string_msg[9:10]
	rf24_addr_hex  := string_msg[10:20]

	// fmt.Println(numpage_hex, keyused_hex, 
	// rst_type, 
	// rf24_setup_hex, rf24_retr_hex, rf24_ch_hex, rf24_aw_hex, rf24_addr_hex)
	
	//===============================================//
	// FStore
	pageErase, err := strconv.ParseInt(numpage_hex, 16, 64)
	if err != nil {
		return
	}
	keyUsed, err := strconv.ParseInt(keyused_hex, 16, 64)
	if err != nil {
		return
	}
	node_data[id].pageErase = pageErase
	node_data[id].keyUsed = keyUsed
	
	//===============================================//
	// check Reset type
	switch rst_type {
		case "0": rst_type = "PwrOn"
		case "1": rst_type = "LowPwr"
		case "2": rst_type = "WDog"
		case "3": rst_type = "WDog"
		case "4": rst_type = "Soft"
		case "5": rst_type = "Hard"
		default : rst_type = "Unknow"
	}
	node_data[id].rstType = rst_type
	
	//===============================================//
	// check RF_SETUP
	rf24_setup, err := strconv.ParseInt(rf24_setup_hex, 16, 64) 
	if err != nil {
		return
	}
	rf24_pwr_str := "min"
	switch (rf24_setup >> 0) & 0x03 {
		case 0x03: rf24_pwr_str = "max"
		case 0x02: rf24_pwr_str = "high"
		case 0x01:	rf24_pwr_str = "low"
	}
	node_data[id].txPower = rf24_pwr_str
	
	rf24_baud := "1M"
	switch (rf24_setup >> 2) & 0x03 {
		case 0x01: rf24_baud = "2M"
		case 0x02: rf24_baud = "250K"
	}

	//===============================================//
	// check RETR
	rf24_retr, err := strconv.ParseInt(rf24_retr_hex, 16, 64) 
	if err != nil {
		return
	}
	rf24_ard := (rf24_retr >> 4) & 0x0F
	rf24_arc := (rf24_retr >> 0) & 0x0F
	node_data[id].ard = rf24_ard
	node_data[id].arc = rf24_arc

	//===============================================//
	// Check CHANNEL
	rf24_ch, err := strconv.ParseInt(rf24_ch_hex, 16, 64) 
	if err != nil {
		return
	}

	//===============================================//
	// Check AW
	rf24_aw, err := strconv.ParseInt(rf24_aw_hex, 16, 64) 
	if err != nil {
		return
	}
	rf24_addr_base := rf24_addr_hex[(10 - rf24_aw * 2):10]

	//===============================================//
	// Uptime minute
	uptime := node_data[id].uptime
	uptime_str := fmt.Sprintf("%dd %02d.%02d", uptime / (24 * 60), (uptime / 60) % 24, uptime % 60)
	bootTime := time.Unix(time_report.Unix() - uptime * 60, 0) 

	//===============================================//
	s := ThingsboardJson.ObjectBegin("")
	fmt.Fprintf(s, `"%s.%s":"`              , node_data[id].name, Config.MySensors.S0)
	fmt.Fprintf(s, `%s<br>`                 , time_format(time_report))
	fmt.Fprintf(s, `%s - %s<br>`            , uptime_str, rst_type)
	fmt.Fprintf(s, `@%02X/%02X<br>`         , node_data[id].parentId, id)
	fmt.Fprintf(s, `%s - %s - [%d %d]<br>`  , node_data[id].name, node_data[id].uid, pageErase, keyUsed)
	fmt.Fprintf(s, `%s<br>`                 , "PA_" + strings.ToUpper(rf24_pwr_str))
	fmt.Fprintf(s, `[%dus, %d]",`           , (rf24_ard + 1) * 250, rf24_arc)
	mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, ThingsboardJson.GetObject())
	
	//===============================================//
	var S strings.Builder
	S.WriteString("{")
	fmt.Fprintf(&S, `"RepTime  " : "%s - %s",`        , time_format_2(time_report), uptime_str)
	fmt.Fprintf(&S, `"BootTime " : "%s - %s",`      	, time_format_2(bootTime), rst_type)
	fmt.Fprintf(&S, `"FStore   " : "%d.%03d",`        , pageErase, keyUsed)	
	fmt.Fprintf(&S, `".txPower.%s":"",`               , rf24_pwr_str)
	fmt.Fprintf(&S, `".txRetry.%d.%d":"",`            , rf24_arc, (rf24_ard + 1) * 250)
	fmt.Fprintf(&S, `".CBA.%d.%s.%s":"",`             , rf24_ch, rf24_baud, rf24_addr_base)
	fmt.Fprintf(&S, `"Parent   " : "%02X",`         	, node_data[id].parentId)
	fmt.Fprintf(&S, `"Id Name  " : "%02X - %s - %s"`  , id, node_data[id].uid, node_data[id].name)
	S.WriteString("}")
	
//fmt.Println( S.String() )
	node_data[id].reportFull = S.String()   // set flag
}


/************************************************************************/
func node_process_report(id int64, string_msg string, time_report time.Time) {
	//log.Println("Report:", id, string_msg)
	parent_hex := string_msg[0:2]
	uptime_hex := string_msg[2:6]
	uid_hex    := string_msg[6:10]
	
	//===============================================//
	// check ParentID
	parent_id, err := strconv.ParseInt(parent_hex, 16, 64) 
	if err != nil {
		return
	}
	node_data[id].parentId = parent_id

	//===============================================//
	// check uptime minute
	uptime, err := strconv.ParseInt(uptime_hex, 16, 64) 
	if err != nil {
		return
	}
	node_data[id].uptime = uptime
	uptime_str := fmt.Sprintf("%dd %02d.%02d", uptime / (24 * 60), (uptime / 60) % 24, uptime % 60)

	//===============================================//
	// check UID:
	check_device(id, uid_hex)
	
	//===============================================//
	s := ThingsboardJson.ObjectBegin("")
	fmt.Fprintf(s, `"%s.%s":"`              , node_data[id].name, Config.MySensors.S0)
	fmt.Fprintf(s, `%s<br>`                 , time_format(time_report))
	fmt.Fprintf(s, `%s - %s<br>`            , uptime_str, node_data[id].rstType)
	fmt.Fprintf(s, `@%02X/%02X<br>`         , parent_id, id)
	fmt.Fprintf(s, `%s - %s - [%d %d]<br>`  , node_data[id].name, node_data[id].uid, node_data[id].pageErase, node_data[id].keyUsed)
	fmt.Fprintf(s, `%s<br>`                 , "PA_" + strings.ToUpper(node_data[id].txPower))
	fmt.Fprintf(s, `[%dus, %d]",`           , (node_data[id].ard + 1) * 250, node_data[id].arc)

	mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, ThingsboardJson.GetObject())
	
	//===============================================//
	var S strings.Builder
	S.WriteString("{")
	fmt.Fprintf(&S, `"UpTime " : "%s",`   , uptime_str)
	fmt.Fprintf(&S, `"Parent " : "%02X"`  , parent_id)
	S.WriteString("}")
	// fmt.Println(S.String())
	node_data[id].report = S.String()   // set flag
	fmt.Println("node report = %s\n", S.String())
}


/************************************************************************/
var NodeProcessReport = map[int] func(int64, string, time.Time) {
	NODE_REPORTFULL_LEN : node_process_reportFull,
	NODE_REPORT_LEN      : node_process_report,
}


/************************************************************************/
func node_process_sensor(id int64, unit string, string_msg string) {
	s := ThingsboardJson.ObjectBegin("")
	
	for i := 0; i < num_sensor; i++ {
		p := i * 4
		number, err := strconv.ParseInt(string_msg[p:(p + 4)], 16, 64)
		if err != nil {
			return
		}
		value := ""
		
		if msg_type[unit][i] == "Temp" {
			temp := float64(number) / 1000.0        // millidegree C
			if temp > Config.MySensors.Max_temp {	  // check temperature max
				temp = Config.MySensors.Max_temp
			}
			value = fmt.Sprintf("%.2f", temp)
		} else {
			if msg_type[unit][i] == "RCWL" {
				if node_data[id].seen == false {  // clear first value
					number = 0  
				} else {
					number = int64( int16( number ) )	              // check negative
					if number > int64(Config.MySensors.Max_RCWL) {  // check RCWL_max
						number = int64(Config.MySensors.Max_RCWL)
					}
				}
			}
			value = fmt.Sprintf("%d", number)
		}
		
		s.WriteString(`"`)
		s.WriteString(node_data[id].name)
		s.WriteString(`.`)
		s.WriteString(msg_type[unit][i])
		fmt.Fprintf(s, `":"%s",`, value)
	}
	
	mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, ThingsboardJson.GetObject())
}


/************************************************************************/
var last_time_sensors_rx int64 = -1;

func MosquittoCallBack_DomoticzOut(c mqtt.Client, message mqtt.Message) {
	go GwChars.BlinkLed_Green()
	if load_config_status == false {
		return
	}
	// fmt.Printf("TOPIC: %s\n", message.Topic())
	// fmt.Printf("MSG:\n %s\n\n", message.Payload())
	dec := json.NewDecoder(bytes.NewReader(message.Payload()))
	var domoticz_msg map[string]interface{}
	if err := dec.Decode(&domoticz_msg); err != nil {
		return
	}
	// fmt.Print("\n")	
	// for key := range domoticz_msg {
		// fmt.Printf("%15s   ", key)
		// fmt.Println(domoticz_msg[key])
	// }
	// return

	//---------------------------------------------------------------------------//
	if domoticz_msg["stype"].(string) != "Text" {
		return  // chỉ xử lý các bản tin kiểu Text
	}

	domoticz_rxtime := time.Now()
	domoticz_svalue := domoticz_msg["svalue1"].(string)
	domoticz_idx    := fmt.Sprintf("%.0f", domoticz_msg["idx"].(float64));
	domoticz_id     := domoticz_msg["id"].(string)
	domoticz_unit   := domoticz_id[6:8]
	
	node_id, err := strconv.ParseInt(domoticz_id[0:6], 16, 64)
	if err != nil {
		return
	}
	
	//---------------------------------------------------------------------------//
	if node_id == 0 {  // Gateway
		if domoticz_unit == "0A" {
			node_data[0].customIdx = domoticz_idx
      
      switch len(domoticz_svalue) {
        case GATEWAY_REPORT_LEN:
          gateway_process_report(domoticz_svalue, domoticz_rxtime)
          
        case GATEWAY_REPORTFULL_LEN:
          gateway_process_reportfull(domoticz_svalue, domoticz_rxtime)
      }
	}

	} else {  // Node
	
		if node_data[node_id] == nil {  // khởi tạo map[key]
			node_data_init(node_id)
		}
	
		switch domoticz_unit {
			case "0A":  // report
				node_data[node_id].customIdx = domoticz_idx
				if fn, ok := NodeProcessReport[len(domoticz_svalue)]; ok {
					fn(node_id, domoticz_svalue, domoticz_rxtime)
				}

			case "0B", "0C", "0D":  // sensors
				if len(domoticz_svalue) != (num_sensor * 4) {  // kiểm tra kích thước bản tin
					return
				}

				// node chưa có trong hệ thống: request UID
				if node_data[node_id].uid == "" {
					node_data[node_id].customIdx = domoticz_idx
					log.Printf("Request UID: %d\n", node_id)
					Send_NodeCmd(node_id, CMD_REPORT)
				} else {
					node_process_sensor(node_id, domoticz_unit, domoticz_svalue)
					last_time_sensors_rx = GwChars.Millis()
					
					if node_data[node_id].seen == false {
						node_data[node_id].seen = true
						Send_NodeCmd(node_id, CMD_REPORT_FULL)
					}
				}
		}
	}
} // end func


/************************************************************************/
func Send_NodeCmd(id int64, cmd_opt string) {
  idx := node_data[id].customIdx
	
	if idx != "" {
		msg := `{"idx":` + idx + `,"svalue":"` + cmd_opt + `"}`
		mqtt_mosquitto.Publish(DOMITICZ_TOPIC_IN, 0, false, msg)
	}
}


/************************************************************************/
func Restart_mysensors() {
	Send_NodeCmd(0, CMD_RESTART)
}


/************************************************************************/
func get_next_node_report() {
	total_node := int64( len(node_data) )

	node_report++
	if node_report >= total_node {
		node_report = 0
	}
	
	for id := range node_data {
		if node_data[id].counter == node_report {
			//log.Println("Get report:", id)
			if id != 0 {
				Send_NodeCmd(id, CMD_REPORT)
			} else {
				Send_NodeCmd(id, CMD_REPORT_FULL)
			}
			return
		}
	}
}


/************************************************************************/
func get_device_id(uid string) (int64) {
	for id := range node_data {
		if node_data[id].uid == uid {
			return id
		}
	}
	return -1
}


/************************************************************************/
func update_custom_device() {
	url := "http://localhost:8088/json.htm?type=devices&filter=utility"
	
	// Get the response from URL
	url_resp, url_err := http.Get(url)
	if url_err != nil {
		fmt.Println(url_err)
		return
	}
	defer url_resp.Body.Close()
	
	// Convert to string
	buf := new(bytes.Buffer)
	buf.ReadFrom(url_resp.Body)
	json_list := buf.String()
	//fmt.Printf("%s\n", json_list)

	var json_list_decode map[string]interface{}
	json_list_decode_err := json.Unmarshal([]byte(json_list), &json_list_decode)
	if json_list_decode_err != nil {
		fmt.Println(json_list_decode_err)
		return
	}
	
	//fmt.Println(json_list_decode["result"])
	if json_list_decode["result"] == nil {
		return
	}
	result := json_list_decode["result"].([]interface{})
	
	//fmt.Println(result)
	//fmt.Println( len(result) )
	//list_node_light = list_node_light[:len(result)]

	for i := 0; i < len(result); i++ {	
		current_device := result[i].(map[string]interface{})

		ID,  ID_ok  := current_device["ID"].(string)
		idx, idx_ok := current_device["idx"].(string)
		
		if (len(ID) == 8) && ID_ok && idx_ok {
			// fmt.Println(ID, idx)  
			nodeUnit, err := strconv.ParseInt(ID[6:8], 16, 64);
			if err == nil && nodeUnit == 10 {
				if nodeId, err:= strconv.ParseInt(ID[0:6], 16, 64); err == nil {
					//fmt.Println(nodeId, nodeUnit, idx)
					if nodeId == 0 {
						//fmt.Println("MySensors gateway custom =", idx)
						node_data[0].customIdx = idx
						return;
					}
				}
			}
		}
	}

}

/************************************************************************/
