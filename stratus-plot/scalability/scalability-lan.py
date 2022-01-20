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
    f, ax = plt.subplots(2,1, figsize=(6,5))
    replicaNo = [16, 32, 64, 128, 256]
    xticks = [14.5,16, 32, 64, 128, 256, 260]
    xticks_label = ["","16", "32", "64", "128", "256", ""]
    thru = [
    ('N-HS',[
        167.2,
        49.3,
        33.2,
        25.1,
        16.1,
    ], 'o', 'coral'),
    ('SMP-HS',[
        175,
        150,
        142,
        89.2,
        78.9,
    ], 'p', 'steelblue'),
    ('S-HS',[
        144,
        135,
        130,
        80.1,
        55.1,
    ], 's', 'purple')
    # ('Tendermint',[
    #     154.3,
    #     133.4,
    #     94.5,
    #     55.6,
    #     30.4,
    # ], '^', 'darkseagreen'),
    # ('Narwhal',[
    #     184.3,
    #     163.4,
    #     154.5,
    #     85.6,
    #     60.4,
    # ], 'h', 'brown')
    ]
    for name, entries, style, color in thru:
        # thru = []
        # for item in entries:
        #     thru.append((item[0]+item[1])/2.0)
        ax[0].plot(replicaNo, entries, marker=style, mec=color, color=color, mfc='none', label='%s'%name, markersize=6)
#         ax[0].errorbar(replicaNo, thru, yerr=errs, marker='s', mfc='red', mec='green', ms=20, mew=4)
        ax[0].set_ylabel("Throughput (KTx/s)")
#         ax[0].set_yscale('log')
        ax[0].legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
        ax[0].set_xticks(xticks)
        ax[0].set_ylim([0,200])
        ax[0].set_xticklabels(xticks_label)
        ax[0].set_xticklabels(("", "", "", "", "", ""))
    lat = [
    ('N-HS',[
        55,
        190,
        700,
        1920,
        5870,
    ], 'o', 'coral'),
    ('SMP-HS',[
        29,
        57,
        98,
        258,
        531,
    ], 'p', 'steelblue'),
    ('S-HS',[
        33,
        72,
        162,
        320,
        1232
    ], 's', 'purple')
    # ('Tendermint',[
    #     154.3,
    #     133.4,
    #     94.5,
    #     55.6,
    #     30.4,
    # ], '^', 'darkseagreen'),
    # ('Narwhal',[
    #     184.3,
    #     163.4,
    #     154.5,
    #     85.6,
    #     60.4,
    # ], 'h', 'brown')
    ]
    for name, entries, style, color in lat:
        ax[1].plot(replicaNo, entries, marker=style, color=color, mec=color, mfc='none', label='%s' % name, markersize=6)
        ax[1].set_ylabel("Latency (ms)")
        ax[1].set_xticks(replicaNo)
        ax[1].set_xticks(xticks)
        ax[1].set_ylim([0,10000])
        ax[1].set_xticklabels(xticks_label)
        ax[1].set_yscale('log')
    ax[0].grid(linestyle='--', alpha=0.2)
    ax[1].grid(linestyle='--', alpha=0.2)
    f.text(0.5, 0.03, 'Number of nodes', ha='center', va='center')
#     plt.subplots_adjust(hspace=0.1)
    plt.savefig('scalability-lan.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
