#!/usr/bin/env bash

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.server.ips_file' service_conf.json | sed 's/\"//g')

kill_all_servers(){
    for line in $(cat $DEPLOY_IPS_FILE)
    do
       ssh -t $DEPLOY_NAME@$line "pkill server ; rm ~/$DEPLOY_FILE/server.pid"
    done
}

# distribute files
kill_all_servers
