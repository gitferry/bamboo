import matplotlib.pyplot as plt

# Measurements from responsiveness.data
dead = [
    ('HS-t10',[
        95.070, 98.324, 99.560,98.824, 98.411, 99.446, 97.400, 96.024, 96.671, 97.400,
        50.661, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        6.317, 27.762, 31.09, 33.496, 40.484, 29.439, 37.503, 28.894, 27.449, 34.587,
        28.168, 36.574, 38.551, 26.734, 38.276, 30.37, 36.025, 37.866, 39.56, 36.533,
#         4.897, 10.737, 13.094, 12.752, 14.86, 9.964, 8.1, 4.052, 7.491, 10.971,
#         10.817, 15.731, 8.282, 11.956, 4.363, 8.721, 7.715, 10.379, 10.394, 6.764,
    ], '-o', 'coral'),
    ('2CHS-t10',[
        90.898, 95.877, 93.347, 100.71, 98.724, 96.41, 99.07, 95.654, 97.21, 98.72,
        32.904, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    ], '-p', 'darkseagreen'),
    ('SL-t10',[
        72.569, 76.752, 71.941, 78.468, 72.385, 69.973, 71.332, 62.685, 57.894, 80.553,
        13.039, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    ], '-s', 'steelblue'),
    ('HS-t100',[
        94.020, 102.38, 100.97,102.47, 100.25, 98.09, 97.78, 99.33, 99.509, 98.573,
        16.066, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6.033, 11.146, 2.659, 10.595, 11.537,
        7.285, 1.605, 6.35, 4.306, 6.259, 5.855, 11.159, 8.71, 6.970, 8.055,
        7.575, 8.013, 7.213, 6.624, 3.852,
    ], '-d', 'orange'),
    ('2CHS-t100',[
        95.488, 99.988, 99.651, 99.161, 100.12, 98.108, 96.982, 101.741, 96.49, 96.01,
        16.608, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5.68, 7.044, 7.30, 1.324,
        10.842, 4.896, 8.366, 2.412, 3.213, 3.691, 2.629, 4.824, 0.612, 4.108,
        2.028, 4.526, 5.176, 3.639, 4.874,
    ], '-h', 'mediumseagreen'),
    ('SL-t100',[
        66.32, 73.61, 68.44, 72.45, 63.72, 59.04, 77.80, 76.48, 66.78, 70.2,
        18.853, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1.544, 1.211, 2.778, 4.337, 2.838,
        3.475, 5.924, 5.603, 6.13, 2.179, 6.628, 2.179, 6.628, 1.816, 4.736,
        2.853, 6.611, 0, 9.468, 2.916,
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
    plt.annotate("",
                xy=(30, 60), xycoords='data',
                xytext=(20, 60), textcoords='data',
                arrowprops=dict(arrowstyle="->",
                                connectionstyle="arc3"),
                )
    plt.text(25, 55, "Silence Attack", ha="center", va="center", fontsize=14)
    plt.text(15,90,'Network\nFluctuation', horizontalalignment='center', verticalalignment='center', fontsize=14)
    plt.tight_layout()
    plt.savefig('responsiveness.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
