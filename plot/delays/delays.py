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
    ('HS-d5',[
        (2.356, 69.8),
        (7.198, 68.8),
        (13.753, 71.7),
        (21.926, 88.9),
        (27.416, 142.3),
        (28.862, 181.5),
        (30.844, 221.1),
        (31.144, 252.1),
        (31.374, 319.1)
    ], '-o', 'coral'),
    ('HS-d10',[
        (1.272, 129.1),
        (3.772, 130.5),
        (7.522, 133.4),
        (11.614, 165.7),
        (15.480, 260.4),
        (15.628, 301.3),
        (16.286, 444.2),
        (16.68, 459.5),
        (17.2, 550.1)
    ], '-^', 'coral'),
    ('2CHS-d5',[
        (2.766, 57.3),
        (8.418, 56.4),
        (15.714, 59.9),
        (23.140, 82.6),
        (28.332, 129.3),
        (30.152, 168.1),
        (30.16, 195.5),
        (30.52, 271.7),
        (31.54, 320.1)
    ], '-p', 'darkseagreen'),
    ('2CHS-d10',[
        (1.48, 107.2),
        (4.366, 108.2),
        (8.468, 113.3),
        (12.646, 150.3),
        (15.698, 247.7),
        (15.826, 303.3),
        (16.698, 446.1),
        (17.116, 637)
    ], '-v', 'darkseagreen'),
    ('SL-d5',[
        (2.870, 55.8),
        (8.702, 55.75),
        (16.206, 59.2),
        (23.198, 81.9),
        (27.266, 136.7),
        (29.330, 160.1),
        (29.510, 214.15),
        (30.162, 313.1)
    ], '-s', 'steelblue'),
    ('SL-d10',[
        (1.576, 102.7),
        (4.684, 102.1),
        (8.86, 107.9),
        (12.77, 147.2),
        (15.976, 263.2),
        (16.748, 420.5),
        (17.456, 506.1),
        (17.014, 605)
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
    plt.legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylim([0,700])
    plt.ylabel('Latency (ms)')
    plt.xlabel('Throughput (KTx/s)')
    plt.savefig('delays.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
