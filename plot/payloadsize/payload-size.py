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
#         (15.637,6.25),
        (42.262,6.83),
        (72.415,7.8),
        (108.774,10.06),
        (129.770,16.8),
        (126.022,25.2),
        (123.610,34.7)
    ], '-o', 'coral'),
    ('HS-p128',[
#         (14.048,6.92),
        (38.650,7.48),
        (65.796,8.56),
        (100.052,10.95),
        (115.621,19.05),
        (115.237,25.2),
        (113.937,31.5),
    ], '-^', 'coral'),
    ('HS-p1024',[
#         (13.598,7.11),
        (36.572,7.68),
        (65.483,8.50),
        (94.390,11.7),
        (102.865,21.5),
        (101.236,28.5),
        (95.555,34.7),
    ], '-*', 'coral'),
    ('2CHS-p0',[
#         (19.670,4.83),
        (52.919,5.36),
        (86.277,6.44),
        (122.119,8.76),
        (130.661,16.7),
        (127.120,21.5),
        (123.102,33.5),
    ], '-p', 'darkseagreen'),
    ('2CHS-p128',[
#         (17.578,5.43),
        (47.579,6.0),
        (80.774,6.85),
        (113.201,9.62),
        (118.210,18.4),
        (117.787,24.6),
        (114.167,31.4),
    ], '-v', 'darkseagreen'),
    ('2CHS-p1024',[
#         (17.526,5.46),
        (46.647,6.05),
        (77.386,7.13),
        (101.410,11.05),
        (108.024,20.9),
        (105.121,28.4),
        (98.592,34.5),
    ], '-d', 'darkseagreen'),
    ('SL-p0',[
#         (15.899,6.0),
        (45.450,6.38),
        (78.615,7.19),
        (110.851,9.85),
        (114.631,19.6),
        (108.519,25.5),
        (114.570,35.47)
    ], '-h', 'steelblue'),
    ('SL-p128',[
#         (15.698,6.13),
        (44.3,6.5),
        (75.576,7.47),
        (99.415,11.0),
        (98.000,22.5),
        (103.988,27.5),
        (100.099,36.2)
    ], '-s', 'steelblue'),
    ('SL-p1024',[
#         (15.843,6.06),
        (44.459,6.36),
        (72.104,7.66),
        (87.135,12.6),
        (92.271,23.7),
        (96.183,30.1),
        (92.531,37.1)
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
