package main

/*
#include <stdlib.h>
#include <mpv_help.h>
typedef struct Point {
int x , y;
} Point;

Point *a(){
struct Point *x = malloc(sizeof(*x));
x->x = 1;
x->y=2;
return x;
}
*/
import "C"
import (
	"fmt"
)

func main() {
	fmt.Println(C.int(1) == 0, C.int(1) == 1)
	point := C.a()
	fmt.Printf("%d\n", point.y)
	fmt.Println(C.test())
}
