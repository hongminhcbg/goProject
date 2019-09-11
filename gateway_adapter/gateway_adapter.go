package main

import (
	"fmt"
	"encoding/json"
	"time"
	"strings"
	"io/ioutil"
	"os/exec"
	"os"
	"runtime"
	"log"
	"path/filepath"

  // custom package
  GwChars "gateway_characteristics"      // rename to "GwChars"
//GwChars "gateway_characteristics_win"  // rename to "GwChars" (Windows)

	"jsonBuffer"
	"gateway_log"
	GP "gateway_parse"
)
/************************************************************************/
var ThingsboardJson = jsonBuffer.JsonBuffer{}

/************************************************************************/
const CONFIG_FILENAME string = "config.json"
const    IOT_FILENAME string = "gateway_adapter"

var buildTime string

/************************************************************************/
type Configuration struct {
	MySensors struct {
		S0              string
		Msg_type        []string
		Max_temp        float64
		Max_RCWL        int
		Report_interval int
		Lost_tb_k1      float64
        Lost_tb_k2      float64
	}
	
	Gateway struct {
		Send_step              int
		Mqtt_reconnect_step    int
		Renew_eth0_interval    int
		Temperature_k1         int
		Temperature_k2         int
	  	MQTTJsonMaxLength      int
		MQTTJsonQueueLength    int
		MQTTJsonBuffCheck      int
		
	}
	
	Thingsboard_1 struct {
		Enable             bool
		Host               string
		MonitorReconnect   bool
		MonitorToken       string
		Send_mysensor      bool
		Send_gateway_avg   int
	}
	
	Thingsboard_2 struct {
		Enable             bool
		Host               string
		RootCA             string
		MonitorReconnect   bool
		MonitorClientCert  string
		MonitorClientKey   string
		Send_mysensor      bool
		Send_gateway_avg   int
	}
	
	Thingspeak_1 struct {
		Enable             bool
    	Reconnect          bool
		Host               string
		ChannelID_1        string
		ChannelWriteKey_1  string
		ChannelID_2        string
		ChannelWriteKey_2  string
		Send_mysensor      bool
		Send_gateway_avg   int
	}
}

/************************************************************************/
var Config = Configuration{}
var load_config_status bool = false

var msg_type = make( map[string][]string )
var num_sensor int = 0
//var num_type   int = 0

/************************************************************************/
func load_config() (bool) {
	file, err := os.Open(CONFIG_FILENAME)
	if err != nil {
		log.Println("Can't open config file:", err)
		return false
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)
	if err != nil {
		log.Println("Can't decode config JSON:", err)
		return false
	}

	for i := 0; i < len(Config.MySensors.Msg_type); i++ {
		list := strings.Split(Config.MySensors.Msg_type[i], ", ")
		msg_type[ list[0] ] = list[1:]
		num_sensor = len(msg_type[list[0]])
	}
//num_type = len(msg_type)
	return true
}


/************************************************************************/
func checkWorkingDir() (string) {
	dir, err := filepath.Abs( filepath.Dir(os.Args[0]) )
	if err != nil {
		panic(err)
	}
	
	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}
		
	log_str := fmt.Sprintf("Change working directory to: %s", dir)
	log.Println(log_str)
	gateway_log.Thingsboard_add_log(log_str)
	return dir
}

/************************************************************************/
func checkFileName() () {
	fullname       := os.Args[0]	                               
	fullname_split := strings.Split(fullname, "/")               
	filename       := fullname_split[ len(fullname_split) - 1 ]  
	//log.Println(fullname, fullname, filename)

	if strings.Contains(filename, "_new") {
		basename := strings.Replace(filename, "_new", "", 1)

		// kill old
		cmd_killall := exec.Command("killall", basename)
		cmd_killall.Run()

		// rename to old name
		cmd_rename := exec.Command("mv", filename, basename)
		cmd_rename.Run()
		
		// start with old name
		cmd_start := exec.Command("sh", "-c", "./gateway_adapter > log_adapter 2>&1 &")
		cmd_start.Start()
		
		// kill new
		os.Exit(0)  // END !!!
	}
}


/************************************************************************/
func checkUpdate() {
	cmd := exec.Command("sh", "-c", "./gateway_checkupdate adapter > log_checkupdate 2>&1 &")
	cmd.Start()
}


/************************************************************************/
func extract_buildTime() {
	list := strings.Fields(buildTime)

	date := list[1]
	month := date[0:2]
	day   := date[3:5]
	year  := date[6:10]
	
	fulltime := list[2]
	time := strings.Split(fulltime, ".")[0]
	if len(time) == 7 {
		time = "0" + time
	}
	
	buildTime = year + "-" + month + "-" + day + "_" + time
}


/************************************************************************/
func main() {
	extract_buildTime()
	gateway_log.Set_header("A")
	gateway_log.Thingsboard_add_log("Program startup ########################")
	gateway_log.Thingsboard_add_log(fmt.Sprintf("BuildTime: %s", buildTime))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	
	log.Println("BuildTime:", buildTime)
	
	checkWorkingDir()	
	checkFileName()
	checkUpdate()
		
	load_config_status = load_config()
	log.Println("Load config status:", load_config_status)
	
	load_config_name()
	GP.ReadFileConfig()

	// Node 0:
	node_data_init(0)
	node_data[0].seen = true
	node_data[0].name = "N0"
	
	update_custom_device()

	
	// Loop:
	gateway_process_mqtt()
}


/************************************************************************/
var listname map[string]interface{}

