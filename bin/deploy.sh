#!/usr/bin/env zsh

N=4

SERVER_PID_FILE=server.pid

SERVER_PID=$(cat "${SERVER_PID_FILE}");

if [ -z "${SERVER_PID}" ]; then
    echo "Process id for servers is written to location: {$SERVER_PID_FILE}"
    go build ../server/
    int=1
    while (( $int<=$N ))
    do
	    ./server -id $int -log_dir=. -log_level=debug -algorithm=hotstuff &
	    echo $! >> ${SERVER_PID_FILE}
	    let "int++"
    done
else
    echo "Servers are already started in this folder."
    exit 0
fi
