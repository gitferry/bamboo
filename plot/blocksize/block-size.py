import matplotlib.pyplot as plt

# Measurements from block-size.data
bsize = [
    ('HS-b100',[
        (15.680,11.21),
        (28.095,11.75),
        (52.799,12.84),
        (54.457,22.74),
        (54.378,26.32),
        (53.818,40.95)
    ], '-o', 'coral'),
    ('HS-b400',[
        (20.0,11.58),
        (40.0,12.69),
        (60.0,14.16),
        (79.8,16.47),
        (99.89,20.28),
        (122.2,35.67),
        (125.4,49.9),
    ], '-^', 'coral'),
    ('HS-b800',[
        (20.0,13.5),
        (59.7,14.2),
        (99.7,17.4),
        (132.4, 23.1),
        (154.3, 31.7),
        (162.4, 37.9),
        (164.1,41.8),
        (164.0,45.8)
    ], '-*', 'coral'),
    ('2CHS-b100',[
        (19.630,9.85),
        (36.824,10.16),
        (54.214,12.08),
        (54.325,19.24),
        (53.214,26.85),
        (53.057,41.31),
    ], '-p', 'darkseagreen'),
    ('2CHS-b400',[
       (19.9, 10.1),
       (39.98, 10.92),
       (59.9, 12.56),
       (79.8, 14.63),
       (99.7, 18.27),
       (112.1, 21.25),
       (118.4, 29.43),
       (125.6, 48.1),
    ], '-v', 'darkseagreen'),
    ('2CHS-b800',[
        (19.9,10.05),
        (59.9,12.2),
        (99.7,15.2),
        (131.8,20.0),
        (152.2,27.9),
        (161.1,36.0),
        (162.2,42.1)
    ], '-d', 'darkseagreen'),
    ('SL-b100',[
        (20.0,14.9),
        (29.9,24.2),
        (31.9,32.8),
        (33.9,42.8),
        (34.9,50.8),
    ], '-h', 'steelblue'),
    ('SL-b400',[
        (20.0,13.3),
        (40.0,15.1),
        (59.9,19.2),
        (79.9,30.52),
        (89.8,45.88),
        (91.8,55.1)
    ], '-s', 'steelblue'),
    ('SL-b800',[
        (20.0, 13.3),
        (59.9, 17.7),
        (99.4, 29.2),
        (108.9, 35.0),
        (117.2, 45.7),
        (118.2, 56.3)
    ], '->', 'steelblue'),
    ('OHS-b100',[
        (9.700,10.194),
        (19.799,12.205),
        (33.7,11.409),
        (38.760,10.0),
        (48.00,19.0),
        (48.00,42.0),
    ], '-8', 'darkmagenta'),
    ('OHS-b800',[
        (17.966,12.14),
        (58.966,12.52),
        (131.544,13.07),
        (141.544,14.07),
        (151.544,15.07),
        (169.542,18.3),
        (172.564,22.4),
        (176.649,37.4),
    ], '-<', 'darkmagenta')]



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
    plt.legend(fancybox=True,frameon=False,framealpha=0.8,mode={"expand", None},ncol=3, loc='upper center')
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylim([0,90])
    plt.ylabel('Latency (ms)')
    plt.xlabel('Throughput (KTx/s)')
    plt.tight_layout()
    plt.savefig('block-size.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
