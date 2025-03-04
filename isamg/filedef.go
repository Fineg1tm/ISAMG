package isamg

import (
	"os"
)

// "errors"
type FDefs map[string]*FileDef

// Master and index file definition record
type FileDef struct {
	Name  string
	Path  string
	Rec   interface{}
	LRecl int
	Keyl  int      // key length
	Keyo  int      // key offset, byte count from start of master record
	Keyu  bool     // key unique or non-unique
	Rridl int      // length of Relative Record ID
	FD    *os.File // opened files for processing
}

// Grouping of related application files' definitions
type FileDefs []FileDef
