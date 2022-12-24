#include "runtime.c"

extern const cz_method_id_t czm_call;
extern const cz_method_id_t czm_bump;

// (f,g,x) -> f.bump(g.call(x), x)
//
// -->
//
// r0 = g.call(x)
// return f.bump(r0, x)
//
void czg_compose(cz_proc_t *p) {

    cz_value_t *kcode = p->frame;
    cz_value_t *r0    = p->frame + 1;
    cz_value_t f      = p->frame[2];
    cz_value_t g      = p->frame[3];
    cz_value_t x      = p->frame[4];

    switch(cz_int_val(*kcode)) {

    case 0:
        *kcode = cz_int(1);

        cz_prepare_call(p, 1);
        p->frame[0] = x;
        cz_call(p, g, czm_call);
        break;

    case 1:
        *r0 = p->ret;

        cz_prepare_tail_call(p, 2);
        p->frame[0] = *r0;
        p->frame[1] = x;
        cz_tail_call(p, f, czm_bump);
        break;
       
    }

}

extern const cz_class_id_t czc_example;

// (y) -> {call() -> y}
void czg_construct(cz_proc_t *p) {

    cz_value_t y = p->frame[0];

    cz_value_t r0 = cz_alloc_object(p, czc_example);
    
    cz_value_t *fields = cz_object_fields(r0);
    fields[0] = y;

    return r0;
}

