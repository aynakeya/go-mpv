#include <mpv/client.h>

// node helpers, since cgo don't have direct access to c union
char * get_node_string(mpv_node * node);
void set_node_string(mpv_node * node,char * val);

int get_node_flag(mpv_node * node);
void set_node_flag(mpv_node * node,int val);

int64_t get_node_int64(mpv_node * node);
void set_node_int64(mpv_node * node,int64_t val);

double get_node_double(mpv_node * node);
void set_node_double(mpv_node * node,double val);

mpv_node_list * get_node_list(mpv_node * node);
void set_node_list(mpv_node * node,mpv_node_list * val);

mpv_byte_array * get_node_byte_array(mpv_node * node);
void set_node_byte_array(mpv_node * node,mpv_byte_array * val);

mpv_node * create_node(mpv_format format);

// node list helpers

mpv_node_list* create_node_list(mpv_format format, int size);

mpv_node * get_node_list_element(mpv_node_list * list,int index);
void set_node_list_element(mpv_node_list * list,int index, mpv_node * node);

char * get_node_list_key(mpv_node_list * list,int index);
void set_node_list_key(mpv_node_list * list,int index, char * key);


// data pointer helper
int64_t * int64_t_ptr(int64_t val);
int * int_ptr(int val);
double * double_ptr(double val);

