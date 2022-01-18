import matplotlib.pyplot as plt

# Measurements from throughput.data

# Measurements from latency.data



def do_plot():
    f, ax = plt.subplots(2,1, figsize=(6,5))
    replicaNo = [16, 32, 64, 128, 256]
    xticks = [0, 16, 32, 64, 128, 256, 300]
    xticks_label = ["","16", "32", "64", "128", "256", ""]
    thru = [
    ('N-HS',[
        153.2,
        133.5,
        94.95,
        55.9,
        30.4,
    ], 'o', 'coral'),
    ('SMP-HS',[
        152,
        89.2,
    ], 'p', 'steelblue'),
    ('S-HS',[
        130,
        80.1
    ], 's', 'purple'),
    ('Tendermint',[
        154.3,
        133.4,
        94.5,
        55.6,
        30.4,
    ], '^', 'darkseagreen'),
    ('Narwhal',[
        184.3,
        163.4,
        154.5,
        85.6,
        60.4,
    ], 'h', 'brown')
    ]
    for name, entries, style, color in thru:
        # thru = []
        # for item in entries:
        #     thru.append((item[0]+item[1])/2.0)
        ax[0].plot(replicaNo, entries, marker=style, mec=color, color=color, mfc='none', label='%s'%name, markersize=6)
#         ax[0].errorbar(replicaNo, thru, yerr=errs, marker='s', mfc='red', mec='green', ms=20, mew=4)
        ax[0].set_ylabel("Throughput (KTx/s)")
#         ax[0].set_yscale('log')
        ax[0].legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
        ax[0].set_xticks(xticks)
        ax[0].set_ylim([0,200])
        ax[0].set_xticklabels(xticks_label)
        ax[0].set_xticklabels(("", "", "", "", "", ""))
    lat = [
    ('N-HS',[
        153.2,
        133.5,
        94.95,
        55.9,
        30.4,
    ], 'o', 'coral'),
    ('SMP-HS',[
        98,
        258,
    ], 'p', 'steelblue'),
    ('S-HS',[
        162,
        320,
    ], 's', 'purple'),
    ('Tendermint',[
        154.3,
        133.4,
        94.5,
        55.6,
        30.4,
    ], '^', 'darkseagreen'),
    ('Narwhal',[
        184.3,
        163.4,
        154.5,
        85.6,
        60.4,
    ], 'h', 'brown')
    ]
    for name, entries, style, color in lat:
        ax[1].plot(replicaNo, entries, marker=style, color=color, mec=color, mfc='none', label='%s' % name, markersize=6)
        ax[1].set_ylabel("Latency (ms)")
        ax[1].set_xticks(replicaNo)
        ax[1].set_xticks(xticks)
        ax[1].set_ylim([0,1000])
        ax[1].set_xticklabels(xticks_label)
        ax[1].set_yscale('log')
    ax[0].grid(linestyle='--', alpha=0.3)
    ax[1].grid(linestyle='--', alpha=0.3)
    f.text(0.5, 0.04, 'Number of Nodes', ha='center', va='center')
#     plt.subplots_adjust(hspace=0.1)
    plt.savefig('scalability.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
