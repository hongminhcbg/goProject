package register_memory

import (
  "reflect"
	"sync"
	"syscall"
	"unsafe"
	"os"
)


/*-----------------------------------------------------------------*/
// Memory offsets, see the spec for more details
const (
	ALLWINNERH3_MEMORY_START       uint32 = 0x01C00000   // System Control base address
	ALLWINNERH3_MEMORY_LENGTH      uint32 = 0x303C00     // System Control -> R_PWM
	
	ALLWINNERH3_CCU_BASE           uint32 = 0x01C20000
	ALLWINNERH3_AHB1_APB1_CFG_REG  uint32 = ALLWINNERH3_CCU_BASE + 0x0054
	ALLWINNERH3_APB2_CFG_REG       uint32 = ALLWINNERH3_CCU_BASE + 0x0058
	ALLWINNERH3_AHB2_CFG_REG       uint32 = ALLWINNERH3_CCU_BASE + 0x005C
	
	ALLWINNERH3_PIO_BASE           uint32 = 0x01C20800
	ALLWINNERH3_PA_DATA            uint32 = ALLWINNERH3_PIO_BASE + 0x0010
	ALLWINNERH3_PB_DATA            uint32 = ALLWINNERH3_PIO_BASE + 0x0058
	ALLWINNERH3_PG_DATA            uint32 = ALLWINNERH3_PIO_BASE + 0x00E8
	
	ALLWINNERH3_PA_CFG1            uint32 = ALLWINNERH3_PIO_BASE + 0x0004
	ALLWINNERH3_PA_CFG2            uint32 = ALLWINNERH3_PIO_BASE + 0x0008
	
	ALLWINNERH3_PG_CFG0            uint32 = ALLWINNERH3_PIO_BASE + 0x00D8
	
	ALLWINNERH3_THS_BASE           uint32 = 0x01C25000
	ALLWINNERH3_THS_DATA           uint32 = ALLWINNERH3_THS_BASE + 0x0080
	
	ALLWINNERH3_SID_BASE           uint32 = 0x01c14200  // http://linux-sunxi.org/SID_Register_Guide
	ALLWINNERH3_SID_KEY0           uint32 = ALLWINNERH3_SID_BASE + 0x0000
	ALLWINNERH3_SID_KEY1           uint32 = ALLWINNERH3_SID_BASE + 0x0004
	ALLWINNERH3_SID_KEY2           uint32 = ALLWINNERH3_SID_BASE + 0x0008
	ALLWINNERH3_SID_KEY3           uint32 = ALLWINNERH3_SID_BASE + 0x000C
)


/*-----------------------------------------------------------------*/
var (
	memBase  uint32  // base address

	// Arrays for 8 / 32 bit access to memory and a semaphore for write locking
	memlock  sync.Mutex
	mem32    []uint32
	mem8     []uint8
)


/*-----------------------------------------------------------------*/
func MemRead8(addr uint32) (uint8) {
	return MemSlice8(addr, 1)[0]
}

func MemRead32(addr uint32) (uint32) {
	return MemSlice32(addr, 1)[0]
}

/*-----------------------------------------------------------------*/
func MemWrite8(addr uint32, data uint8) {
	MemSlice8(addr, 1)[0] = data
}

func MemWrite32(addr uint32, data uint32) {
	MemSlice32(addr, 1)[0] = data
}

/*-----------------------------------------------------------------*/
func MemSlice8(addr uint32, len uint32) ([]uint8) {
	start := (addr - memBase) / 1
	return mem8[start : (start + len)]
}

func MemSlice32(addr uint32, len uint32) ([]uint32) {
	start := (addr - memBase) / 4
	return mem32[start : (start + len)]
}

/*-----------------------------------------------------------------*/
func MemInit() {
	MemOpen(ALLWINNERH3_MEMORY_START, ALLWINNERH3_MEMORY_LENGTH)
}

/*-----------------------------------------------------------------*/
// Open and memory map from /dev/mem
// Some reflection magic is used to convert it to a unsafe []uint32 pointer
func MemOpen(memStart uint32, memSize uint32) (err error) {
	// Open fd for rw mem access
	file, err := os.OpenFile("/dev/mem", os.O_RDWR | os.O_SYNC, 0)
	if err != nil {
		//panic(err)
		return
	}
	defer file.Close()	// FD can be closed after memory mapping

	memlock.Lock()
	defer memlock.Unlock()

	memBase = uint32(memStart)
	mem32, mem8, err = MemMap(file.Fd(), int64(memBase), int(memSize))
	if err != nil {
		//panic(err)
		return
	}
	return nil
}

/*-----------------------------------------------------------------*/
func MemMap(fd uintptr, base int64, length int) (mem []uint32, mem8 []byte, err error) {
	mem8, err = syscall.Mmap(
		int(fd),
		base,
		length,
		syscall.PROT_READ | syscall.PROT_WRITE,
	  syscall.MAP_SHARED,
	)
	if err != nil {
		//panic(err)
		return
	}
	
	// Convert mapped byte memory to unsafe []uint32 pointer, adjust length as needed
	header     := *(*reflect.SliceHeader)(unsafe.Pointer(&mem8))
	header.Len /= (32 / 8) // (32 bit = 4 bytes)
	header.Cap /= (32 / 8)
	mem = *(*[]uint32)(unsafe.Pointer(&header))
	return
}

/*-----------------------------------------------------------------*/
// Close unmaps GPIO memory
func MemClose() (error) {
	memlock.Lock()
	defer memlock.Unlock()
	err := syscall.Munmap(mem8)
	return err
}

/*-----------------------------------------------------------------*/

