package main

import (
	"flag"
	"fmt"
	"github.com/sachinsu/twexport/twitter"
	"io"
	"os"
	"runtime/pprof"
)

const (
	exitFail = 1
)

func main() {
	if err := run(os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}

func run(args []string, stdout io.Writer) error {

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)

	// provider := flags.String("smprovider", "twitter", "which Social Media provider to target")

	cpuprofile := flags.String("cpuprofile", "", "write cpu profile to file")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "can't create profiler: %s", err.Error())
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "can't start profiler: %s", err.Error())
		}
		defer pprof.StopCPUProfile()
	}

	return twitter.SendMessages(args)
}
