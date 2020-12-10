import matplotlib.pyplot as plt

# Measurements from happy-path.data
expt = [
    ('HotStuff',[
        (14.917,11.54),
        (41.649,12.6),
        (62.075,14.15),
        (94.362,18.69),
        (112.436,23.72),
        (124.599,28.59),
        (129.521,33.79),
        (135.073,39.175),
        (140.052,48.7),
        (142.850,59.3)
    ], '-o'),
    ('2C-HS',[
        (17.462,9.6),
        (46.540,10.8),
        (69.698,12.2),
        (101.286,17),
        (113.162,22.8),
        (127.463,27.4),
        (132.674,31.5),
        (136.262,37),
        (139.196,46.3),
        (142.981,57.5)
    ], '--+'),
    ('Streamlet',[
        (16.159,10.15),
        (46.59,10.76),
        (67.20,12.25),
        (101.170,16.63),
        (117.174,21.69),
        (128.625,26.85),
        (132.803,30.55),
        (136.484,36.5),
        (138.231,45.44),
        (144.888,51.7)
    ], '-*'),
    ('Origin-HS',[
        (17.966,12.14),
        (58.966,12.52),
        (131.544,13.07),
        (141.544,14.07),
        (151.544,15.07),
        (169.542,18.3),
        (172.564,22.4),
        (176.649,37.4),
        (176.851,48.4)
    ], '-s')]



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
#     ax.set_yscale("log")
    # plt.ylim([0, 50])
    #plt.xlim([10**3.8, 10**6.4])
    plt.legend(loc='upper left')
    # plt.ylabel('Throughput (Tx per second) in log scale')
    plt.ylabel('Latency (ms)')
    plt.xlabel('Throughput (KTx/s)')
    # plt.xlabel('Requests (Tx) in log scale')
    plt.tight_layout()
    # plt.show()
    plt.savefig('happy-path.pdf', format='pdf', dpi=400)

if __name__ == '__main__':
    do_plot()
