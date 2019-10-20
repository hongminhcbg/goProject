package gateway_parse
import (
	"encoding/json"
	"os"
	"io/ioutil"
	"fmt"
	"strings"
//	"errors"
	"strconv"
)
const CONFIG_FILENAME string = "configcmd.json"
var myMap map[string]interface{}
var string_status []string
var StringReadConfigCmd = ""
var parse_string_function = map[string] func(string, []string) (string, bool){
	"%x"			:parse_addr,
	"%d.%s.%x"		:parse_cba,
	"%s"			:parse_string,
	"%d.%d"			:parse_RF_txretry,
}
func parse_RF_txretry(str string, commands []string) (string, bool) {
	// str = 04 %d.%d
	if(len(commands) != 2) {
		return "error number of parametters", false
	}
	max, _ := strconv.Atoi(commands[0]) //value [1  15]
	timeout, _ := strconv.Atoi(commands[1])  //value [250 4000]
	if max < 0 || max > 15 {
		return "0 <= max <= 15", false
	}
	if timeout < 250 || timeout > 4000 {
		return "250 <= timeout <= 4000", false
	}

	str = strings.Replace(str, " %d", strconv.FormatInt(int64(max), 16), 1)
	str = strings.Replace(str, ".%d", strconv.FormatInt(int64(timeout/250 - 1), 16), 1)

	return str, true

}
func ReadFileConfig() {
	file, err := os.Open(CONFIG_FILENAME)
	if err != nil {
		fmt.Println("Can't open config file:", err)
		return
	}
	defer file.Close()
	byteValue, _ := ioutil.ReadAll(file)
	for i := range(byteValue){
		if byteValue[i] >= 'A' && byteValue[i] <= 'Z'{
			byteValue[i] += 32
		}
	}
	err = json.Unmarshal(byteValue, &myMap)
    StringReadConfigCmd = string(byteValue)
//	fmt.Println("read file config", string(byteValue))
}
func ListOptions(mm map[string]interface{}) string {
	list := " one of"
	for key := range mm {
		list += ` "` + key + `"`
	}
	return list
}
func accert_hex_string(str string) (string, bool){
//	fmt.Printf("accert_hex_string 1\n")
	if len(str) > 10{
		str = str[:10]
	} 
	for i := 0; i < len(str); i++{
		if !((str[i] >= '0' && str[i] <= '9') || (str[i] >= 'a' && str[i] <= 'f')){
//			fmt.Printf("accert_hex_string 2\n")
			return "invalid digit in hex value", false
		}
	}
		return str, true	
}
func parse_addr(str string, p1_t []string) (string, bool){
	p1 := p1_t[0]
	if val, ok := accert_hex_string(p1); ok{
					str = strings.Replace(str, " %x", val, 1)
					return str,true;
				} else {
					return val,false
				}			
}
func parse_cba(str string, cmds []string) (string, bool){
	var p1_chan, p2_bau, p3_add string
	if len(cmds) == 1 {
		p1_chan = cmds[0]
		p2_bau = ""
		p3_add = ""
	} else if len(cmds) == 2 {
		p1_chan = cmds[0]
		p2_bau = cmds[1]
		p3_add = ""
	} else {
		p1_chan = cmds[0]
		p2_bau = cmds[1]
		p3_add = cmds[2]		
	}
	chan_dec,_ := strconv.ParseInt(p1_chan, 10, 64)
	if (chan_dec > 255) || (chan_dec < 0) {
		return "invalid chanel [0 <= chanel <= 255]", false
	}
	if (p1_chan == "") && (p2_bau == "") && (p3_add == ""){
		return "Nothing to execute", false
	}
	if!((p2_bau == "1m") || (p2_bau == "2m") || (p2_bau == "250k") || (p2_bau == "")){
		return "invalid baurate [2M | 1M | 250k]", false
	}
	add_str, ok := accert_hex_string(p3_add)
	if !ok{ //addrees error
		return add_str, false
	}
	//process chanel
	var chan_hex string
	if p1_chan == "" {
		chan_hex = "  "
	} else {
		chan_hex = strconv.FormatInt(chan_dec, 16)
		if len(chan_hex) == 1 {
			chan_hex = "0" + chan_hex
		}
	}
	ans := str
	ans = strings.Replace(ans, " %d", chan_hex, 1)
	
	//process baurate
	var bau_str string
	if p2_bau == ""{
		bau_str = " "
	} else {
		switch p2_bau {
			case "1m":
				bau_str = "0"
			case "250k":
				bau_str = "2"
			case "2m":
				bau_str = "1"
			default:
		}
	}
	ans = strings.Replace(ans, ".%s", bau_str, 1)
	//process address
	ans = strings.Replace(ans, ".%x", add_str, 1) 
	return ans, true
}
func parse_string(str string, cmds []string) (string, bool){
	if len(cmds) > 1 {
		return "command too long", false
	} else{
		return strings.Replace(str, " %s", cmds[0], 1), true
	}
}
func Trace(m map[string]interface{}, cmds []string) (string, bool) {
	temp_interface, ok := m[cmds[0]]
	if !ok {
		return ListOptions(m),false //flase, comment not in map						
	}
	temp_map, ok := temp_interface.(map[string]interface{})
	if ok{
		if len(cmds) == 1{
			return "command is not enough detail, need " + ListOptions(temp_map),false //comment too short, need more infor
		} else {
			return Trace(temp_map,cmds[1:])  //recursive
		}
	} else { // can't make map, left one value
		if len(cmds) == 1{
			return temp_interface.(string),true // true
		} else { // true
			str := temp_interface.(string)
			temp_cmd := strings.Split(str, " ")
			if fn, ok := parse_string_function[temp_cmd[1]]; ok {
				return fn(str, cmds[1:])
			} else {
				return "command too long", false
			}
		}
	}
}
func FindCommand(str string)(string, int64, int64){
		arr_str := strings.Split(strings.ToLower(str), ".")
//		fmt.Printf("length comman = %d\n", len(arr_str))
//		fmt.Println(arr_str)
		id_node := arr_str[1]
		if len(arr_str) < 3 {
			return "command too short", -1, 0
		}

		for index,value := range(arr_str){
			if value == " "{
				arr_str = arr_str[:index]
				break
			}
		}
		//arr_str = strings.Split(strings.ToLower("nodecmd.09.lednew.on"), ".")
		//get node id
		arr_str = append(arr_str[:1], arr_str[2:]...) // merge string, don't process arr_str[1] 
		id_node_int, _ := strconv.ParseInt(id_node, 16, 64)
		if id_node_int > 255 || id_node_int < 0 {
			return "invalid node id", -1, 0
		}
		if val, err := Trace(myMap, arr_str); err {
			if val[0] == 't'{
				// process wait function
				val_arg := strings.Split(val, " ")
				timeout_str := val_arg[0]
				timeout, _ := strconv.ParseInt(timeout_str[1:], 10, 64)
				// fmt.Printf("timeout = %d\n", timeout)
				// fmt.Printf("nodeid = %d\n", id_node_int)
				// fmt.Printf("val = %s\n", val_arg[1])
				return val_arg[1], id_node_int, timeout	
			}
			//timout = 0, have no wait
			return val, id_node_int, 0
		} else {
			return val, -1, 0
		}

}