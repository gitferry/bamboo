import matplotlib.pyplot as plt

SMALL_SIZE = 8
MEDIUM_SIZE = 13
BIGGER_SIZE = 16

plt.rc('font', size=BIGGER_SIZE)          # controls default text sizes
plt.rc('axes', titlesize=BIGGER_SIZE)     # fontsize of the axes title
plt.rc('axes', labelsize=BIGGER_SIZE)    # fontsize of the x and y labels
plt.rc('xtick', labelsize=BIGGER_SIZE)    # fontsize of the tick labels
plt.rc('ytick', labelsize=BIGGER_SIZE)    # fontsize of the tick labels
plt.rc('legend', fontsize=MEDIUM_SIZE)    # legend fontsize

bsize = [
    ('b16K-p128',[
        # (4,703),
        # (8,400),
        (12,300),
        (16,265),
        # (20,280),
        (24,350),
        (25,480),
        (27,650),
        (28,900),
        (29,1428),
    ], '-o', 'coral'),
    ('b64K-p128',[
        # (12,900),
        # (20,580),
        # (28,447),
        # (40,350),
        # (48,315),
        (56,296),
        (72,320),
        (76,330),
        (80,490),
        (81,679),
        (80,1262),
    ], '-^', 'coral'),
    ('b128K-p128',[
        # (28,800),
        # (36,640),
        # (44,540),
        # (52,480),
        # (60,430),
        (68,400),
        (76,380),
        (80,370),
        (84,410),
        (85,612),
        (85,1302),
    ], '-*', 'coral'),
    ('b16K-p512',[
        # (4,380),
        (8,250),
        (12,285),
        (15,634),
        (15,746),
        (14,1582),
    ], '--p', 'steelblue'),
    ('b64K-p512',[
        # (8,653),
        # (16,378),
        (24,295),
        (32,275),
        (40,320),
        (46,500),
        (45,1032),
    ], '--v', 'steelblue'),
    ('b128K-p512',[
        # (12,843),
        # (20,540),
        # (28,423),
        (36,368),
        # (44,320),
        (52,432),
        (51,872),
        (52,1123)
    ], '--d', 'steelblue')]



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
    plt.legend(fancybox=True,frameon=False,framealpha=0.8,mode={"expand", None},ncol=3, loc='best')
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylim([0,2000])
    plt.ylabel('Latency (ms)')
    plt.xlabel('Throughput (KTx/s)')
    plt.tight_layout()
    plt.savefig('batch-size.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
