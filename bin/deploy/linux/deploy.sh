#!/usr/bin/env bash
DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.server.deploy_file' service_conf.json | sed 's/\"//g')

./pkill_server.sh
distribute(){
    for line in $(cat $DEPLOY_IPS_FILE)
    do
      ssh $DEPLOY_NAME@$line "mkdir ~/$DEPLOY_FILE"&
    done
    wait
    echo "Uploading binaries..."
    for line in $(cat $DEPLOY_IPS_FILE)
    do
      scp server closeServer.sh $DEPLOY_NAME@$line:~/$DEPLOY_FILE&
    done
    wait
    echo "Upload success!"
}

# distribute files
distribute
