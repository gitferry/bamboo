import matplotlib.pyplot as plt

# Measurements from happy-path.data
expt = [
    ('HotStuff',[
        # (512,5.96925),
        (1000,3.59),
        (4949,3.78),
        (8714,4.05),
        (12317,4.36),
        (15523,5.05),
        (18698,5.85),
        (20840,7.86),
        (21108,11.7),
        (21022,18.54),
        (20982,29)
        # (262144, 295.122)
    ], '-o'),
    ('2C-HS',[
        (1000,3.08),
        (4961,3.16),
        (8764,3.49),
        (12407,3.73),
        (15829,4.2),
        (18934,5.06),
        (21106,7.4),
        (21508,8.9),
        (21525,14.81),
        (20865,24.54),
        (20906,31.82)
    ], '--+'),
    ('Streamlet',[
        (1000,3.78),
        (4958,2.76),
        (8671,3.42),
        (12257,3.94),
        (15690,5.13),
        (18581,5.83),
        (21139,5.82),
        (21399,10.09),
        (19989,19.43),
        (19793,45.68)
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
    #ax.set_yscale("log")
    plt.ylim([0, 50])
    #plt.xlim([10**3.8, 10**6.4])
    plt.legend(loc='best')
    #plt.ylabel('Throughput (Tx per second) in log scale')
    plt.ylabel('Latency')
    plt.xlabel('Throughput')
    # plt.xlabel('Requests (Tx) in log scale')
    # plt.tight_layout()
    plt.show()
    # plt.savefig('happy-path.pdf', format='pdf', dpi=1000)

if __name__ == '__main__':
    do_plot()