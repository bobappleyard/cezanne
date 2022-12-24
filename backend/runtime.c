#include <stdint.h>
#include <string.h>

typedef uintptr_t cz_method_id_t;
typedef uintptr_t cz_class_id_t;

/**
*/
typedef union {

    struct cz_object_t *object;
    uintptr_t           integer;

} cz_value_t;

/**
*/
typedef struct {

    cz_class_id_t  class_id;
    cz_method_id_t method_id,
                   argc,
                   varc;
    void         (*impl)(struct cz_proc_t *p);

} cz_method_t;

/**
*/
typedef struct cz_object_t {

    cz_class_id_t class_id;
    unsigned char data[];

} cz_object_t;

/**
*/
typedef struct cz_proc_t {

    cz_value_t   ret,
                 recv,
                *frame,
                *stack;
    cz_method_t *method;

} cz_proc_t;

void cz_run(cz_proc_t *p) {

    while (p->method) {
        p->method->impl(p);
    }

}

void cz_stack_alloc(cz_proc_t *p, int n) {

    if ((p->stack + n) > p->frame) {
        cz_fatal("stack overflow");
    }

    p->frame -= n;
    memset(p->frame, 0, sizeof(cz_value_t) * n);

}

void cz_push(cz_proc_t *p, cz_value_t x) {

    cz_stack_alloc(p, 1);
    p->frame[0] = x;

}

void cz_prepare_call(cz_proc_t *p, int argc) {

    cz_push(p, p->recv);
    cz_push(p, cz_method_as_value(p->method));
    cz_stack_alloc(p, argc);

}

void cz_call(cz_proc_t *p, cz_value_t recv, cz_method_id_t method) {

    cz_method_t *m = cz_method_lookup(recv, method);

    cz_stack_alloc(p, m->varc);
    
    p->recv = recv;
    p->method = m;

}

void cz_prepare_tail_call(cz_proc_t *p, int argc) {

    cz_stack_alloc(p, argc);

}

void cz_tail_call(cz_proc_t *p, cz_value_t recv, cz_method_id_t method) {

    cz_method_t *m = cz_method_lookup(recv, method);

    memmove(p->frame, p->frame + p->method->varc, sizeof(cz_value_t) * m->argc);
    p->frame += p->method->varc;

    cz_stack_alloc(p, m->varc);

    p->recv = recv;
    p->method = m;

}

void cz_return(cz_proc_t *p, cz_value_t x) {

    p->frame += p->method->varc + p->method->argc;
    p->method = cz_value_as_method(p->frame[0]);
    p->recv = p->frame[1];
    p->frame += 2;
    
    p->ret = x;

}

cz_value_t cz_int(int x) {
    cz_value_t result;

    if (x < 0) {
        result.integer = ((-x) << 2) | 3;
        return result; 
    }

    result.integer = (x << 2) | 1;
    return result;
}

int cz_int_val(cz_value_t x) {

}

cz_value_t cz_alloc_object(cz_proc_t *p, cz_class_id_t class) {

}

cz_value_t *cz_object_fields(cz_value_t object) {
    
}