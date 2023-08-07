#!/bin/bash

set -e 

CC="gcc"

# clean dir
rm -rf build/*

# test dependency
$CC -c -I. -Wall -Werror -o build/test.o test-pkg.c 
cp test-link.json build/link.json 
ar csr build/test.a build/test.o build/link.json 

# build our example main pkg
go run ./cmd/cz-compile -input . -name main -output build
$CC -c -I. -Wall -Werror -o build/pkg.o build/pkg.c
ar csr build/fac.a build/pkg.o build/link.json 

# link the program
go run ./cmd/cz-link -i build -o build/link.c -p fac
$CC -c -I. -Wall -Werror -o build/main.o main.c
$CC -c -I. -Wall -Werror -o build/link.o build/link.c

$CC -o test build/main.o build/link.o build/fac.a build/test.a
