#!/usr/bin/env bash

go build ../client/
int=1
while (( $int<=$1 ))
do
./client&
let "int++"
done
echo "$1 clients are started"
