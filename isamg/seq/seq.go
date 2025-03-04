package seq

import (
	"io"
	"os"
	"strconv"
	"sync/atomic"
)

// Sequences
var customer_seq atomic.Uint32
var contact_seq atomic.Uint32

func Initialize(seqName string, path string) {
	//retrieve stored values from sequence files (ALL)
	fpath := path + seqName + ".seq"
	data, err := os.ReadFile(fpath)
	if err != nil && err != io.EOF {
		panic(err)
	}
	intVal, err := strconv.ParseInt(string(data), 10, 32)
	if err != nil {
		panic(err)
	}
	switch seqName {
	case "contact":
		contact_seq.Store(uint32(intVal))
	case "customer":
		customer_seq.Store(uint32(intVal))
	}

}
func Next(seqName string, path string) int {
	var nextSeq int
	fpath := path + seqName + ".seq"
	switch seqName {
	case "contact":
		contact_seq.Add(1)
		save(fpath, int(contact_seq.Load()))
		nextSeq = int(contact_seq.Load())
	case "customer":
		customer_seq.Add(1)
		save(fpath, int(customer_seq.Load()))
		nextSeq = int(customer_seq.Load())
	}
	//return int(customer_seq.Load())
	return nextSeq
}
func save(fpath string, seq int) {
	data := []byte(strconv.Itoa(seq))
	err := os.WriteFile(fpath, data, 0777)
	if err != nil {
		panic(err)
	}
}
