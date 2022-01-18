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
        (20.0,11.58),
        (40.0,12.69),
        (60.0,14.16),
        (79.8,16.47),
        (99.89,20.28),
        (122.2,35.67),
        (125.4,49.9),
    ], '-o', 'coral'),
    ('HS-p128',[
        (20.0,12.0),
        (59.9,15.5),
        (79.8,18.5),
        (99.5,25.4),
        (109.4,38.4),
        (113.0,50.1),
        (115.8,88.5)
    ], '-^', 'coral'),
    ('HS-p1024',[
        (20.0,15.0),
        (59.9,19.3),
        (79.7,30.5),
        (89.4,82.5),
    ], '-*', 'coral'),
    ('2CHS-p0',[
       (19.9, 10.1),
       (39.98, 10.92),
       (59.9, 12.56),
       (79.8, 14.63),
       (99.7, 18.27),
       (112.1, 21.25),
       (118.4, 29.43),
       (125.6, 48.1),
    ], '-p', 'darkseagreen'),
    ('2CHS-p128',[
        (20.0,10.3),
        (59.9,13.2),
        (99.5,22.8),
        (108.3,35.4),
        (114.1,56.1),
        (115.8,80.5),
    ], '-v', 'darkseagreen'),
    ('2CHS-p1024',[
        (20.0,12.5),
        (59.4,19.1),
        (79.4,31.1),
        (89.5,79.6)
    ], '-d', 'darkseagreen'),
    ('SL-p0',[
        (20.0,13.3),
        (40.0,15.1),
        (59.9,19.2),
        (79.9,30.52),
        (89.8,45.88),
        (91.8,68.67),
    ], '-h', 'steelblue'),
    ('SL-p128',[
        (20.0,13.9),
        (39.9,16.67),
        (59.8,23.4),
        (69.8,31.2),
        (75.7,52.5),
        (79.4,70.1)
    ], '-s', 'steelblue'),
    ('SL-p1024',[
        (19.9,15.1),
        (39.9,25.7),
        (49.8,48.9),
        (51.0,79.8)
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
    plt.legend(loc='best', fancybox=True,frameon=False,framealpha=0.8,ncol=2)
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylabel('Latency (ms)')
    plt.ylim([0,130])
    plt.xlim([10,130])
    plt.xlabel('Throughput (KTx/s)')
    plt.savefig('payload-size.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