func load_config_name() () {
  file, err := ioutil.ReadFile("config_name.json")
	if err != nil {
		return
	}
	
	var totalname map[string]interface{}
	json.Unmarshal([]byte(file), &totalname)
	
	uid := GwChars.UID
	
	if totalname[uid] != nil {
		listname = totalname[uid].(map[string]interface{})
		listname[uid] = listname["Gateway"]
		delete(listname, "Gateway");
		//fmt.Printf("%v\n", listname)
	}
}


func get_config_name(uid string) (string) {
	if listname[uid] != nil {
		return listname[uid].(string)
	}
	return "??"
}


/************************************************************************/
// func thingsboard_process_buffer() {
	// ThingsboardJson.GetString()  // flush buffer
	
	// // total_msg := ThingsboardJson.GetString()
	// // if Config.Thingsboard_1.Enable == true {
		// // mqtt_thingsboard.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, total_msg)
	// // }
	// // if Config.Thingsboard_2.Enable == true {
		// // mqtt_amazon.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, total_msg)
	// // }
	
	// //mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, total_msg)
// }


/************************************************************************/
func thingsboard_process_monitor_msg() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ThingsboardJson.ObjectBegin("")
	ThingsboardJson.AddKeyValue(`"Adapter.heapsys"`, "%.2f", float64(m.HeapSys) / (1024 * 1024) )
	mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, ThingsboardJson.GetObject())
	
}


/************************************************************************/
func thingsboard_process_startup_msg() {
	ThingsboardJson.ObjectBegin("")
	ThingsboardJson.AddKeyValue("??.Report", "%s", " ")
	ThingsboardJson.AddKeyValue("??.Light" , "%s", " ")
	ThingsboardJson.AddKeyValue("??.RCWL"  , "%s", " ")
	ThingsboardJson.AddKeyValue("Adapter.heapsys", "%d", 0)
	
	mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, ThingsboardJson.GetObject())
}


/************************************************************************/
func gateway_process_mqtt() {
//fmt.Println("gateway_process_mqtt")

	// Setup:						
	// if Config.Thingsboard_1.Enable  == true {
	// 	mqtt_thingsboard_reconnect()
	// }
	// if Config.Thingsboard_2.Enable  == true {
	// 	mqtt_amazon_reconnect()
	// }

	mqtt_mosquitto_reconnect()
	
	//====================================================//
	now_ms := GwChars.Millis()
	// mqtt_reconnect_prev_ms      := now_ms
	domoticz_get_report_prev_ms := now_ms
	// renew_eth0_prev_ms          := now_ms

	thingsboard_process_startup_msg()
	
	get_next_node_report()
	
//fmt.Println("MQTTJsonBuffCheck", Config.Gateway.MQTTJsonBuffCheck)
	
	
	/// Loop:
	for {
		GwChars.Sleep_ms( Config.Gateway.MQTTJsonBuffCheck )

		thingsboard_process_monitor_msg()
		thingsboard_process_log_msg()
		
		now_ms = GwChars.Millis()


		//====================================================//
		if (now_ms - domoticz_get_report_prev_ms) >= int64(Config.MySensors.Report_interval) {
			domoticz_get_report_prev_ms = now_ms
			get_next_node_report()
		}
		
		//====================================================//
		// Nếu timeout = 2p thì Restart gateway mysensors:
		if (last_time_sensors_rx != -1 ) && ((now_ms - last_time_sensors_rx) >= int64(60000 * 2)) {
			Restart_mysensors()
			
			// Nếu chênh lệch hơn 3p thì mới gửi Log
			if ((now_ms - last_time_sensors_rx) >= int64(60000 * 3)) {
				gateway_log.Thingsboard_add_log("Restart gateway mysensors")
			}
			
			last_time_sensors_rx = now_ms
		}
		
		//====================================================//
		
	} // end loop
} // end func


/************************************************************************/
func get_current_time() string {
	return time_format( time.Now() )
}


/************************************************************************/
func time_format(t time.Time) (string) {
	_ , month, day := t.Date()  // year, month, day
	hour, min, sec := t.Clock()
	return fmt.Sprintf("%02d:%02d:%02d - %02d/%02d", hour, min, sec, day, month)
}


func time_format_2(t time.Time) (string) {
	// _ , month, day := t.Date()  // year, month, day
	// hour, min, sec := t.Clock()
	// return fmt.Sprintf("%02d.%02d.%02d - %02d/%02d", hour, min, sec, day, month)
	
	year , month, day := t.Date()
	hour, min, sec := t.Clock()
	return fmt.Sprintf("%04d.%02d.%02d %02d.%02d.%02d", year, month, day, hour, min, sec)
}


/************************************************************************/
func time_format_log(t time.Time) (string) {
	_ , month, day := t.Date()  // year, month, day
	return fmt.Sprintf("%02d-%02d %s ", month, day, t.Format("15:04:05.000"))
}


/************************************************************************/
//var thingsboard_log = make(chan string, 50)


/************************************************************************/
func thingsboard_process_log_msg() {
	log_len := len(gateway_log.Thingsboard_log)
	if log_len == 0 {
		return
	}

	log_str := make([]string, log_len)
	for i := 0; i < log_len; i++ {
		log_str[i] = <-gateway_log.Thingsboard_log
	}

	s := ThingsboardJson.ObjectBegin("")
	s.WriteString(`"Monitor.log":"`)
	for i := log_len - 1; i >= 0; i-- {
		s.WriteString(log_str[i])
		s.WriteString(`<br>`)
	}
	s.WriteString(`",`)
	
	mqtt_mosquitto.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, ThingsboardJson.GetObject())
}


/************************************************************************/
