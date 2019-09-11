package gateway_checkMos
import(
    "github.com/eclipse/paho.mqtt.golang"
    "time"
    "fmt"
	"math/rand"    
)

var cMosSta chan string = make(chan string, 1)

/*************************************************************************/
// callback function when mosquitto received message check startup status
func mosquittoCallBack_CheckMos(c mqtt.Client, message mqtt.Message){
    fmt.Println("check mosquitto mgs: " + string(message.Payload()))
    cMosSta<-string(message.Payload())
}
/********************************************************************/

func GetStatus(c mqtt.Client, topic string, timeout time.Duration) bool{
    if(len(cMosSta) > 0){
        <-cMosSta
    }
	c.Subscribe(topic, 0, mosquittoCallBack_CheckMos)
	checkMosStr := fmt.Sprintf("%d", rand.Intn(100))   
    fmt.Println("checkMosStr = " + checkMosStr)
    c.Publish(topic, 0, false, checkMosStr)
    chanTimeOut := time.After(timeout * time.Millisecond)    
    for{
        select {
            case <-chanTimeOut:
                c.Unsubscribe(topic)
                //fmt.Println("case timeout")
                return false
            case msg := <-cMosSta:
                c.Unsubscribe(topic)
                //fmt.Println("case receives msg")
                return (msg == checkMosStr)
        }
    }
}
