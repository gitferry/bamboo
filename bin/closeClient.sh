#!/usr/bin/env bash

PID_FILE=client.pid

if [ ! -f "${PID_FILE}" ]; then
    echo "No client is running."
else
	while read pid; do
		if [ -z "${pid}" ]; then
			echo "No client is running."
		else
			kill -15 "${pid}"
			echo "Client with PID ${pid} shutdown."
    	fi
	done < "${PID_FILE}"
	rm "${PID_FILE}"
fi
