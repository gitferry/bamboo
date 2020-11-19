import statistics
import os
path = "."
s = []
durations = []
totalThroughput = 0.0
totalLatency = 0.0
f = open("server.30314.log")
counter = 0
for line in iter(f):
    if "millisecond" in line:
        for item in line.strip().split(" "):
            if "ms" in item:
                duration = float(item[:-2])
                durations.append(duration)
                counter += 1
    if counter == 1000000:
        break
        
f.close()
print("mean is:", statistics.mean(durations))
print("var is:", statistics.variance(durations))