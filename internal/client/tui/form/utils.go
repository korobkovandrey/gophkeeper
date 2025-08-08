package form

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func renderInput(name string, input textinput.Model) string {
	view := input.View()
	if input.Focused() {
		if input.Err == nil {
			view = focusedStyle.Render(view)
		} else {
			view = focusedInvalidStyle.Render(view)
		}
	} else {
		if input.Err == nil {
			view = unfocusedStyle.Render(view)
		} else {
			view = unfocusedInvalidStyle.Render(view)
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Center,
		labelStyle.Render(name+":"), view,
	)
}

func requiredValidator(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("required")
	}
	return nil
}

func ccnValidator(s string) error {
	if len(s) > 16+3 {
		return fmt.Errorf("CCN is too long")
	}
	if s == "" || len(s)%5 != 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		return fmt.Errorf("CCN is invalid")
	}

	if len(s)%5 == 0 && s[len(s)-1] != ' ' {
		return fmt.Errorf("CCN must separate groups with spaces")
	}
	c := strings.ReplaceAll(s, " ", "")
	_, err := strconv.ParseInt(c, 10, 64)
	return err
}

func expValidator(s string) error {
	e := strings.ReplaceAll(s, "/", "")
	_, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return fmt.Errorf("EXP is invalid")
	}
	if len(s) >= 3 && (strings.Index(s, "/") != 2 || strings.LastIndex(s, "/") != 2) {
		return fmt.Errorf("EXP is invalid")
	}
	return nil
}

func cvvValidator(s string) error {
	_, err := strconv.ParseInt(s, 10, 64)
	return err
}
