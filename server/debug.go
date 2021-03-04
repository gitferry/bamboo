package main

import (
	"fmt"
	"github.com/gitferry/bamboo/config"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"time"
)

// Debug related config keys
const (
	DefaultURL          = "0.0.0.0"
	PprofEnable         = "pprof.enable"
	PprofPort           = "pprof.port"
	PprofDetailEnable   = "pprof.detail"
	PprofRecordDuration = "pprof.duration"

	MemsizeEnable = "memsize.enable"
	MemsizePort   = "memsize.port"
)

func setupDebug() {
	if config.GetConfig().Pprof {
		addr := DefaultURL + ":" + "10001"
		go func() {
			_ = http.ListenAndServe(addr, nil)
		}()
		go recordPProf(5 * time.Second)
	}
}

func recordPProf(duration time.Duration) {
	var (
		cpuProfile string
		memProfile string
		cpuFile    *os.File
		memFile    *os.File
	)

	dir := "./debug"
	exist, err := pathExists(dir)
	if err != nil {
		return
	}
	if !exist {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return
		}
	}
	cpuProfile = fmt.Sprint("./debug/cpu_", time.Now().Format("2006-01-02-15-04-05"))
	memProfile = fmt.Sprint("./debug/mem_", time.Now().Format("2006-01-02-15-04-05"))
	cpuFile, _ = os.Create(cpuProfile)
	_ = pprof.StartCPUProfile(cpuFile)
	tick := time.NewTicker(duration)

	for {
		select {
		case <-tick.C:
			pprof.StopCPUProfile()
			_ = cpuFile.Close()
			memFile, _ = os.Create(memProfile)
			_ = pprof.WriteHeapProfile(memFile)
			_ = memFile.Close()

			cpuProfile = fmt.Sprint("./debug/cpu_", time.Now().Format("2006-01-02-15-04-05"))
			memProfile = fmt.Sprint("./debug/mem_", time.Now().Format("2006-01-02-15-04-05"))
			cpuFile, _ = os.Create(cpuProfile)
			_ = pprof.StartCPUProfile(cpuFile)
		}
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
