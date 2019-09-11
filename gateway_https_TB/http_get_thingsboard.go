package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
   "strings"
  "time"
  "encoding/json"
  "bytes"
  "strconv"
)


/// Setup
const THINGSBOARD_HOST  string = "http://demo.thingsboard.io"
const THINGSBOARD_TOKEN string = "Q8ryIjV0hn1DJvNsGUfg"

// const THINGSBOARD_CURL  string = THINGSBOARD_HOST + "/api/v1/" + THINGSBOARD_TOKEN + "/telemetry"
const THINGSBOARD_CURL  string = THINGSBOARD_HOST + "/api/v1/" + THINGSBOARD_TOKEN + "/rpc?timeout=2000000"
/********************************/
func ThingsboardGet() {
  fmt.Println("fuck")
  req, _ := http.NewRequest("GET", THINGSBOARD_CURL, nil)
  req.Header.Add("Content-Type", "application/json")
  
  res, _ := http.DefaultClient.Do(req) //hang on here, wait event
   
  defer res.Body.Close()
  body, _ := ioutil.ReadAll(res.Body)
   
    fmt.Println(string(body))
    fmt.Printf("type of body is %T\n", body)
	dec := json.NewDecoder(bytes.NewReader(body))
    
	var json_decode map[string]interface{}
	if err := dec.Decode(&json_decode); err != nil {
        fmt.Println("error")
        fmt.Println(err)
        return
	}
  fmt.Printf("type of json_decode[id] = %T\n", json_decode["id"])
	if id_respond_1, ok := json_decode["id"].(float64); ok{
        id_respond := int(id_respond_1)
        fmt.Printf("id receive = %d\n", id_respond)
        http_ThingsboardPost("{\"status\":\"ok\"}", strconv.Itoa(id_respond))	
        fmt.Println("Done")
	} else {
        fmt.Println(json_decode)
    }
}

/*************************/
// respond function

func http_ThingsboardPost(msg string, id string) {
fmt.Println("begin respond data")	
http_res_THINGSBOARD_HOST   := "http://demo.thingsboard.io:80"
http_res_THINGSBOARD_TOKEN  := "Q8ryIjV0hn1DJvNsGUfg"
http_res_THINGSBOARD_CURL   := http_res_THINGSBOARD_HOST + "/api/v1/" + http_res_THINGSBOARD_TOKEN + "/rpc/" + id

  req, _ := http.NewRequest("POST", http_res_THINGSBOARD_CURL, strings.NewReader(msg))
  req.Header.Add("Content-Type", "application/json")
  http.DefaultClient.Do(req)
  fmt.Println("End respond data")	
}
/****************************/
func Sleep_ms(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))  // sleep(nano second)
}
func thres_ThingsboardGet(){
    for{
        ThingsboardGet()
        Sleep_ms(100)
    }
}

func main() {
  go thres_ThingsboardGet()
  
  for {
    //ThingsboardPost(msg)
    Sleep_ms(1000)
  }
}
//curl -v -X POST -d "{\"status\":\"ok\"}" http://demo.thingsboard.io:80/Q8ryIjV0hn1DJvNsGUfg/rpc/1 --header "Content-Type:application/json"