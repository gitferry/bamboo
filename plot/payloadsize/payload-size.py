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

# Measurements from payload-size.data
psize = [
    ('HS-p0',[
        (14.404,11.4),
        (40.494,12.7),
        (69.092,15.1),
        (96.863,21.8),
        (107.034,38.1),
        (111.545,56.8),
        (111.553,67.5),
        (110.863,83.4)
    ], '-o', 'coral'),
    ('HS-p128',[
        (13.502,11.8),
        (42.035,12.4),
        (67.242,15.6),
        (87.784,23.8),
        (97.557,42.3),
        (100.438,65),
        (101.349,85),
    ], '-^', 'coral'),
    ('HS-p1024',[
        (7.911,16.4),
        (12.038,23.1),
        (11.005,37.1),
        (7.142,97.8),
    ], '-*', 'coral'),
    ('2CHS-p0',[
        (17.073,9.6),
        (46.553,10.9),
        (73.281,13.9),
        (99.096,20.9),
        (108.383,38.5),
        (112.514,62),
        (113.417,77)
    ], '-p', 'darkseagreen'),
    ('2CHS-p128',[
        (16.689,9.9),
        (43.730,11.5),
        (68.247,14.7),
        (85.869,23.9),
        (96.815,43.5),
        (98.072,67.8),
        (102.178,85.3)
    ], '-v', 'darkseagreen'),
    ('2CHS-p1024',[
        (8.685,13.7),
        (19.006,15.1),
        (24.598,30.37),
        (11.905,80)
    ], '-d', 'darkseagreen'),
    ('SL-p0',[
        (16.74,9.7),
        (48.838,10.15),
        (77.275,13),
        (96.486,21.2),
        (107.166,38),
        (106.177,64),
        (103.788,83)
    ], '-h', 'steelblue'),
    ('SL-p128',[
        (14.556,10.5),
        (43.915,11.1),
        (70.228,14.2),
        (88.007,22.8),
        (93.720,35.5),
        (95.334,53.1),
        (98.194,70),
        (100.44,78)
    ], '-s', 'steelblue'),
    ('SL-p1024',[
        (11.046,12.1),
        (11.695,18.88),
        (11.519,36.7),
        (9.857,62.1),
    ], '->', 'steelblue')]



def do_plot():
    f = plt.figure(1, figsize=(8,6))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style, color in psize:
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
    plt.savefig('payload-size.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
