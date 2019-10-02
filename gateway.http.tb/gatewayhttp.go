package gatewayhttptb

import (
//	"gateway_log"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"bytes"
	GwChars "gateway_characteristics"
	"strings"
)

/*************************/

// Client http
type Client struct {
	urlPost 			string
	urlGet 				string
	urlRes 				string
	idDev 				string
}
/************************************/

// NewClient create new client
func NewClient(host string, monitorTocken string, idDev string) *Client {
	c := &Client{}
	c.urlGet = host + "/api/v1/" + monitorTocken + "/rpc?timeout=2000000"
	c.urlPost = host + "/api/v1/" + monitorTocken + "/telemetry"
	c.urlRes = host + "/api/v1/" + monitorTocken + "/rpc/" // + id
	c.idDev = idDev
	return c
}
/**********************************************/

func (c *Client) SetupCallback(CB func(interface{}, string, string)){
	go func(){
		fmt.Println(c.idDev, " on connect HTTP")
		for {
			req, _ := http.NewRequest("GET", c.urlGet, nil)
			req.Header.Add("Content-Type", "application/json")
			res, _ := http.DefaultClient.Do(req)
			body, _ := ioutil.ReadAll(res.Body)
			dec := json.NewDecoder(bytes.NewReader(body))
			var jsonDecode map[string]interface{}
			if err := dec.Decode(&jsonDecode); err != nil {
				fmt.Println("decode error")
			} else if idRes1, ok := jsonDecode["id"].(float64); ok {
				idRes := fmt.Sprintf("%d", int(idRes1)) //float to string
				//fmt.Printf("id receive = %s\n", idRes)
				if method, ok := jsonDecode["method"].(string); ok {
					//fmt.Println(method)
					//callbackFunction(idRes, method, idDev)
					CB(c, idRes, method)
				}
			}
			GwChars.Sleep_ms(100)
		}
	}()
}
/**************************************/

// Post post msg to host
func (c *Client) Post(msg string) {
	req, _ := http.NewRequest("POST", c.urlPost, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)	
}
/*************************************/

// Respond to host
func (c *Client) Respond(idRes string, msg string){
	THINGSBOARDCURL := c.urlRes + idRes
	req, _ := http.NewRequest("POST", THINGSBOARDCURL, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}
/***************************************/

