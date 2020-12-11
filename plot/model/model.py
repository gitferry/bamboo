import matplotlib.pyplot as plt

# Measurements from block-size.data
bsize = [
    ('Bamboo-SL',[
       (16.159, 10.15),
       (46.159, 10.76),
       (67.320, 12.25),
       (101.170, 16.63),
       (117.174, 21.69),
       (128.625, 26.85),
       (132.803, 30.55),
       (136.484, 36.5),
       (138.231, 45.44),
       (144.888, 51.7)
    ], '-o', 'coral'),
    ('Model-SL',[
        (16.519,5.53),
        (46.159,7.50),
        (67.32,9.86),
        (101.170,19.475),
        (128.625,26.33),
        (132.803,33.99),
        (136.484,44.74),
    ], '-s', 'steelblue')]

def do_plot():
    f = plt.figure(1, figsize=(7,5))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style, color in bsize:
        throughput = []
        latency = []
        for t, l in entries:
            throughput.append(t)
            latency.append(l)
        ax.plot(throughput, latency, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
    plt.legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylabel('Latency (ms)')
    plt.xlabel('Throughput (KTx/s)')
    plt.tight_layout()
    plt.ylim([0,60])
    plt.savefig('model-implementation.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
