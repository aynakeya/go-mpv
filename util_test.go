package mpv

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestUnsafePointer(t *testing.T) {
	a := "ABCDEFGH"
	fmt.Printf("%x %p\n", unsafe.Pointer(&a), unsafe.Pointer(&a))
	bptr := unsafe.Pointer(&append([]byte(a), 0)[0])
	//bptr := unsafe.Pointer(&[]byte(a)[0])
	a0 := []byte("AA")
	a1 := []byte("BBBB")
	a2 := []byte("CCCCCC")
	a3 := []byte("DDDDDD")
	a4 := []byte("EEEEEE")
	a5 := []byte("FFFFFF")
	fmt.Printf("%p %p\n", bptr, unsafe.Pointer(uintptr(bptr)+uintptr(1)))
	fmt.Printf("%p %p %p %p %p %p\n", a0, a1, a2, a3, a4, a5)
	for i := 0; i < 64; i++ {
		fmt.Printf("0x%x\n", *(*byte)(unsafe.Pointer(uintptr(bptr) + uintptr(i))))
	}
	fmt.Println(a0, a1, a2, a3, a4, a5)

}

func TestInt64(t *testing.T) {
	var a int = 13
	var b interface{} = a
	fmt.Println(b.(int64))

}

func TestGetPtr(t *testing.T) {
	fmt.Println(getMpvDataPointer(FORMAT_NONE, nil))
	fmt.Println(getMpvDataPointer(FORMAT_STRING, "asdfasdf"))
	//fmt.Println(*(*int64)(getMpvDataPointer(FORMAT_INT64, 1234)))
}
