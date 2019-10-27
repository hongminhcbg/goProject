package main

import (
	"io/ioutil"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
//	"io/ioutil"
	b64 "encoding/base64"
	RegMem "gatewayPackage/register_memory"	
)

var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwOJK1RJBUwRu/5aCyktTaietXFMOAAkElhSq1M6BocVWs7yD
y592CX30Bl0Ul4faWM9EZSlhak8Ay1CdMNis+lBZanKmAO2bPmSIIYBDQE2BzLIo
MoJWi/Cd5PevioKSRPytqVB/S4+xz1IOD8Y407SZM3LfZ5XMfqC+VHpcnAycQ8iT
FK0s3yjImathFNF3U7fiEzU4G7PJRn8e9ntubDd1pXYABqrVF/REcd/3Rs/qrlhG
v3b7tAXZb2lkSLdCq3Md+BMksxUCoH3rZijSphbZSCdIrzofg+IG0y5WtdsBz6uw
Ol2QX/EUoEdO+xhLgaOFykUoWz037ZzkLEhKkQIDAQABAoIBAB+1lAPPSnnxYqYW
Ak5rb70l5LQm20haMyzRHPx7Loh/vq8xsKELCAardDCPoNEAfn7XJDFVSjSF5GWI
TS84j8de6jQ7wNqqNTleoZqQUX4Cv/H83+rdzoiW9/4qUet9Z7p7p7kMCMFNUDf7
D2C8f58eM4lnux52W/X9SwzsSMlGaGHcAKPz4vXUFWyt3naVtANhdkHjgKxA0Ev4
W7yRgpbOKruPKzBNTRXAqq+yHZj/pONtXl8do+plwhHU8CW0BPyvkU4DH52lxWza
mM71ow8UJC30FXF/NZ+wthFnRZO3/dhaeuNYgX7yAs3DhNn7Q8nzU4ujd8ug2OGf
iJ9C8YECgYEA32KthV7VTQRq3VuXMoVrYjjGf4+z6BVNpTsJAa4kF+vtTXTLgb4i
jpUrq6zPWZkQ/nR7+CuEQRUKbky4SSHTnrQ4yIWZTCPDAveXbLwzvNA1xD4w4nOc
JgG/WYiDtAf05TwC8p/BslX20Ox8ZAXUq6pkAeb1t8M2s7uDpZNtBMkCgYEA3QuU
vrrqYuD9bQGl20qJI6svH875uDLYFcUEu/vA/7gDycVRChrmVe4wU5HFErLNFkHi
uifiHo75mgBzwYKyiLgO5ik8E5BJCgEyA9SfEgRHjozIpnHfGbTtpfh4MQf2hFsy
DJbeeRFzQs4EW2gS964FK53zsEtnr7bphtvfY4kCgYEAgf6wr95iDnG1pp94O2Q8
+2nCydTcgwBysObL9Phb9LfM3rhK/XOiNItGYJ8uAxv6MbmjsuXQDvepnEp1K8nN
lpuWN8rXTOG6yG1A53wWN5iK0WrHk+BnTA7URcwVqJzAvO3RYVPqqlcwTKByOtrR
yhxcGmdHMusdWDaVA7PpS1ECgYATCGs/XQLPjsXje+/XCPzz+Epvd7fi12XpwfQd
Z5j/q82PsxC+SQCqR38bwwJwELs9/mBSXRrIPNFbJEzTTbinswl9YfGNUbAoT2AK
GmWz/HBY4uBoDIgEQ6Lu1o0q05+zV9LgaKExVYJSL0EKydRQRUimr8wK0wNTivFi
rk322QKBgHD3aEN39rlUesTPX8OAbPD77PcKxoATwpPVrlH8YV7TxRQjs5yxLrxL
S21UkPRxuDS5CMXZ+7gA3HqEQTXanNKJuQlsCIWsvipLn03DK40nYj54OjEKYo/F
UgBgrck6Zhxbps5leuf9dhiBrFUPjC/gcfyHd/PYxoypHuQ3JUsJ
-----END RSA PRIVATE KEY-----
`)

func RsaDecrypt(ciphertext []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}


func getUID() string{
	RegMem.MemInit()
	SID_128 := make([]uint8, 256) 
	SID_128 = RegMem.MemSlice8(RegMem.ALLWINNERH3_SID_KEY0, 16)
	return fmt.Sprintf("%08X", SID_128)
}

func CheckCopyRight() bool {
	// get real UID
	realUid := getUID()

	fmt.Println("real UID = ", realUid)
	// read data encode in secrect file
	data, err := ioutil.ReadFile("/etc/log123.txt")
	if err != nil {
		return false
	}
	fmt.Println("string read in secrect file", string(data))
	sEnc := string(data)
	
	// decode base64
	sDec, err := b64.StdEncoding.DecodeString(sEnc)
	if err != nil {
		return false
	}

	// decode RSA
	uidDecodeInFile, err := RsaDecrypt(sDec)
	if err != nil {
		return false
	}
	fmt.Println("uidDecodeInFile = ", string(uidDecodeInFile))
	return string(uidDecodeInFile) == realUid
}
func main(){
	fmt.Println("Hello World")
	if CheckCopyRight() {
		fmt.Println("check Copyright success")
	} else {
		fmt.Println("check Copyright flase")
	}
}