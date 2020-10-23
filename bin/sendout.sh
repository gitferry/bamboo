#!/usr/bin/env zsh
    echo "sending exutables into remote sites"
    go build ../server/
    go build ../client/
    scp server client gaify@sp1:~/github/bamboo/
    scp server client gaify@sp2:~/github/bamboo/
    scp server client gaify@sp3:~/github/bamboo/
    scp server client gaify@sp4:~/github/bamboo/
    scp server client gaify@sp5:~/github/bamboo/
    echo "Servers are already started in this folder."
