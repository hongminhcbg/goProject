package tbclient
import (
	"fmt"
)

// TbClient interface monitor connect to host
type TbClient interface{
	Post(string)
	Respond(string, string)
	//SetupCallback(func(TbClient, string, string), string)
}
// Disable use for thingsboard disable
type Disable struct {
    
}

// Start abc
func Start(idDev string) *Disable{
    c := &Disable{}
	fmt.Println(idDev, " disable")
	return c
}

// SetupCallback and do nothng
// func (c *Disable) SetupCallback(callbackFunc func(TbClient, string, string), idDev string){
//     fmt.Println(idDev, " disable")
// }
/*****************************************************************/

// Post do nothing
func (c *Disable) Post(msg string){

}
/********************************************************/

// Respond nothings to host
func (c *Disable) Respond(idRes, mgs string){

}
/*****************************************************************/
