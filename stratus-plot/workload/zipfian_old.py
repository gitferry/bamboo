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
plt.rc('legend', fontsize=BIGGER_SIZE)    # legend fontsize

# s=1.01, v=5
data1 = [64246,53529,45437,40505,35474,32199,29379,26662,24545,22872,21524,19897,18619,17641,16857,16011,15281,14491,13701,13377,12643,12205,11784,11381,10843,10578,10278,10208,9502,9494,9043,8799,8530,8284,8099,7823,7859,7701,7428,7045,7165,6754,6637,6648,6515,6335,6029,6189,6052,5879,5700,5624,5335,5506,5380,5182,5087,5159,4976,4848,4839,4681,4724,4617,4571,4538,4566,4402,4269,4182,4361,4068,4055,3973,3847,3993,3858,3857,3790,3621,3625,3661,3651,3612,3575,3472,3457,3385,3428,3345,3292,3186,3134,3184,3131,3065,3168,3087,3036]

# s=1.01, v=1
data2 = [196116,97756,64886,48887,38326,32322,27867,23880,21445,19176,17581,15947,14693,14056,12789,12041,11196,10598,10065,9431,9208,8895,8334,7881,7652,7207,7098,6861,6582,6164,6185,6068,5775,5609,5448,5167,5119,4965,4926,4607,4602,4584,4385,4272,4250,4037,3952,3941,3891,3759,3699,3736,3695,3484,3408,3366,3332,3369,3180,3079,3077,3007,2892,2980,2923,2824,2822,2752,2746,2641,2619,2591,2562,2595,2557,2507,2480,2432,2341,2368,2282,2317,2341,2252,2191,2189,2137,2054,2048,2029,2053,2081,1939,2026,1995,2001,1927,1904,1813]

def do_plot():
    f = plt.figure(1, figsize=(7,5))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    replicaNo = range(1,100)
    x1 = np.array(data1)
    x2 = np.array(data2)
    
    ax.plot(replicaNo, x1/(x1.sum() * 1.0), label='zipfian s=1.01, v=5')
    ax.set_ylabel("Workload distribution")
    ax.set_xlabel("Replica ID")
    ax.plot(replicaNo, x2/(x2.sum() * 1.0), label='zipfian s=1.01, v=1')
    ax.legend(loc='best', fancybox=True,frameon=True,framealpha=0.3)
    plt.tight_layout()
    plt.savefig('zipfian.pdf', format='pdf')
    plt.show()

if __name__ == '__main__':
    do_plot()
