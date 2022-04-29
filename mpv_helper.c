#include <mpv/client.h>
#include <stdlib.h>
#include <string.h>

// node helpers, since cgo don't have direct access to c union
char * get_node_string(mpv_node * node) {
    return (char*)node->u.string;
}

void set_node_string(mpv_node * node,char * val) {
    node->u.string = val;
}

int get_node_flag(mpv_node * node) {
    return (int)node->u.flag;
}

void set_node_flag(mpv_node * node,int val) {
    node->u.flag = val;
}

int64_t get_node_int64(mpv_node * node) {
    return (int64_t)node->u.int64;
}

void set_node_int64(mpv_node * node,int64_t val) {
    node->u.int64 = val;
}

double get_node_double(mpv_node * node) {
    return (double)node->u.double_;
}

void set_node_double(mpv_node * node,double val) {
    node->u.double_ = val;
}

mpv_node_list * get_node_list(mpv_node * node) {
    return (mpv_node_list *)node->u.list;
}

void set_node_list(mpv_node * node,mpv_node_list * val) {
    node->u.list = val;
}

mpv_byte_array * get_node_byte_array(mpv_node * node) {
    return (mpv_byte_array *)node->u.ba;
}

void set_node_byte_array(mpv_node * node,mpv_byte_array * val) {
    node->u.ba = val;
}

// node_list helper

mpv_node_list* create_node_list(mpv_format format, int size) {
    mpv_node_list* list = malloc(sizeof(* list));
    list->keys = NULL;
    list->num = size;
    list->values = calloc(size, sizeof(mpv_node));
    if (format == MPV_FORMAT_NODE_MAP) {
    	list->keys = calloc(size, sizeof(char *));
    }
    return list;
}


mpv_node * get_node_list_element(mpv_node_list * list,int index) {
    return &list->values[index];
}

void set_node_list_element(mpv_node_list * list,int index, mpv_node * node) {
    list->values[index] = *node;
}

char * get_node_list_key(mpv_node_list * list,int index) {
    return list->keys[index];
}
void set_node_list_key(mpv_node_list * list,int index, char * key) {
    list->keys[index] = malloc(strlen(key)+1);
    strcpy(list->keys[index],key);
}




mpv_node * create_node(mpv_format format) {
	mpv_node * node = malloc(sizeof(*node));
	node->format = format;
    return node;
}

// data pointer helper
int64_t * int64_t_ptr(int64_t val) {
    int64_t * ptr = malloc(sizeof(*ptr));
    *ptr = val;
    return ptr;
}

int * int_ptr(int val) {
    int * ptr = malloc(sizeof(*ptr));
    *ptr = val;
    return ptr;
}

double * double_ptr(double val) {
    double * ptr = malloc(sizeof(*ptr));
    *ptr = val;
    return ptr;
}