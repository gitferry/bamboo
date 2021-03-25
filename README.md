## What is Bamboo?

**Bamboo** is a prototyping and evaluation framework that studies the next generation BFT (Byzantine fault-tolerant) protocols specific for blockchains, namely chained-BFT, or cBFT.
By leveraging Bamboo, developers can prototype a brand new cBFT protocol in around 300 LoC and evaluate using rich benchmark facilities.

Bamboo is designed based on an observation that the core of cBFT protocols can be abstracted into 4 rules: **Proposing**, **Voting**, **State Updating**, and **Commit**.
Therefore, Bamboo abstracts the 4 rules into a *Safety* module and provides implementations of the rest of the components that can be shared across cBFT protocols, leaving the safety module to be specified by developers.

*Warning*: **Bamboo** is still under heavy development, with more features and protocols to include.

Bamboo details can be found in this [technical report](https://arxiv.org/abs/2103.00777). The paper is to appear at [ICDCS 2021](https://icdcs2021.us/).

## What is cBFT?
At a high level, cBFT protocols share a unifying *propose-vote* paradigm in which they assign transactions coming from the clients a unique order in the global ledger.
A blockchain is a sequence of blocks cryptographically linked together by hashes.
Each block in a blockchain contains a hash of its parent block along with a batch of transactions and other metadata.  

Similar to classic BFT protocols, cBFT protocols are driven by leader nodes and operate in a view-by-view manner.
Each participant takes actions on receipt of messages according to four protocol-specific rules: **Proposing**, **Voting**, **State Updating**, and **Commit**.
Each view has a designated leader chosen at random, which proposes a block according to the **Proposing** rule and populates the network.
On receiving a block, replicas take actions according to the **Voting** rule and update their local state according to the **State Updating** rule.
For each view, replicas should certify the validity of the proposed block by forming a *Quorum Certificate* (or QC) for the block.
A block with a valid QC is considered certified.
The basic structure of a blockchain is depicted in the figure below.

![blockchain](https://github.com/gitferry/bamboo/blob/master/doc/propose-vote.jpeg?raw=true)

Forks happen because of conflicting blocks, which is a scenario in which two blocks do not extend each other.
Conflicting blocks might arise because of network delays or proposers deliberately ignoring the tail of the blockchain.
Replicas finalize a block whenever the block satisfies the **Commit** rule based on their local state.
Once a block is finalized, the entire prefix of the chain is also finalized. Rules dictate that all finalized blocks remain in a single chain.
Finalized blocks can be removed from memory to persistent storage for garbage collection.

## What is included?

Protocols:
- [x] [HotStuff and two-chain HotStuff](https://dl.acm.org/doi/10.1145/3293611.3331591)
- [x] [Streamlet](https://dl.acm.org/doi/10.1145/3419614.3423256)
- [x] [Fast-HotStuff](https://arxiv.org/abs/2010.11454)
- [ ] [LBFT](https://arxiv.org/abs/2012.01636)
- [ ] [SFT](https://arxiv.org/abs/2101.03715)

Features:
- [x] Benchmarking
- [x] Fault injection


# How to build

1. Install [Go](https://golang.org/dl/).

2. Download Bamboo source code.

3. Build `server` and `client`.
```
cd bamboo/bin
go build ../server
go build ../client
```

# How to run

Users can run Bamboo-based cBFT protocols in simulation (single process) or deployment.

## Simulation
In simulation mode, replicas are running in separate Goroutines and messages are passing via Go channel.
1. ```cd bamboo/bin```.
2. Modify `ips.txt` with a set of IPs of each node. The number of IPs equals to the number of nodes. Here, the local IP is `127.0.0.1`. Each node will be assigned by an increasing port from `8070`.
3. Modify configuration parameters in `config.json`.
4. Modify `simulation.sh` to specify the name of the protocol you are going to run.
5. Run `server` and then run `client` using scripts.
```
bash simulation.sh
```
```
bash runClient.sh
```
6. close the simulation by stopping the client and the server in order.
```
bash closeClient.sh
bash stop.sh
```
Logs are produced in the local directory with the name of `client/server.xxx.log` where `xxx` is the pid of the process.

## Deploy
Bamboo can be deployed in a real network.
1. ```cd bamboo/bin/deploy```.
2. Build `server` and `client`.
3. Specify external IPs and internal IPs of server nodes in `pub_ips.txt` and `ips.txt`, respectively.
4. IPs of machines running as clients are specified in `clients.txt`.
5. The type of the protocol is specified in `run.sh`.
6. Modify configuration parameters in `config.json`.
7. Modify `deploy.sh` and `setup_cli.sh` to specify the username and password for logging onto the server and client machines. 
8. Upload binaries and config files onto the remote machines.
```
bash deploy.sh
bash setup_cli.sh
```
9. Upload/Update config files onto the remote machines.
```
bash update_conf.sh
```
10. Start the server nodes.
```
bash start.sh
```
11. Log onto the client machine (assuming only one) via ssh and start the client.
```
bash ./runClient.sh
```
The number of concurrent clients can be specified in `runClient.sh`.
12. Stop the client and server.
```
bash ./closeClient.sh
bash ./pkill.sh
```

# Monitor
During each run, one can view the statistics (throughput, latency, view number, etc.) at a node via a browser.
```
http://127.0.0.1:8070/query
``` 
where `127.0.0.1:8070` can be replaced with the actual node address.
