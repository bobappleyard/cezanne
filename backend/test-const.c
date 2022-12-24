#include <stdio.h>

extern const int test;

int main(int argc, char **argv) {
    printf("test value: %d\n", test);
    return 0;
}
