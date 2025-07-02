package mpv

/*
#include <stdlib.h>
#include <stdint.h>

// data pointer helper
int64_t * make_int64_t_ptr(int64_t val);
int * make_int_ptr(int val);
double * make_double_ptr(double val);
char ** make_string_ptr(char * val);

int64_t * make_int64_t_ptr(int64_t val) {
    int64_t * ptr = malloc(sizeof(*ptr));
    *ptr = val;
    return ptr;
}

int * make_int_ptr(int val) {
    int * ptr = malloc(sizeof(*ptr));
    *ptr = val;
    return ptr;
}

double * make_double_ptr(double val) {
    double * ptr = malloc(sizeof(*ptr));
    *ptr = val;
    return ptr;
}

char ** make_string_ptr(char * val) {
    char ** ptr = malloc(sizeof(*ptr));
    *ptr = val;
    return ptr;
}
*/
import "C"
import (
	"unsafe"
)

// mallocMpvDataPointer malloc data in C memory space
// don't forget to free the memory after calling this function
func mallocMpvDataPointer(format Format, data interface{}) (ptr unsafe.Pointer) {
	switch format {
	case FORMAT_NONE:
		ptr = nil
	case FORMAT_STRING, FORMAT_OSD_STRING:
		// add null terminate bytes.
		//panic("using GetPropertyString or GetOptionString instead")
		ptr = unsafe.Pointer(C.make_string_ptr(C.CString(data.(string))))
	case FORMAT_FLAG:
		ptr = unsafe.Pointer(C.make_int_ptr(boolToCInt(data.(bool))))
	case FORMAT_INT64:
		i, ok := data.(int64)
		if !ok {
			i = int64(data.(int))
		}
		ptr = unsafe.Pointer(C.make_int64_t_ptr(C.int64_t(i)))
	case FORMAT_DOUBLE:
		ptr = unsafe.Pointer(C.make_double_ptr(C.double(data.(float64))))
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

func freeMpvDataPointer(format Format, ptr unsafe.Pointer) {
	if ptr == nil {
		panic("free a nil pointer")
	}
	switch format {
	case FORMAT_STRING, FORMAT_OSD_STRING:
		C.free(unsafe.Pointer(*(**C.char)(ptr)))
		C.free(ptr)
	default:
		C.free(ptr)
	}
}

func boolToCInt(b bool) (i C.int) {
	i = 0
	if b {
		i = 1
	}
	return
}
