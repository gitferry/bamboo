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
    f, ax = plt.subplots(2,1, figsize=(6,5))
    replicaNo = [100, 200, 300, 400]
    xticks = [100, 200, 300, 400]
    # bandwidth = 100Mbps, batchsize = 12K
    # s=1.01 v=1
    thru1 = [
    ('SMP-HS',[
        [5.7, 8.5],
        [1.5, 3.1],
        [1.4, 2.4], # 20000
        [0.9, 1.4],
    ], 'p', 'steelblue'),
    ('GS-HS',[
        [17.8, 18.8], # 18.4, 21.4
        [12.8, 14.5],
        [11.6, 13.2],
        [9.4, 10.8],
    ], 's', 'purple')
    ]
    for name, entries, style, color in thru1:
        # thru = []
        # for item in entries:
        #     thru.append((item[0]+item[1])/2.0)
        ax[0].plot(replicaNo, entries, marker=style, mec=color, color=color, mfc='none', label='%s'%name, markersize=6)
#         ax[0].errorbar(replicaNo, thru, yerr=errs, marker='s', mfc='red', mec='green', ms=20, mew=4)
        ax[0].set_ylabel("Throughput (KTx/s)")
#         ax[0].set_yscale('log')
        ax[0].legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
        ax[0].set_xticks(xticks)
        ax[0].set_ylim([0,250])
        # ax[0].set_xticklabels(xticks_label)
        ax[0].set_xticklabels(("", "", "", "", "", ""))
    # s=1.1 v=5
    thru2 = [
    ('SMP-HS',[
        [7.4, 10.5],
        [4.1, 5.7],
        [2.9, 5.1],
        [2.3, 2.7],
    ], 'p', 'steelblue'),
    ('GS-HS',[
        [16.5, 18.2], # [17.1, 19.2]
        [11.1, 13.1], # 3411
        [10.2, 12.0], # 6165
        [7.8, 9.2],
    ], 's', 'purple')
    ]
    for name, entries, style, color in lat:
        ax[1].plot(replicaNo, entries, marker=style, color=color, mec=color, mfc='none', label='%s' % name, markersize=6)
        ax[1].set_ylabel("Latency (ms)")
        ax[1].set_xticks(replicaNo)
        ax[1].set_xticks(xticks)
        ax[1].set_ylim([0,10000])
        # ax[1].set_xticklabels(xticks_label)
        ax[1].set_yscale('log')
    ax[0].grid(linestyle='--', alpha=0.2)
    ax[1].grid(linestyle='--', alpha=0.2)
    f.text(0.5, 0.03, 'Number of nodes', ha='center', va='center')
#     plt.subplots_adjust(hspace=0.1)
    plt.savefig('unbalance_workload.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
