package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// THINGSBOARDHOST is url of host
const THINGSBOARDHOST string = "http://192.168.0.101:8080"                                             //1

// THINGSBOARDTOKEN is monitor tocken
const THINGSBOARDTOKEN string = "zffE7RwyeOpMDFqfYu1M"                                                  //2

// THINGSBOARDCURL is address wait command of user
const THINGSBOARDCURL string = THINGSBOARDHOST + "/api/v1/" + THINGSBOARDTOKEN + "/rpc?timeout=2000000" //3

// ThingsboardGet wait command of user and send result {"status":"ok"}
func ThingsboardGet() {
	fmt.Println("Begin wait command")
	req, _ := http.NewRequest("GET", THINGSBOARDCURL, nil)
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req) //hang on here, wait event

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	//fmt.Println(string(body))
	//fmt.Printf("type of body is %T\n", body)
	dec := json.NewDecoder(bytes.NewReader(body))

	var jsonDecode map[string]interface{}
	if err := dec.Decode(&jsonDecode); err != nil {
		fmt.Println("error")
		fmt.Println(err)
		return
	}
	//fmt.Printf("type of jsonDecode[id] = %T\n", jsonDecode["id"])
	if idrespond1, ok := jsonDecode["id"].(float64); ok {
		idRespond := int(idrespond1)
		fmt.Printf("id receive = %d\n", idRespond)
		httpThingsboardPost("{\"status\":\"ok\"}", strconv.Itoa(idRespond))
		//fmt.Println("Done")
	} else {
		fmt.Println(jsonDecode)
	}
}
/*************************/

// respond function
func httpThingsboardPost(msg string, id string) {
	fmt.Println("begin respond data")
	httpResTHINGSBOARDURL := THINGSBOARDHOST + "/api/v1/" + THINGSBOARDTOKEN + "/rpc/" + id
	req, _ := http.NewRequest("POST", httpResTHINGSBOARDURL, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
	fmt.Println("End respond data")
}
/*********************/

// Sleepms sleep in ms
func Sleepms(ms int) {
	time.Sleep(time.Millisecond * time.Duration(ms)) // sleep(nano second)
}
/**************************/

// thresThingsboardGet make a thres, alway wait command of user
func thresThingsboardGet() {
	for {
		ThingsboardGet()
		Sleepms(100)
	}
}
/*******************************/

func main() {
	go thresThingsboardGet()
	for {
		//ThingsboardPost(msg)
		Sleepms(1000)
	}
}

// curl -v -X POST -d "{\"status\":\"ok\"}" http://demo.thingsboard.io:80/Q8ryIjV0hn1DJvNsGUfg/rpc/1 --header "Content-Type:application/json"
// curl -v -X POST -d "{\"status\":\"ok\"}" http://192.168.0.101:8080:8080/zffE7RwyeOpMDFqfYu1M/rpc/1 --header "Content-Type:application/json"
// curl -v -X GET http://192.168.0.101:8080/api/v1/zffE7RwyeOpMDFqfYu1M/rpc?timeout=20000
