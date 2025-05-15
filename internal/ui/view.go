package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m *Model) View() string {
	// If a dialog is active, render it
	if m.activeDialog != NoDialog {
		return m.renderDialog()
	}

	// Render the main UI
	var sb strings.Builder

	// Header
	sb.WriteString(m.renderHeader())
	sb.WriteString("\n")

	// Lists (with their own titles)
	sb.WriteString(m.renderLists())
	sb.WriteString("\n")

	// Footer
	sb.WriteString(m.renderFooter())

	return sb.String()
}

// renderHeader renders the header
func (m *Model) renderHeader() string {
	// Use cached style with updated width but without bottom border
	headerStyle := m.styles.header.Width(m.width).BorderBottom(false)

	// Title style similar to dialog titles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	serverStatus := "Stopped"
	if m.Server.IsRunning() {
		serverStatus = fmt.Sprintf("Running (%s)", m.Server.GetAddress())
	}

	proxyTarget := m.ProxyManager.GetTargetURL()
	header := fmt.Sprintf("Server: %s | Proxy: %s", serverStatus, proxyTarget)

	return headerStyle.Render(titleStyle.Render("Climock") + " - " + header)
}

// renderLists renders the feature and endpoint lists
func (m *Model) renderLists() string {
	// Calculate widths accounting for borders (subtract border width)
	// Border takes 2 characters (1 on each side)
	featureWidth := m.width/4 - 2
	endpointWidth := 3*m.width/4 - 2
	
	// Use cached styles with adjusted widths
	featuresStyle := m.styles.features.Width(featureWidth)
	endpointsStyle := m.styles.endpoints.Width(endpointWidth)

	// Apply border styling to both panels consistently
	// Use a highlighted border for the active panel
	featuresStyle = featuresStyle.
		BorderStyle(lipgloss.RoundedBorder())
	
	endpointsStyle = endpointsStyle.
		BorderStyle(lipgloss.RoundedBorder())
	
	// Highlight the active panel with a different border color
	// Use a much lighter color (253) for inactive borders
	if m.activePanel == FeaturesPanel {
		featuresStyle = featuresStyle.
			BorderForeground(lipgloss.Color("63"))
		endpointsStyle = endpointsStyle.
			BorderForeground(lipgloss.Color("253"))
	} else {
		featuresStyle = featuresStyle.
			BorderForeground(lipgloss.Color("253"))
		endpointsStyle = endpointsStyle.
			BorderForeground(lipgloss.Color("63"))
	}

	featuresView := featuresStyle.Render(m.featuresList.View())
	endpointsView := endpointsStyle.Render(m.endpointsList.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, featuresView, endpointsView)
}

// renderFooter renders the footer
func (m *Model) renderFooter() string {
	// Use cached style with updated width but without top border
	footerStyle := m.styles.footer.Width(m.width).BorderTop(false)

	// Style for the footer content
	footerContentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Create a panel-specific keymap that only shows relevant shortcuts
	panelKeyMap := NewPanelKeyMap(m.keyMap, m.activePanel)
	
	// Get shortcuts in two rows for consistent height
	shortcutRows := panelKeyMap.ShortHelpInRows()
	
	// Create a copy of the first row to add conditional shortcuts
	row1 := make([]key.Binding, len(shortcutRows[0]))
	copy(row1, shortcutRows[0])
	
	// Check if items are available for selection
	hasFeatures := len(m.featuresList.Items()) > 0
	hasEndpoints := m.selectedFeature != "" && len(m.endpointsList.Items()) > 0
	
	// Add New option conditionally
	if m.activePanel == FeaturesPanel {
		// Always show New in features panel
		row1 = append(row1, m.keyMap.New)
	} else if m.activePanel == EndpointsPanel && m.selectedFeature != "" {
		// Only show New in endpoints panel if a feature is selected
		row1 = append(row1, m.keyMap.New)
	}
	
	// Add panel-specific actions
	if m.activePanel == EndpointsPanel && hasEndpoints {
		// Only show toggle and response options if endpoints are available
		row1 = append(row1, m.keyMap.Toggle, m.keyMap.Response)
	}
	
	// Add Open and Delete options based on selection state
	if (m.activePanel == FeaturesPanel && hasFeatures) ||
	   (m.activePanel == EndpointsPanel && hasEndpoints) {
		row1 = append(row1, m.keyMap.Open, m.keyMap.Delete)
	}
	
	// Render each row of shortcuts
	var sb strings.Builder
	
	// First row
	for i, binding := range row1 {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(binding.Help().Key)
		sb.WriteString(" ")
		sb.WriteString(binding.Help().Desc)
	}
	
	// Add a newline between rows
	sb.WriteString("\n")
	
	// Second row
	for i, binding := range shortcutRows[1] {
		if i > 0 {
			sb.WriteString("  ")
		}
		sb.WriteString(binding.Help().Key)
		sb.WriteString(" ")
		sb.WriteString(binding.Help().Desc)
	}
	
	return footerStyle.Render(footerContentStyle.Render(sb.String()))
}

