import matplotlib.pyplot as plt

n4bs100 = [
    ('Bamboo-HS',[
        (10.0,10.07),
        (20.0,10.07),
        (30.0,10.7),
        (40.0,12.8),
        (49.8,17.5),
        (60.6,44.25),
    ], '-o', 'coral'),
    ('Bamboo-2CHS',[
        (10.0,8.54),
        (20.0,8.63),
        (30.0,9.5),
        (40.0,11.27),
        (49.8,13.3),
        (60.6,35.21),
    ], '-s', 'darkseagreen'),
    ('Bamboo-SL',[
        (10.0,9.54),
        (20.0,9.44),
        (30.0,11.4),
        (40.0,14.42),
        (46.7,24.28),
        (47.9,29.07),
    ], '-d', 'steelblue'),
    ('Model-HS',[
        (10.0,10.07),
        (20.0,10.07),
        (30.0,10.7),
        (40.0,12.8),
        (49.8,17.5),
        (60.6,44.25),
    ], '--<', 'coral'),
    ('Model-2CHS',[
        (10.0,8.54),
        (20.0,8.63),
        (30.0,9.5),
        (40.0,11.27),
        (49.8,13.3),
        (60.6,35.21),
    ], '--^', 'darkseagreen'),
    ('Model-SL',[
        (10.0,9.54),
        (20.0,9.44),
        (30.0,11.4),
        (40.0,14.42),
        (46.7,24.28),
        (47.9,29.07),
    ], '-->', 'steelblue')]

n4bs400 = [
    ('Bamboo-HS',[
        (20.0,10.6),
        (40.0,12.24),
        (60.0,13.36),
        (80.0,15.67),
        (100.0,17.54),
        (119.2,25.34),
        (131.3,36.3),
    ], '-o', 'coral'),
    ('Bamboo-2CHS',[
       (20.0, 9.16),
       (40.0, 10.12),
       (60.0, 11.36),
       (79.8, 12.9),
       (100.0, 16.24),
       (121.1, 21.3),
       (130.4, 33.74),
    ], '-s', 'darkseagreen'),
    ('Bamboo-SL',[
        (20.0,9.3),
        (40.0,10.29),
        (59.84,12.44),
        (80.0,17.1),
        (99.1,26.5),
        (109.8,35.0),
    ], '-d', 'steelblue'),
    ('Model-HS',[
        (20.0,10.6),
        (40.0,12.24),
        (60.0,13.36),
        (80.0,15.67),
        (100.0,17.54),
        (119.2,25.34),
        (131.3,36.3),
    ], '--<', 'coral'),
    ('Model-2CHS',[
       (20.0, 9.16),
       (40.0, 10.12),
       (60.0, 11.36),
       (79.8, 12.9),
       (100.0, 16.24),
       (121.1, 21.3),
       (130.4, 33.74),
    ], '--^', 'darkseagreen'),
    ('Model-SL',[
        (20.0,9.3),
        (40.0,10.29),
        (59.84,12.44),
        (80.0,17.1),
        (99.1,26.5),
        (109.8,35.0),
    ], '-->', 'steelblue')]

n8bs100 = [
    ('Bamboo-HS',[
        (4.0,29.0),
        (8.0,29.5),
        (12.0,30.86),
        (16.0,34.04),
        (20.0,38.4),
        (24.0,50.7),
    ], '-o', 'coral'),
    ('Bamboo-2CHS',[
       (4.0, 26.4),
       (8.0, 29.14),
       (12.0, 29.4),
       (16.0, 31.27),
       (20.0, 36.8),
       (23.4, 41.48),
    ], '-s', 'darkseagreen'),
    ('Bamboo-SL',[
        (4.0,32.19),
        (8.0,31.1),
        (12.0,33.87),
        (16.1,39.9),
        (19.4,53.34),
        (22.0,72.0),
    ], '-d', 'steelblue'),
    ('Model-HS',[
        (4.0,29.0),
        (8.0,29.5),
        (12.0,30.86),
        (16.0,34.04),
        (20.0,38.4),
        (24.0,50.7),
    ], '--<', 'coral'),
    ('Model-2CHS',[
       (4.0, 26.4),
       (8.0, 29.14),
       (12.0, 29.4),
       (16.0, 31.27),
       (20.0, 36.8),
       (23.4, 41.48),
    ], '--^', 'darkseagreen'),
    ('Model-SL',[
        (4.0,32.19),
        (8.0,31.1),
        (12.0,33.87),
        (16.1,39.9),
        (19.4,53.34),
        (22.0,72.0),
    ], '-->', 'steelblue')]

