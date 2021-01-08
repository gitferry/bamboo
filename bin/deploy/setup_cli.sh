#!/usr/bin/env bash

addkey(){
    expect <<EOF
        set timeout 60
        spawn ssh-copy-id $2@$1
        expect {
            "yes/no" {send "yes\r";exp_continue }
            "password:" {send "$3\r";exp_continue }
            eof
        }
EOF
}

# add ssh-key
add_ssh_key(){
	SERVER_ADDR=(`cat clients.txt`)
    echo "Add your local ssh public key into all nodes"
    for (( j=1; j<=$1; j++ ))
    do
            addkey ${SERVER_ADDR[j-1]} $2 $3
	    wait
    done
}

distribute(){
    SERVER_ADDR=(`cat clients.txt`)
    for (( j=1; j<=$1; j++))
    do
       ssh -t $2@${SERVER_ADDR[j-1]} mkdir bamboo
       echo -e "---- upload client ${j}: $2@${SERVER_ADDR[j-1]} \n ----"
       scp client ips.txt config.json runClient.sh closeClient.sh $2@${SERVER_ADDR[j-1]}:/home/$2/bamboo
       ssh -t $2@${SERVER_ADDR[j-1]} chmod 777 /home/$2/bamboo/runClient.sh
       ssh -t $2@${SERVER_ADDR[j-1]} chmod 777 /home/$2/bamboo/closeClient.sh
       wait
    done
}

USERNAME="gaify"
PASSWD="GaiFY#1"
MAXPEERNUM=(`wc -l clients.txt | awk '{ print $1 }'`)
FIRST=true

if ${FIRST}; then
    add_ssh_key $MAXPEERNUM $USERNAME $PASSWD
fi

# distribute files
distribute $MAXPEERNUM $USERNAME
