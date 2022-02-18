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

# batchsize = 512000
def do_plot():
    f, ax = plt.subplots(1,2,figsize=(10,4),constrained_layout=True)
    replicaNo = [16, 32, 64, 128, 256, 400]
    xticks = [14.5,16, 32, 64, 128, 256, 400, 410]
    xticks_label = ["","16", "", "64", "128", "256", "400", ""]
    thru = [
    ('N-HS',[
        167.2,
        49.3,
        33.2,
        25.1,
        16.1,
        0,
    ], 'o', 'coral'),
    ('N-SL',[
        28.2,
        8.3,
        1.2,
        0,
        0,
        0,
    ], '>', 'olive'),
    ('SMP-HS',[
        # 175,
        # 150,
        # 142,
        # 89.2,
        # 78.9,
        # 184,
        # 162,
        # 160,
        # 112,
        # 102,
        # 193,
        # 151,
        207, # batchsize = 128000
        195,
        185,
        131.7,
        101, # batchsize = 512000
        88,
    ], 'p', 'steelblue'),
    ('S-HS',[
        # 144,
        # 135,
        # 130,
        # 80.1,
        # 55.1,
        # 181,
        # 155,
        # 152,
        # 101,
        # 92,
        # 146,
        202,
        186,
        176.9,
        125.3,
        93,
        82,
    ], 's', 'purple'),
    ('S-SL',[
        167,
        149.3,
        110.9,
        70.3,
        30,
        10,
    ], '<', 'brown'),
    ('Narwhal',[
        254.3, # 173.3, 16
        167.8, # 32
        139.2, # 64
        130.2, # 128
        4.8, # 256
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
        ax[0].set_ylim([0, 280])
        ax[0].set_xticklabels(xticks_label)
        # ax[0].set_xticklabels(("", "", "", "", "", ""))
    lat = [
    ('N-HS',[
        55,
        190,
        700,
        1920,
        5870,
        1000000,
    ], 'o', 'coral'),
    ('N-SL',[
        100,
        532,
        1739,
        20391,
        1923919,
        1992913,
    ], '>', 'olive'),
    ('SMP-HS',[
        # 29,
        # 57,
        # 98,
        # 258,
        # 531,
        # 38,
        # 72,
        # 140,
        # 220,
        # 804,
        # 450,
        # 1012,
        49,
        96,
        203, # batchsize = 256000
        820, # batchsize = 400000
        2682, # batchsize = 512000
        4862,
    ], 'p', 'steelblue'),
    ('S-HS',[
        # 33,
        # 72,
        # 162,
        # 320,
        # 1232,
        # 97,
        # 279,
        # 893,
        # 2891,
        # 5613,
        # 878,
        # 1093,
        53,
        104,
        277,
        965,
        4573,
        7649,
    ], 's', 'purple'),
    ('S-SL',[
        129,
        349.3,
        1032.9,
        4932.3,
        9821,
        102311,
    ], '<', 'brown'),
    ('Narwhal',[
        832, # 16
        1031, #32
        1241, # 64
        6125, # 128
        40379, # 256
        100000, # 400
    ], '^', 'darkseagreen')
    ]
    for name, entries, style, color in lat:
        ax[1].plot(replicaNo, entries, marker=style, color=color, mec=color, mfc='none', label='%s' % name, markersize=8)
        ax[1].set_ylabel("Latency (ms)")
        ax[1].set_xticks(replicaNo)
        ax[1].set_xticks(xticks)
        ax[1].set_xticklabels(xticks_label)
    ax[1].set_yscale('log')
    ax[0].grid(linestyle='--', alpha=0.5)
    ax[1].grid(linestyle='--', alpha=0.5)
    ax[1].set_ylim([0,10000])
    # f.text(0.5, 0.03, 'Number of nodes', ha='center', va='center')
    # f.supxlabel("Number of replicas")
    f.supxlabel('# of replicas')
    # plt.subplots_adjust(hspace=0.1)
    plt.savefig('scalability-lan.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
