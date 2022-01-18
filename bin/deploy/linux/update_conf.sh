#!/usr/bin/env bash

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.server.deploy_file' service_conf.json | sed 's/\"//g')

update(){
    for line in $(cat $DEPLOY_IPS_FILE)
    do
       scp config.json run.sh ips.txt $DEPLOY_NAME@$line:~/$DEPLOY_FILE
       ssh $DEPLOY_NAME@$line "chmod 777 ~/$DEPLOY_FILE/run.sh"
    done
}

# update config.json to replicas
update
