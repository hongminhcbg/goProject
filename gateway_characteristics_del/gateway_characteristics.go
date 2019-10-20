package gateway_characteristics

import (
	"fmt"
	"time"
	"strconv"
	"strings"
	"os"
	"os/exec"
	"io/ioutil"
	"net"
	"hash/crc32"
	"math"
	"bytes"
	RegMem "register_memory"  // rename "register_memory" to "RegMem"
	"gateway_log"
)

/************************************************************************/
var (
	led_red    string = ""
	led_green  string = ""
  UID        string = ""
	Mem_total  int64  = 0
)


/************************************************************************/
var list_led_green = []string {
	"/sys/class/leds/orangepi:green:pwr/brightness",
	"/sys/class/leds/green_led/brightness",
}

var list_led_red = []string {
	"/sys/class/leds/orangepi:red:pwr/brightness",
  "/sys/class/leds/red_led/brightness",
}

/************************************************************************/
func Get_file_exist(listfile []string) (string) {
	for i := 0; i < (len(listfile)); i++ {
		if _, err := os.Stat(listfile[i]); err == nil {
			return listfile[i]
		}
	}
	
	return ""
}

/************************************************************************/
func init() {	
	led_red   = Get_file_exist(list_led_red)
	led_green = Get_file_exist(list_led_green)
	
	RegMem.MemInit()  // map memory to array
	UID = Get_UID()
	Mem_total = Get_MemTotal()
}

/************************************************************************/
func Get_CpuInfo() (int64, int64, int64) {
	cpu_info, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0, 0
	}
	cpu_info_words := strings.Fields( string(cpu_info) )
	
	cpu_user, err := strconv.ParseInt(cpu_info_words[1], 10, 64)
	if err != nil {
		return 0, 0, 0
	}
	
	cpu_system, err := strconv.ParseInt(cpu_info_words[3], 10, 64)
	if err != nil {
		return 0, 0, 0
	}

	cpu_idle, err := strconv.ParseInt(cpu_info_words[4], 10, 64)
	if err != nil {
		return 0, 0, 0
	}

	return cpu_user, cpu_system, cpu_idle
}

/************************************************************************/
type Cpu_data struct {
	prev_user    int64
	prev_system  int64
	prev_idle    int64
}

func Get_CPU(p *Cpu_data) (float64) {
	user, system, idle := Get_CpuInfo()
	delta_user   := user   - p.prev_user
	delta_system := system - p.prev_system
	delta_idle   := idle   - p.prev_idle
	p.prev_user   = user
	p.prev_system = system
	p.prev_idle   = idle
	cpu_avg := float64(delta_user + delta_system) / 
	           float64(delta_user + delta_system + delta_idle) * 100.0
	return cpu_avg
}


func AddCPU(w *strings.Builder, p *Cpu_data) {
	user, system, idle := Get_CpuInfo()
	delta_user   := user   - p.prev_user
	delta_system := system - p.prev_system
	delta_idle   := idle   - p.prev_idle
	p.prev_user   = user
	p.prev_system = system
	p.prev_idle   = idle
	cpu_avg := float64(delta_user + delta_system) / 
	           float64(delta_user + delta_system + delta_idle) * 100.0

	w.WriteString(`"Monitor.cpu":`)
	fmt.Fprintf(w, `"%.2f",`, cpu_avg)
}


/************************************************************************/
func Get_MemTotal() (int64) {
	mem_info, err := exec.Command("free").Output()
	if err != nil {
		return -1
	}
	mem_info_words := strings.Fields( string(mem_info) )
	number, err := strconv.ParseInt(mem_info_words[7], 10, 32)
	if err != nil {
		return -1
	}	
	return number
}


func Get_MemUse() (int64) {
	mem_info, err := exec.Command("free").Output()
	if err != nil {
		return 0
	}
	mem_info_words := strings.Fields( string(mem_info) )
	
	mem_use, err := strconv.ParseInt(mem_info_words[8], 10, 64)
	if err != nil {
		return 0
	}
	return mem_use
}


func AddMemUse(w *strings.Builder) {
	w.WriteString(`"Monitor.mem":`)
	fmt.Fprintf(w, `"%.2f",`, float64(Get_MemUse()) / float64(Mem_total) * 100.0)
}

/************************************************************************/
var Temperature_tb1 float64 = 0.0
var Temperature_k1  float64 = 0.0
var Temperature_k2  float64 = 0.0

func Temperature_Init(k1 int, k2 int) {
	Temperature_k1 = float64(k1)
	Temperature_k2 = float64(k2)
}


