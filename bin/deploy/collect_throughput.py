
import os
path = "."
files= os.listdir(path)
s = []
throughputs = []
latencies = []
totalThroughput = 0.0
totalLatency = 0.0
fileNo = 0
for file in files:
     if not os.path.isdir(file):
         if file.endswith(".log") and file.startswith("client."):
             fileNo += 1
             f = open(path+"/"+file)
             for line in iter(f):
                 if "Throughput" in line:
                     throughput = line.strip().split(" ")[-1]
                     throughputs.append(throughput)
                     totalThroughput += float(throughput)
                 if "median" in line:
                     latency = line.strip().split(" ")[-1]
                     latencies.append(latency)
                     totalLatency += float(latency)
             f.close()

with open(path+'/'+str(fileNo)+".log", 'w') as w_f:
    w_f.write("Experiment with " + str(fileNo) +" clients:" + "\n")
    for i in range(fileNo):
        w_f.write("["+str(i)+"]"+" throughput: "+str(throughputs[i])+", latency: "+str(latencies[i])+"\n")

    w_f.write("Total throughput: "+str(totalThroughput))
    w_f.write("\nAverage latency: "+str(totalLatency/fileNo))
