import matplotlib.pyplot as plt

# Measurements from block-size.data
bsize = [
    ('Bamboo-HS',[
       (2.0, 15.02),
       (6.0, 13.3),
       (12.0, 13.09),
       (24.0, 14.6),
       (47.9, 16.5),
       (78.8, 18.97),
       (94.5, 23.59),
       (108.5, 34.59),
       (110.8, 50.7),
    ], '-o', 'coral'),
    ('Bamboo-2CHS',[
       (2.0, 12.10),
       (6.0, 10.95),
       (12.0, 11.62),
       (24.0, 12.05),
       (47.9, 12.79),
       (79.8, 16.82),
       (98.0, 19.52),
       (110.9, 29.0),
       (109.9, 48.89),
    ], '-^', 'sandybrown'),
    ('Bamboo-SL',[
       (2.0, 12.12),
       (6.0, 11.2),
       (12.0, 11.15),
       (24.0, 11.57),
       (47.9, 12.4),
       (79.8, 15.14),
       (98.0, 19.02),
       (110.3, 29.9),
       (109.9, 52.89),
    ], '-<', 'indianred'),
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
    plt.ylim([0,70])
    plt.savefig('model-implementation.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
