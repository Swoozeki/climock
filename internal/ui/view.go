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

	// Panel titles
	sb.WriteString(m.renderPanelTitles())
	sb.WriteString("\n")

	// Lists
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

	serverStatus := "Stopped"
	if m.Server.IsRunning() {
		serverStatus = fmt.Sprintf("Running (%s)", m.Server.GetAddress())
	}

	proxyTarget := m.ProxyManager.GetTargetURL()
	header := fmt.Sprintf("Server: %s | Proxy: %s", serverStatus, proxyTarget)

	return headerStyle.Render("Mockoho - " + header)
}

// renderPanelTitles renders the panel titles
func (m *Model) renderPanelTitles() string {
	// Use cached styles with updated widths
	featuresTitleStyle := m.styles.featureTitle.Copy().Width(m.width/4)
	endpointsTitleStyle := m.styles.endpointsTitle.Copy().Width(3*m.width/4)

	// Apply bold styling based on active panel
	if m.activePanel == FeaturesPanel {
		featuresTitleStyle = featuresTitleStyle.Bold(true)
	} else {
		endpointsTitleStyle = endpointsTitleStyle.Bold(true)
	}

	featureTitle := featuresTitleStyle.Render("Features")
	endpointsTitle := endpointsTitleStyle.Render(fmt.Sprintf("Endpoints (%s)", m.selectedFeature))

	return lipgloss.JoinHorizontal(lipgloss.Top, featureTitle, endpointsTitle)
}

// renderLists renders the feature and endpoint lists
func (m *Model) renderLists() string {
	// Use cached styles with updated widths
	featuresStyle := m.styles.features.Copy().Width(m.width/4)
	endpointsStyle := m.styles.endpoints.Copy().Width(3*m.width/4)

	featuresView := featuresStyle.Render(m.featuresList.View())
	endpointsView := endpointsStyle.Render(m.endpointsList.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, featuresView, endpointsView)
}

// renderFooter renders the footer
func (m *Model) renderFooter() string {
	// Use cached style with updated width
	footerStyle := m.styles.footer.Copy().Width(m.width)

	return footerStyle.Render(m.help.View(m.keyMap))
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
		return m.renderMainUI()
	}
}

// renderMainUI renders the main UI (without dialogs)
func (m *Model) renderMainUI() string {
	var sb strings.Builder

	// Header
	sb.WriteString(m.renderHeader())
	sb.WriteString("\n")

	// Panel titles
	sb.WriteString(m.renderPanelTitles())
	sb.WriteString("\n")

	// Lists
	sb.WriteString(m.renderLists())
	sb.WriteString("\n")

	// Footer
	sb.WriteString(m.renderFooter())

	return sb.String()
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
  Tab       - Switch between Features and Endpoints panels
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

// We're now reusing renderInputDialog for ProxyConfigDialog, so this function is no longer needed

// This function is no longer needed as we're using lipgloss.Place for centering