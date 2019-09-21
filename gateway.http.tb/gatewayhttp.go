package gatewayhttptb

import (
	"gateway_log"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"bytes"
	GwChars "gateway_characteristics"
	"strings"
)
/*************************/
type client struct {
	urlPost string
	urlGet string
	urlRes string
}

var httpClient = [2]client{{urlPost:"", urlGet:"", urlRes:""}, {urlPost:"", urlGet:"", urlRes:""}}

var statusSetup = [2]bool{false, false}
/******************************/
func setupCallbackfunc(c client, callbackFunc func(string, string)){
	for {
		req, _ := http.NewRequest("GET", c.urlGet, nil)
		req.Header.Add("Content-Type", "application/json")
	
		res, _ := http.DefaultClient.Do(req)
	
		//defer res.Body.Close()
		//fmt.Println("debug 1", c.urlGet)
		body, _ := ioutil.ReadAll(res.Body)
		dec := json.NewDecoder(bytes.NewReader(body))
		var jsonDecode map[string]interface{}
		if err := dec.Decode(&jsonDecode); err != nil {
			fmt.Println("decode error")
			return
		}
		if idRes1, ok := jsonDecode["id"].(float64); ok {
			idRes := fmt.Sprintf("%d", int(idRes1)) //float to string
			//c.idRes = idRes
			//fmt.Printf("id receive = %s\n", idRes)
			if method, ok := jsonDecode["method"].(string); ok {
				fmt.Println(method)
				callbackFunc(idRes, method)
			}
		}
		GwChars.Sleep_ms(1000)
	}
}
/***********************************/

// Setup setup connect to host and device
func Setup(host, monitorTocken string, callbackFunc func(string, string), index int){
	httpClient[index].urlGet = host + "/api/v1/" + monitorTocken + "/rpc?timeout=2000000"
	httpClient[index].urlPost = host + "/api/v1/" + monitorTocken + "/telemetry"
	httpClient[index].urlRes = host + "/api/v1/" + monitorTocken + "/rpc/" // + id
	statusSetup[index] = true
	gateway_log.Thingsboard_add_log("Thingsboard.OnConnect http")
	fmt.Println("Begin loop in function", httpClient[index], index)
	go setupCallbackfunc(httpClient[index], callbackFunc)
}
/*********************************/

// RespondMsg respond data to host, push data to url {host}/api/v1/{monitorTocken}/rpc/id (id is int number TB send)
func RespondMsg(idRes, msg string){
	if statusSetup[0] == true {
		THINGSBOARDCURL := httpClient[0].urlRes + idRes
		req, _ := http.NewRequest("POST", THINGSBOARDCURL, strings.NewReader(msg))
		req.Header.Add("Content-Type", "application/json")
		http.DefaultClient.Do(req)
	}
	if statusSetup[1] == true {
		THINGSBOARDCURL := httpClient[1].urlRes + idRes
		req, _ := http.NewRequest("POST", THINGSBOARDCURL, strings.NewReader(msg))
		req.Header.Add("Content-Type", "application/json")
		http.DefaultClient.Do(req)
	}
}
/************************************/

// PostMsg post msg to thingsboard
func PostMsg(msg string){
	if statusSetup[0] == true {
		req, _ := http.NewRequest("POST", httpClient[0].urlPost, strings.NewReader(msg))
		req.Header.Add("Content-Type", "application/json")
		http.DefaultClient.Do(req)
	}

	if statusSetup[1] == true {
		req, _ := http.NewRequest("POST", httpClient[1].urlPost, strings.NewReader(msg))
		req.Header.Add("Content-Type", "application/json")
		http.DefaultClient.Do(req)
	}
}