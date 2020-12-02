#!/usr/bin/env bash

SERVER_PID_FILE=server.pid

if [ -z "${SERVER_PID}" ]; then
    nohup ./server -id $1 -log_dir=. -log_level=info -algorithm=hotstuff &
    sleep 0.1
    echo $! >> ${SERVER_PID_FILE}
else
    echo "Servers are already started in this folder."
    exit 0
fi
