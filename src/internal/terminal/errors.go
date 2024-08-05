package terminal

import (
	"desktop-cleaner/internal/utils"
	"fmt"
	"os"
	"strings"
)

func (t *Terminal) OutputSimpleError(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintln(os.Stderr, ColorHiGreen.Bold(true).Render("🚨 "+utils.Capitalize(msg)))
}

func (t *Terminal) OutputErrorAndExit(msg string, args ...interface{}) {
	t.ToggleSpinner(false, "")
	msg = fmt.Sprintf(msg, args...)

	displayMsg := ""
	errorParts := strings.Split(msg, ": ")

	addedErrors := map[string]bool{}

	if len(errorParts) > 1 {
		var lastPart string
		i := 0
		for _, part := range errorParts {
			// don't repeat the same error message
			if _, ok := addedErrors[strings.ToLower(part)]; ok {
				continue
			}

			if len(lastPart) < 10 && i > 0 {
				lastPart = lastPart + ": " + part
				displayMsg += ": " + part
				addedErrors[strings.ToLower(lastPart)] = true
				addedErrors[strings.ToLower(part)] = true
				continue
			}

			if i != 0 {
				displayMsg += "\n"
			}

			// indent the error message
			for n := 0; n < i; n++ {
				displayMsg += "  "
			}
			if i > 0 {
				displayMsg += "→ "
			}

			s := utils.Capitalize(part)
			if i == 0 {
				s = ColorHiRed.Bold(true).Render("🚨 " + s)
			}

			displayMsg += s

			addedErrors[strings.ToLower(part)] = true
			lastPart = part
			i++
		}
	} else {
		displayMsg = ColorHiRed.Bold(true).Render("🚨 " + msg)
	}

	fmt.Fprintln(os.Stderr, ColorHiRed.Bold(true).Render(displayMsg))
	os.Exit(1)
}

func (t *Terminal) OutputUnformattedErrorAndExit(msg string) {
	t.ToggleSpinner(false, "")
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func (t *Terminal) OutputInfo(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintln(os.Stdout, ColorHiBlue.Render("🔵 "+utils.Capitalize(msg)))
}

func (t *Terminal) OutputSuccess(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintln(os.Stdout, ColorHiGreen.Render("✅ "+utils.Capitalize(msg)))
}

func (t *Terminal) OutputWarning(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	fmt.Fprintln(os.Stdout, ColorHiYellow.Render("⚠️ "+utils.Capitalize(msg)))
}

func (t *Terminal) ConfirmYesNo(msg string) bool {
	fmt.Print(ColorHiBlue.Render("🔵 " + utils.Capitalize(msg) + " (y/n): "))
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(response)
	return response == "y" || response == "yes"
}
