package terminal

import (
	"time"

	"github.com/briandowns/spinner"
)

const withMessageMinDuration = 700 * time.Millisecond
const withoutMessageMinDuration = 350 * time.Millisecond

var spnr = spinner.New(spinner.CharSets[33], 100*time.Millisecond)
var startedAt time.Time

var lastMessage string
var active bool

func (t *Terminal) startSpinner(msg string) {
	if active {
		if msg == lastMessage {
			return
		}

		spnr.Stop()
	}

	startedAt = time.Now()
	spnr.Prefix = msg + " "
	lastMessage = msg
	spnr.Start()
	active = true
}

func (t *Terminal) stopSpinner() {
	elapsed := time.Since(startedAt)

	if lastMessage != "" && elapsed < withMessageMinDuration {
		time.Sleep(withMessageMinDuration - elapsed)
	} else if elapsed < withoutMessageMinDuration {
		time.Sleep(withoutMessageMinDuration - elapsed)
	}

	spnr.Stop()
	t.ClearCurrentLine()

	active = false
}

func (t *Terminal) ToggleSpinner(toggle bool, msg string) {
	if !toggle {
		t.stopSpinner()
		return
	}

	t.startSpinner(msg)
}

func (t *Terminal) ResumeSpinner() {
	if !active {
		t.startSpinner(lastMessage)
	}
}
