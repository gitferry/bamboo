package main

import (
	"encoding/binary"
	"flag"

	"github.com/gitferry/bamboo"
	"github.com/gitferry/bamboo/benchmark"
	"github.com/gitferry/bamboo/db"
)

var load = flag.Bool("load", false, "Load K keys into DB")

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

func (d *Database) Write(k, v int) error {
	key := db.Key(k)
	value := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(value, uint64(v))
	err := d.Put(key, value)
	return err
}

func main() {
	bamboo.Init()

	d := new(Database)
	d.Client = bamboo.NewHTTPClient()
	b := benchmark.NewBenchmark(d)
	b.Run()
}
