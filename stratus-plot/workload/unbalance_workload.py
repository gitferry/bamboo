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
plt.rc('legend', fontsize=MEDIUM_SIZE)    # legend fontsize

# batchsize = 512000
def do_plot():
    # f, ax = plt.subplots(2,1, figsize=(6,5))
    f, ax = plt.subplots(1,2,figsize=(10,4), sharey=True, frameon=False, constrained_layout=True)
    # plt.tick_params(labelcolor="none", bottom=False, left=False)
    replicaNo = np.arange(100,500,100)
    tick_label=['100','200','300','400']
    error_params=dict(ecolor='black',capsize=2, elinewidth=1)
    # bandwidth = 100Mbps, batchsize = 12K
    # s=1.01 v=1
    data1 = [
    (
        # SMP-HS
        [5.7, 8.5],
        [1.5, 3.1],
        [1.4, 2.4], # 20000
        [0.9, 1.4]
    ),
    (
        # S-HS-d1
        [17.8, 18.8], # 18.4, 21.4
        [12.8, 14.5],
        [11.6, 13.2],
        [9.4, 10.8]
    ),
    (
        # S-HS-d2
        [20.8, 21.8], # 18.4, 21.4
        [15.8, 17.5],
        [14.6, 16.2],
        [12.4, 13.8]
    ),
    (
        # S-HS-d3
        [21.8, 22.8], # 18.4, 21.4
        [16.8, 18.5],
        [15.6, 17.2],
        [13.4, 14.8]
    )
    ]
    bar_width = 20
    thru1_smp = []
    err1_smp = []
    corlor1_smp = 'steelblue'
    for entries in data1[0]:
        ave = (entries[0] + entries[1])/2.0
        thru1_smp.append(ave)
        err1_smp.append(entries[1]-entries[0])
    thru1_d1 = []
    err1_d1 = []
    corlor1_d1 = 'orangered'
    for entries in data1[1]:
        ave = (entries[0] + entries[1])/2.0
        thru1_d1.append(ave)
        err1_d1.append(entries[1]-entries[0])
    thru1_d2 = []
    err1_d2 = []
    corlor1_d2 = 'coral'
    for entries in data1[2]:
        ave = (entries[0] + entries[1])/2.0
        thru1_d2.append(ave)
        err1_d2.append(entries[1]-entries[0])
    thru1_d3 = []
    err1_d3 = []
    corlor1_d3 = 'lightcoral'
    for entries in data1[3]:
        ave = (entries[0] + entries[1])/2.0
        thru1_d3.append(ave)
        err1_d3.append(entries[1]-entries[0])
    ax[0].bar(replicaNo-bar_width, thru1_smp, bar_width, color=corlor1_smp, yerr=err1_smp, error_kw=error_params, label='SMP-HS', edgecolor='black', alpha=0.6, hatch='/')
    ax[0].bar(replicaNo, thru1_d1, bar_width, color=corlor1_d1, yerr=err1_d1, error_kw=error_params, label='S-HS-d1', edgecolor='black', alpha=0.6, hatch='//')
    ax[0].bar(replicaNo+bar_width, thru1_d2, bar_width, color=corlor1_d2, yerr=err1_d2, error_kw=error_params, label='S-HS-d2', edgecolor='black', alpha=0.6, hatch='\\')
    ax[0].bar(replicaNo+2*bar_width, thru1_d3, bar_width, color=corlor1_d3, yerr=err1_d3, error_kw=error_params, label='S-HS-d3', edgecolor='black', alpha=0.6, hatch='/\\')
    ax[0].set_ylabel("Throughput (KTx/s)")
    # plt.set_xticklabels(tick_label)
    # plt.set_xticks(replicaNo+2*bar_width)
    # s=1.1 v=5
    data2 = [
    (
        # SMP-HS
        [7.4, 10.5],
        [4.1, 5.7],
        [2.9, 5.1],
        [2.3, 2.7],
    ),
    (
        # S-HS-d1
        [16.5, 18.2], # [17.1, 19.2]
        [11.1, 13.1], # 3411
        [10.2, 12.0], # 6165
        [7.8, 9.2],
    ),
    (
        # S-HS-d2
        [19.5, 21.2], # [17.1, 19.2]
        [14.1, 16.1], # 3411
        [13.2, 15.0], # 6165
        [10.8, 12.2],
    ),
    (
        # S-HS-d3
        [20.5, 22.2], # [17.1, 19.2]
        [15.1, 17.1], # 3411
        [14.2, 16.0], # 6165
        [11.8, 13.2],
    )
    ]
    thru2_smp = []
    err2_smp = []
    corlor2_smp = 'steelblue'
    for entries in data2[0]:
        ave = (entries[0] + entries[1])/2.0
        thru2_smp.append(ave)
        err2_smp.append(entries[1]-entries[0])
    thru2_d1 = []
    err2_d1 = []
    corlor2_d1 = 'orangered'
    for entries in data2[1]:
        ave = (entries[0] + entries[1])/2.0
        thru2_d1.append(ave)
        err2_d1.append(entries[1]-entries[0])
    thru2_d2 = []
    err2_d2 = []
    corlor2_d2 = 'coral'
    for entries in data1[2]:
        ave = (entries[0] + entries[1])/2.0
        thru2_d2.append(ave)
        err2_d2.append(entries[1]-entries[0])
    thru2_d3 = []
    err2_d3 = []
    corlor2_d3 = 'lightcoral'
    for entries in data1[3]:
        ave = (entries[0] + entries[1])/2.0
        thru2_d3.append(ave)
        err2_d3.append(entries[1]-entries[0])
    ax[1].bar(replicaNo-bar_width, thru2_smp, bar_width, color=corlor2_smp, yerr=err2_smp, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='/')
    ax[1].bar(replicaNo, thru2_d1, bar_width, color=corlor2_d1, yerr=err2_d1, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='//')
    ax[1].bar(replicaNo+bar_width, thru2_d2, bar_width, color=corlor2_d2, yerr=err2_d2, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='\\')
    ax[1].bar(replicaNo+2*bar_width, thru2_d3, bar_width, color=corlor2_d3, yerr=err2_d3, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='/\\')
    # plt.bar(replicaNo-2*bar_width, thru1_smp, bar_width, color=corlor1_smp, yerr=err1_smp, error_kw=error_params, label='SMP-HS-Zipf1')
    # plt.bar(replicaNo-bar_width, thru2_smp, bar_width, color=corlor2_smp, yerr=err2_smp, error_kw=error_params, label='SMP-HS-Zipf2')
    # plt.bar(replicaNo, thru1_gs, bar_width, color=corlor1_gs, yerr=err1_gs, error_kw=error_params, label='GS-HS-Zipf1')
    # plt.bar(replicaNo+bar_width, thru2_gs, bar_width, color=corlor2_gs, yerr=err2_gs, error_kw=error_params, label='GS-HS-Zipf2')
    # plt.xlabel("Number of nodes")
    # plt.ylabel("Throughput (KTx/s)")
    f.supxlabel('# of replicas')
    # ax[0].legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
    # plt.set_ylabel("Throughput (KTx/s)")
    # ax[1].legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
    ax[0].set_title('(a) Zipfian s=1.01 v=1')
    ax[1].set_title('(b) Zipfian s=1.01 v=5')
    ax[0].set_ylim([0,27])
    ax[1].set_ylim([0,27])
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
    ax[0].grid(linestyle='--', alpha=0.5)
    ax[1].grid(linestyle='--', alpha=0.5)
    plt.xticks(replicaNo,tick_label)
    # plt.text(0.5, 0.03, 'Number of nodes', ha='center', va='center')
    # plt.subplots_adjust(wspace=0.08)
    ax[0].legend(loc='upper right', fancybox=True, ncol=2)
    # plt.tight_layout()
    plt.savefig('unbalance_workload.pdf', format='pdf')
    # plt.xlabel("# of replicas")
    # plt.text(0.04, 0.5, '# of replicas', ha='center', va='center')
    plt.show()

if __name__ == '__main__':
    do_plot()
