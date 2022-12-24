Basic Approach
--------------

Intermediate form a bit like ANF:

    data IR = 
          Arg(Int)
        | StrLit(String)  
        | Return(Int)
        | Tail  (Var, String, [Int])
        | Invoke(Var, String, [Int])

Each instruction is referred to by its index to mean the result of executing that instruction.

1.  Convert the code to `[IR]`
2.  Assign a register for each non final step, along with a step to represent the position in the 
    function (`k`)
3.  Each step gets a case in a switch statement
4.  A non-initial step begins by assigning the current result value to the previous step's register
5.  Then it sets k to the next function position
6.  It then issues its code and breaks from the switch.

Example:

    def example(f, g, x) = f.bump(g.call(x), x)

    -->

        [
            Arg(1),                    // g
            Arg(2),                    // x
            Invoke(0, "call", [1]),
            Arg(0),                    // f
            Tail(3, "bump", [2, 1])
        ]

    -->

    void czg_example(cz_proc_t *p) {

        cz_value_t *k = p->frame + 0;
        cz_value_t *r0 = p->frame + 1;
        cz_value_t *r1 = p->frame + 2;
        cz_value_t *r2 = p->frame + 3;
        cz_value_t *r3 = p->frame + 4;

        switch(cz_value_to_int(*k)) {

        case 0:
            *k = cz_int_to_value(1);

            p->result = p->frame[6];
            break;

        case 1:
            *r0 = p->result;
            *k = cz_int_to_value(2);

            p->result = p->frame[7];
            break;

        case 2:
            *r1 = p->result;
            *k = cz_int_to_value(3);

            cz_prepare_call(p, *r0, czm_call);
            p->frame[0] = *r1;
            cz_call(p);
            break;

        case 3:
            *r1 = p->result;
            *k = cz_int_to_value(4);

            p->result = p->frame[5];
            break;

        case 4:
            *r3 = p->result;

            cz_prepare_tail_call(p, *r3, czm_bump);
            p->frame[0] = *r2;
            p->frame[1] = *r1;
            cz_tail_call(p);
            break;

        }

    }

This is a very simple approach that should work. It has many low-hanging optimisations, which I will
leave until it's running.

Objects
-------

No need for types at runtime, they are checked in the compiler. We do need to remember what methods
are available for which objects though. Everywhere we create an object we establish a class and
replace the object creation syntax with a call to a constructor function that we generate. That
constructor associates the created object with a class_id that can be used to locate methods.

When we compile a module, as well as creating a `.o` file via generated C, we store some metadata
about the objects and methods in that module (we can probably store some other stuff as well in
there, but that isn't relevant to this discussion). When we link a program together, we use this
metadata to generate a C file describing a method table. This method table is a sparse matrix where
each method is a column and each class is a row. Because it is sparse we can interleave the rows to
save space.
