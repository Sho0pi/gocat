package cmd

import (
	"context"
	"fmt"
	"github.com/sho0pi/gocat/cmd/version"
	"github.com/sho0pi/gocat/internal/filter"
	"github.com/sho0pi/gocat/internal/logreader"
	"github.com/sho0pi/gocat/internal/printer"
	"github.com/sho0pi/gocat/internal/types"
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
	ignoreTags   []string
	minLevel     types.LogLevel
	processNames []string
	showTime     bool
	dump         bool
	clear        bool
}

func newGocatCommand() *cobra.Command {
	opts := &gocatOptions{
		minLevel: types.LevelVerbose,
	}
	cmd := &cobra.Command{
		Use:   "gocat",
		Short: "A beautiful logcat wrapper for Android",
		Long: `Gocat is a CLI tool that wraps adb logcat with filtering and colorful output.
It can parse logs either from an input file or directly from adb logcat.`,
		Version: version.Version,
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

			logFilter := filter.NewLogFilter(logCh, filteredCh, opts.tags, opts.minLevel, opts.processNames)
			go logFilter.Start(ctx)

			// Create and start the Printer
			logPrinter := printer.NewPrinter(filteredCh, metaCh, errCh, cmd.OutOrStdout(), cmd.ErrOrStderr(), opts.showTime)
			if err := logPrinter.Start(ctx); err != nil {
				return fmt.Errorf("error from printer: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	cmd.SetVersionTemplate("GoCat version {{.Version}}\n")

	cmd.Flags().VarP(&opts.minLevel, "min-level", "l", "Minimum level to be displayed")
	cmd.Flags().BoolVar(&opts.showTime, "show-time", false, "Show times")
	cmd.Flags().BoolVarP(&opts.dump, "dump", "d", false, "Dump the log and then exit (don't block).")
	cmd.Flags().BoolVarP(&opts.clear, "clear", "c", false, "Clear the entire log before running")
	cmd.Flags().StringSliceVarP(&opts.tags, "tags", "t", []string{}, "Filter output by specified tag(s)")
	cmd.Flags().StringSliceVarP(&opts.ignoreTags, "ignore-tags", "i", []string{}, "Filter output by ignoring specified tag(s)")

	return cmd
}

func Execute() {
	cmd := newGocatCommand()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
