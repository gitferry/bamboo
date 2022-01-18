#!/usr/bin/env bash

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.client.ips_file' service_conf.json | sed 's/\"//g')

kill_all_clients(){
    echo "Killing clients"
    for line in $(cat $DEPLOY_IPS_FILE)
    do
       ssh $DEPLOY_NAME@$line "pkill client"&
    done
    wait
    echo "All clients are stopped"
}

# distribute files
kill_all_clients
