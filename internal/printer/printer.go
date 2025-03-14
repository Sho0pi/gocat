package printer

import (
	"context"
	"fmt"
	"github.com/sho0pi/gocat/internal/logreader"
	"io"
	"sync"

	"github.com/fatih/color"
)

// Printer is responsible for displaying log entries with color formatting
type Printer struct {
	logCh  <-chan *logreader.LogEntry
	metaCh <-chan string
	errCh  <-chan error
	out    io.Writer
	errOut io.Writer
	mu     sync.Mutex // Protects concurrent writes to output
}

// NewPrinter creates a new Printer with default settings
func NewPrinter(
	logCh <-chan *logreader.LogEntry,
	metaCh <-chan string,
	errCh <-chan error,
	out io.Writer,
	errOut io.Writer,
) *Printer {
	return &Printer{
		logCh:  logCh,
		metaCh: metaCh,
		errCh:  errCh,
		out:    out,
		errOut: errOut,
	}
}

// Start begins printing log entries from the channels until the context is cancelled
func (p *Printer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case entry, ok := <-p.logCh:
			if !ok {
				return nil
			}
			p.printLogEntry(entry)
		case metaLine, ok := <-p.metaCh:
			if !ok {
				continue
			}
			p.printMetadata(metaLine)
		case err, ok := <-p.errCh:
			if !ok {
				continue
			}
			p.printError(err)
		}
	}
}

// printLogEntry formats and prints a log entry with appropriate colors
func (p *Printer) printLogEntry(entry *logreader.LogEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if entry.IsStackLine {
		// Stack trace lines are indented and use the same color as their parent entry
		c := entry.LogLevel.Color()
		msg := c.Sprint(entry.Message)
		fmt.Fprintln(p.out, msg)
		//c.Fprintln(p.out, entry.Message)
		return
	}

	fmt.Fprint(p.out, entry.LogLevel.String())

	// Format: TIME PID-TID/TAG LEVEL: MESSAGE
	timestamp := entry.Timestamp.Format("15:04:05.000")

	// First print the prefix with default color
	fmt.Fprintf(p.out, "%s %d-%d/", timestamp, entry.ProcessID, entry.ThreadID)

	// Print the tag with a special color
	tagColor := color.New(color.FgHiBlue)
	tagColor.Fprintf(p.out, "%s ", entry.Tag)

	// Print the log level with its specific color
	//logLevelColor.Fprintf(p.out, "%s: ", entry.LogLevel)

	// Finally print the message with the same color as the log level
	fmt.Fprintln(p.out, entry.LogLevel.Sprint(entry.Message))
	//logLevelColor.Fprintln(p.out, entry.Message)
}

// printMetadata prints metadata lines with a special format
func (p *Printer) printMetadata(metaLine string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	metaColor := color.New(color.FgHiMagenta, color.Bold)
	metaColor.Fprintln(p.out, metaLine)
}

// printError prints error messages to stderr
func (p *Printer) printError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	errColor := color.New(color.FgHiRed)
	errColor.Fprintln(p.errOut, "ERROR:", err.Error())
}
