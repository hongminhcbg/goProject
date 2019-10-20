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
	GwChars "gatewayPackage/gateway_characteristics" // rename to "GwChars"
	gateway_log "gatewayPackage/gateway_log"
	GP "gatewayPackage/gateway_parse"
	jsonBuffer "gatewayPackage/jsonBuffer"
	GMT "gatewayPackage/tbClient.mqtt" //gateway mqtt thingsboard
	GHT "gatewayPackage/tbClient.http" //gateway http thingsboard
	//GDT "gatewayPackage/tbClient.disable" // gateway disable
	tbclient "gatewayPackage/tbClient"
)

/************************************************************************/

var ThingsboardJson = jsonBuffer.JsonBuffer{}
/************************************************************************/

const CONFIG_FILENAME string = "config.json"
const FILENAME string = "gateway_monitor"

var buildTime string
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
/************************************/

// tbClient interface monitor connect to host
// type tbClient interface{
// 	Post(string)
// 	Respond(string, string)
// 	SetupCallback(func(interface{}, string, string))
// }
/************************************/

// Config struct
var Config = Configuration{}

// ConfigCmd string read in file config
var ConfigCmd = ""

// loadConfigStatus status read data in config.json, false when readfile falselure
var loadConfigStatus = false

// useMqttTB1 true is MQTT and flase is HTTP
var useMqttTB1 = true 

// useMqttTB1 true is MQTT and flase is HTTP
var useMqttTB2 = true

var tb1 tbclient.TbClient
var tb2 tbclient.TbClient
/************************************************************************/

func loadConfig() {
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
	loadConfigStatus = true
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

	logStr := fmt.Sprintf("Change working directory to: %s", dir)
	log.Println(logStr)
	gateway_log.Thingsboard_add_log(logStr)
	return dir
}

/************************************************************************/

func checkFileName() {
	fullname := os.Args[0]
	fullnameSplit := strings.Split(fullname, "/")
	filename := fullnameSplit[len(fullnameSplit)-1]
	//log.Println(fullname, fullname, filename)

	if strings.Contains(filename, "_new") {
		basename := strings.Replace(filename, "_new", "", 1)

		// kill old program
		cmdKillall := exec.Command("killall", basename)
		cmdKillall.Run()

		// rename to old name
		cmdRename := exec.Command("mv", filename, basename)
		cmdRename.Run()

		// start with old name
		cmdStart := exec.Command("sh", "-c", "./gateway_monitor > log_monitor 2>&1 &")
		cmdStart.Start()

		// kill new
		os.Exit(0) // END !!!
	}
}
/************************************************************************/

// checkUpdate check data form dropbox, if change download new file
func checkUpdate() {
	cmdCheckupdate := exec.Command("sh", "-c", "./gateway_checkupdate monitor > log_checkupdate 2>&1 &")
	cmdCheckupdate.Start()
}
/************************************************************************/

// extractBuildTime parse data and get build time
func extractBuildTime() {
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
	extractBuildTime()
	gateway_log.Thingsboard_add_log("Program startup ########################")
	gateway_log.Thingsboard_add_log(fmt.Sprintf("BuildTime: %s", buildTime))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	checkWorkingDir()
	checkFileName()
	checkUpdate()
	loadConfig()
	loadConfigName()
	GP.ReadFileConfig()
	ConfigCmd = GP.StringReadConfigCmd
	useMqttTB1 = strings.Contains(Config.Thingsboard_1.Host, "tcp")
	useMqttTB2 = strings.Contains(Config.Thingsboard_2.Host, "tcp")

	// Loop:
	mainMonitor()
}

/************************************************************************/
func thingsboardProcessDebugMsg() {
	s := ThingsboardJson.ObjectBegin("[")
	s.WriteString(`Monitor.debug:"`)
	s.WriteString(getCurrentTime())
	fmt.Fprintf(s, `<br>%d pkt / %d s<br>`, domoticz_rx_count, Config.Gateway.Debug_domoticz)
	s.WriteString(domoticz_rx_msg)
	s.WriteString(`",`)

	ThingsboardJson.ObjectEndCheckLength()

	// clear
	domoticz_rx_count = 0
	domoticz_rx_msg = ""
}

/************************************************************************/

