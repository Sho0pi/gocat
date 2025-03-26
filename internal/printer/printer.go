package printer

import (
	"context"
	"fmt"
	"github.com/sho0pi/gocat/internal/logreader"
	"io"
	"strings"
	"sync"

	"github.com/gookit/color"
)

const maxTagLength = 25
const stacktracePrefixLength = maxTagLength + 5

var tagColors = []*color.RGBStyle{
	color.HEXStyle("#FFFFFF"),
	color.HEXStyle("#ff785f"),
	color.HEXStyle("#ffd063"),
	color.HEXStyle("#00dbb4"),
	color.HEXStyle("#00bcc4"),
	color.HEXStyle("#856bf5"),
	color.HEXStyle("#ce8ff9"),
	color.HEXStyle("#ff65bf"),
}

// Printer is responsible for displaying log entries with color formatting
type Printer struct {
	logCh         <-chan *logreader.LogEntry
	metaCh        <-chan string
	errCh         <-chan error
	out           io.Writer
	errOut        io.Writer
	showTime      bool
	mu            sync.Mutex // Protects concurrent writes to output
	previousEntry *logreader.LogEntry
}

// NewPrinter creates a new Printer with default settings
func NewPrinter(
	logCh <-chan *logreader.LogEntry,
	metaCh <-chan string,
	errCh <-chan error,
	out io.Writer,
	errOut io.Writer,
	showTime bool,
) *Printer {
	return &Printer{
		logCh:    logCh,
		metaCh:   metaCh,
		errCh:    errCh,
		out:      out,
		showTime: showTime,
		errOut:   errOut,
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

func (p *Printer) sprintTag(tag string, maxLength int) string {
	if p.previousEntry != nil && p.previousEntry.Tag == tag {
		return strings.Repeat(" ", maxLength)
	}
	chosenColor := tagColors[len(tag)%len(tagColors)]
	if len(tag) > maxLength {
		tag = tag[:maxLength-1] + "…"
	} else {
		tag = strings.Repeat(" ", maxLength-len(tag)) + tag
	}

	return chosenColor.Sprint(tag)
}

// printLogEntry formats and prints a log entry with appropriate tagColors
func (p *Printer) printLogEntry(entry *logreader.LogEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if entry.IsStackLine {
		// Stack trace lines are indented and use the same color as their parent entry
		fmt.Printf(
			"%s %s\n",
			strings.Repeat(" ", stacktracePrefixLength),
			entry.LogLevel.Sprint(entry.Message),
		)
		return
	}

	logPrefix := fmt.Sprintf(
		"%s %s",
		p.sprintTag(entry.Tag, maxTagLength),
		entry.LogLevel.Pretty(),
	)
	message := entry.SprintMessage()

	fmt.Fprintf(
		p.out,
		"%s %s\n",
		logPrefix,
		message,
	)

	p.previousEntry = entry
}

// printMetadata prints metadata lines with a special format
func (p *Printer) printMetadata(metaLine string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	metaColor := color.New(color.FgMagenta, color.Bold)
	fmt.Fprintln(
		p.out,
		metaColor.Sprint(metaLine),
	)

	p.previousEntry = nil
}

// printError prints error messages to stderr
func (p *Printer) printError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Fprintln(
		p.errOut,
		color.FgRed.Sprintf("ERROR: %s", err.Error()),
	)

	p.previousEntry = nil
}
