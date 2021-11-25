#!/usr/bin/env bash
sh ./pkill.sh

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.server.ips_file' service_conf.json | sed 's/\"//g')

start(){
  count=0
  for line in $(cat $DEPLOY_IPS_FILE)
  do
    count=$((count+1))
    ssh -t $DEPLOY_NAME@$line "cd ~/$DEPLOY_FILE ; nohup ./run.sh $count"
    sleep 0.1
    echo replica $count is launched!
  done
}

# update config.json to replicas
start
