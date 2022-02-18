import matplotlib.pyplot as plt

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
    f, ax = plt.subplots(1,2,figsize=(10,3),constrained_layout=True)
    replicaNo = [0, 10, 20, 30]
    # xticks = [0, 10, 20, 30]
    # xticks_label = ["0", "10", "20", "30"]
    thru = [
    ('SMP-HS',[
        142.8,
        126.7,
        91.8,
        50.9,
    ], 'o', 'steelblue'),
    ('S-HS-f',[
        137.3,
        135.5,
        126.0,
        105,
    ], 'p', 'coral'),
    ('S-HS-2f',[
        137.3,
        136.5,
        134.0,
        129.1,
    ], '^', 'darkseagreen'),
    ]
    for name, entries, style, color in thru:
        ax[0].plot(replicaNo, entries, marker=style, mec=color, color=color, mfc='none', label='%s'%name, markersize=8)
        ax[0].set_ylabel("Throughput (KTx/s)")
        # ax[0].set_xticks(xticks)
        ax[0].set_ylim([0,200])
        # ax[0].set_xticklabels(xticks_label)
        # ax[0].set_xticklabels(("", "", "", "", "", ""))
    lat = [
    ('SMP-HS',[
        603,
        850,
        4096,
        9669,
    ], 'o', 'steelblue'),
    ('S-HS-f',[
        616,
        622,
        618,
        616,
    ], 'p', 'coral'),
    ('S-HS-2f',[
        916,
        922,
        918,
        916,
    ], '^', 'darkseagreen')
    ]
    for name, entries, style, color in lat:
        ax[1].plot(replicaNo, entries, marker=style, color=color, mec=color, mfc='none', label='%s' % name, markersize=8)
        ax[1].set_ylabel("Latency (ms)")
        ax[1].set_xticks(replicaNo)
        # ax[1].set_xticks(xticks)
        ax[1].set_ylim([0,10000])
        # ax[1].set_xticklabels(xticks_label)
        # ax[1].set_yscale('log')
    ax[0].grid(linestyle='--', alpha=0.5)
    ax[1].grid(linestyle='--', alpha=0.5)
    ax[1].legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
    # f.text(0.5, 0.03, 'Number of Byzantine nodes', ha='center', va='center')
    # plt.tight_layout()
    f.supxlabel('# of Byz. replicas')
    plt.savefig('byz-100.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
