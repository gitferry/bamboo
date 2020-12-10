import matplotlib.pyplot as plt

# Measurements from forking-attack.data
cgr = [
    ('HotStuff',[
        1.0, 0.873, 0.766, 0.658, 0.562, 0.476
    ], '-o', 'coral'),
    ('2C-HS',[
        1.0, 0.933, 0.853, 0.789, 0.718, 0.659
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        1.0, 0.9375, 0.875, 0.812, 0.75, 0.6875
    ], '-s', 'steelblue')
    ]

bi = [
    ('HotStuff',[
        3.0, 3.557, 4.286, 5.122, 6.676, 9.531
    ], '-o', 'coral'),
    ('2C-HS',[
        2.0, 2.149, 2.395, 2.632, 2.964, 3.383
    ], '-^', 'darkseagreen'),
    ('Streamlet',[
        2.0, 2.29, 2.662, 3.153, 3.820, 5.688
    ], '-s', 'steelblue')
    ]

thru = [
    ('HotStuff',[

    ], '-o', 'coral'),
    ('2C-HS',[

    ], '-^', 'coral'),
    ('Streamlet',[

    ], '-s', 'coral')
    ]

lat = [
    ('HotStuff',[

    ], '-o', 'coral'),
    ('2C-HS',[

    ], '-^', 'coral'),
    ('Streamlet',[

    ], '-s', 'coral')
    ]



def do_plot():
    f, ax = plt.subplots(1,2, figsize=(8,5))
    byzNo = [0, 2, 4, 6, 8, 10]
    for name, entries, style, color in cgr:
        cgrs = []
        for item in entries:
            cgrs.append(item)
        ax[0].plot(byzNo, cgrs, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[0].set_ylabel("chain growth rate")
        ax[0].set_ylim([0.4,1.0])
    for name, entries, style, color in bi:
        bis = []
        for item in entries:
            bis.append(item)
        ax[1].plot(byzNo, bis, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[1].set_ylabel("block intervals")
        ax[1].yaxis.set_label_position("right")
        ax[1].yaxis.tick_right()
        ax[1].set_ylim([1.0,6.0])
    plt.legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
    f.text(0.5, 0.04, 'Byz. number', ha='center', va='center')
    plt.subplots_adjust(wspace=0.1)
    plt.savefig('silence-attack-data.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