func Temperature_ReadFile() (int) {
	temp_out, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		return 0
	}
	
	temp_split := strings.Split(string(temp_out), "\n")	// remove \n
	temp_num, err := strconv.ParseInt(temp_split[0], 10, 32)
	if err != nil {
		return 0
	}
	return int(temp_num)
}


func Temperature_ReadReg() (float64) {
	reading := RegMem.MemRead32(RegMem.ALLWINNERH3_THS_DATA) & 0xFFF // 12 bit

	// https://forum.armbian.com/topic/2145-improved-sunxi_tp_temp-for-h3/
	//temp := ((float64(reading) - 2794.0) / -14.882)  // Allwinner H3 datasheet formula
	temp := ((float64(reading) - 2048.0) / -14.882)  // my formula

	return temp
}


func Get_Temperature() (float64) {
	current_temp := Temperature_ReadReg()
	
	delta_temp_1 := current_temp - Temperature_tb1
	invK1 := Temperature_k2 * math.Exp(-Temperature_k1 * delta_temp_1 * delta_temp_1)
	Temperature_tb1 += delta_temp_1 / (1.0 + invK1)
	
	return Temperature_tb1
}


func AddTemperature(w *strings.Builder) {
	current_temp := Temperature_ReadReg()
	
	delta_temp_1 := current_temp - Temperature_tb1
	invK1 := Temperature_k2 * math.Exp(-Temperature_k1 * delta_temp_1 * delta_temp_1)
	Temperature_tb1 += delta_temp_1 / (1.0 + invK1)
	
	w.WriteString(`"Monitor.temp":`)
	fmt.Fprintf(w, `"%.2f",`, Temperature_tb1)
}

func AddUpTime(w *strings.Builder) {
	uptime_raw, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return
	}

	uptime_split := strings.Split(string(uptime_raw), ".")
	uptime_sec, err := strconv.ParseInt(uptime_split[0], 10, 64)
	if err != nil {
		return
	}
	uptime_hour   := (uptime_sec / 3600)
	uptime_minute := (uptime_sec % 3600) / 60
	uptime_draw   := (uptime_hour * 60) + uptime_minute
	uptime_day    :=  uptime_hour / 24
	uptime_hour   %= 24
	if uptime_draw >= 60 {
		uptime_draw = (uptime_draw % 30) + 30
	}
	
	w.WriteString(`"Monitor.uptime":`)
	fmt.Fprintf(w, `"%d",`, uptime_draw)
	w.WriteString(`"Monitor.uptimedisplay":`)
	fmt.Fprintf(w, `"%dd%02dh%02dm",`, uptime_day, uptime_hour, uptime_minute)
}


/************************************************************************/

func Get_UID() (string) {
	SID_128 := RegMem.MemSlice8(RegMem.ALLWINNERH3_SID_KEY0, 16)
	SID_32  := crc32.ChecksumIEEE( SID_128 )
	return fmt.Sprintf("%08X", SID_32)
}


/************************************************************************/

func Get_Overlayroot() (float64) {
	out, err := exec.Command("df", "--output=used", "/").Output()
	if err != nil {
		return 0
	}
  words := strings.Fields( string(out) )
	if len(words) != 2 {
		return 0
	}
	
	used_KB, err := strconv.ParseFloat(words[1], 64)
	if err != nil {
		return 0
	}
	return (used_KB / 1024)  // MB
}


func AddOverlayroot(w *strings.Builder) {
	out, err := exec.Command("df", "--output=used", "/").Output()
	if err != nil {
		return
	}
  words := strings.Fields( string(out) )
	if len(words) != 2 {
		return
	}
	
	used_KB, err := strconv.ParseFloat(words[1], 64)
	if err != nil {
		return
	}
	
	w.WriteString(`"Monitor.overlay":`)
	fmt.Fprintf(w, `"%.2f",`, used_KB / 1024) // MB
}


/************************************************************************/
func Get_IpAddress() (string) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "0"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	
	ip := localAddr.IP.String()
	if strings.Contains(ip, ".") {
		return ip
	}
	return "0"
}


func AddIpAddress(w *strings.Builder) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	w.WriteString(`"Monitor.ip":`)
	fmt.Fprintf(w, `"%s",`, localAddr.IP.String())
}


/************************************************************************/
func Micros() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

func Millis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func Sleep_ms(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))  // sleep(nano second)
}

