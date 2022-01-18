#!/usr/bin/env bash

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.server.deploy_file' service_conf.json | sed 's/\"//g')
DEPLOY_CLIENT_FILE=$(jq '.client.ips_file' service_conf.json | sed 's/\"//g')

kill_all_servers(){
    for line in $(cat $DEPLOY_IPS_FILE)
    do
       ssh -t -T $DEPLOY_NAME@$line "rm -rf ~/$DEPLOY_FILE/server.* ~/$DEPLOY_FILE/nohup.out ~/$DEPLOY_FILE/debug"&
    done
    wait
    for line in $(cat $DEPLOY_CLIENT_FILE)
    do
       ssh -t -T $DEPLOY_NAME@$line "rm -rf ~/$DEPLOY_FILE/client.* ~/$DEPLOY_FILE/nohup.out"&
    done
    wait
}

# distribute files
kill_all_servers
