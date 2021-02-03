package goworker

import (
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/cihub/seelog"
	"vitess.io/vitess/go/pools"
)

var (
	logger      seelog.LoggerInterface
	pool        *pools.ResourcePool
	ctx         context.Context
	initMutex   sync.Mutex
	initialized bool
)

var workerSettings WorkerSettings

type WorkerSettings struct {
	QueuesString   string
	Queues         queuesFlag
	IntervalFloat  float64
	Interval       intervalFlag
	Concurrency    int
	Connections    int
	URI            string
	Namespace      string // 不指定 namespace，默认会用 resque:
	ExitOnComplete bool
	IsStrict       bool
	UseNumber      bool
	SkipTLSVerify  bool
	TLSCertPath    string
}

func SetSettings(settings WorkerSettings) {
	workerSettings = settings
}

// Init initializes the goworker process. This will be
// called by the Work function, but may be used by programs
// that wish to access goworker functions and configuration
// without actually processing jobs.
func Init() error {
	initMutex.Lock()
	defer initMutex.Unlock()
	if !initialized {
		var err error
		logger, err = seelog.LoggerFromWriterWithMinLevel(os.Stdout, seelog.InfoLvl)
		if err != nil {
			return err
		}

		if err := flags(); err != nil {
			return err
		}
		ctx = context.Background()

		pool = newRedisPool(workerSettings.URI, workerSettings.Connections, workerSettings.Connections, time.Minute)

		initialized = true
	}
	return nil
}

// GetConn returns a connection from the goworker Redis
// connection pool. When using the pool, check in
// connections as quickly as possible, because holding a
// connection will cause concurrent worker functions to lock
// while they wait for an available connection. Expect this
// API to change drastically.
func GetConn() (*RedisConn, error) {
	resource, err := pool.Get(ctx)

	if err != nil {
		return nil, err
	}
	return resource.(*RedisConn), nil
}

// PutConn puts a connection back into the connection pool.
// Run this as soon as you finish using a connection that
// you got from GetConn. Expect this API to change
// drastically.
func PutConn(conn *RedisConn) {
	pool.Put(conn)
}

// Close cleans up resources initialized by goworker. This
// will be called by Work when cleaning up. However, if you
// are using the Init function to access goworker functions
// and configuration without processing jobs by calling
// Work, you should run this function when cleaning up. For
// example,
//
//	if err := goworker.Init(); err != nil {
//		fmt.Println("Error:", err)
//	}
//	defer goworker.Close()
func Close() {
	initMutex.Lock()
	defer initMutex.Unlock()
	if initialized {
		pool.Close()
		initialized = false
	}
}

// Work starts the goworker process. Check for errors in
// the return value. Work will take over the Go executable
// and will run until a QUIT, INT, or TERM signal is
// received, or until the queues are empty if the
// -exit-on-complete flag is set.
// 有看到 fan-in, fan-out 的影子，看起来像是 pipeline concurrency pattern
func Work() error {
	err := Init()
	if err != nil {
		return err
	}
	defer Close()

	// quit channel, 用于通知 poller 退出
	quit := signals()

	// workerSettings.Queues   : poller 只会从 workerSettings.Queues 这几个 queue 中 poll job
	// workerSettings.IsStrict : strict 模式下会按 queue 注册顺序去从每个 queue pull job，否则会打乱 queue 顺序再从乱序的 queues 去 pull job
	poller, err := newPoller(workerSettings.Queues, workerSettings.IsStrict)
	if err != nil {
		return err
	}

	// workerSettings.Interval : 并不是每次 poll 的间隔，而是当某次 poll 没有取到任何 job 时才会休眠 Interval 后再次 poll。
	jobs, err := poller.poll(time.Duration(workerSettings.Interval), quit)
	if err != nil {
		return err
	}

	var monitor sync.WaitGroup // 用于阻塞 main.main

	// workerSettings.Concurrency : worker pool 中的 worker 数量
	// id = workerID
	for id := 0; id < workerSettings.Concurrency; id++ {
		worker, err := newWorker(strconv.Itoa(id), workerSettings.Queues)
		if err != nil {
			return err
		}
		// jobs 为 read only channel
		worker.work(jobs, &monitor)
	}

	monitor.Wait()

	return nil
}