/************************************************************************/
func Write_file(file string, data []byte) (error) {
	if file == "" {
		return nil
	}
	
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}
	
	err = ioutil.WriteFile(file, data, stat.Mode())
	return err
}


func SetLed_Red(state string) {
	Write_file(led_red, []byte(state))
}

func SetLed_Green(state string) {
	Write_file(led_green, []byte(state))
}

func BlinkLed_Green() {
	SetLed_Green("1")
	Sleep_ms(5)
	SetLed_Green("0")
}

func BlinkLed_Red() {
	SetLed_Red("1")
	Sleep_ms(5)
	SetLed_Red("0")
}

/************************************************************************/
// func detect_gateway() (bool) {
  // // nếu kết nối đc tới Gateway thì sẽ xuất hiện cờ "UG"
  // flag, err := exec.Command("sh", "-c", "route -n").Output()
	
	// if err == nil && strings.Contains( string(flag), "UG") {
		// return true
	// }
	// return false
// }

/************************************************************************/
func Renew_eth0() (bool) {
	// cmd_down := exec.Command("ifconfig", "eth0", "down")
	// cmd_down.Run()
	
	// Sleep_ms(100) // idle
	
	// cmd_up := exec.Command("ifconfig", "eth0", "up")
	// cmd_up.Run()

	//Sleep_ms(5000) // idle
	
	// wait_timeout:
	// for i := 0; i < 30; i++ {
		// Sleep_ms(1000)
		
		// if detect_gateway() == true {
			// return true
		// }
	// }
	
	// return false
	
	return true
}

/************************************************************************/
func Reboot() {
	cmd := exec.Command("sudo", "reboot")
	cmd.Start()  // END !!! 
}

func Poweroff() {
	cmd := exec.Command("sudo", "poweroff")
	cmd.Start()  // END !!!
}

/************************************************************************/
func Adapter_start() {
	cmd_start := exec.Command("sh", "-c", "./gateway_adapter > log_adapter 2>&1 &")
	cmd_start.Start()
}

func Monitor_start() {
	cmd_start := exec.Command("sh", "-c", "./gateway_monitor > log_monitor 2>&1 &")
	cmd_start.Start()
}

/************************************************************************/
func Adapter_restart_adapter() {
	Adapter_start()
	os.Exit(0)  // END !!!
}

func Adapter_restart_monitor() {
	cmd_kill := exec.Command("killall", "gateway_monitor")
	cmd_kill.Run()
	
	Monitor_start()
}


/************************************************************************/
func Monitor_restart_monitor() {
	gateway_log.Thingsboard_add_log("Monitor_restart_monitor")
	Sleep_ms(6000)    
	Monitor_start()
	os.Exit(0)  // END !!!
}

//var Test_global = 0

func Monitor_restart_adapter() {
	gateway_log.Thingsboard_add_log("Monitor_restart_adapter")
	Sleep_ms(6000)    
	
//	Test_global = 1
	cmd_kill := exec.Command("killall", "gateway_adapter")
	cmd_kill.Run()
	
	Adapter_start()
}


/************************************************************************/
func Restart_mosquitto() {
	gateway_log.Thingsboard_add_log("Restart_mosquitto()")
	Sleep_ms(6000)
	cmd := exec.Command("service", "mosquitto", "restart")
	cmd.Start()
}

func Restart_domoticz() {
	gateway_log.Thingsboard_add_log("Restart_domoticz()")
	Sleep_ms(6000)
	cmd := exec.Command("service", "domoticz", "restart")
	cmd.Start()
}

/************************************************************************/
func CheckUpdate() {
	cmd := exec.Command("sh", "-c", "./gateway_checkupdate download > log_checkupdate 2>&1 &")
	cmd.Run()
}

/************************************************************************/
func Exec_Cmd(c string) (string, string) {
	cmd := exec.Command("sh", "-c", c)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	cmd.Run()
	
	// remove "\n"
	out := strings.Replace(outb.String(), "\n", "", -1)
	err := strings.Replace(errb.String(), "\n", "", -1)
	
	return out, err
}

/************************************************************************/
func Commit() (string) {
	// Mount R-W
	Exec_Cmd("mount  -o  remount,rw  /dev/mmcblk0p1")
	
	// Copy
	Exec_Cmd("cp  gateway_checkupdate gateway_adapter gateway_monitor config.json config_name.json  /media/root-ro/root")

	// Mount R-O
	_, err := Exec_Cmd("mount  -o  remount,ro  /dev/mmcblk0p1")
	
	if err != "" {
		return err
	}
	return "Commit done"
}

/************************************************************************/
