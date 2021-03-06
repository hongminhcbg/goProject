package main // setup
import (
	"fmt"
//	RegMem "gatewayPackage/register_memory"
	b64 "encoding/base64"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
//	"os/exec"
	"strings"
	"io/ioutil"
	"os"
)

// backup
// uid = 8100C00220460004080467780E09233C
// var publicKey = []byte(`
// -----BEGIN PUBLIC KEY-----
// MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwOJK1RJBUwRu/5aCyktT
// aietXFMOAAkElhSq1M6BocVWs7yDy592CX30Bl0Ul4faWM9EZSlhak8Ay1CdMNis
// +lBZanKmAO2bPmSIIYBDQE2BzLIoMoJWi/Cd5PevioKSRPytqVB/S4+xz1IOD8Y4
// 07SZM3LfZ5XMfqC+VHpcnAycQ8iTFK0s3yjImathFNF3U7fiEzU4G7PJRn8e9ntu
// bDd1pXYABqrVF/REcd/3Rs/qrlhGv3b7tAXZb2lkSLdCq3Md+BMksxUCoH3rZijS
// phbZSCdIrzofg+IG0y5WtdsBz6uwOl2QX/EUoEdO+xhLgaOFykUoWz037ZzkLEhK
// kQIDAQAB
// -----END PUBLIC KEY-----
// `)

// new 
var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCzSsKrGq/mb+tnyaEnCxwapTJZ
Vg3CNTyfTOaXakFzlpG88qXoJXMhfn/8ktoA0EFSNQXwmCwFI/Xf8wc+vo1SyySM
DXAKZCAyEmkISC76+crB7BPXzArvbbKpnvBs9He2XWhl8xCBhfoAzz3isl4I3SD6
tV2z8pGVRIq2J6GiAQIDAQAB
-----END PUBLIC KEY-----
`)

// func getUID() string{
// 	RegMem.MemInit()
// 	SID_128 := make([]uint8, 256) 
// 	SID_128 = RegMem.MemSlice8(RegMem.ALLWINNERH3_SID_KEY0, 16)
// 	return fmt.Sprintf("%08X", SID_128)
// }

func RsaEncrypt(origData []byte) ([]byte, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// executeCommand break freeze to write uid to secret file
// func executeCommand(commandString string) (error) {
// 	arr := strings.Split(commandString, " ")
// 	cmd := exec.Command(arr[0], arr[1:]...)
// 	_, err := cmd.CombinedOutput()
// 	if(err != nil){
// 		return err
// 	}
// 	return nil
// }

// writeData2ConfigFile write data after encode base64 to secrect file
func writeData2ConfigFile(fileNane, Key string) bool {
	data, err := ioutil.ReadFile(fileNane) 
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(data))
	orgStr := string(data)
	if ok := strings.Index(orgStr, `"GatewayToken"`); ok != -1 {
		preFix := ""
		posFix := ""
		next := -1
		for i := ok + len(`"GatewayToken"`); i < len(orgStr); i++ {
			if orgStr[i] == '"' {
				preFix = orgStr[:i+1]
				next = i + 1
				break
			}
		}
		if next == -1 {
			return false
		}
		for i := next; i < len(orgStr); i++ {
			if orgStr[i] == '"'{
				posFix = orgStr[i:]
				break
			}
		}

		newStr := preFix + Key + posFix
		fmt.Println(newStr)
		ioutil.WriteFile(fileNane, []byte(newStr), 0666)
		return true

	} else if ok = strings.Index(orgStr, `"MQTTJsonBuffCheck"`); ok != -1{
		for i := ok; i < len(orgStr); i++ {
			if orgStr[i] == '\n' {
				preFix := orgStr[:i]
				posFix := orgStr[i:]
				//fmt.Println("preFix = ", preFix, "posFix = ", posFix)
				newStr := preFix + ",\n    \"GatewayToken\":\"" + Key + "\"" + posFix
				fmt.Println(newStr)
				ioutil.WriteFile(fileNane, []byte(newStr), 0666)
				return true
			}
		}
	}
	return false
}

func main(){
	var uid string
	if len(os.Args) > 1 {
		uid = os.Args[1]
	} else {
		fmt.Scan(&uid)
	}
	fmt.Println("uid: ", uid)

	encodeRSA, err := RsaEncrypt([]byte(uid))
	if err != nil {
		panic(err)
	}

	// enocde base64
	sEnc := b64.StdEncoding.EncodeToString([]byte(encodeRSA))
	
	fmt.Println("gatewayTocken = ", sEnc)

	if writeData2ConfigFile("config.json", sEnc) {
		fmt.Println("\n\n =======> Done Active <=======")
	} else {
		fmt.Println("\n\n =======> Active False <=======")
	}
}