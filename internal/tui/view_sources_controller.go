package tui

import (
	"github.com/arcaven/ThreeDoors/internal/core/connection"
	tea "github.com/charmbracelet/bubbletea"
)

// handleSourceViewMessage handles Update() messages for source/sync views
// (Sources, SourceDetail, SyncLog, SyncLogDetail, ConnectWizard, Disconnect, Reauth).
// Returns (model, cmd, handled). If handled is false, the caller should
// continue processing in the main Update switch.
func (m *MainModel) handleSourceViewMessage(msg tea.Msg) (tea.Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case ShowSourcesMsg:
		if m.connMgr != nil {
			m.sourcesView = NewSourcesView(m.connMgr)
			m.sourcesView.SetWidth(m.width)
			m.sourcesView.SetHeight(m.height)
			m.previousView = m.viewMode
			m.setViewMode(ViewSources)
		}
		return m, nil, true

	case ShowSourceDetailMsg:
		if m.connMgr != nil {
			conn, err := m.connMgr.Get(msg.ConnectionID)
			if err == nil {
				m.sourceDetailView = NewSourceDetailView(conn, m.connMgr)
				m.sourceDetailView.SetWidth(m.width)
				m.sourceDetailView.SetHeight(m.height)
				m.previousView = m.viewMode
				m.setViewMode(ViewSourceDetail)
			}
		}
		return m, nil, true

	case ShowSyncLogDetailMsg:
		if m.syncEventLog != nil {
			events, err := m.syncEventLog.SyncLog(msg.ConnectionID, 0)
			if err != nil {
				events = nil
			}
			m.syncLogDetailView = NewSyncLogDetailView(msg.ConnectionID, events)
			m.syncLogDetailView.SetWidth(m.width)
			m.syncLogDetailView.SetHeight(m.height)
			m.previousView = m.viewMode
			m.setViewMode(ViewSyncLogDetail)
		}
		return m, nil, true

	case SourceActionMsg:
		if msg.Action == "sync_log" && m.syncEventLog != nil {
			events, err := m.syncEventLog.SyncLog(msg.ConnectionID, 0)
			if err != nil {
				events = nil
			}
			m.syncLogDetailView = NewSyncLogDetailView(msg.ConnectionID, events)
			m.syncLogDetailView.SetWidth(m.width)
			m.syncLogDetailView.SetHeight(m.height)
			m.previousView = m.viewMode
			m.setViewMode(ViewSyncLogDetail)
		}
		if msg.Action == "reauth" && m.connMgr != nil {
			conn, err := m.connMgr.Get(msg.ConnectionID)
			if err == nil {
				if conn.State != connection.StateAuthExpired {
					return m, func() tea.Msg {
						return FlashMsg{Text: "Re-authentication only available for expired connections"}
					}, true
				}
				authType := AuthNone
				tokenHelp := ""
				for _, spec := range DefaultProviderSpecs() {
					if spec.Name == conn.ProviderName {
						authType = spec.AuthType
						tokenHelp = spec.TokenHelp
						break
					}
				}
				if authType == AuthOAuth {
					return m, func() tea.Msg {
						return FlashMsg{Text: "OAuth re-authentication not yet supported — disconnect and reconnect"}
					}, true
				}
				m.reauthDialog = NewReauthDialog(conn, tokenHelp)
				m.reauthDialog.SetWidth(m.width)
				m.reauthDialog.SetHeight(m.height)
				m.previousView = m.viewMode
				m.setViewMode(ViewReauth)
				return m, m.reauthDialog.Init(), true
			}
		}
		if msg.Action == "disconnect" && m.connMgr != nil {
			conn, err := m.connMgr.Get(msg.ConnectionID)
			if err == nil {
				m.disconnectDialog = NewDisconnectDialog(conn)
				m.disconnectDialog.SetWidth(m.width)
				m.disconnectDialog.SetHeight(m.height)
				m.previousView = m.viewMode
				m.setViewMode(ViewDisconnect)
				return m, m.disconnectDialog.Init(), true
			}
		}
		return m, nil, true

	case ShowConnectWizardMsg:
		if m.connMgr != nil {
			m.connectWizard = NewConnectWizard(DefaultProviderSpecs(), m.connMgr)
			m.connectWizard.SetWidth(m.width)
			m.connectWizard.SetHeight(m.height)
			m.previousView = m.viewMode
			m.setViewMode(ViewConnectWizard)
			return m, m.connectWizard.Init(), true
		}
		return m, nil, true

	case ConnectWizardCompleteMsg:
		if m.connMgr != nil {
			conn, err := m.connMgr.Add(msg.ProviderName, msg.Label, msg.Settings)
			if err == nil {
				conn.SyncMode = msg.SyncMode
				conn.PollInterval = msg.PollInterval
				_ = m.connMgr.Transition(conn.ID, connection.StateConnected)
			}
		}
		m.connectWizard = nil
		m.setViewMode(ViewSources)
		return m, nil, true

	case ConnectWizardCancelMsg:
		m.connectWizard = nil
		if m.previousView == ViewSources {
			m.setViewMode(ViewSources)
		} else {
			m.setViewMode(ViewDoors)
		}
		return m, nil, true

	case ShowDisconnectDialogMsg:
		if m.connMgr != nil {
			conn, err := m.connMgr.Get(msg.ConnectionID)
			if err == nil {
				m.disconnectDialog = NewDisconnectDialog(conn)
				m.disconnectDialog.SetWidth(m.width)
				m.disconnectDialog.SetHeight(m.height)
				m.previousView = m.viewMode
				m.setViewMode(ViewDisconnect)
				return m, m.disconnectDialog.Init(), true
			}
		}
		return m, nil, true

	case DisconnectConfirmedMsg:
		if m.connMgr != nil {
			err := m.connMgr.Disconnect(msg.ConnectionID, msg.KeepTasks)
			if err == nil {
				flashText := "Connection disconnected — tasks kept locally"
				if !msg.KeepTasks {
					flashText = "Connection disconnected — synced tasks removed"
				}
				m.disconnectDialog = nil
				m.setViewMode(ViewSources)
				m.sourcesView = NewSourcesView(m.connMgr)
				m.sourcesView.SetWidth(m.width)
				m.sourcesView.SetHeight(m.height)
				return m, func() tea.Msg { return FlashMsg{Text: flashText} }, true
			}
		}
		m.disconnectDialog = nil
		m.setViewMode(ViewSources)
		return m, nil, true

	case DisconnectCancelledMsg:
		m.disconnectDialog = nil
		if m.previousView == ViewSourceDetail {
			m.setViewMode(ViewSourceDetail)
		} else {
			m.setViewMode(ViewSources)
		}
		return m, nil, true

	case ReauthCompleteMsg:
		if m.connSvc != nil {
			err := m.connSvc.ReAuthenticate(msg.ConnectionID, msg.NewToken)
			if err == nil {
				m.reauthDialog = nil
				m.setViewMode(ViewSourceDetail)
				return m, func() tea.Msg {
					return FlashMsg{Text: "Re-authenticated successfully"}
				}, true
			}
			m.reauthDialog = nil
			m.setViewMode(ViewSourceDetail)
			return m, func() tea.Msg {
				return FlashMsg{Text: "Re-authentication failed: " + err.Error()}
			}, true
		}
		m.reauthDialog = nil
		m.setViewMode(ViewSourceDetail)
		return m, nil, true

	case ReauthCancelledMsg:
		m.reauthDialog = nil
		if m.previousView == ViewSourceDetail {
			m.setViewMode(ViewSourceDetail)
		} else {
			m.setViewMode(ViewSources)
		}
		return m, nil, true

	case ShowSyncLogMsg:
		sv := NewSyncLogView(msg.Entries)
		sv.SetWidth(m.width)
		sv.SetHeight(m.height)
		m.syncLogView = sv
		m.previousView = m.viewMode
		m.setViewMode(ViewSyncLog)
		return m, nil, true
	}

	return m, nil, false
}

