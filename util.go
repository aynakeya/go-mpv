package mpv

/*
#include <mpv_helper.h>
*/
import "C"
import "unsafe"

// getMpvDataPointer malloc data in C memory space
// don't forget to free the memory after calling this function
func getMpvDataPointer(format Format, data interface{}) (ptr unsafe.Pointer) {
	switch format {
	case FORMAT_NONE:
		ptr = nil
	case FORMAT_STRING, FORMAT_OSD_STRING:
		// add null terminate bytes.
		ptr = unsafe.Pointer(C.CString(data.(string)))
	case FORMAT_FLAG:
		ptr = unsafe.Pointer(C.int_ptr(boolToCInt(data.(bool))))
	case FORMAT_INT64:
		i, ok := data.(int64)
		if !ok {
			i = int64(data.(int))
		}
		ptr = unsafe.Pointer(C.int64_t_ptr(C.int64_t(i)))
	case FORMAT_DOUBLE:
		ptr = unsafe.Pointer(C.double_ptr(C.double(data.(float64))))
	case FORMAT_NODE:
		//FORMAT_NODE_ARRAY, FORMAT_NODE_MAP only used under FORMAT_NODE
		ptr = unsafe.Pointer(data.(Node).CNode())
	case FORMAT_BYTE_ARRAY:
		panic("not implement yet")
	default:
		ptr = nil
	}

	return
}

func boolToCInt(b bool) (i C.int) {
	i = 0
	if b {
		i = 1
	}
	return
}
