#ifndef CZ_H
#define CZ_H

#include <stdint.h>

/* Type declarations */

typedef uintptr_t cz_value_t;

typedef struct {

    void (*impl)();
    int k;
    int data_pos;

} cz_call_stack_entry_t;

typedef struct {

    int id;
    int fieldc; 

} cz_class_t;

typedef struct {

    int method_id;
    void (*impl)();

} cz_impl_t;

/* Runtime library */

extern cz_value_t cz_alloc(const cz_class_t *c);
extern int cz_class_id_of(cz_value_t object);

/* Global state */

extern cz_value_t cz_value;
extern int cz_stack_pos;
extern int cz_data_pos;
extern cz_value_t *cz_data_stack;
extern cz_call_stack_entry_t *cz_call_stack;

extern const cz_class_t cz_classes[];
extern const cz_impl_t cz_impls[];

/* ASM */

#define CZ_PROLOG() switch (cz_call_stack[cz_stack_pos].k) { case 0:
#define CZ_EPILOG() }

#define CZ_RETURN() do {                                            \
    cz_stack_pos--;                                                 \
    return;                                                         \
} while(0)

#define CZ_CALL(f, p) do {                                          \
    cz_call_stack[cz_stack_pos].k = __LINE__;                       \
    cz_stack_pos++;                                                 \
    cz_call_stack[cz_stack_pos].k = 0;                              \
    cz_call_stack[cz_stack_pos].impl = (f);                         \
    cz_call_stack[cz_stack_pos].data_pos = cz_data_pos + (p);       \
    return;                                                         \
    case __LINE__:                                                  \
} while(0) 

#define CZ_CREATE(id, p) cz_value = cz_alloc(cz_classes + (id))

#define CZ_METHOD_LOOKUP(id) cz_call_stack[cz_stack_pos].impl = cz_impls[(id) + cz_class_id_of(cz_value)].impl

#endif