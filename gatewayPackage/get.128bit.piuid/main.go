package main
import (
	"fmt"
	RegMem "gatewayPackage/register_memory"
)

func main(){
	RegMem.MemInit()
	SID_128 := make([]uint8, 256) 
	SID_128 = RegMem.MemSlice8(RegMem.ALLWINNERH3_SID_KEY0, 16)
	fmt.Println("uid 128 bits: ", fmt.Sprintf("%08X", SID_128))
}