#!/usr/bin/env bash

start(){
    SERVER_ADDR=(`cat ips.txt`)
    for (( j=1; j<=$1; j++))
    do
      ssh -t $2@${SERVER_ADDR[j-1]} 'cd ~/bamboo ; ./run.sh $j'
    done
}

USERNAME="gaify"
ALGORITHM="hotstuff"
MAXPEERNUM=(`wc -l ips.txt | awk '{ print $1 }'`)

# update config.json to replicas
start $MAXPEERNUM $USERNAME
