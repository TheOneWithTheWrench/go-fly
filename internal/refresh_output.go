package internal

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

type refreshOutput struct {
	writer   io.Writer
	file     *os.File
	terminal bool

	mu        sync.Mutex
	indicator *spinner.Spinner
}

func NewRefreshOutput(writer io.Writer) RefreshOutput {
	if writer == nil {
		writer = io.Discard
	}

	file, _ := writer.(*os.File)
	return &refreshOutput{
		writer:   writer,
		file:     file,
		terminal: isTerminalWriter(writer),
	}
}

func (o *refreshOutput) Write(p []byte) (int, error) {
	if !o.terminal {
		return o.writer.Write(p)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.stopSpinnerLocked()
	return o.writer.Write(p)
}

func (o *refreshOutput) SetStatus(status string) error {
	if !o.terminal {
		_, err := fmt.Fprintln(o.writer, status)
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if status == "" {
		o.stopSpinnerLocked()
		return nil
	}

	if o.indicator == nil {
		o.indicator = spinner.New(
			spinner.CharSets[9],
			100*time.Millisecond,
			spinner.WithHiddenCursor(false),
			spinner.WithWriterFile(o.file),
		)
		o.indicator.Suffix = " " + status
		o.indicator.Start()
		return nil
	}

	o.indicator.Lock()
	o.indicator.Suffix = " " + status
	o.indicator.Unlock()

	return nil
}

func (o *refreshOutput) ClearStatus() error {
	if !o.terminal {
		return nil
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.stopSpinnerLocked()
	return nil
}

func (o *refreshOutput) stopSpinnerLocked() {
	if o.indicator == nil {
		return
	}

	o.indicator.Stop()
	o.indicator = nil
}

func isTerminalWriter(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
