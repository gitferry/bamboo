import matplotlib.pyplot as plt

# Measurements from batch-size.data
expt = [
    ('HS-100',[
        (15.680,6.21),
        (28.095,6.75),
        (52.799,7.84),
        (54.457,17.74),
        (54.378,21.32),
        (53.818,35.95)
    ], '-o', 'coral'),
    ('HS-400',[
        (15.637,6.25),
        (42.262,6.83),
        (72.415,7.8),
        (108.774,10.06),
        (129.770,16.8),
        (126.022,25.2),
        (123.610,34.7)
    ], '-^', 'coral'),
    ('HS-800',[
        (15.266,6.29),
        (41.880,6.89),
        (86.973,8.52),
        (119.637,11.82),
        (139.025,19.2),
        (135.160,26.2),
        (130.272,34.7)
    ], '-*', 'coral'),
    ('2CHS-100',[
        (19.630,4.85),
        (36.824,5.16),
        (54.214,7.08),
        (54.325,14.24),
        (53.214,21.85),
        (53.057,36.31),
    ], '-p', 'darkseagreen'),
    ('2CHS-400',[
        (19.670,4.83),
        (52.919,5.36),
        (86.277,6.44),
        (122.119,8.76),
        (130.661,16.7),
        (127.120,21.5),
        (123.102,33.5),
    ], '-v', 'darkseagreen'),
    ('2CHS-800',[
        (20.465,4.69),
        (52.999,5.31),
        (86.609,6.33),
        (122.310,8.72),
        (129.505,16.7),
        (128.118,22.6),
        (123.155,34.6)
    ], '-d', 'darkseagreen'),
    ('SL-100',[
        (15.976,6.0),
        (29.575,6.16),
        (43.034,6.37),
        (44.482,12.92),
        (46.037,21.25),
        (45.859,36.46)
    ], '-h', 'steelblue'),
    ('SL-400',[
        (15.899,6.0),
        (45.450,6.38),
        (78.615,7.19),
        (110.851,9.85),
        (114.631,19.6),
        (108.519,25.5),
        (114.570,35.47)
    ], '-s', 'steelblue'),
    ('SL-800',[
        (16.527,5.82),
        (45.058,6.4),
        (80.645,7.05),
        (111.050,9.95),
        (128.599,17.1),
        (136.775,20.0),
        (135.536,26.1),
        (132.186,34.8)
    ], '->', 'steelblue')]



def do_plot():
    f = plt.figure(1, figsize=(7,5))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style, color in expt:
        throughput = []
        latency = []
        for t, l in entries:
            # batch.append(N*ToverN)
            # throughput.append(ToverN*(N-t) / latency)
            throughput.append(t)
            latency.append(l)
        ax.plot(throughput, latency, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
    #ax.set_xscale("log")
    # ax.set_yscale("log")
    # plt.ylim([0, 50])
    #plt.xlim([10**3.8, 10**6.4])
    plt.legend(loc='upper left')
    # plt.ylabel('Throughput (Tx per second) in log scale')
    plt.ylabel('Latency (ms)')
    plt.xlabel('Throughput (KTx/s)')
    # plt.xlabel('Requests (Tx) in log scale')
    plt.tight_layout()
#     plt.show()
    plt.savefig('batch-size.pdf', format='pdf', dpi=400)

if __name__ == '__main__':
    do_plot()
