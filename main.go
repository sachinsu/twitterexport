package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/sachinsu/twexport/twitter"
)

const (
	exitFail = 1
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // second signal, hard exit
		os.Exit(exitCodeInterrupt)
	}()

	if err := run(ctx, os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}

func run(ctx context.Context, args []string, stdout io.Writer) error {

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

	return twitter.SendMessages(ctx, args)
}
