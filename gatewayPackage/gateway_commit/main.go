package gateway_commit 
import(
"os/exec"
"strings"
"strconv"
)
/***************************************************************************/
const rw_strings string = "mount -o remount,rw /dev/mmcblk0p1"
//const cp_strings string = "cp -ru /root/iot_gateway /media/root-ro/root"
const cp_strings string = "rsync -ruv /root/iot_gateway/ /media/root-ro/root/iot_gateway"
const ro_strings string = "mount -o remount,ro /dev/mmcblk0p1"
const cd_string string = "cd /media/root-ro/root/iot_gateway"
const li_string string = "ls -lh /media/root-ro/root/iot_gateway"

/***************************************************************************/

var ans strings.Builder
var temp_ok int
func execute_command(command_string string) (error) {
	ans.WriteString("> " + command_string + "\n")
	arr := strings.Split(command_string, " ")
	cmd := exec.Command(arr[0], arr[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()
	if(err != nil){
		ans.WriteString(string(stdoutStderr))
	} else{
		ans.WriteString("ok (" +strconv.Itoa(temp_ok) + ") " + string(stdoutStderr))
		temp_ok++;
	}
	ans.WriteString("\n")
//	time.Sleep(1000*time.Millisecond)
	return err		
}
func Commit() string {
	ans.Reset()
	temp_ok = 1;
	//_ = execute_command(li_string)
	err := execute_command(rw_strings)
	if err != nil{
		_ = execute_command(ro_strings)
		return ans.String()
	}
	err = execute_command(cp_strings)
	if err != nil{
		_= execute_command(ro_strings)
		return ans.String()		
	}
	_= execute_command(ro_strings)
	//_ = execute_command(li_string)
	return ans.String()	
}
