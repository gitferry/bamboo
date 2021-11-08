#!/usr/bin/env bash

# add ssh-key
add_ssh_key(){
	SERVER_ADDR=(`cat public_ips.txt`)
    echo "Add your local ssh public key into all nodes"
    for (( j=1; j<=$1; j++ ))
    do
            addkey ${SERVER_ADDR[j-1]} $2 $3
	    wait
    done
}

distribute(){
    SERVER_ADDR=(`cat public_ips.txt`)
    for (( j=1; j<=$1; j++))
    do 
       ssh -t $2@${SERVER_ADDR[j-1]} mkdir bamboo
       echo -e "---- upload replica ${j}: $2@${SERVER_ADDR[j-1]} \n ----"
       scp server ips.txt $2@${SERVER_ADDR[j-1]}:/root/bamboo
    done
}

USERNAME='root'
MAXPEERNUM=(`wc -l public_ips.txt | awk '{ print $1 }'`)

# distribute files
distribute $MAXPEERNUM $USERNAME
