import matplotlib.pyplot as plt

# Measurements from forking-attack.data
cgr = [
    ('HotStuff',[
        1.0, 0.93, 0.864, 0.807, 0.738, 0.637
    ], '-o', 'coral'),
    ('2CHS',[
        1.0, 0.935, 0.872, 0.810, 0.742, 0.643
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        1.0, 1.0, 1.0, 1.0, 1.0, 1.0
    ], '-s', 'steelblue')
    ]

bi = [
    ('HotStuff',[
        3.0, 3.557, 4.286, 5.122, 6.676, 9.531
    ], '-o', 'coral'),
    ('2CHS',[
        2.0, 2.218, 2.496, 2.854, 3.371, 4.496
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        2.0, 2.29, 2.662, 3.153, 3.820, 5.688
    ], '-s', 'steelblue')
    ]

thru = [
    ('HotStuff',[
        [49.14, 49.1],
        [36.474, 37.31],
        [25.279, 26.14],
        [18.2, 18.5],
        [13.4, 13.5],
        [9.9, 10.0]
    ], '-o', 'coral'),
    ('2CHS',[
        [50.90, 50.99],
        [36.0, 36.5],
        [25.6, 25.9],
        [17.5, 17.7],
        [12.6, 12.7],
        [9.5, 9.7]
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        [14.1, 14.2],
        [12.5, 12.5],
        [11.6, 11.7],
        [9.5, 9.7],
        [7.9, 8.1],
        [5.7, 5.8],
    ], '-s', 'steelblue')
    ]

lat = [
    ('HotStuff',[
        [213, 222],
        [416, 443],
        [735, 776],
        [1073, 1160],
        [1537, 1581],
        [2137, 2221],
    ], '-o', 'coral'),
    ('2C-HS',[
        [216, 220],
        [402, 419],
        [576, 612],
        [1029, 1074],
        [1488, 1535],
        [2053, 2097],
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        [597, 630],
        [707, 740],
        [765, 826],
        [856, 911],
        [1050, 1155],
        [1400, 1438],
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
    plt.savefig('silence-attack-data.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
