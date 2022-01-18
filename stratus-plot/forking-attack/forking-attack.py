import matplotlib.pyplot as plt

# Measurements from forking-attack.data
cgr = [
    ('HotStuff',[
        1.0, 0.873, 0.766, 0.658, 0.562, 0.476
    ], '-o', 'coral'),
    ('2CHS',[
        1.0, 0.933, 0.853, 0.789, 0.718, 0.659
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        1.0, 1.0, 1.0, 1.0, 1.0, 1.0
    ], '-s', 'steelblue')
    ]

bi = [
    ('HotStuff',[
        3.0, 3.231, 3.491, 3.859, 4.324, 5.086
    ], '-o', 'coral'),
    ('2CHS',[
        2.0, 2.149, 2.395, 2.632, 2.964, 3.383
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        2.0, 2.0, 2.0, 2.0, 2.0, 2.0
    ], '-s', 'steelblue')
    ]

thru = [
    ('HotStuff',[
        [49.14, 49.1],
        [43.564, 43.625],
        [38.465, 38.617],
        [33.537, 33.791],
        [29.099, 29.195],
        [24.795, 24.914]
    ], '-o', 'coral'),
    ('2CHS',[
        [50.90, 50.99],
        [48.28, 48.32],
        [45.04, 45.23],
        [42.79, 42.90],
        [39.9, 40.1],
        [37.0, 37.2]
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        [14.1, 14.2],
        [14.1, 14.2],
        [14.1, 14.2],
        [14.1, 14.2],
        [14.1, 14.2],
        [14.1, 14.2],
    ], '-s', 'steelblue')
    ]

lat = [
    ('HotStuff',[
        [213, 222],
        [270, 276],
        [330, 335],
        [376, 397],
        [495, 517],
        [738, 753]
    ], '-o', 'coral'),
    ('2CHS',[
        [216, 220],
        [249, 255],
        [267, 274],
        [293, 299],
        [324, 331],
        [361, 384]
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        [597, 630],
        [597, 630],
        [597, 630],
        [597, 630],
        [597, 630],
        [597, 630],
    ], '-s', 'steelblue')
    ]

def do_plot():
    f, ax = plt.subplots(2,2, figsize=(8,6))
    byzNo = [0, 2, 4, 6, 8, 10]
    for name, entries, style, color in cgr:
        cgrs = []
        for item in entries:
            cgrs.append(item)
        ax[1][0].plot(byzNo, cgrs, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[1][0].set_ylabel("Chain growth rate")
        ax[1][0].set_ylim([0,1.0])
        ax[1][0].legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
    for name, entries, style, color in bi:
        bis = []
        for item in entries:
            bis.append(item)
        ax[1][1].plot(byzNo, bis, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[1][1].set_ylabel("Block intervals")
        ax[1][1].yaxis.set_label_position("right")
        ax[1][1].yaxis.tick_right()
        ax[1][1].set_ylim([0,10.0])
    for name, entries, style, color in thru:
        throughput = []
        errs = []
        for item in entries:
            throughput.append((item[0]+item[1])/2.0)
            errs.append(abs(item[0]-item[1]))
        ax[0][0].errorbar(byzNo, throughput, yerr=errs, fmt=style, mec=color, color=color, mfc='none', label='%s'%name, markersize=6)
        ax[0][0].set_ylabel("Throughput (KTx/s)")
        ax[0][0].legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
#         a0[00[1].set_xticks(xticks)
        ax[0][0].set_ylim([0,60])
        ax[0][0].set_xticklabels(("", "", "", "", "", ""))
        ax[0][0].set_xlim([0,10])
#         a1[00[1].set_xticklabels(xticks_label)
    for name, entries, style, color in lat:
        latency = []
        errs = []
        for item in entries:
            latency.append((item[0]+item[1])/2.0)
            errs.append(abs(item[0]-item[1]))
        ax[0][1].errorbar(byzNo, latency, yerr=errs, fmt=style, mec=color, color=color, mfc='none', label='%s'%name, markersize=6)
        ax[0][1].set_ylabel("Latency (ms)")
#         ax[0][1].legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
        ax[0][1].yaxis.set_label_position("right")
        ax[0][1].yaxis.tick_right()
        ax[0][1].set_xticklabels(("", "", "", "", "", ""))
#         a0[1][1].set_xticks(xticks)
        ax[0][1].set_xlim([0,10])
#         ax[0][1].set_ylim([100,1000])
#         ax[1][1].set_xticklabels(xticks_label)
#     plt.legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
    f.text(0.5, 0.04, 'Byz. number', ha='center', va='center')
    plt.subplots_adjust(wspace=0.1)
    plt.subplots_adjust(hspace=0.1)
    ax[0][0].grid(linestyle='--', alpha=0.3)
    ax[1][0].grid(linestyle='--', alpha=0.3)
    ax[0][1].grid(linestyle='--', alpha=0.3)
    ax[1][1].grid(linestyle='--', alpha=0.3)
    plt.savefig('forking-attack-data.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
