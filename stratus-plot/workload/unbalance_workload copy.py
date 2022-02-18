import matplotlib.pyplot as plt
import numpy as np

SMALL_SIZE = 8
MEDIUM_SIZE = 13
BIGGER_SIZE = 16

plt.rc('font', size=BIGGER_SIZE)          # controls default text sizes
plt.rc('axes', titlesize=BIGGER_SIZE)     # fontsize of the axes title
plt.rc('axes', labelsize=BIGGER_SIZE)    # fontsize of the x and y labels
plt.rc('xtick', labelsize=BIGGER_SIZE)    # fontsize of the tick labels
plt.rc('ytick', labelsize=BIGGER_SIZE)    # fontsize of the tick labels
plt.rc('legend', fontsize=BIGGER_SIZE)    # legend fontsize

# batchsize = 512000
def do_plot():
    # f, ax = plt.subplots(2,1, figsize=(6,5))
    replicaNo = np.arange(100,500,100)
    tick_label=['100','200','300','400']
    error_params=dict(ecolor='black',capsize=3)
    # bandwidth = 100Mbps, batchsize = 12K
    # s=1.01 v=1
    data1 = [
    (
        [5.7, 8.5],
        [1.5, 3.1],
        [1.4, 2.4], # 20000
        [0.9, 1.4]
    ),
    (
        [17.8, 18.8], # 18.4, 21.4
        [12.8, 14.5],
        [11.6, 13.2],
        [9.4, 10.8]
    )]
    bar_width = 20
    thru1_smp = []
    err1_smp = []
    corlor1_smp = 'steelblue'
    for entries in data1[0]:
        ave = (entries[0] + entries[1])/2.0
        thru1_smp.append(ave)
        err1_smp.append(entries[1]-entries[0])
    thru1_gs = []
    err1_gs = []
    corlor1_gs = 'orangered'
    for entries in data1[1]:
        ave = (entries[0] + entries[1])/2.0
        thru1_gs.append(ave)
        err1_gs.append(entries[1]-entries[0])
    # plt.set_ylabel("Throughput (KTx/s)")
    # plt.set_xticklabels(tick_label)
    # plt.set_xticks(replicaNo+2*bar_width)
    # s=1.1 v=5
    data2 = [
    (
        [7.4, 10.5],
        [4.1, 5.7],
        [2.9, 5.1],
        [2.3, 2.7],
    ),
    (
        [16.5, 18.2], # [17.1, 19.2]
        [11.1, 13.1], # 3411
        [10.2, 12.0], # 6165
        [7.8, 9.2],
    )]
    thru2_smp = []
    err2_smp = []
    corlor2_smp = 'lightskyblue'
    for entries in data2[0]:
        ave = (entries[0] + entries[1])/2.0
        thru2_smp.append(ave)
        err2_smp.append(entries[1]-entries[0])
    thru2_gs = []
    err2_gs = []
    corlor2_gs = 'lightcoral'
    for entries in data2[1]:
        ave = (entries[0] + entries[1])/2.0
        thru2_gs.append(ave)
        err2_gs.append(entries[1]-entries[0])
    plt.bar(replicaNo-2*bar_width, thru1_smp, bar_width, color=corlor1_smp, yerr=err1_smp, error_kw=error_params, label='SMP-HS-Zipf1')
    plt.bar(replicaNo-bar_width, thru2_smp, bar_width, color=corlor2_smp, yerr=err2_smp, error_kw=error_params, label='SMP-HS-Zipf2')
    plt.bar(replicaNo, thru1_gs, bar_width, color=corlor1_gs, yerr=err1_gs, error_kw=error_params, label='GS-HS-Zipf1')
    plt.bar(replicaNo+bar_width, thru2_gs, bar_width, color=corlor2_gs, yerr=err2_gs, error_kw=error_params, label='GS-HS-Zipf2')
    plt.xlabel("Number of nodes")
    plt.ylabel("Throughput (KTx/s)")
    plt.legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
    # plt.set_ylabel("Throughput (KTx/s)")
    # ax[1].legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
    # ax[1].set_xticklabels(tick_label)
    # ax[1].set_xticks(replicaNo+2*bar_width)
    # thru2 = [
    # ('SMP-HS',[
    #     [7.4, 10.5],
    #     [4.1, 5.7],
    #     [2.9, 5.1],
    #     [2.3, 2.7],
    # ], 'p', 'steelblue'),
    # ('GS-HS',[
    #     [16.5, 18.2], # [17.1, 19.2]
    #     [11.1, 13.1], # 3411
    #     [10.2, 12.0], # 6165
    #     [7.8, 9.2],
    # ], 's', 'purple')
    # ]
    # for name, entries, style, color in thru2:
    #     ax[1].plot(replicaNo, entries, marker=style, color=color, mec=color, mfc='none', label='%s' % name, markersize=6)
    #     ax[1].set_ylabel("Latency (ms)")
    #     ax[1].set_xticks(replicaNo)
    #     # ax[1].set_xticks(xticks)
    #     ax[1].set_ylim([0,10000])
    #     # ax[1].set_xticklabels(xticks_label)
    #     ax[1].set_yscale('log')
    plt.grid(linestyle='--', alpha=0.2)
    plt.grid(linestyle='--', alpha=0.2)
    plt.xticks(replicaNo,tick_label)
    # plt.text(0.5, 0.03, 'Number of nodes', ha='center', va='center')
#     plt.subplots_adjust(hspace=0.1)
    plt.savefig('unbalance_workload.pdf', format='pdf')
    # plt.ylabel("Throughput (KTx/s)")
    plt.show()

if __name__ == '__main__':
    do_plot()
