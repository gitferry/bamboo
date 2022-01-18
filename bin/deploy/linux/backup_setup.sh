#!/usr/bin/env bash

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.server.deploy_file' service_conf.json | sed 's/\"//g')

rand(){
        echo $(( $RANDOM % 10 ))
}

update(){
    for line in $(cat $DEPLOY_IPS_FILE)
    do
	    {
	    rnd=$(rand)
	    if [ ${rnd} -gt 5 ]
		then
		bw=10
		else
		bw=30
	    fi
	    echo "setting ${bw}Mbps to ${line}"
	    ssh $DEPLOY_NAME@$line "tc qdisc add dev eth0 root handle 1:0 htb default 1; tc class add dev eth0 parent 1:0 classid 1:1 htb rate ${bw}Mbps; tc qdisc add dev eth0 parent 1:1 handle 2:0 netem delay 20ms"
    }&
    done
    wait
}

# update config.json to replicas
update
