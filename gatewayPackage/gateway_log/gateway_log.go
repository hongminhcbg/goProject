package gateway_log

import(
	"fmt"
	"time"
)


var Thingsboard_log = make(chan string, 50)

var Header string = "."

func Set_header(head string){
	Header = "[" + head + "] "
}


func time_format_log(t time.Time) (string) {
	_ , month, day := t.Date()  // year, month, day
	return fmt.Sprintf("%02d-%02d %s ", month, day, t.Format("15:04:05.000"))
}


func Thingsboard_add_log(logmsg string) {
	Thingsboard_log <- (time_format_log(time.Now()) + Header + logmsg)
}
