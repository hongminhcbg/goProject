package tbclienthttp

import (
//	gateway_log "gatewayPackage/gateway_log"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"bytes"
	GwChars "gatewayPackage/gateway_characteristics"
	"strings"
	tbclient "gatewayPackage/tbClient"
	"time"
	"net"
)

/*************************/

// Client http
type Client struct {
	urlPost 			string
	urlGet 				string
	urlRes 				string
	idDev 				string
	hostAlive 			chan bool // len chan == 0 hostAlive
}
/*********************************/

// parseURL split host to url, EX tcp://192.168.0.102:1883 => 192.168.0.102:1883
func parseURL(host string) string{
	arr := strings.Split(host, "//")
	return arr[1]
}
/*************************************/

// testURLCanReach check url can reach, if can't reach, pi will kill thread http and not reconnect
func testURLCanReach(url string) bool {
	timeout := time.Duration(1 * time.Second)
	_, err := net.DialTimeout("tcp", url, timeout)
	if err != nil {
		return false
	}
	return true
}
/************************************/

// Start create new client
func Start(host string, monitorTocken string, CB func(tbclient.TbClient, string, string), idDev string) *Client {
	c := &Client{}
	c.urlGet = host + "/api/v1/" + monitorTocken + "/rpc?timeout=2000000"
	c.urlPost = host + "/api/v1/" + monitorTocken + "/telemetry"
	c.urlRes = host + "/api/v1/" + monitorTocken + "/rpc/" // + id
	c.hostAlive = make(chan bool, 1)
	// ping test
	url := parseURL(host);
	if !testURLCanReach(url){
		fmt.Println(idDev, "HTTP can't connect, wait 30s and reconnect")
		c.hostAlive <- false
	} else {
		fmt.Println(idDev, "HTTP on connect")
	}
	go func(){
		for {
			req, _ := http.NewRequest("GET", c.urlGet, nil)
			req.Header.Add("Content-Type", "application/json")
			res, _ := http.DefaultClient.Do(req)
			if res == nil { // host die
				if len(c.hostAlive) == 0 {
					c.hostAlive <- false
				}
				fmt.Println("host die, wait 30s and reconnect")
				time.Sleep(30000 * time.Millisecond)
				continue
			}
			if len(c.hostAlive) != 0 {
				<-c.hostAlive
			}

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Println("[lhm log] read body error")
			}
			dec := json.NewDecoder(bytes.NewReader(body))
			var jsonDecode map[string]interface{}
			if err = dec.Decode(&jsonDecode); err != nil {
				fmt.Println("decode error")
			} else if idRes1, ok := jsonDecode["id"].(float64); ok {
				idRes := fmt.Sprintf("%d", int(idRes1)) //float to string
				//fmt.Printf("id receive = %s\n", idRes)
				if method, ok := jsonDecode["method"].(string); ok {
					//fmt.Println(method)
					CB(tbclient.TbClient(c), idRes, method)
				}
			}
			GwChars.Sleep_ms(100)
		}
	}()
	return c
}
/**************************************/

// Post post msg to host
func (c *Client) Post(msg string) {
	if len(c.hostAlive) != 0 {
		fmt.Println("host not alive, can't post msg")
		return
	}

	req, _ := http.NewRequest("POST", c.urlPost, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}
/*************************************/

// Respond to host
func (c *Client) Respond(idRes string, msg string){
	if len(c.hostAlive) != 0 {
		fmt.Println("host not alive, can't respond msg")
		return
	}
	THINGSBOARDCURL := c.urlRes + idRes
	req, _ := http.NewRequest("POST", THINGSBOARDCURL, strings.NewReader(msg))
	req.Header.Add("Content-Type", "application/json")
	http.DefaultClient.Do(req)
}
/***************************************/

