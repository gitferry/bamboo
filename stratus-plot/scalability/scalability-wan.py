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

def do_plot():
    f, ax = plt.subplots(1,2,figsize=(10,4),constrained_layout=True)
    replicaNo = [16, 32, 64, 128, 256, 400]
    xticks = [14.5, 16, 32, 64, 128, 256, 400, 410]
    xticks_label = ["","16", "", "64", "128", "256", "400", ""]
    thru = [
    ('N-HS',[
        # 167.2,
        # 49.3,
        # 33.2,
        # 25.1,
        # 16.1,
        4.5,
        2.7,
        1.2,
        0.8,
        0,
        0,
    ], 'o', 'coral'),
    ('N-SL',[
        1.2,
        0.2,
        0,
        0,
        0,
        0,
    ], '>', 'olive'),
    ('SMP-HS',[
        # 34.1,
        66.2,
        61.1,
        51.8,
        22.0,
        17.1,
        13.8,
    ], 'p', 'steelblue'),
    ('S-HS',[
        # 50.1,
        64.6,
        59.8,
        49.9,
        20.9,
        15.5,
        12.4,
    ], 's', 'purple'),
    ('S-SL',[
        60.1,
        50.3,
        32.1,
        8.1,
        2.3,
        0.6,
    ], '<', 'brown'),
    ('Narwhal',[
        33.9, #16
        29, #32
        24.1, #64
        18, #128
        4.4, # 256
        0, # 400
    ], 'h', 'darkseagreen')
    ]
    for name, entries, style, color in thru:
        # thru = []
        # for item in entries:
        #     thru.append((item[0]+item[1])/2.0)
        ax[0].plot(replicaNo, entries, marker=style, mec=color, color=color, mfc='none', label='%s'%name, markersize=8)
#         ax[0].errorbar(replicaNo, thru, yerr=errs, marker='s', mfc='red', mec='green', ms=20, mew=4)
        ax[0].set_ylabel("Throughput (KTx/s)")
#         ax[0].set_yscale('log')
        ax[0].legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
        ax[0].set_xticks(xticks)
        # ax[0].set_ylim([0,200])
        ax[0].set_ylim([0, 80])
        ax[0].set_xticklabels(xticks_label)
        # ax[0].set_xticklabels(("", "", "", "", "", ""))
    lat = [
    ('N-HS',[
        982,
        1342,
        2931,
        35810,
        1021432,
        2132193,
    ], 'o', 'coral'),
    ('N-SL',[
        1521,
        3321,
        12433,
        50321,
        1923919,
        1992913,
    ], '>', 'olive'),
    ('SMP-HS',[
        # 5.398,
        1031,
        1532,
        2312,
        4018,
        7386,
        15321,
    ], 'p', 'steelblue'),
    ('S-HS',[
        # 5.6,
        1282,
        1721,
        3321,
        5666,
        8732,
        18021,
    ], 's', 'purple'),
    ('S-SL',[
        2231,
        3499,
        7811,
        19011,
        29111,
        102311,
    ], '<', 'brown'),
    ('Narwhal',[
        3588, #16
        3605, #32
        4171, #64
        8888, #128
        52306, #256
        500000, #400
    ], '^', 'darkseagreen')
    ]
    for name, entries, style, color in lat:
        ax[1].plot(replicaNo, entries, marker=style, color=color, mec=color, mfc='none', label='%s' % name, markersize=8)
        ax[1].set_ylabel("Latency (ms)")
        ax[1].set_xticks(replicaNo)
        ax[1].set_xticks(xticks)
        ax[1].set_xticklabels(xticks_label)
        ax[1].set_yscale('log')
    ax[1].set_ylim([0,100000])
    ax[0].grid(linestyle='--', alpha=0.5)
    ax[1].grid(linestyle='--', alpha=0.5)
    f.supxlabel('# of replicas')
    # f.text(0.5, 0.03, 'Number of nodes', ha='center', va='center')
#     plt.subplots_adjust(hspace=0.1)
    plt.savefig('scalability-wan.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
