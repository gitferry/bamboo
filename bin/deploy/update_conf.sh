#!/usr/bin/env bash

update(){
    SERVER_ADDR=(`cat public_ips.txt`)
    for (( j=1; j<=$1; j++))
    do
       scp config.json run.sh ips.txt $2@${SERVER_ADDR[j-1]}:/home/$2/bamboo
       ssh -t $2@${SERVER_ADDR[j-1]} 'chmod 777 ~/bamboo/run.sh'
    done
}

USERNAME="gaify"
MAXPEERNUM=(`wc -l public_ips.txt | awk '{ print $1 }'`)

# update config.json to replicas
update $MAXPEERNUM $USERNAME
