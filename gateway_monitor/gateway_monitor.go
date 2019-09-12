package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	// "net/http"
	GwChars "gateway_characteristics" // rename to "GwChars"
	"gateway_log"
	GP "gateway_parse"
	"jsonBuffer"
	// "github.com/eclipse/paho.mqtt.golang"
)

/************************************************************************/
var ThingsboardJson = jsonBuffer.JsonBuffer{}

/************************************************************************/
const CONFIG_FILENAME string = "config.json"
const FILENAME string = "gateway_monitor"

var buildTime string

//var lastMsg string
/************************************************************************/
type Configuration struct {
	MySensors struct {
		S0              string
		Msg_type        []string
		Max_temp        float64
		Max_RCWL        int
		Report_interval int
	}

	Gateway struct {
		Send_step           int
		Temperature_k1      int
		Temperature_k2      int
		Debug_domoticz      int
		NetworkTimeout      int
		MQTTJsonMaxLength   int
		MQTTJsonQueueLength int
		MQTTJsonBuffCheck   int
	}

	Thingsboard_1 struct {
		Enable           bool
		Host             string
		MonitorReconnect bool
		MonitorToken     string
		Send_gateway_avg int
	}

	Thingsboard_2 struct {
		Enable           bool
		Host             string
		MonitorReconnect bool
		MonitorToken     string
		Send_gateway_avg int
	}

	Thingspeak_1 struct {
		Enable            bool
		Reconnect         bool
		Host              string
		ChannelID_1       string
		ChannelWriteKey_1 string
		ChannelID_2       string
		ChannelWriteKey_2 string
		Send_gateway_avg  int
	}
}

