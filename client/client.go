package main

import (
	"encoding/binary"
	"flag"

	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/benchmark"
	"github.com/gitferry/zeitgeber/db"
	"github.com/gitferry/zeitgeber/identity"
)

var id = flag.String("id", "", "node id this client connects to")
var load = flag.Bool("load", false, "Load K keys into DB")

// Database implements Zeitgeber.DB interface for benchmarking
type Database struct {
	zeitgeber.Client
}

func (d *Database) Init() error {
	return nil
}

func (d *Database) Stop() error {
	return nil
}

func (d *Database) Read(k int) (int, error) {
	key := db.Key(k)
	v, err := d.Get(key)
	if len(v) == 0 {
		return 0, nil
	}
	x, _ := binary.Uvarint(v)
	return int(x), err
}

func (d *Database) Write(k, v int) error {
	key := db.Key(k)
	value := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(value, uint64(v))
	err := d.Put(key, value)
	return err
}

func main() {
	zeitgeber.Init()

	d := new(Database)
	d.Client = zeitgeber.NewHTTPClient(identity.NodeID(*id))
	b := benchmark.NewBenchmark(d)
	if *load {
		b.Load()
	} else {
		b.Run()
	}
}
