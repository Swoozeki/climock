package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// featureItem represents a feature in the features list
type featureItem struct {
	name string
}

// endpointItem represents an endpoint in the endpoints list
type endpointItem struct {
	id              string
	method          string
	path            string
	active          bool
	defaultResponse string
	responses       []string
}

// FilterValue implements the list.Item interface
func (i featureItem) FilterValue() string {
	return i.name
}

// Title returns the title of the feature item
func (i featureItem) Title() string {
	return i.name
}

// Description returns the description of the feature item
func (i featureItem) Description() string {
	return ""
}

// FilterValue implements the list.Item interface
func (i endpointItem) FilterValue() string {
	return fmt.Sprintf("%s %s %s", i.id, i.method, i.path)
}

// Title returns the title of the endpoint item
func (i endpointItem) Title() string {
	methodStyle := lipgloss.NewStyle().
		Width(7).
		Align(lipgloss.Left)

	// Use emojis for active/inactive status
	active := "🟢"
	if !i.active {
		active = "🔴"
	}

	return fmt.Sprintf("%s %s %s",
		methodStyle.Render(i.method),
		i.path,
		active)
}

// Description returns the description of the endpoint item
func (i endpointItem) Description() string {
	var responses []string
	for _, r := range i.responses {
		if r == i.defaultResponse {
			// Make default response more obvious with ★ symbol and bold style
			defaultStyle := lipgloss.NewStyle().Bold(true)
			responses = append(responses, defaultStyle.Render("★"+r))
		} else {
			responses = append(responses, r)
		}
	}
	
	return fmt.Sprintf("[%s]", strings.Join(responses, " | "))
}