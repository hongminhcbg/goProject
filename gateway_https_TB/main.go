package main

import (
  "fmt"
  "net/http"
  // "io/ioutil"
 // "bytes"
  "strings"
  "time"
)


/// Setup
const THINGSBOARD_HOST  string = "http://demo.thingsboard.io:80"
const THINGSBOARD_TOKEN string = "Q8ryIjV0hn1DJvNsGUfg"
const THINGSBOARD_CURL  string = THINGSBOARD_HOST + "/api/v1/" + THINGSBOARD_TOKEN + "/telemetry"


func ThingsboardPost(msg string) {
  req, _ := http.NewRequest("POST", THINGSBOARD_CURL, strings.NewReader(msg))
  req.Header.Add("Content-Type", "application/json")
  http.DefaultClient.Do(req)
  
  /// check res & err:
  // res, _ := http.DefaultClient.Do(req)
  // body, err := ioutil.ReadAll(res.Body)
  // fmt.Println(string(body), err)
}


func Sleep_ms(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms))  // sleep(nano second)
}


func main() {
  counter := 0
  
  for {
    
    key := "golang_http_test"
    val := fmt.Sprintf("%d", counter)

    msg := "{" + key + ":" + val + "}"
    //fmt.Printf(THINGSBOARD_CURL)
    ThingsboardPost(msg)
    counter++
    fmt.Println("publish msg -> done")
    Sleep_ms(1000)
    
  }
}