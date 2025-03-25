package printer

import (
	"context"
	"fmt"
	"github.com/sho0pi/gocat/internal/logreader"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/gookit/color"
)

const maxTagLength = 25

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

func trueLen(coloredString string) int {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
	return len(re.ReplaceAllString(coloredString, ""))
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

func getTag(tag string, maxLength int) string {
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

	tag := getTag(entry.Tag, maxTagLength)
	if p.previousEntry != nil {
		if p.previousEntry.Tag == entry.Tag {
			tag = strings.Repeat(" ", maxTagLength)
		}
	}

	entryMsg := fmt.Sprintf(" %s %s", tag, entry.LogLevel.String())

	if entry.IsStackLine {
		// Stack trace lines are indented and use the same color as their parent entry
		fmt.Fprint(p.out, strings.Repeat(" ", trueLen(entryMsg)+1))
		fmt.Fprint(p.out, "    ")
		fmt.Fprintln(p.out, entry.LogLevel.Sprint(entry.Message))
		return
	}

	message := entry.Message
	if entry.LogLevel.Repr == "F" || entry.LogLevel.Repr == "E" {
		//message = entry.LogLevel.LevelStyle.Sprint(entry.Message)
		message = entry.LogLevel.Sprint(entry.Message)
	}

	fmt.Fprintf(
		p.out,
		"%s %s\n",
		entryMsg,
		message,
	)

	//// Print the log level with its specific color
	//fmt.Fprint(p.out, entry.LogLevel.String())
	//
	//// Format: TIME PID-TID/TAG LEVEL: MESSAGE
	//timestamp := entry.Timestamp.Format("15:04:05.000")
	//
	//// First print the prefix with default color
	//fmt.Fprintf(p.out, "%s %d-%d/", timestamp, entry.ProcessID, entry.ThreadID)
	//
	//// Print the tag with a special color
	//tagColor := color.New(color.FgHiBlue)
	//tagColor.Fprintf(p.out, "%s ", entry.Tag)
	//
	//// Finally print the message with the same color as the log level
	//fmt.Fprintln(p.out, entry.LogLevel.Sprint(entry.Message))

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
