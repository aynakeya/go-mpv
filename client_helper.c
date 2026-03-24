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

const void * get_byte_array_data(mpv_byte_array * ba) {
    return ba ? ba->data : NULL;
}

size_t get_byte_array_size(mpv_byte_array * ba) {
    return ba ? ba->size : 0;
}

mpv_byte_array * create_byte_array(const void *data, size_t size) {
    mpv_byte_array *ba = malloc(sizeof(*ba));
    if (!ba) {
        return NULL;
    }
    ba->size = size;
    ba->data = NULL;
    if (size == 0) {
        return ba;
    }
    ba->data = malloc(size);
    if (!ba->data) {
        free(ba);
        return NULL;
    }
    if (data) {
        memcpy(ba->data, data, size);
    } else {
        memset(ba->data, 0, size);
    }
    return ba;
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

static void free_node_list(mpv_node_list *list, int is_map);

void free_node(mpv_node *node) {
    if (!node) {
        return;
    }
    switch (node->format) {
    case MPV_FORMAT_STRING:
        free((void *)node->u.string);
        break;
    case MPV_FORMAT_NODE_ARRAY:
        free_node_list((mpv_node_list *)node->u.list, 0);
        break;
    case MPV_FORMAT_NODE_MAP:
        free_node_list((mpv_node_list *)node->u.list, 1);
        break;
    case MPV_FORMAT_BYTE_ARRAY: {
        mpv_byte_array *ba = (mpv_byte_array *)node->u.ba;
        if (ba) {
            free(ba->data);
            free(ba);
        }
        break;
    }
    case MPV_FORMAT_NONE:
    case MPV_FORMAT_FLAG:
    case MPV_FORMAT_INT64:
    case MPV_FORMAT_DOUBLE:
    default:
        break;
    }
    node->format = MPV_FORMAT_NONE;
    node->u.string = NULL;
}

static void free_node_list(mpv_node_list *list, int is_map) {
    if (!list) {
        return;
    }
    for (int i = 0; i < list->num; i++) {
        free_node(&list->values[i]);
        if (is_map && list->keys) {
            free(list->keys[i]);
        }
    }
    free(list->values);
    free(list->keys);
    free(list);
}
