package cmd

import (
	"context"
	"fmt"
	"github.com/sho0pi/gocat/internal/filter"
	"github.com/sho0pi/gocat/internal/logreader"
	"github.com/sho0pi/gocat/internal/printer"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"os/exec"
)

const (
	logChanSize     = 1000
	regularChanSize = 10
)

type gocatOptions struct {
	tags         []string
	logLevels    []string
	processNames []string
	showTime     bool
}

func newGocatCommand() *cobra.Command {
	opts := &gocatOptions{}
	cmd := &cobra.Command{
		Use:   "gocat",
		Short: "A beautiful logcat wrapper for Android",
		Long: `Gocat is a CLI tool that wraps adb logcat with filtering and colorful output.
It can parse logs either from an input file or directly from adb logcat.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var reader io.ReadCloser
			var err error

			info, err := os.Stdin.Stat()
			if err != nil {
				return fmt.Errorf("os.Stdin.Stat() failed: %v", err)
			}

			if info.Mode()&os.ModeCharDevice == 0 {
				reader = os.Stdin
			} else {
				adbCmd := exec.CommandContext(ctx, "adb", "logcat")
				reader, err = adbCmd.StdoutPipe()
				if err != nil {
					return fmt.Errorf("adbCmd.StdoutPipe() failed: %w", err)
				}

				if err := adbCmd.Start(); err != nil {
					return fmt.Errorf("adbCmd.Start() failed: %w", err)
				}

				defer func() {
					if adbCmd.Process != nil {
						adbCmd.Process.Kill()
					}
				}()
			}

			logCh := make(chan *logreader.LogEntry, logChanSize)
			filteredCh := make(chan *logreader.LogEntry, logChanSize)
			metaCh := make(chan string, regularChanSize)
			errCh := make(chan error, regularChanSize)

			logReader := logreader.NewLogReader(reader, logCh, metaCh, errCh)
			go logReader.Start(ctx)

			logFilter := filter.NewLogFilter(logCh, filteredCh, opts.tags, opts.logLevels, opts.processNames)
			go logFilter.Start(ctx)

			// Create and start the Printer
			logPrinter := printer.NewPrinter(filteredCh, metaCh, errCh, cmd.OutOrStdout(), cmd.ErrOrStderr())
			if err := logPrinter.Start(ctx); err != nil {
				return fmt.Errorf("error from printer: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func Execute() {
	cmd := newGocatCommand()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