var Config = Configuration{}
var ConfigCmd = ""
var load_config_status bool = false
var useMqttTB1 bool = true //true is MQTT and flase is HTTP
var useMqttTB2 bool = true //true is MQTT and flase is HTTP
/************************************************************************/
func load_config() {
	file, err := os.Open(CONFIG_FILENAME)
	if err != nil {
		log.Println("Can't open config file:", err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Config)
	if err != nil {
		log.Println("Can't decode config.json:", err)
		return
	}
	load_config_status = true
}

/************************************************************************/
func checkWorkingDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
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
func checkFileName() {
	fullname := os.Args[0]
	fullname_split := strings.Split(fullname, "/")
	filename := fullname_split[len(fullname_split)-1]
	//log.Println(fullname, fullname, filename)

	if strings.Contains(filename, "_new") {
		basename := strings.Replace(filename, "_new", "", 1)

		// kill old program
		cmd_killall := exec.Command("killall", basename)
		cmd_killall.Run()

		// rename to old name
		cmd_rename := exec.Command("mv", filename, basename)
		cmd_rename.Run()

		// start with old name
		cmd_start := exec.Command("sh", "-c", "./gateway_monitor > log_monitor 2>&1 &")
		cmd_start.Start()

		// kill new
		os.Exit(0) // END !!!
	}
}

/************************************************************************/
func checkUpdate() {
	cmd_checkupdate := exec.Command("sh", "-c", "./gateway_checkupdate monitor > log_checkupdate 2>&1 &")
	cmd_checkupdate.Start()
}

/************************************************************************/
func extract_buildTime() {
	list := strings.Fields(buildTime)

	date := list[1]
	month := date[0:2]
	day := date[3:5]
	year := date[6:10]

	fulltime := list[2]
	time := strings.Split(fulltime, ".")[0]
	if len(time) == 7 {
		time = "0" + time
	}

	buildTime = year + "-" + month + "-" + day + "_" + time
}

/************************************************************************/
func main() {
	// for i := 0; i < 3; i++ {
	// GwChars.SetLed_Red("1")
	// GwChars.SetLed_Green("0")
	// GwChars.Sleep_ms(500)

	// GwChars.SetLed_Red("0")
	// GwChars.SetLed_Green("1")
	// GwChars.Sleep_ms(500)
	// }
	gateway_log.Set_header("M")
	extract_buildTime()
	gateway_log.Thingsboard_add_log("Program startup ########################")
	gateway_log.Thingsboard_add_log(fmt.Sprintf("BuildTime: %s", buildTime))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	checkWorkingDir()

	checkFileName()
	checkUpdate()

	load_config()
	load_config_name()
	GP.ReadFileConfig()
	ConfigCmd = GP.StringReadConfigCmd
	useMqttTB1 = strings.Contains(Config.Thingsboard_1.Host, "tcp")
	useMqttTB2 = strings.Contains(Config.Thingsboard_2.Host, "tcp")

	if true == true {
		gateway_log.Thingsboard_add_log("MQTT " + Config.Thingsboard_1.Host)
	} else {
		gateway_log.Thingsboard_add_log("HTTP " + Config.Thingsboard_1.Host)
	}

	// Loop:
	main_monitor()
}

/************************************************************************/
func thingsboard_process_debug_msg() {
	s := ThingsboardJson.ObjectBegin("[")
	s.WriteString(`Monitor.debug:"`)
	s.WriteString(get_current_time())
	fmt.Fprintf(s, `<br>%d pkt / %d s<br>`, domoticz_rx_count, Config.Gateway.Debug_domoticz)
	s.WriteString(domoticz_rx_msg)
	s.WriteString(`",`)

	ThingsboardJson.ObjectEndCheckLength()

	// clear
	domoticz_rx_count = 0
	domoticz_rx_msg = ""
}

/************************************************************************/
func thingsboard_process_monitor_msg() {
	s := ThingsboardJson.ObjectBegin("[")
	GwChars.AddCPU(s, &thingsboard_cpu)
	GwChars.AddMemUse(s)
	GwChars.AddTemperature(s)
	GwChars.AddIpAddress(s)
	GwChars.AddUpTime(s)
	GwChars.AddOverlayroot(s)
	AddHeapmonitor(s)
	AddReporttime(s)
	ThingsboardJson.AddKeyValue(`"Monitor.grow"`, "%.2f", float64(ThingsboardJson.S.Len()+1)/1024.0)
	ThingsboardJson.ObjectEndCheckLength()
}

/************************************************************************/
func ThingsboardSendMsg(msg string) {
	if Config.Thingsboard_1.Enable == true {
		if useMqttTB1 == true { //switch protocol to send message
			//lastMsg = msg
			mqtt_thingsboard.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, msg)
		} else {
			//fmt.Println("[lhm log] ThingsboardSendMsg http")
			ThingsboardPostHTTP(TB1HttpClient, msg)
		}
	}
	if Config.Thingsboard_2.Enable == true {
		if useMqttTB2 == true {
			mqtt_amazon.Publish(THINGSBOARD_TOPIC_TELEMETRY, 0, false, msg)
		} else {
			ThingsboardPostHTTP(TB2HttpClient, msg)
		}

	}
}

/// Setup
/************************************************************************/
func thingsboard_process_buffer() {
	//fmt.Println("[lhm log] begin thingsboard_process_buffer")
	if thingsboard_1_connected == false && thingsboard_2_connected == false {
		return // mất hết kết nối
	}

	// kiểm tra Queue: nếu có dữ liệu thì gửi bớt đi 1
	//fmt.Printf("[lhm log] begin check queue and len queue = %d\n", len(ThingsboardJson.Queue))
	if len(ThingsboardJson.Queue) > 0 {
		QueueMsg := <-ThingsboardJson.Queue
		ThingsboardSendMsg(QueueMsg)
		gateway_log.Thingsboard_add_log(fmt.Sprintf("Queue: [%d / %d] [%d + %d]\n",
			len(ThingsboardJson.Queue),
			Config.Gateway.MQTTJsonQueueLength,
			Config.Gateway.MQTTJsonMaxLength,
			len(QueueMsg)-Config.Gateway.MQTTJsonMaxLength))
	}

	ThingsboardSendMsg(ThingsboardJson.GetString())
}

