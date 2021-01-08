import matplotlib.pyplot as plt

# Measurements from responsiveness.data
dead = [
    ('HS-t50',[
        95.070, 98.324, 99.560,98.824, 98.411, 99.446, 97.400, 96.024, 96.671, 97.400,
        50.661, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        4.897, 10.737, 13.094, 12.752, 14.86, 9.964, 8.1, 4.052, 7.491, 10.971,
        10.817, 15.731, 8.282, 11.956, 4.363, 8.721, 7.715, 10.379, 10.394, 6.764,
    ], '-o', 'coral'),
    ('2CHS-t50',[
        90.898, 95.877, 93.347, 100.71, 98.724, 96.41, 99.07, 95.654, 97.21, 98.72,
        32.904, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    ], '-p', 'darkseagreen'),
    ('SL-t50',[
        72.569, 76.752, 71.941, 78.468, 72.385, 69.973, 71.332, 62.685, 57.894, 80.553,
        13.039, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    ], '-s', 'steelblue'),
    ('2CHS-t200',[
        95.488, 99.988, 99.651, 99.161, 100.12, 98.108, 96.982, 101.741, 96.49, 96.01,
        18.608, 0, 0, 3.249, 0.845, 0.38, 1.768, 2.813, 3.188, 1.754,
        0.922, 2.563, 2.728, 4.48, 1.871, 4.525, 4.629, 3.033, 2.353, 6.672,
        5.451, 0, 0.859, 4.683, 2.662, 4.531, 1.753, 2.522, 5.422, 3.421,
    ], '-h', 'mediumseagreen'),
    ('SL-t200',[
        66.32, 73.61, 68.44, 72.45, 63.72, 59.04, 77.80, 76.48, 66.78, 70.2,
        18.853, 0, 2.747, 0, 2.633, 1.544, 1.211, 2.778, 4.337, 2.838,
        3.475, 5.924, 5.603, 6.13, 2.179, 6.628, 2.179, 6.628, 1.816, 4.736,
        2.853, 6.611, 0, 9.468, 2.916, 5.094, 2.717, 2.496, 7.421, 3.423,
    ], '-*', 'lightskyblue')]



def do_plot():
    f = plt.figure(1, figsize=(7,5))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    time = xrange(1,41)
    for name, entries, style, color in dead:
        ax.plot(time, entries, style, color=color, label='%s' % name, markersize=6, alpha=0.6)
    plt.legend(fancybox=True,frameon=False,framealpha=0.8,loc='best')
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylabel('Throughput (KTx/s)')
    plt.xlabel('Time (s)')
    plt.axvline(x=10,ls="--",c="black")
    plt.axvline(x=20,ls="--",c="black")
    plt.text(15,90,'Fault Injection', horizontalalignment='center', verticalalignment='center', fontsize=14)
    plt.tight_layout()
    plt.savefig('responsiveness.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
