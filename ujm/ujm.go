package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mattn/ujihisa"
)

var debug = flag.Bool("d", false, "debug")

func main() {
	flag.Parse()
	var f *os.File = os.Stdin
	var err error
	var b []byte

	switch {
	case flag.NArg() == 0:
		b, err = ioutil.ReadAll(f)
	case flag.NArg() == 1:
		b, err = ioutil.ReadFile(flag.Arg(0))
	default:
		flag.Usage()
		os.Exit(1)
	}

	if err == nil {
		vm := ujihisa.NewVM()
		vm.Debug = *debug
		err = vm.Run(string(b))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
