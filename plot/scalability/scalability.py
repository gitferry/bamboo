import matplotlib.pyplot as plt

# Measurements from throughput.data

# Measurements from latency.data



def do_plot():
    f, ax = plt.subplots(2,1, figsize=(7,5))
    replicaNo = [4, 8, 16, 32, 64]
    xticks = [0,4, 8, 16, 32, 64,70]
    xticks_label = ["","4", "8", "16", "32", "64", ""]
    thru = [
    ('HotStuff',[
        [119.2815, 118.939],
        [90.721, 83.114],
        [65.721, 68.168],
        [39.309, 36.254],
        [21.741, 20.710]
    ], '-o', 'coral'),
    ('2CHS',[
        [119.565, 115.069],
        [108.572, 109.417],
        [80.644, 79.838],
        [49.601, 52.035],
        [26.917, 27.486]
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        [122.572, 114.046],
        [114.186, 117.320],
        [66.533, 69.830],
        [40.558, 42.116],
        [22.331, 23.164]
    ], '-s', 'steelblue')
    ]
    for name, entries, style, color in thru:
        thru = []
        errs = []
        for item in entries:
            thru.append((item[0]+item[1])/2.0)
            errs.append(abs(item[0]-item[1]))
        ax[0].errorbar(replicaNo, thru, yerr=errs, fmt=style, mec=color, color=color, mfc='none', label='%s'%name, markersize=6)
#         ax[0].errorbar(replicaNo, thru, yerr=errs, marker='s', mfc='red', mec='green', ms=20, mew=4)
        ax[0].set_ylabel("Throughput (KTx/s)")
        ax[0].legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
        ax[0].set_xticks(xticks)
        ax[0].set_ylim([0,140])
        ax[0].set_xticklabels(xticks_label)
    lat = [
    ('HotStuff',[
        [7.29, 7.16],
        [9.27, 9.28],
        [12.86, 12.86],
        [21.99, 21.98],
        [39.038, 38.8]
    ], '-o', 'coral'),
    ('2C-HS',[
        [6.16, 6.18],
        [7.66, 7.69],
        [10.31, 10.70],
        [17.13, 17.2],
        [29.99, 29.96]
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        [5.6, 5.61],
        [6.39, 6.26],
        [12.50, 12.84],
        [20.38, 20.43],
        [35.51, 35.38]
    ], '-s', 'steelblue')
    ]
    for name, entries, style, color in lat:
        lat = []
        errs = []
        for item in entries:
            lat.append((item[0]+item[1])/2.0)
            errs.append(abs(item[0]-item[1]))
        ax[1].errorbar(replicaNo, lat, yerr=errs, fmt=style, color=color, mec=color, mfc='none', label='%s' % name, markersize=6)
        ax[1].set_ylabel("Latency (ms)")
        ax[1].set_xticks(replicaNo)
        ax[1].set_xticks(xticks)
        ax[1].set_ylim([0,50])
        ax[1].set_xticklabels(xticks_label)
    f.text(0.5, 0.04, 'Number of Nodes', ha='center', va='center')
    plt.subplots_adjust(wspace=0.1)
    plt.savefig('scalability.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
