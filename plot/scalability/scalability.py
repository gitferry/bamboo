import matplotlib.pyplot as plt

# Measurements from throughput.data
thru = [
    ('HotStuff',[
        119.2815
    ], '-o', 'coral'),
    ('2C-HS',[

    ], '-^', 'coral'),
    ('Streamlet',[

    ], '-s', 'coral')
    ]

# Measurements from throughput.data
lat = [
    ('HotStuff',[

    ], '-o', 'coral'),
    ('2C-HS',[

    ], '-^', 'coral'),
    ('Streamlet',[

    ], '-s', 'coral')
    ]



def do_plot():
    f, ax = plt.subplots(2,1, figsize=(5,8))
    replicaNo = [4, 8, 16, 32, 64, 128]
    for name, entries, style, color in thru:
        thru = []
        for item in entries:
            thru.append(item)
        ax[0].plot(byzNo, cgrs, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[0].set_ylabel("Throughput (KTx/s)")
        ax[0].set_ylim([0.4,1.0])
    for name, entries, style, color in lat:
        bis = []
        for item in entries:
            bis.append(item)
        ax[1].plot(byzNo, bis, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[1].set_ylabel("Latency (ms)")
        ax[1].yaxis.set_label_position("right")
        ax[1].yaxis.tick_right()
        ax[1].set_ylim([1.0,6.0])
    plt.legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
    f.text(0.5, 0.04, 'Byz. number', ha='center', va='center')
    plt.subplots_adjust(wspace=0.1)
    plt.savefig('scalability.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
