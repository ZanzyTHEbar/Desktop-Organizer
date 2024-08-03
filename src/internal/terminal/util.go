package terminal

import (
	errsx "desktop-cleaner/internal/errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

func (t *Terminal) AlternateScreen() {
	fmt.Print("\x1b[?1049h\x1b[?25l")
}

func (t *Terminal) ClearScreen() {
	fmt.Print("\x1b[2J")
}

func (t *Terminal) ClearCurrentLine() {
	fmt.Print("\033[2K")
	fmt.Print("\033[0G")
}

func (t *Terminal) MoveCursorToTopLeft() {
	fmt.Print("\033[H")
}

func (t *Terminal) MoveCursorUpLines(numLines int) {
	fmt.Printf("\033[%dA", numLines)
}

func (t *Terminal) BackToMainScreen() {
	fmt.Print("\x1b[?1049l\x1b[?25h")
}

func (t *Terminal) PageOutput(output string, reverse bool) {

	var reverseFlag string
	if reverse {
		reverseFlag = "+G"
	}

	cmd := exec.Command("less", "-R", reverseFlag)
	cmd.Env = append(os.Environ(), "LESS=FRX", "LESSCHARSET=utf-8")
	cmd.Stdin = strings.NewReader(output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		errBuilder := errsx.GenericErr(ColorHiRed.Bold(true).Render("Failed to page output"), err)
		t.OutputErrorAndExit(errBuilder.Error())
	}
}

func (t *Terminal) GetDivisionLine() string {
	terminalWidth, err := t.getTerminalWidth()
	if err != nil {
		errsBuilder := errsx.GenericErr(ColorHiRed.Bold(true).Render("Failed to get terminal width"), err)
		slog.Error(errsBuilder.Error())
		terminalWidth = 50
	}

	return strings.Repeat("─", terminalWidth)
}

func (t *Terminal) getTerminalWidth() (int, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, err
	}

	return width, nil
}
