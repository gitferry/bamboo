#!/usr/bin/env bash

distribute(){
    SERVER_ADDR=(`cat clients.txt`)
    for (( j=1; j<=$1; j++))
    do
       ssh -t $2@${SERVER_ADDR[j-1]} mkdir bamboo
       echo -e "---- upload client ${j}: $2@${SERVER_ADDR[j-1]} \n ----"
       scp client ips.txt config.json runClient.sh closeClient.sh $2@${SERVER_ADDR[j-1]}:/root/bamboo
       ssh -t $2@${SERVER_ADDR[j-1]} chmod 777 /root/bamboo/runClient.sh
       ssh -t $2@${SERVER_ADDR[j-1]} chmod 777 /root/bamboo/closeClient.sh
       wait
    done
}

USERNAME='root'
MAXPEERNUM=(`wc -l clients.txt | awk '{ print $1 }'`)

# distribute files
distribute $MAXPEERNUM $USERNAME
