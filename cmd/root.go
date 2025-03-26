package cmd

import (
	"context"
	"fmt"
	"github.com/sho0pi/gocat/cmd/version"
	"github.com/sho0pi/gocat/internal/completion"
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
	processNames []string
	minLevel     types.LogLevel
	showTime     bool
	dump         bool
	clear        bool
}

func newGocatCommand() *cobra.Command {
	opts := &gocatOptions{
		minLevel: types.VerboseLevel,
	}
	cmd := &cobra.Command{
		Use:   "gocat",
		Short: "A beautiful logcat wrapper for Android",
		Long: `Gocat is a powerful CLI tool that enhances Android's logcat experience.

It provides advanced log filtering, real-time parsing, and colorful output 
for Android developers. Gocat supports multiple log sources including 
direct ADB logcat, log files, and piped input.`,
		Example: `
  # Filter logs by specific tags
  gocat -t "NetworkManager" -t "DatabaseHelper" -l warn

  # Ignore certain log tags
  gocat --ignore-tags "AndroidRuntime" --ignore-tags "System"

  # Dump logs and exit immediately
  gocat -d

  # Read logs from a file or pipe
  cat logfile.log | gocat`,

		Version:           version.Version,
		ValidArgsFunction: noArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			reader, err := getLogReader(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to get log reader: %w", err)
			}
			defer reader.Close()

			logCh := make(chan *logreader.LogEntry, logChanSize)
			filteredCh := make(chan *logreader.LogEntry, logChanSize)
			metaCh := make(chan string, regularChanSize)
			errCh := make(chan error, regularChanSize)

			logReader := logreader.NewLogReader(reader, logCh, metaCh, errCh)
			go logReader.Start(ctx)

			logFilter := filter.NewLogFilter(logCh, filteredCh, opts.tags, opts.ignoreTags, opts.minLevel, opts.processNames)
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
	cmd.SetVersionTemplate("Gocat version {{.Version}}\n")

	cmd.Flags().BoolVarP(&opts.dump, "dump", "d", false, "Capture and display logs, then exit without blocking")
	cmd.Flags().BoolVarP(&opts.clear, "clear", "c", false, "Clear existing Android device logs before capturing")

	cmd.Flags().StringSliceVarP(&opts.tags, "tags", "t", []string{}, "Filter output by specified tag(s)")
	cmd.Flags().StringSliceVarP(&opts.ignoreTags, "ignore-tags", "i", []string{}, "Exclude logs with specified tag(s)")
	cmd.Flags().VarP(&opts.minLevel, "min-level", "l", "Set minimum log level for display (verbose, debug, info, warn, error, fatal)")
	_ = cmd.RegisterFlagCompletionFunc("min-level", completion.LogLevels())
	cmd.Flags().StringSliceVar(&opts.processNames, "process-name", []string{}, "Filter logs by process name(s)")
	_ = cmd.Flags().MarkHidden("process-name")
	_ = cmd.RegisterFlagCompletionFunc("process-name", completion.RunningProcesses())

	return cmd
}

func getLogReader(ctx context.Context, opts *gocatOptions) (io.ReadCloser, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not check stdin: %w", err)
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		return os.Stdin, nil
	}

	if opts.clear {
		if err := exec.CommandContext(ctx, "adb", "logcat", "-c").Run(); err != nil {
			return nil, fmt.Errorf("failed to clear logcat: %w", err)
		}
	}

	adbArgs := []string{"logcat"}
	if opts.dump {
		adbArgs = append(adbArgs, "-d")
	}

	adbCmd := exec.CommandContext(ctx, "adb", adbArgs...)
	reader, err := adbCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not get adb stdout pipe: %w", err)
	}

	if err := adbCmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start adb: %w", err)
	}

	// Ensure adb process is killed when we're done
	go func() {
		<-ctx.Done()
		if adbCmd.Process != nil {
			adbCmd.Process.Kill()
		}
	}()

	return reader, nil
}

func noArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func Execute() {
	cmd := newGocatCommand()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
