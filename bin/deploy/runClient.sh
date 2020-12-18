#!/usr/bin/env bash

N=1

PID_FILE=client.pid

if [ -z "${PID}" ]; then
    echo "Process id for clients is written to location: {$PID_FILE}"
    int=1
    while (( $int<=$N ))
    do
    ./client&
    echo $! >> ${PID_FILE}
    let "int++"
    done
else
    echo "Clients are already started in this folder."
    exit 0
fi
