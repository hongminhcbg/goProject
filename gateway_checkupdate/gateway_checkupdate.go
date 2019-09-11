package main

import (
	"fmt"
	"io"
    "os"
	"encoding/json"
	"time"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"hash/crc32"
	"register_memory"
)

/************************************************************************/
const AUTOUPDATE_CONFIG  string = "autoupdate.json"

var AUTOUPDATE_FOLDER string = ""
var         MY_FOLDER string = ""


/************************************************************************/
var generalFolder = map[string]BOXFILE{}
var myFolder         = map[string]BOXFILE{}

var timeStamp map[string]interface{}

var arg         string  = ""
var SID_32_str  string  = ""


/************************************************************************/
const (
	FILE_CONFIG      = "config.json"
    FILE_CONFIG_CMD  = "configcmd.json"
	FILE_CONFIG_NAME = "config_name.json"
	FILE_ADAPTER     = "gateway_adapter"
	FILE_MONITOR     = "gateway_monitor"
	FILE_CHECKUPDATE = "gateway_checkupdate"
)

const FILE_NEW = "_new"

/************************************************************************/
var file_general = []string{FILE_ADAPTER, 
														FILE_MONITOR, 
														FILE_CHECKUPDATE,
														FILE_CONFIG_NAME,
														"mqttserver.pub.pem"}
														
var file_specific = []string{   FILE_CONFIG,
                                FILE_CONFIG_CMD,
                                "mqttclient_monitor.cer.pem",
								"mqttclient_monitor.key.pem"}


/************************************************************************/
var	ts_config       int64 = 0
var ts_adapter      int64 = 0
var ts_monitor      int64 = 0
var	ts_config_name  int64 = 0


/************************************************************************/
func change_working_dir() {
	dir, err := filepath.Abs( filepath.Dir(os.Args[0]) )
	if err != nil {
		panic(err)
	}
	
	err = os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}


/************************************************************************/
// main
/************************************************************************/
func main() {
	TimeStamp_Load()
	fmt.Println("end load timestamp")
	change_working_dir()
	fmt.Println("end change_working_dir")
	register_memory.MemInit()  // map memory to array
	fmt.Println("end register_memory.MemInit")
	defer register_memory.MemClose()
	SID_128 := register_memory.MemSlice8(register_memory.ALLWINNERH3_SID_KEY0, 16)
	SID_32  := crc32.ChecksumIEEE( SID_128 )
	SID_32_str = fmt.Sprintf("%08X", SID_32)

	// SID_32_str = "4A9980F3"
	
	argsWithoutProg := os.Args[1:]
	if (len(argsWithoutProg) == 1) {
		arg = argsWithoutProg[0]
	}
	
	load_autoupdate_config()
	
	generalFolder = Dropbox_CheckFolder(AUTOUPDATE_FOLDER)
	fmt.Println(AUTOUPDATE_FOLDER)
	fmt.Println(generalFolder)
	fmt.Println()
	
	MY_FOLDER = Dropbox_GetLink(AUTOUPDATE_FOLDER, SID_32_str)
	myFolder = Dropbox_CheckFolder(MY_FOLDER)
	fmt.Println(MY_FOLDER)
	fmt.Println(myFolder)
	fmt.Println()
	
	
	switch arg {
		case "download"   : download_all()
		case "all"        : checkupdate_all()
		case FILE_ADAPTER : checkupdate_execute(FILE_ADAPTER)
		case FILE_MONITOR : checkupdate_execute(FILE_MONITOR)
	}
	
	TimeStamp_Save()
}

/************************************************************************/
func download_all() {
	fmt.Println("Download file_general:")
	for i := 0; i < len(file_general); i++ {
		filename := file_general[i]
		
		fmt.Print("Downloading ", filename, " ... ")
		Dropbox_Download(AUTOUPDATE_FOLDER, filename)
		fmt.Println("done.")
		
		timeStamp[filename] = generalFolder[filename].Ts
	}

	fmt.Println("Download file_specific: ", SID_32_str)
	for i := 0; i < len(file_specific); i++ {
		filename := file_specific[i]
		fmt.Print("Downloading ", filename, " ... ")
		Dropbox_Download(MY_FOLDER, filename)
		fmt.Println("done.")
		
		timeStamp[filename] = myFolder[filename].Ts
	}
}


/************************************************************************/
func checkupdate_all() {
	for i := 0; i < len(file_general); i++ {
		filename := file_general[i]
		
		fmt.Println("Checking", filename)
		
		ts, ok := timeStamp[filename].(float64)
		if !ok || int64(ts) != generalFolder[filename].Ts{
			
			timeStamp[filename] = generalFolder[filename].Ts

			switch filename {
				case FILE_ADAPTER: checkupdate_execute(FILE_ADAPTER)
				case FILE_MONITOR: checkupdate_execute(FILE_MONITOR)
				
				default:
				  fmt.Println("  Download", filename)
				  Dropbox_Download(AUTOUPDATE_FOLDER, filename)
			}	
		}
	}

	for i := 0; i < len(file_specific); i++ {
		filename := file_specific[i]
		fmt.Println("Checking", filename)

		ts, ok := timeStamp[filename].(float64)
		if ! ok || int64(ts) != myFolder[filename].Ts {
			timeStamp[filename] = myFolder[filename].Ts
			
			fmt.Println("  Download", filename)
			Dropbox_Download(MY_FOLDER, filename)
		}
	}
}


/************************************************************************/
func checkupdate_execute(filename string) {
	file_new := filename + FILE_NEW
	fmt.Println("  Download", file_new)

	Dropbox_DownloadNew(AUTOUPDATE_FOLDER, filename)

	// check file size:
	stat, err := os.Stat(file_new)
	if err == nil && stat.Size() == generalFolder[filename].Size {
		timeStamp[filename] = generalFolder[filename].Ts
		
		fmt.Println("Add execute")
		cmd_chmod := exec.Command("chmod", "+x", file_new)
		cmd_chmod.Run()
		sleep_ms(10)
		
		fmt.Println("Start", file_new)
		
		file_log := "log_" + filename
		cmd_start := exec.Command("sh", "-c", "./" + file_new + " > " + file_log + " 2>&1 &")
		cmd_start.Start()  // Start starts the specified command but does not wait for it to complete	
	}
}


/************************************************************************/
func sleep_ms(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))  // sleep(nano second)
}


/************************************************************************/
func TimeStamp_Load() {
	timeStamp = make(map[string]interface{})
	
  file, err := ioutil.ReadFile("timestamp.json")
	if err != nil {
		return
	}
	// fmt.Printf("%s\n", file)
	
	json_err := json.Unmarshal(file, &timeStamp)
	if json_err != nil {
		fmt.Println("TimeStamp_Load err:", json_err)
		return
	}
	//fmt.Println(timeStamp)
}


/************************************************************************/
func TimeStamp_Save() {
	json, err := json.MarshalIndent(timeStamp, "", "  ")
	//fmt.Println(string(json))
	
	// Create new file
	out, err := os.Create("timestamp.json")
	if err != nil {
		return
	}
	defer out.Close()
	
	// Write the body to new file
	io.WriteString(out, string(json))
}


/************************************************************************/
func load_autoupdate_config() {
  autoupdate, err := ioutil.ReadFile(AUTOUPDATE_CONFIG)
	if err != nil {
		return
	}
	var m map[string]interface{}
	json.Unmarshal([]byte(autoupdate), &m)
	
	if m["autoupdate"] != nil {
		AUTOUPDATE_FOLDER = m["autoupdate"].(string)
	}
	// fmt.Println(AUTOUPDATE_FOLDER)
}


/************************************************************************/
