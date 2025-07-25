package tui

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/killallgit/dick/internal/config"
	"github.com/killallgit/dick/internal/tui/messages"
	"github.com/killallgit/dick/internal/tui/views"
)

// Model is the main TUI model that manages different views
type Model struct {
	// Current active view
	activeView messages.ViewType
	views      map[messages.ViewType]views.View
	
	// Shared state
	config      *config.Config
	width       int
	height      int
	lastUpdate  time.Time
	refreshRate time.Duration
	
	// Navigation history for ESC to go back
	viewHistory []messages.ViewType
	
	// Watch mode flag
	watchMode bool
}

// NewModel creates a new main TUI model
func NewModel(cfg *config.Config, watchMode bool) *Model {
	m := &Model{
		config:      cfg,
		views:       make(map[messages.ViewType]views.View),
		refreshRate: time.Second,
		lastUpdate:  time.Now(),
		watchMode:   watchMode,
		viewHistory: []messages.ViewType{},
	}
	
	// Initialize all views
	m.views[messages.StatusView] = views.NewStatusView(cfg)
	m.views[messages.MonitorView] = views.NewMonitorView(cfg)
	m.views[messages.HelpView] = views.NewHelpView()
	
	// Set initial view based on mode
	if watchMode {
		m.activeView = messages.MonitorView
	} else {
		m.activeView = messages.StatusView
	}
	
	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.EnterAltScreen,
		m.tick(),
	}
	
	// Initialize active view
	if view, ok := m.views[m.activeView]; ok {
		if cmd := view.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	return tea.Batch(cmds...)
}

// Update handles messages for the main model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle global messages first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update context for all views
		for viewType, view := range m.views {
			if v, cmd := view.Update(msg); cmd != nil {
				m.views[viewType] = v
				cmds = append(cmds, cmd)
			}
		}
		return m, tea.Batch(cmds...)
		
	case tea.KeyMsg:
		// Global navigation keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			// Only quit if not in confirm view
			if m.activeView != messages.ConfirmView {
				return m, tea.Quit
			}
		case "?":
			// Toggle help view
			if m.activeView != messages.HelpView {
				m.navigateTo(messages.HelpView)
			} else {
				m.navigateBack()
			}
			return m, nil
		case "1":
			m.navigateTo(messages.StatusView)
			return m, nil
		case "2":
			m.navigateTo(messages.MonitorView)
			return m, nil
		case "3":
			m.navigateTo(messages.SettingsView)
			return m, nil
		case "tab":
			m.navigateNext()
			return m, nil
		case "shift+tab":
			m.navigatePrev()
			return m, nil
		case "esc":
			if len(m.viewHistory) > 0 {
				m.navigateBack()
				return m, nil
			}
		}
		
	case messages.TickMsg:
		m.lastUpdate = msg.Time
		cmds = append(cmds, m.tick())
		
	case messages.NavigateMsg:
		m.navigateTo(msg.To)
		return m, nil
		
	case messages.ConfigReloadMsg:
		// Reload config
		if cfg, err := config.LoadConfig(); err == nil {
			m.config = cfg
			// Notify all views of config update
			for viewType, view := range m.views {
				if v, cmd := view.Update(msg); cmd != nil {
					m.views[viewType] = v
					cmds = append(cmds, cmd)
				}
			}
		}
		return m, tea.Batch(cmds...)
	}
	
	// Pass message to active view
	if view, ok := m.views[m.activeView]; ok {
		updatedView, cmd := view.Update(msg)
		m.views[m.activeView] = updatedView
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	
	return m, tea.Batch(cmds...)
}

// View renders the current active view
func (m *Model) View() string {
	if view, ok := m.views[m.activeView]; ok {
		return view.View()
	}
	return "Loading..."
}

// Helper methods for navigation

func (m *Model) navigateTo(viewType messages.ViewType) {
	if m.activeView != viewType {
		// Add current view to history
		m.viewHistory = append(m.viewHistory, m.activeView)
		m.activeView = viewType
		
		// Initialize view if needed
		if view, ok := m.views[viewType]; ok {
			if cmd := view.Init(); cmd != nil {
				// We can't return a command from here, so we'll handle init in Update
			}
		}
	}
}

func (m *Model) navigateBack() {
	if len(m.viewHistory) > 0 {
		// Pop last view from history
		m.activeView = m.viewHistory[len(m.viewHistory)-1]
		m.viewHistory = m.viewHistory[:len(m.viewHistory)-1]
	}
}

func (m *Model) navigateNext() {
	views := []messages.ViewType{messages.StatusView, messages.MonitorView, messages.SettingsView}
	for i, v := range views {
		if v == m.activeView {
			m.navigateTo(views[(i+1)%len(views)])
			break
		}
	}
}

func (m *Model) navigatePrev() {
	views := []messages.ViewType{messages.StatusView, messages.MonitorView, messages.SettingsView}
	for i, v := range views {
		if v == m.activeView {
			prev := i - 1
			if prev < 0 {
				prev = len(views) - 1
			}
			m.navigateTo(views[prev])
			break
		}
	}
}

// tick returns a command that triggers a refresh
func (m *Model) tick() tea.Cmd {
	return tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
		return messages.TickMsg{Time: t}
	})
}