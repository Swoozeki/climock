package ui

import (
	"fmt"
	"strings"

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
	// Use cached style with updated width
	headerStyle := m.styles.header.Copy().Width(m.width)

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

	return headerStyle.Render(titleStyle.Render("Mockoho") + " - " + header)
}

// renderLists renders the feature and endpoint lists
func (m *Model) renderLists() string {
	// Calculate widths accounting for borders (subtract border width)
	// Border takes 2 characters (1 on each side)
	featureWidth := m.width/4 - 2
	endpointWidth := 3*m.width/4 - 2
	
	// Use cached styles with adjusted widths
	featuresStyle := m.styles.features.Copy().Width(featureWidth)
	endpointsStyle := m.styles.endpoints.Copy().Width(endpointWidth)

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
	// Use cached style with updated width
	footerStyle := m.styles.footer.Copy().Width(m.width)

	// Style for the footer content
	footerContentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	return footerStyle.Render(footerContentStyle.Render(m.help.View(m.keyMap)))
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

	// Content for the help dialog
	helpContent := `
Navigation:
	 ←/→       - Switch between Features and Endpoints panels
	 ↑/↓       - Navigate up/down in the current panel
	 Enter     - Select a feature or endpoint

Actions:
  t         - Toggle endpoint active/inactive
  r         - Cycle through available responses (sets as default)
  o         - Open configuration file in default editor
  n         - Create new endpoint or feature
  d         - Delete selected endpoint or feature
  p         - Change proxy target
  s         - Start/stop server
  q         - Quit application
  h         - Show this help screen
  /         - Search for endpoints
  Ctrl+r    - Reload configurations from disk

Press Esc or any key to return...`

	// Combine title and content
	content := titleStyle.Render("Mockoho Help") + helpContent

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