// renderDialog renders the active dialog
func (m *Model) renderDialog() string {
	switch m.activeDialog {
	case HelpDialog:
		return m.renderHelpDialog()
	case NewFeatureDialog, NewEndpointDialog:
		return m.renderInputDialog()
	case DeleteConfirmDialog:
		return m.renderConfirmDialog()
	case ProxyConfigDialog:
		return m.renderInputDialog() // Reuse input dialog renderer
	default:
		// If we somehow get here with NoDialog, render the main UI
		return m.View()
	}
}


// renderHelpDialog renders the help dialog
func (m *Model) renderHelpDialog() string {
	// Create a box for the dialog - make it even narrower
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 1).
		Width(m.width - 60)

	// Style for the title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))

	// Style for section headers
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("111"))

	// Key style
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63"))

	// Create a grid layout for maximum compactness
	navSection := sectionStyle.Render("Navigation:")
	
	// Navigation keys in a compact grid
	navKeys := fmt.Sprintf(
		"%s Switch panels  %s Move up/down  %s Select",
		keyStyle.Render("←/→"), keyStyle.Render("↑/↓"), keyStyle.Render("Enter"))

	// Actions in a compact grid with 3 columns
	actionsSection := sectionStyle.Render("Actions:")
	
	// First row of actions
	actionsRow1 := fmt.Sprintf(
		"%s Toggle endpoint  %s Cycle responses  %s Open config",
		keyStyle.Render("t"), keyStyle.Render("r"), keyStyle.Render("o"))
	
	// Second row of actions
	actionsRow2 := fmt.Sprintf(
		"%s New item        %s Delete item     %s Proxy target",
		keyStyle.Render("n"), keyStyle.Render("d"), keyStyle.Render("p"))
	
	// Third row of actions
	actionsRow3 := fmt.Sprintf(
		"%s Start/stop      %s Quit           %s Help screen",
		keyStyle.Render("s"), keyStyle.Render("q"), keyStyle.Render("h"))
	
	// Fourth row of actions - removed search (/) since it doesn't work
	actionsRow4 := fmt.Sprintf(
		"%s Reload configs",
		keyStyle.Render("Ctrl+r"))

	// Footer text
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center)
	footer := footerStyle.Render("Press Esc to return")

	// Combine title and content with minimal spacing
	content := titleStyle.Render("Climock Help") + "\n" +
		navSection + "\n" +
		navKeys + "\n\n" +
		actionsSection + "\n" +
		actionsRow1 + "\n" +
		actionsRow2 + "\n" +
		actionsRow3 + "\n" +
		actionsRow4 + "\n\n" +
		footer

	// Create the dialog box
	dialog := box.Render(content)

	// Position the dialog in the center of the screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// renderInputDialog renders an input dialog
func (m *Model) renderInputDialog() string {
	// Create a box for the dialog
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(m.width - 20)

	// Style for the title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	// Style for the instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginBottom(1)

	// Style for the buttons
	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)

	// Build the dialog content
	var sb strings.Builder
	sb.WriteString(titleStyle.Render(m.dialogTitle))
	sb.WriteString("\n")
	
	// Add navigation instructions if we have multiple inputs
	if len(m.textInputs) > 1 {
		sb.WriteString(instructionStyle.Render("Use [Tab] to navigate between fields"))
		sb.WriteString("\n\n")
	} else {
		sb.WriteString("\n")
	}
	
	// Handle case where textInputs might be nil
	if len(m.textInputs) > 0 {
		for i, ti := range m.textInputs {
			sb.WriteString(ti.View())
			if i < len(m.textInputs)-1 {
				sb.WriteString("\n\n")
			}
		}
	} else {
		sb.WriteString("Loading inputs...")
	}
	
	sb.WriteString("\n\n")
	sb.WriteString(buttonStyle.Render("[Enter] Confirm  [Esc] Cancel"))

	// Create the dialog box
	dialog := box.Render(sb.String())

	// Position the dialog in the center of the screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// renderConfirmDialog renders a confirmation dialog
func (m *Model) renderConfirmDialog() string {
	// Create a box for the dialog
	box := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(m.width - 20).
		Align(lipgloss.Center)

	// Style for the title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	// Style for the content
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		MarginTop(1).
		MarginBottom(1)

	// Style for the buttons
	buttonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)

	// Build the dialog content
	var sb strings.Builder
	sb.WriteString(titleStyle.Render(m.dialogTitle))
	sb.WriteString("\n\n")
	sb.WriteString(contentStyle.Render(m.dialogContent))
	sb.WriteString("\n\n")
	sb.WriteString(buttonStyle.Render("[Enter] Confirm  [Esc] Cancel"))

	// Create the dialog box
	dialog := box.Render(sb.String())

	// Position the dialog in the center of the screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}