/************************************************************************/
type Gateway_data struct {
	average    int
	count_step int
}

var thingsboard_data = Gateway_data{}
var thingsboard_cpu = GwChars.Cpu_data{}

/************************************************************************/
func AddReporttime(w *strings.Builder) {
	t := time.Now()
	_, month, day := t.Date()
	hour, min, sec := t.Clock()

	w.WriteString(`"Monitor.reptime":`)
	fmt.Fprintf(w, `"%02d:%02d:%02d - %02d/%02d",`, hour, min, sec, day, month)
}

func AddHeapmonitor(w *strings.Builder) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	w.WriteString(`"Monitor.heapsys":`)
	fmt.Fprintf(w, `%.2f,`, float64(m.HeapSys)/(1024*1024))
	//w.WriteString(`Monitor.heapalloc:`)
	//fmt.Fprintf(w, `%.2f,`, float64(m.HeapAlloc) / (1024 * 1024))
}

/************************************************************************/
var listname map[string]interface{}

func load_config_name() {
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
		delete(listname, "Gateway")
	}
}

/************************************************************************/
func get_config_name(uid string) string {
	if listname[uid] != nil {
		return listname[uid].(string)
	}
	return "Unknow"
}

/************************************************************************/
func thingsboard_process_startup_msg() {
	ThingsboardJson.Queue = make(chan string, Config.Gateway.MQTTJsonQueueLength)

	ThingsboardJson.MaxLength = Config.Gateway.MQTTJsonMaxLength
	ThingsboardJson.Size = Config.Gateway.MQTTJsonMaxLength
	ThingsboardJson.S.Grow(Config.Gateway.MQTTJsonMaxLength)

	ThingsboardJson.ObjectBegin("[")
	ThingsboardJson.AddKeyValue(`"Monitor.modtime"`, "%s", buildTime)
	ThingsboardJson.AddKeyValue(`"Monitor.name"`, "%s", get_config_name(GwChars.UID))
	ThingsboardJson.AddKeyValue(`"Monitor.uid"`, "%s", GwChars.UID)
	ThingsboardJson.AddKeyValue(`"Monitor.heapsys"`, "%d", 0)
	ThingsboardJson.AddKeyValue(`"Monitor.debug"`, "%s", get_current_time()+`<br>Debug disable`)
	ThingsboardJson.ObjectEndCheckLength()
}

/************************************************************************/
func setup_mqtt() {
	if Config.Thingsboard_1.Enable == true && useMqttTB1 == true {
		mqtt_thingsboard_reconnect()
	}
	if Config.Thingsboard_2.Enable == true && useMqttTB2 == true {
		mqtt_amazon_reconnect()
	}
}

/************************************************************************/
func setupHTTP() {
	if Config.Thingsboard_1.Enable == true && useMqttTB1 == false {
		httpTB1Setup()
	}
	if Config.Thingsboard_2.Enable == true && useMqttTB2 == false {
		httpTB2Setup()
	}
}

/************************************************************************/

func setup_msg() {
	thingsboard_process_startup_msg()

	GwChars.Temperature_Init(Config.Gateway.Temperature_k1, Config.Gateway.Temperature_k2)

	thingsboard_data.average = Config.Thingsboard_1.Send_gateway_avg
	thingsboard_data.count_step = thingsboard_data.average
}

