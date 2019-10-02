package gatewayhttptb
import(
    "fmt"
)

// Disable use for thingsboard disable
type Disable struct {
    idDev   string
    
}

func NewClient(idDev string) *Disable{
    c := &Disable{idDev:idDev}
    return c
}

// Setup and do nothng
func (c *Disable) SetupCallback(callbackFunc func(interface{}, string, string)){
    fmt.Println(c.idDev, " disable")
}
/*****************************************************************/

// Post do nothing
func (c *Disable) Post(msg string){

}
/********************************************************/

// Respond nothings to host
func (c *Disable) Respond(idRes, mgs string){

}
/*****************************************************************/
