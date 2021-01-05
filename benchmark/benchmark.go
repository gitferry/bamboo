package benchmark

import (
	"math/rand"
	"sync"
	"time"

	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/log"
)

var count uint64

// DB is general interface implemented by client to call client library
type DB interface {
	Init() error
	Write(key int, value []byte) error
	Stop() error
}

// DefaultBConfig returns a default benchmark config
func DefaultBConfig() config.Bconfig {
	return config.Bconfig{
		T:           60,
		N:           0,
		Throttle:    0,
		Concurrency: 1,
	}
}

// Benchmark is benchmarking tool that generates workload and collects operation history and latency
type Benchmark struct {
	db DB // read/write operation interface
	config.Bconfig
	*History

	rate      *Limiter
	latency   []time.Duration // latency per operation
	startTime time.Time
	counter   int

	wait sync.WaitGroup // waiting for all generated keys to complete
}

// NewBenchmark returns new Benchmark object given implementation of DB interface
func NewBenchmark(db DB) *Benchmark {
	b := new(Benchmark)
	b.db = db
	b.Bconfig = config.Configuration.Benchmark
	b.History = NewHistory()
	if b.Throttle > 0 {
		b.rate = NewLimiter(b.Throttle)
	}
	return b
}

// Run starts the main logic of benchmarking
func (b *Benchmark) Run() {
	var genCount, sendCount, confirmCount uint64

	b.latency = make([]time.Duration, 0)
	keys := make(chan int, b.Concurrency)
	latencies := make(chan time.Duration, 1000)
	defer close(latencies)
	go b.collect(latencies)

	for i := 0; i < b.Concurrency; i++ {
		go b.worker(keys, latencies)
	}

	b.db.Init()
	b.startTime = time.Now()
	if b.T > 0 {
		timer := time.NewTimer(time.Second * time.Duration(b.T))
	loop:
		for {
			select {
			case <-timer.C:
				log.Infof("Benchmark stops")
				break loop
			default:
				b.wait.Add(1)
				//log.Debugf("is generating key No.%v", j)
				k := b.next()
				genCount++
				keys <- k
				sendCount++
				//log.Debugf("generated key No.%v", j-1)
			}
		}
	} else {
		for i := 0; i < b.N; i++ {
			b.wait.Add(1)
			keys <- b.next()
		}
		b.wait.Wait()
	}

	t := time.Now().Sub(b.startTime)

	b.db.Stop()
	close(keys)
	stat := Statistic(b.latency)
	confirmCount = uint64(len(b.latency))
	log.Infof("Concurrency = %d", b.Concurrency)
	log.Infof("Benchmark Time = %v\n", t)
	log.Infof("Throughput = %f\n", float64(len(b.latency))/t.Seconds())
	log.Infof("genCount: %d, sendCount: %d, confirmCount: %d", genCount, sendCount, confirmCount)
	log.Info(stat)

	//stat.WriteFile("latency")
	//b.History.WriteFile("history")
}

func (b *Benchmark) worker(keys <-chan int, result chan<- time.Duration) {
	//var s time.Time
	//var e time.Time
	//var v int
	//var err error
	for k := range keys {
		//op := new(operation)
		//v = rand.Int()
		//s = time.Now()
		value := make([]byte, config.GetConfig().PayloadSize)
		rand.Read(value)
		//rand.Read(value)
		_ = b.db.Write(k, value)
		//res, err := strconv.Atoi(r)
		//log.Debugf("latency is %v", time.Duration(res)*time.Nanosecond)
		//e = time.Now()
		//op.input = v
		//op.start = s.Sub(b.startTime).Nanoseconds()
		//if err == nil {
		//op.end = e.Sub(b.startTime).Nanoseconds()
		//result <- e.Sub(s)
		//result <- time.Duration(res) * time.Nanosecond
		//} else {
		//op.end = math.MaxInt64
		//log.Error(err)
		//}
		//b.History.AddOperation(k, op)
	}
}

// generates key based on distribution
func (b *Benchmark) next() int {
	var key int
	switch b.Distribution {
	case "uniform":
		key = int(count)
		count += uint64(config.GetConfig().N() - config.GetConfig().ByzNo)
	default:
		log.Fatalf("unknown distribution %s", b.Distribution)
	}

	if b.Throttle > 0 {
		b.rate.Wait()
	}

	return key
}

func (b *Benchmark) collect(latencies <-chan time.Duration) {
	for t := range latencies {
		b.latency = append(b.latency, t)
		b.wait.Done()
	}
}