n8bs400 = [
    ('Bamboo-HS',[
        (8.0,26.15),
        (20.0,28.74),
        (40.0,33.14),
        (60.0,41.57),
        (71.58,53.9),
        (79.4,72.9),
    ], '-o', 'coral'),
    ('Bamboo-2CHS',[
       (8.0, 23.9),
       (20.0, 26.45),
       (40.0, 28.9),
       (60.8, 39.56),
       (70.66, 48.48),
       (79.79, 61.25),
    ], '-s', 'darkseagreen'),
    ('Bamboo-SL',[
        (8.0,29.23),
        (19.628,31.63),
        (32.2,31.73),
        (39.9,42.74),
        (52.6,50.4),
        (59.52,62.64),
    ], '-d', 'steelblue'),
    ('Model-HS',[
        (8.0,26.15),
        (20.0,28.74),
        (40.0,33.14),
        (60.0,41.57),
        (71.58,53.9),
        (79.4,72.9),
    ], '--<', 'coral'),
    ('Model-2CHS',[
       (8.0, 23.9),
       (20.0, 26.45),
       (40.0, 28.9),
       (60.8, 39.56),
       (70.66, 48.48),
       (79.79, 61.25),
    ], '--^', 'darkseagreen'),
    ('Model-SL',[
        (8.0,29.23),
        (19.628,31.63),
        (32.2,31.73),
        (39.9,42.74),
        (52.6,50.4),
        (59.52,62.64),
    ], '-->', 'steelblue')]

def do_plot():
    f,ax = plt.subplots(2,2, figsize=(8,6))
    for name, entries, style, color in n4bs100:
        throughput = []
        latency = []
        for t, l in entries:
            throughput.append(t)
            latency.append(l)
        ax[0][0].plot(throughput, latency, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[0][0].title.set_text('4/100')
        ax[0][0].set_xlim([0,70])
        ax[0][0].set_ylim([0,50])
#         ax[0][0].legend(loc='best', fancybox=True,frameon=False,framealpha=0.8)
#         ax[0][0].legend(bbox_to_anchor=(0, 1, 1, 0), loc="lower left", mode="expand", ncol=2)
    for name, entries, style, color in n4bs400:
        throughput = []
        latency = []
        for t, l in entries:
            throughput.append(t)
            latency.append(l)
        ax[1][0].plot(throughput, latency, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[1][0].title.set_text('4/400')
        ax[1][0].set_xlim([10,140])
    for name, entries, style, color in n8bs100:
        throughput = []
        latency = []
        for t, l in entries:
            throughput.append(t)
            latency.append(l)
        ax[0][1].plot(throughput, latency, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[0][1].title.set_text('8/100')
    for name, entries, style, color in n8bs400:
        throughput = []
        latency = []
        for t, l in entries:
            throughput.append(t)
            latency.append(l)
        ax[1][1].plot(throughput, latency, style, color=color, label='%s' % name, markersize=8, alpha=0.8)
        ax[1][1].title.set_text('8/400')
    ax[0][0].grid(linestyle='--', alpha=0.3)
    ax[1][0].grid(linestyle='--', alpha=0.3)
    ax[0][1].grid(linestyle='--', alpha=0.3)
    ax[1][1].grid(linestyle='--', alpha=0.3)
    f.text(0.5, 0.04, 'Throughput tps', ha='center', va='center')
    plt.subplots_adjust(wspace=0.2)
    plt.subplots_adjust(hspace=0.3)
#     plt.legend(bbox_to_anchor=(0, 1, 1, 0), loc="lower left", mode="expand", ncol=2)
    plt.savefig('model-implementation.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
