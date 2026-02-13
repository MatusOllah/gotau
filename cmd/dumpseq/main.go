package main

import (
	"fmt"
	"os"
	"time"

	"github.com/SladkyCitron/gotau/sequence/ust"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <input ust>\n", os.Args[0])
		return
	}

	before := time.Now()

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	ustFile, err := ust.Decode(f)
	if err != nil {
		panic(err)
	}

	seq := ustFile.Sequence()
	after := time.Since(before)
	spew.Fdump(os.Stdout, seq)
	fmt.Fprintf(os.Stderr, "took %v\n", after)
}