// sourceViewContent returns the View() content for source/sync views.
// Returns (view, showValuesFooter, handled).
func (m *MainModel) sourceViewContent() (string, bool, bool) {
	switch m.viewMode {
	case ViewSources:
		if m.sourcesView != nil {
			return m.sourcesView.View(), false, true
		}
		return "", false, true
	case ViewSourceDetail:
		if m.sourceDetailView != nil {
			return m.sourceDetailView.View(), false, true
		}
		return "", false, true
	case ViewSyncLogDetail:
		if m.syncLogDetailView != nil {
			return m.syncLogDetailView.View(), false, true
		}
		return "", false, true
	case ViewConnectWizard:
		if m.connectWizard != nil {
			return m.connectWizard.View(), false, true
		}
		return "", false, true
	case ViewDisconnect:
		if m.disconnectDialog != nil {
			return m.disconnectDialog.View(), false, true
		}
		return "", false, true
	case ViewReauth:
		if m.reauthDialog != nil {
			return m.reauthDialog.View(), false, true
		}
		return "", false, true
	case ViewSyncLog:
		if m.syncLogView != nil {
			return m.syncLogView.View(), false, true
		}
		return "", false, true
	}
	return "", false, false
}

// Update delegate methods for source/sync views.

func (m *MainModel) updateSources(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.sourcesView == nil {
		return m, nil
	}
	cmd := m.sourcesView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateSourceDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.sourceDetailView == nil {
		return m, nil
	}
	cmd := m.sourceDetailView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateSyncLogDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.syncLogDetailView == nil {
		return m, nil
	}
	cmd := m.syncLogDetailView.Update(msg)
	return m, cmd
}

func (m *MainModel) updateConnectWizard(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.connectWizard == nil {
		return m, nil
	}
	cmd := m.connectWizard.Update(msg)
	return m, cmd
}

// resizeSourceViews updates dimensions for source/sync views on window resize.
func (m *MainModel) resizeSourceViews(width, height int) {
	if m.sourcesView != nil {
		m.sourcesView.SetWidth(width)
		m.sourcesView.SetHeight(height)
	}
	if m.sourceDetailView != nil {
		m.sourceDetailView.SetWidth(width)
		m.sourceDetailView.SetHeight(height)
	}
	if m.disconnectDialog != nil {
		m.disconnectDialog.SetWidth(width)
		m.disconnectDialog.SetHeight(height)
	}
	if m.reauthDialog != nil {
		m.reauthDialog.SetWidth(width)
		m.reauthDialog.SetHeight(height)
	}
	if m.syncLogDetailView != nil {
		m.syncLogDetailView.SetWidth(width)
		m.syncLogDetailView.SetHeight(height)
	}
}

func (m *MainModel) updateDisconnect(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.disconnectDialog == nil {
		return m, nil
	}
	cmd := m.disconnectDialog.Update(msg)
	return m, cmd
}

func (m *MainModel) updateReauth(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.reauthDialog == nil {
		return m, nil
	}
	cmd := m.reauthDialog.Update(msg)
	return m, cmd
}

func (m *MainModel) updateSyncLog(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.syncLogView == nil {
		return m, nil
	}
	cmd := m.syncLogView.Update(msg)
	return m, cmd
}
