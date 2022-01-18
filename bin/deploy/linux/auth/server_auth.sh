#!/usr/bin/env bash

echo "================== basic info ================="

AUTH_NAME=$(jq '.user' server_auth.json | sed 's/\"//g')
AUTH_PASS=$(jq '.password' server_auth.json | sed 's/\"//g')

AUTH_PATH=~/.ssh/server_auth.pub

echo 'user: ' "$AUTH_NAME"
echo 'pass: ' "$AUTH_PASS"
echo 'path: ' "$AUTH_PATH"
echo "================== start auth ================="

for line in $(cat server_auth.txt)
do
	{
	echo 'auth: ' "$line"
  expect << __EOF
	spawn ssh-copy-id $AUTH_NAME@$line

	expect {
    "yes/no" {send "yes\r";exp_continue }
    "password:" {send "$AUTH_PASS\r";exp_continue }
    eof
  }
__EOF
}&
done
wait

echo "================== finished! ================="
