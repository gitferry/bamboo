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
       scp server ips.txt $2@${SERVER_ADDR[j-1]}:/home/$2/bamboo
    done
}

USERNAME="gaify"
PASSWD="GaiFY#1"
FIRST=true
MAXPEERNUM=(`wc -l public_ips.txt | awk '{ print $1 }'`)

if ${FIRST}; then
    add_ssh_key $MAXPEERNUM $USERNAME $PASSWD
fi

# distribute files
distribute $MAXPEERNUM $USERNAME