/************************************************************************/
func main_monitor() {
	/// Setup:
	setup_msg()
	mqtt_mosquitto_reconnect()
	setup_mqtt()
	setupHTTP()
	now_ms := GwChars.Millis()
	check_buff_prev_ms := now_ms
	domoticz_debug_prev_ms := now_ms
	network_last_connected := now_ms

	var network_timeout int64
	if Config.Gateway.NetworkTimeout > 0 {
		network_timeout = int64(Config.Gateway.NetworkTimeout)
	} else {
		network_timeout = 9223372036854775807 // maxInt64
	}

	var debug_timeout int64
	if Config.Gateway.Debug_domoticz > 0 {
		debug_timeout = int64(Config.Gateway.Debug_domoticz)
	} else {
		debug_timeout = 9223372036854775807 // maxInt64
	}

	/// Loop:
	for {
		GwChars.Sleep_ms(Config.Gateway.Send_step) // delay step
		// thingsboard_data.count_step--
		// if thingsboard_data.count_step <= 0 {
		// thingsboard_data.count_step = thingsboard_data.average
		// }

		now_ms = GwChars.Millis()

		if (now_ms - domoticz_debug_prev_ms) >= debug_timeout {
			domoticz_debug_prev_ms = now_ms
			thingsboard_process_debug_msg()
		}

		if (now_ms - check_buff_prev_ms) >= int64(Config.Gateway.MQTTJsonBuffCheck) {
			check_buff_prev_ms = now_ms

			thingsboard_process_log_msg()
			thingsboard_process_monitor_msg()
			thingsboard_process_buffer()
		}

		if thingsboard_1_connected == false && thingsboard_2_connected == false {
			if (now_ms - network_last_connected) >= network_timeout {
				GwChars.Reboot()
			}
		} else {
			network_last_connected = now_ms
		}

	} // end loop
} // end function

/************************************************************************/
func get_log_file(logfile string, num_line_input interface{}) string {
	num_line_default := 10
	line_request, ok := num_line_input.(float64)
	if ok == true {
		num_line_default = int(line_request)
	}

	num_line_str := fmt.Sprintf("%d", num_line_default)

	log, err := exec.Command("tail", "-n", num_line_str, logfile).Output()
	if err != nil {
		return "err"
	}

	lines := strings.Split(string(log), "\n")

	log_file := "{"
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) > 27 {
			log_file += fmt.Sprintf(`"%s":"%s",`, lines[i][0:26], lines[i][27:])
		}
	}
	log_file = log_file[:len(log_file)-1] + "}" // remove last ","

	return log_file
}

/************************************************************************/
func get_current_time() string {
	return time_format(time.Now())
}

/************************************************************************/
func time_format(t time.Time) string {
	_, month, day := t.Date() // year, month, day
	hour, min, sec := t.Clock()
	return fmt.Sprintf("%02d:%02d:%02d - %02d/%02d", hour, min, sec, day, month)
}

/************************************************************************/
func time_format_log(t time.Time) string {
	_, month, day := t.Date() // year, month, day
	//hour, min, sec := t.Clock()
	//us := t.UnixNano() / int64(time.Microsecond)
	// return fmt.Sprintf("%02d-%02d-%02d %s", year , month, day, t.Format("15:04:05.000000"))
	return fmt.Sprintf("%02d-%02d %s ", month, day, t.Format("15:04:05.000"))
}

/************************************************************************/
// func GetModTime() (string) {
// // get last modified time
// stat, err := os.Stat(FILENAME)
// if err != nil {
// log.Println("Err get stat new file: ", err)
// return ""
// }
// return time_format(stat.ModTime())
// }

/************************************************************************/
//var thingsboard_log = make(chan string, 50)

//func thingsboard_add_log(logmsg string) {
//	gateway_log.Thingsboard_log <- (time_format_log(time.Now()) + " [M] " + logmsg)
//}

/************************************************************************/
func thingsboard_process_log_msg() {
	// if ThingsboardJson.S.Len() >= Config.Gateway.MaxGrow {
	// return
	// }

	log_len := len(gateway_log.Thingsboard_log)
	for log_len == 0 {
		return
	}

	log_str := make([]string, log_len)
	for i := 0; i < log_len; i++ {
		log_str[i] = <-(gateway_log.Thingsboard_log)
	}

	s := ThingsboardJson.ObjectBegin("[")
	s.WriteString(`"Monitor.log":"`)
	for i := log_len - 1; i >= 0; i-- {
		s.WriteString(log_str[i])
		s.WriteString(`<br>`)
	}
	s.WriteString(`",`)

	ThingsboardJson.ObjectEndCheckLength()
}

/************************************************************************/
