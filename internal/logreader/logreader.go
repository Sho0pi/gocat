package logreader

import (
	"bufio"
	"context"
	"github.com/sho0pi/gocat/internal/types"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LogEntry represents a single parsed logcat entry
type LogEntry struct {
	Timestamp   time.Time
	ProcessID   int
	ThreadID    int
	LogLevel    types.LogLevel
	Tag         string
	Message     string
	ProcessName string
	IsStackLine bool
}

// LogReader parses adb logcat output into structured LogEntry objects
type LogReader struct {
	reader io.ReadCloser
	logCh  chan<- *LogEntry
	metaCh chan<- string
	errCh  chan<- error
}

// NewLogReader creates a new LogReader with the given reader
func NewLogReader(
	reader io.ReadCloser,
	logCh chan<- *LogEntry,
	metaCh chan<- string,
	errCh chan<- error,
) *LogReader {
	return &LogReader{
		reader: reader,
		logCh:  logCh,
		metaCh: metaCh,
		errCh:  errCh,
	}
}

// Start begins reading logs from the reader and sending them to the appropriate channels
func (lr *LogReader) Start(ctx context.Context) {
	defer func() {
		if err := lr.reader.Close(); err != nil {
			lr.errCh <- err
		}
	}()
	defer close(lr.logCh)
	defer close(lr.metaCh)
	defer close(lr.errCh)

	scanner := bufio.NewScanner(lr.reader)
	var currentEntry *LogEntry

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()

			// Check if it's a metadata line (like "--------- beginning of main")
			if strings.HasPrefix(line, "--------- ") {
				lr.metaCh <- line
				continue
			}

			// Try to parse the line as a log entry
			entry, err := lr.parseLine(line)
			if err != nil {
				// If we have a current entry and this line is a continuation (stack trace)
				if currentEntry != nil && strings.HasPrefix(line, "\t") {
					stackEntry := &LogEntry{
						Timestamp:   currentEntry.Timestamp,
						ProcessID:   currentEntry.ProcessID,
						ThreadID:    currentEntry.ThreadID,
						LogLevel:    currentEntry.LogLevel,
						Tag:         currentEntry.Tag,
						Message:     line,
						ProcessName: currentEntry.ProcessName,
						IsStackLine: true,
					}
					lr.logCh <- stackEntry
				} else {
					lr.errCh <- err
				}
				continue
			}

			// We successfully parsed a new log entry
			currentEntry = entry
			lr.logCh <- entry
		}
	}

	if err := scanner.Err(); err != nil {
		lr.errCh <- err
	}
}

// Regular expression for parsing a logcat line
var logEntryRegex = regexp.MustCompile(`^(\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3})\s+(\d+)\s+(\d+)\s+([VDIWEF])\s+([^:]+):\s+(.*)$`)

// parseLine attempts to parse a line from logcat into a LogEntry
func (lr *LogReader) parseLine(line string) (*LogEntry, error) {
	matches := logEntryRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil, &ParseError{Line: line}
	}

	// Parse timestamp
	// Format: MM-DD HH:MM:SS.sss
	currentYear := time.Now().Year()
	timestamp, err := time.Parse("2006-01-02 15:04:05.000", strconv.Itoa(currentYear)+"-"+matches[1])
	if err != nil {
		return nil, &ParseError{Line: line, Err: err}
	}

	// Parse process and thread IDs
	pid, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, &ParseError{Line: line, Err: err}
	}

	tid, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, &ParseError{Line: line, Err: err}
	}

	return &LogEntry{
		Timestamp:   timestamp,
		ProcessID:   pid,
		ThreadID:    tid,
		LogLevel:    types.LogLevel(matches[4]),
		Tag:         strings.TrimSpace(matches[5]),
		Message:     matches[6],
		ProcessName: "", // We don't have process name in the log format, only PID
		IsStackLine: false,
	}, nil
}

// ParseError represents an error that occurred during log parsing
type ParseError struct {
	Line string
	Err  error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return "failed to parse line: " + e.Line + ": " + e.Err.Error()
	}
	return "failed to parse line: " + e.Line
}
