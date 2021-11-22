#!/usr/bin/env bash
sh ./pkill.sh

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.server.ips_file' service_conf.json | sed 's/\"//g')

distribute(){
    for line in $(cat $DEPLOY_IPS_FILE)
    do
      ssh $DEPLOY_NAME@$line "mkdir ~/$DEPLOY_FILE"
      echo "---- upload replica: $DEPLOY_NAME@$line \n ----"
      scp server $DEPLOY_NAME@$line:~/$DEPLOY_FILE
    done
}

# distribute files
distribute
