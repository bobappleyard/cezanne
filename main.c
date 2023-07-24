#include "cz.h"
#include <stdlib.h>
#include <stdio.h>

cz_value_t cz_value;
int cz_stack_pos = 0;
int cz_data_pos = 0;
cz_value_t *cz_data_stack;
cz_call_stack_entry_t *cz_call_stack;

extern void cz_init();

extern cz_value_t cz_alloc(const cz_class_t *c) {
    cz_value_t *object = malloc((1 + c->fieldc) * sizeof(cz_value_t));
    object[0] = (cz_value_t) c->id;
    return (cz_value_t) object;
}

extern int cz_class_id_of(cz_value_t object) {
    return (int) *((cz_value_t*) object);
}

int main() {
    cz_data_stack = malloc(1024 * sizeof(cz_value_t));
    cz_call_stack = malloc(1024 * sizeof(cz_call_stack_entry_t));
    
    cz_call_stack[0].data_pos = 0;
    cz_call_stack[0].k = 0;
    cz_call_stack[0].impl = cz_init;

    while (cz_stack_pos >= 0) {
        cz_data_pos = cz_call_stack[cz_stack_pos].data_pos;
        cz_call_stack[cz_stack_pos].impl();
    }
}
