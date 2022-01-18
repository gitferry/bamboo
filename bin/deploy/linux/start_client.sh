#!/usr/bin/env bash

DEPLOY_NAME=$(jq '.auth.user' service_conf.json | sed 's/\"//g')
DEPLOY_FILE=$(jq '.server.dir' service_conf.json | sed 's/\"//g')
DEPLOY_IPS_FILE=$(jq '.client.ips_file' service_conf.json | sed 's/\"//g')

echo "Updating configuration..."
for line in $(cat $DEPLOY_IPS_FILE)
do
  scp config.json ips.txt runClient.sh $DEPLOY_NAME@$line:~/$DEPLOY_FILE&
done
wait
echo "Starting $1 clients..."
for line in $(cat $DEPLOY_IPS_FILE)
do
  ssh -t $DEPLOY_NAME@$line "chmod 777 ~/$DEPLOY_FILE/runClient.sh; cd ~/$DEPLOY_FILE ; nohup ./runClient.sh $1"; sleep 0.1&
done
wait
echo "$1 clients are started"
