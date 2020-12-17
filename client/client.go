package main

import (
	"github.com/gitferry/bamboo"
	"github.com/gitferry/bamboo/benchmark"
	"github.com/gitferry/bamboo/db"
)

// Database implements bamboo.DB interface for benchmarking
type Database struct {
	bamboo.Client
}

func (d *Database) Init() error {
	return nil
}

func (d *Database) Stop() error {
	return nil
}

func (d *Database) Write(k int, v []byte) (string, error) {
	key := db.Key(k)
	r, err := d.Put(key, v)
	return r, err
}

func main() {
	bamboo.Init()

	d := new(Database)
	d.Client = bamboo.NewHTTPClient()
	b := benchmark.NewBenchmark(d)
	b.Run()
}
