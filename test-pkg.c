#include <cz.h>
#include <stdio.h>

extern const int cz_classes_test;

extern void cz_m_true();
extern void cz_m_false();

static cz_value_t cz_true, cz_false;

void cz_impl_test() {
    CZ_CREATE(cz_classes_test, 0);
    cz_true = cz_value;
    CZ_CREATE(cz_classes_test + 1, 0);
    cz_false = cz_value;
    CZ_RETURN();
}

void cz_impl_test_print() {
    printf("%lu\n", cz_value >> 2);
    CZ_RETURN();
}

void cz_impl_test_lte() {
    CZ_LOAD(0);
    int left = cz_value >> 2;
    CZ_LOAD(1);
    int right = cz_value >> 2;

    cz_value = left <= right ? cz_true : cz_false ;
    CZ_RETURN();
}

void cz_impl_test_sub() {
    CZ_LOAD(0);
    int left = cz_value >> 2;
    CZ_LOAD(1);
    int right = cz_value >> 2;

    CZ_INT(left - right);
    CZ_RETURN();
}

void cz_impl_test_mul() {
    CZ_LOAD(0);
    int left = cz_value >> 2;
    CZ_LOAD(1);
    int right = cz_value >> 2;

    CZ_INT(left * right);
    CZ_RETURN();
}

void cz_impl_test_0_match() {
    CZ_LOAD(0);
    CZ_CALL_TAIL(cz_m_true);
}

void cz_impl_test_1_match() {
    CZ_LOAD(0);
    CZ_CALL_TAIL(cz_m_false);
}
