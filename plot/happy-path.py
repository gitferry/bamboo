import matplotlib.pyplot as plt

# Measurements from happy-path.data
expt = [
    ('HotStuff',[
        (14917,11.54),
        (41649,12.6),
        (62075,14.15),
        (94362,18.69),
        (112436,23.72),
        (124599,28.59),
        (129521,33.79),
        (135073,39.175),
        (140052,48.7),
        (142850,59.3)
    ], '-o'),
    ('2C-HS',[
        (17462,9.6),
        (46540,10.8),
        (69698,12.2),
        (101286,17),
        (113162,22.8),
        (127463,27.4),
        (132674,31.5),
        (136262,37),
        (139196,46.3),
        (142981,57.5)
    ], '--+'),
    ('Streamlet',[
        (16159,10.15),
        (46159,10.76),
        (67320,12.25),
        (101170,16.63),
        (117174,21.69),
        (128625,26.85),
        (132803,30.55),
        (136484,36.5),
        (138231,45.44),
        (144888,51.7)
    ], '-*'),
    ('Origin-HS',[
        (17966,12.14),
        (18966,12.52),
        (131544,13.07),
        (141544,14.07),
        (151544,15.07),
        (169542,18.3),
        (172564,22.4),
        (176649,37.4),
        (176851,48.4)
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
    # ax.set_yscale("log")
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