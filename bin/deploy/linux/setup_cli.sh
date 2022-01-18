#!/usr/bin/env bash

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.client.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.client.ips_file' service_conf.json | sed 's/\"//g')

distribute(){
    for line in $(cat $DEPLOY_IPS_FILE)
    do
       ssh $DEPLOY_NAME@$line "mkdir $DEPLOY_FILE"
       echo "---- upload client: $DEPLOY_NAME@$line\n ----"
       scp client ips.txt config.json runClient.sh closeClient.sh $DEPLOY_NAME@$line:~/$DEPLOY_FILE
       ssh $DEPLOY_NAME@$line "chmod 777 ~/$DEPLOY_FILE/runClient.sh"
       ssh $DEPLOY_NAME@$line "chmod 777 ~/$DEPLOY_FILE/closeClient.sh"
    done
}

# distribute files
distribute
