import matplotlib.pyplot as plt

# Measurements from batch-size.data
cgr = [
    ('HotStuff',[
        1.0, 0.873, 0.766, 0.658, 0.562, 0.476
    ], '-o', 'coral'),
    ('2C-HS',[
        1.0, 0.933, 0.853, 0.789, 0.718, 0.659
    ], '-^', 'coral'),
    ('Streamlet',[
        1.0, 0.9375, 0.875, 0.812, 0.75, 0.6875
    ], '-*', 'coral')
    ]

bi = [
    ('HotStuff',[
        3.0, 3.231, 3.491, 3.859, 4.324, 5.086
    ], '-o', 'coral'),
    ('2C-HS',[
        2.0, 2.149, 2.395, 2.632, 2.964, 3.383
    ], '-^', 'coral'),
    ('Streamlet',[
        2.29, 3.0, 3.0, 3.0, 3.0, 3.0
    ], '-*', 'coral')
    ]

thru = [
    ('HotStuff',[

    ], '-o', 'coral'),
    ('2C-HS',[

    ], '-^', 'coral'),
    ('Streamlet',[

    ], '-*', 'coral')
    ]

lat = [
    ('HotStuff',[

    ], '-o', 'coral'),
    ('2C-HS',[

    ], '-^', 'coral'),
    ('Streamlet',[

    ], '-*', 'coral')
    ]



def do_plot():
    f, ax = plt.subplots(1,2, figsize=(7,5))
#     plt.clf()
    byzNo = [0, 2, 4, 6, 8, 10]
    for name, entries, style, color in cgr:
        cgrs = []
        for item in entries:
            cgrs.append(item)
        ax[0].plot(byzNo, cgrs, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
    for name, entries, style, color in bi:
        bis = []
        for item in entries:
            bis.append(item)
        ax[1].plot(byzNo, bis, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
    #ax.set_xscale("log")
    # ax.set_yscale("log")
    # plt.ylim([0, 50])
#     plt.xlim([0, 10])
    plt.legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
#     plt.grid(linestyle='--', alpha=0.3)
    # plt.ylabel('Throughput (Tx per second) in log scale')
#     plt.ylabel('Latency (ms)')
    plt.xlabel('Byz. number')
    # plt.xlabel('Requests (Tx) in log scale')
#     plt.tight_layout()
    plt.show()
#     plt.savefig('batch-size.pdf', format='pdf', dpi=400)

if __name__ == '__main__':
    do_plot()
