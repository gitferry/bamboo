import matplotlib.pyplot as plt

# Measurements from happy-path.data
expt = [
    ('HotStuff',[
        # (512,5.96925),
        (1000,12.46),
        (4944,13.28),
        (8677,13.14),
        (12047,15.17),
        (14775,19.53),
        (16585,57.06),
	    (16684,96.41),
        (17244,233.72)
        # (262144, 295.122)
    ], '-o'),
    ('2C-HS',[
        (1000,10.56),
        (4934,10.74),
        (8640,11.13),
        (12017,12.80),
        (14900,16.05),
        (16833,32),
        (17527,185.78)
    ], '--+'),
    ('Streamlet',[
        (1000,10.79),
        (4947,11.08),
        (8625,11.61),
        (11952,12.74),
        (14721,16.84),
	    (16256,25.19),
        (16817,119.03),
	    (17034,149.32)
    ], '-*')]



def do_plot():
    f = plt.figure(1, figsize=(7,5));
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style in expt:
        throughput = []
        latency = []
        for t, l in entries:
            # batch.append(N*ToverN)
            # throughput.append(ToverN*(N-t) / latency)
            throughput.append(t)
            latency.append(l)
        ax.plot(throughput, latency, style, label='%s' % name)
    #ax.set_xscale("log")
    ax.set_yscale("log")
    # plt.ylim([0, 50])
    #plt.xlim([10**3.8, 10**6.4])
    plt.legend(loc='upper left')
    # plt.ylabel('Throughput (Tx per second) in log scale')
    plt.ylabel('Latency (ms)')
    plt.xlabel('Throughput (txn/s)')
    # plt.xlabel('Requests (Tx) in log scale')
    plt.tight_layout()
    # plt.show()
    plt.savefig('happy-path.png', format='png', dpi=100)

if __name__ == '__main__':
    do_plot()