// thingsboardProcessMonitorMsg add data of monitor into ThingsboardJson
func thingsboardProcessMonitorMsg() {
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

// ThingsboardSendMsg send data to thingsboard
func ThingsboardSendMsg(msg string) {
	tb1.Post(msg)
	tb2.Post(msg)
}
/************************************************************************/

// thingsboardProcessBuffer check queue and S, if have data post to thingsboard
func thingsboardProcessBuffer() {
	//fmt.Println("[lhm log] begin thingsboardProcessBuffer")
	if thingsboard1connected == false && thingsboard2connected == false {
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
	// get data form S end send to TB
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

// AddReporttime add report time to header massage
func AddReporttime(w *strings.Builder) {
	t := time.Now()
	_, month, day := t.Date()
	hour, min, sec := t.Clock()

	w.WriteString(`"Monitor.reptime":`)
	fmt.Fprintf(w, `"%02d:%02d:%02d - %02d/%02d",`, hour, min, sec, day, month)
}
/***********************/

// AddHeapmonitor add monitor heapsys 
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

func loadConfigName() {
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

// getConfigName get name of monitor
func getConfigName(uid string) string {
	if listname[uid] != nil {
		return listname[uid].(string)
	}
	return "Unknow"
}
/************************************************************************/

// thingsboardProcessStartupMsg massage send to TB when start up
func thingsboardProcessStartupMsg() {
	ThingsboardJson.Queue = make(chan string, Config.Gateway.MQTTJsonQueueLength)

	ThingsboardJson.MaxLength = Config.Gateway.MQTTJsonMaxLength
	ThingsboardJson.Size = Config.Gateway.MQTTJsonMaxLength
	ThingsboardJson.S.Grow(Config.Gateway.MQTTJsonMaxLength)

	ThingsboardJson.ObjectBegin("[")
	ThingsboardJson.AddKeyValue(`"Monitor.modtime"`, "%s", buildTime)
	ThingsboardJson.AddKeyValue(`"Monitor.name"`, "%s", getConfigName(GwChars.UID))
	ThingsboardJson.AddKeyValue(`"Monitor.uid"`, "%s", GwChars.UID)
	ThingsboardJson.AddKeyValue(`"Monitor.heapsys"`, "%d", 0)
	ThingsboardJson.AddKeyValue(`"Monitor.debug"`, "%s", getCurrentTime()+`<br>Debug disable`)
	ThingsboardJson.ObjectEndCheckLength()
}
/************************************************************************/

func setupMsg() {
	thingsboardProcessStartupMsg()

	GwChars.Temperature_Init(Config.Gateway.Temperature_k1, Config.Gateway.Temperature_k2)

	thingsboard_data.average = Config.Thingsboard_1.Send_gateway_avg
	thingsboard_data.count_step = thingsboard_data.average
}
/************************************************************************/

func setupTbClient(mqttTb1, mqtttb2 bool){
	if Config.Thingsboard_1.Enable {
		if useMqttTB1 {
			tb1 = GMT.Start(Config.Thingsboard_1.Host, Config.Thingsboard_1.MonitorToken, processAllCommand, "TB1")
			gateway_log.Thingsboard_add_log("TB1 MQTT")
		} else {
			tb1 = GHT.Start(Config.Thingsboard_1.Host, Config.Thingsboard_1.MonitorToken, processAllCommand, "TB1")
			gateway_log.Thingsboard_add_log("TB1 HTTP")
		}
	} else {
		tb1 = tbclient.Start("TB1")
		gateway_log.Thingsboard_add_log("TB1 Disable")
	}

	if Config.Thingsboard_2.Enable {
		if useMqttTB2 {
			tb2 = GMT.Start(Config.Thingsboard_2.Host, Config.Thingsboard_2.MonitorToken, processAllCommand, "TB2")
			gateway_log.Thingsboard_add_log("TB2 MQTT")
		} else {
			tb2 = GHT.Start(Config.Thingsboard_2.Host, Config.Thingsboard_2.MonitorToken, processAllCommand, "TB2")
			gateway_log.Thingsboard_add_log("TB2 HTTP")
		}
	} else {
		tb2 = tbclient.Start("TB2")
		gateway_log.Thingsboard_add_log("TB2 Disable")
	}
}
/*************************************************************************/
// mainMonitor loop here
func mainMonitor() {
	/// Setup:
	setupMsg()
	mqttMosquittoReconnect()
	setupTbClient(useMqttTB1, useMqttTB2)
	nowMs := GwChars.Millis()
	checkBuffPrevMs := nowMs
	domoticzDebugPrevMs := nowMs
	networkLastConnected := nowMs

	var networkTimeout int64
	if Config.Gateway.NetworkTimeout > 0 {
		networkTimeout = int64(Config.Gateway.NetworkTimeout)
	} else {
		networkTimeout = 9223372036854775807 // maxInt64
	}

	var debugTimeout int64
	if Config.Gateway.Debug_domoticz > 0 {
		debugTimeout = int64(Config.Gateway.Debug_domoticz)
	} else {
		debugTimeout = 9223372036854775807 // maxInt64
	}

	/// Loop:
	for {
		GwChars.Sleep_ms(Config.Gateway.Send_step) // delay step
		// thingsboard_data.count_step--
		// if thingsboard_data.count_step <= 0 {
		// thingsboard_data.count_step = thingsboard_data.average
		// }

		nowMs = GwChars.Millis()

		if (nowMs - domoticzDebugPrevMs) >= debugTimeout {
			domoticzDebugPrevMs = nowMs
			thingsboardProcessDebugMsg()
		}

		if (nowMs - checkBuffPrevMs) >= int64(Config.Gateway.MQTTJsonBuffCheck) {
			checkBuffPrevMs = nowMs

			thingsboardProcessLogMsg() // S now [{"ts":"1234", "data":{"Monitor.log":"abc", "":""}},
			thingsboardProcessMonitorMsg() 	// s now [{"ts":"1234", "data":{"Monitor.log":"abc", "":""}}, 
												// 		{"ts":"1235", "data":{"key1":"value1"}},{}]
			thingsboardProcessBuffer()
		}

		if thingsboard1connected == false && thingsboard2connected == false {
			if (nowMs - networkLastConnected) >= networkTimeout {
				GwChars.Reboot()
			}
		} else {
			networkLastConnected = nowMs
		}

	} // end loop
} // end function

/************************************************************************/
// getLogFile noone call him
func getLogFile(logfile string, numLineInput interface{}) string {
	numLineDefault := 10
	lineRequest, ok := numLineInput.(float64)
	if ok == true {
		numLineDefault = int(lineRequest)
	}

	numLineStr := fmt.Sprintf("%d", numLineDefault)

	log, err := exec.Command("tail", "-n", numLineStr, logfile).Output()
	if err != nil {
		return "err"
	}

	lines := strings.Split(string(log), "\n")

	logFile := "{"
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) > 27 {
			logFile += fmt.Sprintf(`"%s":"%s",`, lines[i][0:26], lines[i][27:])
		}
	}
	logFile = logFile[:len(logFile)-1] + "}" // remove last ","

	return logFile
}

/************************************************************************/
func getCurrentTime() string {
	return timeFormat(time.Now())
}

/************************************************************************/
func timeFormat(t time.Time) string {
	_, month, day := t.Date() // year, month, day
	hour, min, sec := t.Clock()
	return fmt.Sprintf("%02d:%02d:%02d - %02d/%02d", hour, min, sec, day, month)
}

/************************************************************************/
func timeFormatLog(t time.Time) string {
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
// return timeFormat(stat.ModTime())
// }

/************************************************************************/
//var thingsboard_log = make(chan string, 50)

//func thingsboard_add_log(logmsg string) {
//	gateway_log.Thingsboard_log <- (timeFormat_log(time.Now()) + " [M] " + logmsg)
//}

/************************************************************************/
func thingsboardProcessLogMsg() {
	// if ThingsboardJson.S.Len() >= Config.Gateway.MaxGrow {
	// return
	// }

	logLen := len(gateway_log.Thingsboard_log)
	for logLen == 0 {
		return
	}

	logStr := make([]string, logLen)
	for i := 0; i < logLen; i++ {
		logStr[i] = <-(gateway_log.Thingsboard_log)
	}

	s := ThingsboardJson.ObjectBegin("[") //s is string Builder
	s.WriteString(`"Monitor.log":"`)
	for i := logLen - 1; i >= 0; i-- {
		s.WriteString(logStr[i])
		s.WriteString(`<br>`)
	}
	s.WriteString(`",`)
	ThingsboardJson.ObjectEndCheckLength()
	// S now [{"ts":"1234", "data":{"Monitor.log":"abc", "":""}},
}
/************************************************************************/
