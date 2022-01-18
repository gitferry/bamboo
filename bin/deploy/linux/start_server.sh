#!/usr/bin/env bash

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.server.ips_file' service_conf.json | sed 's/\"//g')

start(){
  echo "Updating configuration..."
  for line in $(cat $DEPLOY_IPS_FILE)
  do
    scp config.json run.sh ips.txt $DEPLOY_NAME@$line:~/$DEPLOY_FILE&
  done
  wait
  count=0
  for line in $(cat $DEPLOY_IPS_FILE)
  do
  {
    count=$((count+1))
    ssh -t $DEPLOY_NAME@$line "chmod 777 ~/$DEPLOY_FILE/run.sh; cd ~/$DEPLOY_FILE ; nohup ./run.sh $count"; sleep 0.1&
    echo "Replica $count is launched!"
  }
  done
  wait
}

# update config.json to replicas
start
