import matplotlib.pyplot as plt

SMALL_SIZE = 8
MEDIUM_SIZE = 10
BIGGER_SIZE = 16

plt.rc('font', size=BIGGER_SIZE)          # controls default text sizes
plt.rc('axes', titlesize=BIGGER_SIZE)     # fontsize of the axes title
plt.rc('axes', labelsize=BIGGER_SIZE)    # fontsize of the x and y labels
plt.rc('xtick', labelsize=BIGGER_SIZE)    # fontsize of the tick labels
plt.rc('ytick', labelsize=BIGGER_SIZE)    # fontsize of the tick labels
plt.rc('legend', fontsize=BIGGER_SIZE)    # legend fontsize

# Measurements from delays.data
delays = [
    ('HS-d0',[
        (20.0,12.0),
        (59.9,15.5),
        (79.8,18.5),
        (99.5,25.4),
        (109.4,38.4),
        (113.0,50.1),
        (115.8,88.5)
    ], '-o', 'coral'),
    ('HS-d5',[
        (2.356, 69.8),
        (7.198, 68.8),
        (13.753, 71.7),
        (21.926, 88.9),
        (28.862, 181.5),
        (30.844, 221.1),
        (31.374, 319.1)
    ], '-o', 'coral'),
    ('HS-d10',[
        (1.272, 129.1),
        (3.772, 130.5),
        (7.522, 133.4),
        (11.614, 165.7),
        (16.68, 459.5),
        (18.1, 550.1)
    ], '-^', 'coral'),
    ('2CHS-d0',[
        (20.0,10.3),
        (59.9,13.2),
        (99.5,22.8),
        (108.3,35.4),
        (114.1,56.1),
        (115.8,80.5),
    ], '-p', 'darkseagreen'),
    ('2CHS-d5',[
        (2.766, 57.3),
        (8.418, 56.4),
        (15.714, 59.9),
        (23.140, 82.6),
        (30.152, 168.1),
        (31.54, 320.1)
    ], '-p', 'darkseagreen'),
    ('2CHS-d10',[
        (1.48, 107.2),
        (4.366, 108.2),
        (12.646, 150.3),
        (16.698, 446.1),
        (17.816, 537)
    ], '-v', 'darkseagreen'),
    ('SL-d0',[
        (20.0,13.3),
        (40.0,15.1),
        (59.9,19.2),
        (79.9,30.52),
        (89.8,45.88),
        (91.8,68.67),
    ], '-s', 'steelblue'),
    ('SL-d5',[
        (2.0,62.1),
        (6.0,61),
        (11.9,63.6),
        (19.9,76.9),
        (29.9,197),
        (31.5,344),
    ], '-s', 'steelblue'),
    ('SL-d10',[
        (1.97,111.7),
        (5.9,112.7),
        (12.0,141),
        (16.0,235),
        (17.9,550)
    ], '->', 'steelblue')]

def do_plot():
    f = plt.figure(1, figsize=(8,6))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style, color in delays:
        throughput = []
        latency = []
        for t, l in entries:
            throughput.append(t)
            latency.append(l)
        ax.plot(throughput, latency, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
    plt.legend(loc='upper right', fancybox=True,frameon=False,framealpha=0.8,ncol=2)
    plt.grid(linestyle='--', alpha=0.3)
#     plt.ylim([0,700])
    plt.yscale('log')
    plt.ylabel('Latency (ms)')
    plt.xlabel('Throughput (KTx/s)')
    plt.savefig('delays.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
