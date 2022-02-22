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
        # S-HS-G
        [17.8, 18.8], # 18.4, 21.4
        [10.8, 12.5],
        [9.6, 11.2],
        [7.4, 8.8]
    ),
    (
        # S-HS-d1
        [27.8, 28.8], # 18.4, 21.4
        [14.8, 16.5],
        [12.6, 14.2],
        [10.4, 11.8]
    ),
    (
        # S-HS-d2
        [28.8, 30.8], # 18.4, 21.4
        [16.8, 17.5],
        [13.6, 15.2],
        [11.9, 12.8]
    ),
    (
        # S-HS-d3
        [29.5, 30.8], # 18.4, 21.4
        [17.2, 18.1],
        [13.9, 15.3],
        [12.3, 13.1]
    )
    ]
    even_workload = [
        31.9,
        19.9,
        16.5,
        13.4,
    ]
    bar_width = 18
    thru1_smp = []
    err1_smp = []
    corlor1_smp = 'steelblue'
    for entries in data1[0]:
        ave = (entries[0] + entries[1])/2.0
        thru1_smp.append(ave)
        err1_smp.append(entries[1]-entries[0])
    thru1_g = []
    err1_g = []
    corlor1_g = 'darkseagreen'
    for entries in data1[1]:
        ave = (entries[0] + entries[1])/2.0
        thru1_g.append(ave)
        err1_g.append(entries[1]-entries[0])
    thru1_d1 = []
    err1_d1 = []
    corlor1_d1 = 'orangered'
    for entries in data1[2]:
        ave = (entries[0] + entries[1])/2.0
        thru1_d1.append(ave)
        err1_d1.append(entries[1]-entries[0])
    thru1_d2 = []
    err1_d2 = []
    corlor1_d2 = 'coral'
    for entries in data1[3]:
        ave = (entries[0] + entries[1])/2.0
        thru1_d2.append(ave)
        err1_d2.append(entries[1]-entries[0])
    thru1_d3 = []
    err1_d3 = []
    corlor1_d3 = 'lightcoral'
    for entries in data1[4]:
        ave = (entries[0] + entries[1])/2.0
        thru1_d3.append(ave)
        err1_d3.append(entries[1]-entries[0])
    ax[0].bar(replicaNo-2*bar_width, thru1_smp, bar_width, color=corlor1_smp, yerr=err1_smp, error_kw=error_params, label='SMP-HS', edgecolor='black', alpha=0.6, hatch='/')
    ax[0].bar(replicaNo-1*bar_width, thru1_g, bar_width, color=corlor1_g, yerr=err1_g, error_kw=error_params, label='S-HS-G', edgecolor='black', alpha=0.6, hatch='//')
    ax[0].bar(replicaNo, thru1_d1, bar_width, color=corlor1_d1, yerr=err1_d1, error_kw=error_params, label='S-HS-d1', edgecolor='black', alpha=0.6, hatch='\\')
    ax[0].bar(replicaNo+bar_width, thru1_d2, bar_width, color=corlor1_d2, yerr=err1_d2, error_kw=error_params, label='S-HS-d2', edgecolor='black', alpha=0.6, hatch='\\\\')
    ax[0].bar(replicaNo+2*bar_width, thru1_d3, bar_width, color=corlor1_d3, yerr=err1_d3, error_kw=error_params, label='S-HS-d3', edgecolor='black', alpha=0.6, hatch='\\\\\\')
    ax[0].plot(replicaNo, even_workload, marker='s', linestyle='--', mec='purple', color='purple', mfc='none', label='S-HS-Even', markersize=8)
    ax[0].set_ylabel("Throughput (KTx/s)")
    # plt.set_xticklabels(tick_label)
    # plt.set_xticks(replicaNo+2*bar_width)
    # s=1.1 v=10
    data2 = [
    (
        # SMP-HS
        [7.4, 10.5],
        [4.1, 5.7],
        [2.9, 5.1],
        [2.3, 2.7],
    ),
    (
        # S-HS-G
        [10.5, 11.2], # [17.1, 19.2]
        [5.1, 6.1], # 3411
        [3.2, 4.0], # 6165
        [2.8, 3.2],
    ),
    (
        # S-HS-d1
        [19.5, 21.2], # [17.1, 19.2]
        [9.1, 10.1], # 3411
        [8.2, 9.0], # 6165
        [7.8, 8.2],
    ),
    (
        # S-HS-d2
        [26.5, 27.2], # [17.1, 19.2]
        [15.1, 17.1], # 3411
        [11.2, 13.0], # 6165
        [10.8, 12.2],
    ),
    (
        # S-HS-d3
        [29.5, 30.2], # [17.1, 19.2]
        [18.1, 19.5], # 3411
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
    thru2_g = []
    err2_g = []
    corlor2_g = 'darkseagreen'
    for entries in data2[1]:
        ave = (entries[0] + entries[1])/2.0
        thru2_g.append(ave)
        err2_g.append(entries[1]-entries[0])
    thru2_d1 = []
    err2_d1 = []
    corlor2_d1 = 'orangered'
    for entries in data2[2]:
        ave = (entries[0] + entries[1])/2.0
        thru2_d1.append(ave)
        err2_d1.append(entries[1]-entries[0])
    thru2_d2 = []
    err2_d2 = []
    corlor2_d2 = 'coral'
    for entries in data2[3]:
        ave = (entries[0] + entries[1])/2.0
        thru2_d2.append(ave)
        err2_d2.append(entries[1]-entries[0])
    thru2_d3 = []
    err2_d3 = []
    corlor2_d3 = 'lightcoral'
    for entries in data2[4]:
        ave = (entries[0] + entries[1])/2.0
        thru2_d3.append(ave)
        err2_d3.append(entries[1]-entries[0])
    ax[1].bar(replicaNo-2*bar_width, thru2_smp, bar_width, color=corlor2_smp, yerr=err2_smp, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='/')
    ax[1].bar(replicaNo-1*bar_width, thru2_g, bar_width, color=corlor2_g, yerr=err2_g, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='//')
    ax[1].bar(replicaNo, thru2_d1, bar_width, color=corlor2_d1, yerr=err2_d1, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='\\')
    ax[1].bar(replicaNo+bar_width, thru2_d2, bar_width, color=corlor2_d2, yerr=err2_d2, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='\\\\')
    ax[1].bar(replicaNo+2*bar_width, thru2_d3, bar_width, color=corlor2_d3, yerr=err2_d3, error_kw=error_params, edgecolor='black', alpha=0.6, hatch='\\\\\\')
    ax[1].plot(replicaNo, even_workload, linestyle='--', marker='s', mec='purple', color='purple', mfc='none', markersize=8)
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
    ax[0].set_title('(a) Zipf1 s=1.01 v=1 (highly skewed)')
    ax[1].set_title('(b) Zipf10 s=1.01 v=10 (lightly skewed)')
    ax[0].set_ylim([0,40])
    ax[1].set_ylim([0,40])
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
