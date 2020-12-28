import matplotlib.pyplot as plt

# Measurements from block-size.data
bsize = [
    ('Bamboo-HS',[
       (2.0, 13.32),
       (6.0, 12.3),
       (12.0, 12.09),
       (24.0, 13.2),
       (48.0, 15.2),
       (78.8, 19.17),
       (94.5, 23.59),
       (108.5, 34.59),
       (110.8, 43.2),
    ], '-o', 'coral'),
    ('Bamboo-2CHS',[
       (2.0, 12.10),
       (6.0, 10.95),
       (12.0, 11.62),
       (24.0, 12.05),
       (47.9, 12.79),
       (79.0, 17.82),
       (98.0, 19.92),
       (110.9, 29.0),
       (109.9, 48.89),
    ], '-s', 'darkseagreen'),
    ('Bamboo-SL',[
       (2.0, 12.26),
       (6.0, 12.06),
       (12.0, 11.38),
       (24.0, 11.63),
       (48.0, 12.4),
       (80.0, 14.5),
       (97.4, 19.2),
       (110.0, 29.5),
       (109.9, 46.9),
    ], '-d', 'steelblue'),
    ('Model-HS',[
       (2.0, 13.23),
       (6.0, 11.8),
       (12.0, 11.9),
       (24.0, 12.98),
       (48.0, 15.19),
       (79.0, 19.44),
       (94.5, 23.19),
       (108.5, 30.7),
       (110.0, 35.14),
    ], '--<', 'coral'),
    ('Model-2CHS',[
       (2.0, 11.17),
       (6.0, 10.14),
       (12.0, 10.29),
       (24.0, 11.33),
       (48.0, 13.06),
       (79.0, 17.2),
       (98.0, 20.12),
       (110.9, 29.32),
       (109.0, 42.32),
    ], '--^', 'darkseagreen'),
    ('Model-SL',[
       (2.0, 12.26),
       (6.0, 12.06),
       (12.0, 11.38),
       (24.0, 11.63),
       (48.0, 12.4),
       (80.0, 14.5),
       (97.4, 19.2),
       (110.0, 29.5),
       (109.9, 46.9),
    ], '-->', 'steelblue')]

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
    plt.ylim([10,40])
    plt.savefig('model-implementation.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
