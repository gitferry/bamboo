#!/usr/bin/env bash

construct_node1_and_frigate(){
	rm -rf ${DUMP_PATH}/node1
    mkdir -p ${DUMP_PATH}/node1
    cp -rf  ${DUMP_PATH}/node/* ${DUMP_PATH}/node1/

	rm -rf test.tar.gz
    mkdir -p test
    cp -rf  ./frigate/* ./test/

	# init dynamic.hosts, ns_dynamic.nodes and frigate.hosts
	> python ./hosts.py
	tar czf test.tar.gz test
	rm -rf test
}


f_distribute(){

	allIPs=(`cat ips.txt`)
	num=${#allIPs[@]}
	echo "The number of nodes is ${num}"
	for (( j=1; j<=num; j++ ))
	do  
		if ((j>1)); then
		    rm -rf ${DUMP_PATH}/node${j}
		    mkdir -p ${DUMP_PATH}/node${j}
		    cp -rf  ${DUMP_PATH}/node1/* ${DUMP_PATH}/node${j}/
		fi
        
        # NOTE!!!s ""
		if ((j==1)); then
			sed -i ""    "s/domain1 127.0.0.1/domain1 ${allIPs[j-1]}/g" ${DUMP_PATH}/node${j}/configuration/dynamic.toml   #将dynamic.toml中selfaddr的ip,改为自己的ip.
		else
			sed -i ""    "s/domain1 ${allIPs[0]}/domain1 ${allIPs[j-1]}/g" ${DUMP_PATH}/node${j}/configuration/dynamic.toml #将dynamic.toml中selfaddr的ip,改为自己的ip.
	        sed -i "" "s/self = \"node1\"/self = \"node${j}\"/g" ${DUMP_PATH}/node${j}/configuration/dynamic.toml #将dynamic.toml中的self = "node1"为自己的编号
			sed -i ""  "s/hostname =  \"node1\"/hostname = \"node${j}\"/g" ${DUMP_PATH}/node${j}/configuration/global/ns_dynamic.toml #将ns_dynamic.toml 中的self 中的hostname="node1"和你=4改为相应的内容
	    fi

	    sed -i ""  "s/n = 4/n = ${num}/g" ${DUMP_PATH}/node${j}/configuration/global/ns_dynamic.toml
	done

	echo "SUCCESS"
}


# output root dir
DUMP_PATH="./nodes"

construct_node1_and_frigate
f_distribute