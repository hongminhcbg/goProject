package main

/************************************************************************/
import (
	"fmt"
	"io"
	"os"
	"bytes"
	"strings"
	"net/http"
	"golang.org/x/net/html"
	"encoding/json"
)


/************************************************************************/
// Thông tin các file trong folder Dropbox:
type BOXFILE struct {
	Ts    int64   // timestamp
	Size  int64
}


/************************************************************************/
func FilePrint(fileName string, data string) {
	if file, err := os.Create(fileName); err == nil {
		defer file.Close()
		io.WriteString(file, data);
	} else {
		fmt.Println("[ERR.FilePrint.os.Create]", err)
	}
}


/************************************************************************/
func getScriptContains(doc *html.Node, substr string) (string) {
	var result string = ""
	var f func(*html.Node)
	
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			check := checkScriptContains(n, substr)
			if check != "" {
				result = check
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return result
}


func checkScriptContains(n *html.Node, substr string) (string) {
	if n.Data == "script" {   // check <script>
		var buf bytes.Buffer
		w := io.Writer(&buf)
		html.Render(w, n)
		
		if strings.Contains(buf.String(), substr) {
			return buf.String()
		}
	}

	return ""
}


/************************************************************************/
func Dropbox_CheckFolder(url string) (map[string]BOXFILE) {
	// Get the response from URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("http Get:", err)
		return nil
	}
	defer resp.Body.Close()

	// Convert to string
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(resp.Body)
	bodyString := buffer.String()
	
	//fmt.Printf("%s\n", s)
	//FilePrint("response.html", bodyString)
	
	//----------------------------------------------//
	// Chuyển đổi html: func Parse(r io.Reader) (*Node, error)
	htmlDoc, err := html.Parse( strings.NewReader(bodyString) )
	if err != nil {
		fmt.Println("ERR.html.Parse", err)
		return nil
	}

	// tìm script có chứa thông của các file trong folder:
  script := getScriptContains(htmlDoc, "responseReceived")
	if script == "" {
		fmt.Println("ERR.getScriptContains")
		return nil
	}
	
	//FilePrint("script.html", script)

	//----------------------------------------------//
	// Tách phần Json chứa thông tin của từng file:
	jsonSplit := strings.Split(script, "responseReceived")  // cắt chuỗi trong thân hàm

	start   := strings.Index(    jsonSplit[1], `"`)
	stop    := strings.LastIndex(jsonSplit[1], `"`)
	jsonStr := jsonSplit[1][(start + 1) : stop]         // bỏ 2 kí tự " ở đầu & cuối
	jsonStr  = strings.Replace(jsonStr, `\`, "", -1)    // bỏ hết các kí tự \

	var jsonDecode map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonDecode); err != nil {
		fmt.Println("ERR.json.Unmarshal", err)
		return nil
	}
		
	//----------------------------------------------//
	//fileInfo := jsonDecode["file_info"].([]interface{})
	fileInfo := jsonDecode["entries"].([]interface{})
	
	
	// Lấy thông tin của từng file trong folder:
	var boxFile = map[string]BOXFILE{}
	
	for i := 0; i < len(fileInfo); i++ {
		file := fileInfo[i].(map[string]interface{})
		
		//fileName := file["fq_path"].(string)
		fileName := file["filename"].(string)
		
		//fileName  = fileName[1:len(fileName)]  // remove [0]
		
		if file["is_dir"].(bool) == false {  // nếu là folder thì ["ts"] = nil
			boxFile[fileName] = BOXFILE{ int64(file["ts"].(float64)), int64(file["bytes"].(float64)) }
		} else {
			boxFile[fileName] = BOXFILE{ 0.0, 0 }
		}
	}
	
	return boxFile
}


/************************************************************************/
func Dropbox_GetLink(url string, fileName string) (string) {
	url = url + "&preview=" + fileName
	
	// Get the response from URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("http Get:", err)
		return ""
	}
	defer resp.Body.Close()

	// Convert to string
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(resp.Body)
	bodyString := buffer.String()
	
	//fmt.Printf("%s\n", s)
//FilePrint("Dropbox_Download.response.html", bodyString)
	
	//----------------------------------------------//
	// Chuyển đổi html: func Parse(r io.Reader) (*Node, error)
	htmlDoc, err := html.Parse( strings.NewReader(bodyString) )
	if err != nil {
		fmt.Println("ERR.html.Parse", err)
		return ""
	}

	// tìm script có chứa thông của file cần tìm:
	script := getScriptContains(htmlDoc, "InitReact.mountComponent")
	if script == "" {
		fmt.Println("ERR.getScriptContains")
		return ""
	}
	
//FilePrint("script.html", script)
	
	// Tách phần Json chứa thông tin của từng file:
	jsonSplit := strings.Split(script, "InitReact.mountComponent")  // cắt chuỗi trong thân hàm
	start   := strings.Index(jsonSplit[1], "{")
	stop    := strings.Index(jsonSplit[1], ")")
	jsonStr := jsonSplit[1][start : stop]  // tách chuỗi Json

//FilePrint("jsonStr.json", jsonStr)

	var jsonDecode map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonDecode); err != nil {
		fmt.Println("ERR.json.Unmarshal", err)
		return ""
	}
	
	// for key := range jsonDecode {
		// fmt.Println(key)
	// }
	
	props          := jsonDecode["props"].(map[string]interface{})
	preview        := props["preview"].(map[string]interface{})
	sharedLinkInfo := preview["sharedLinkInfo"].(map[string]interface{})
	
//FilePrint("sharedLinkInfo.json", fmt.Sprintf("%q", sharedLinkInfo))
	//fmt.Println(sharedLinkInfo)
	
	fileUrl := sharedLinkInfo["url"].(string)
	//fmt.Println(fileUrl)
		
	return fileUrl
}


/************************************************************************/
func Dropbox_Download(url string, fileName string) {
	fileLink := Dropbox_GetLink(url, fileName)
		
	Http_GetFile(fileLink, fileName)
}


/************************************************************************/
func Dropbox_DownloadNew(url string, fileName string) {
	fileLink := Dropbox_GetLink(url, fileName)
		
	Http_GetFile(fileLink, fileName + "_new")
}

/************************************************************************/
func Http_GetFile(url string, fileName string) {
	urlRaw := strings.Replace(url, "dl=0", "raw=1", 1)
	
	// Get the data from URL
	resp, err := http.Get(urlRaw)
	if err != nil {
		fmt.Println("[ERR.Http_GetFile.http.Get]", err)
		return
	}
	defer resp.Body.Close()
	
	// Convert to string
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()

	FilePrint(fileName, s)
}

/************************************************************************/
