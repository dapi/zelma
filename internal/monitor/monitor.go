package monitor

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dapi/zelma/internal/live"
	statusbackend "github.com/dapi/zelma/internal/status"
)

const DefaultRefreshInterval = 5 * time.Second

type Provider interface {
	Snapshot(context.Context) (statusbackend.Snapshot, error)
}

type Focuser interface {
	Focus(context.Context, int) error
}

type App struct {
	ctx             context.Context
	provider        Provider
	focuser         Focuser
	refreshInterval time.Duration

	snapshot       statusbackend.Snapshot
	rows           []Row
	selected       int
	showOther      bool
	statusMessage  string
	recoveryNotice string
	refreshing     bool
	tickGeneration uint64
}

type Row struct {
	Session   statusbackend.Session
	Group     string
	Focusable bool
}

type Option func(*App)

func WithRefreshInterval(interval time.Duration) Option {
	return func(app *App) {
		app.refreshInterval = interval
	}
}

func New(ctx context.Context, provider Provider, focuser Focuser, opts ...Option) *App {
	if ctx == nil {
		ctx = context.Background()
	}
	app := &App{
		ctx:             ctx,
		provider:        provider,
		focuser:         focuser,
		refreshInterval: DefaultRefreshInterval,
		showOther:       true,
		statusMessage:   "loading",
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func Run(ctx context.Context, provider Provider, focuser Focuser, output io.Writer, opts ...Option) error {
	app := New(ctx, provider, focuser, opts...)
	programOptions := []tea.ProgramOption{tea.WithContext(app.ctx)}
	if output != nil {
		programOptions = append(programOptions, tea.WithOutput(output))
	}
	_, err := tea.NewProgram(app, programOptions...).Run()
	return err
}

func (app *App) Init() tea.Cmd {
	return app.beginRefresh()
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return app, tea.Quit
		case "up", "k":
			app.Move(-1)
			return app, nil
		case "down", "j":
			app.Move(1)
			return app, nil
		case "r":
			return app, app.beginRefresh()
		case "tab", "t":
			app.ToggleOther()
			return app, nil
		case "enter", "f":
			return app, app.focusCmd()
		}
	case tickMsg:
		if msg.generation != app.tickGeneration {
			return app, nil
		}
		return app, app.beginRefresh()
	case snapshotMsg:
		app.refreshing = false
		app.applySnapshot(msg.snapshot, msg.err)
		return app, app.tickCmd()
	case focusMsg:
		app.applyFocusResult(msg.id, msg.err)
		return app, nil
	}
	return app, nil
}

func (app *App) View() string {
	var builder strings.Builder
	degraded := "no"
	if app.snapshot.Degraded {
		degraded = "yes"
	}
	fmt.Fprintf(
		&builder,
		"zelma monitor                         live %d  stale %d  blocked %d  degraded %s\n\n",
		app.snapshot.Summary.Live,
		app.snapshot.Summary.Stale,
		app.snapshot.Summary.Blocked,
		degraded,
	)

	liveRows := app.rowsForGroup("live")
	otherRows := app.rowsForGroup("other")

	builder.WriteString("LIVE\n")
	if len(liveRows) == 0 {
		if app.snapshot.Degraded {
			builder.WriteString("  Live state unavailable.\n")
		} else {
			builder.WriteString("  No live zelma instances.\n")
		}
	} else {
		for _, row := range liveRows {
			builder.WriteString(app.renderRow(row))
		}
	}

	if app.showOther {
		builder.WriteString("\nOTHER\n")
		if len(otherRows) == 0 {
			builder.WriteString("  No stale or non-active records.\n")
		} else {
			for _, row := range otherRows {
				builder.WriteString(app.renderRow(row))
			}
		}
	}

	if app.recoveryNotice != "" {
		fmt.Fprintf(&builder, "\nrecovery: %s\n", app.recoveryNotice)
	}
	if app.statusMessage != "" {
		fmt.Fprintf(&builder, "\nstatus: %s\n", app.statusMessage)
	}
	return builder.String()
}

func (app *App) Move(delta int) {
	if len(app.rows) == 0 {
		app.selected = 0
		return
	}
	app.selected += delta
	if app.selected < 0 {
		app.selected = 0
	}
	if app.selected >= len(app.rows) {
		app.selected = len(app.rows) - 1
	}
}

func (app *App) ToggleOther() {
	app.showOther = !app.showOther
	app.rebuildRows()
}

func (app *App) Selected() (Row, bool) {
	if app.selected < 0 || app.selected >= len(app.rows) {
		return Row{}, false
	}
	return app.rows[app.selected], true
}

func (app *App) applySnapshot(snapshot statusbackend.Snapshot, err error) {
	if err != nil {
		app.statusMessage = fmt.Sprintf("refresh failed: %v", err)
		return
	}
	app.snapshot = snapshot
	app.statusMessage = fmt.Sprintf("refreshed %s", time.Now().Format("15:04:05"))
	app.recoveryNotice = firstRecoveryHint(snapshot)
	app.rebuildRows()
}

func (app *App) rebuildRows() {
	previousID := 0
	if row, ok := app.Selected(); ok {
		previousID = row.Session.ID
	}

	rows := make([]Row, 0, len(app.snapshot.Sessions))
	for _, session := range app.snapshot.Sessions {
		if isLiveActive(session) {
			rows = append(rows, Row{Session: session, Group: "live", Focusable: true})
		}
	}
	if app.showOther {
		for _, session := range app.snapshot.Sessions {
			if !isLiveActive(session) {
				rows = append(rows, Row{Session: session, Group: "other"})
			}
		}
	}
	app.rows = rows
	app.selected = 0
	if previousID != 0 {
		for index, row := range app.rows {
			if row.Session.ID == previousID {
				app.selected = index
				return
			}
		}
	}
}

func (app *App) rowsForGroup(group string) []Row {
	rows := make([]Row, 0)
	for _, row := range app.rows {
		if row.Group == group {
			rows = append(rows, row)
		}
	}
	return rows
}

func (app *App) renderRow(row Row) string {
	prefix := " "
	if selected, ok := app.Selected(); ok && selected.Session.ID == row.Session.ID {
		prefix = ">"
	}
	session := row.Session
	return fmt.Sprintf(
		"%s %-3d %-9s %-11s %-24s %-14s %-8s %-12s %s\n",
		prefix,
		session.ID,
		emptyDash(session.DashboardStatus),
		emptyDash(session.LiveStatus),
		emptyDash(session.OpenedPath),
		emptyDash(session.ZellijSession),
		emptyDash(session.ZellijTab),
		emptyDash(session.ZellijPane),
		codexLabel(session.CodexSession),
	)
}

func (app *App) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		if app.provider == nil {
			return snapshotMsg{err: fmt.Errorf("monitor status provider is not configured")}
		}
		snapshot, err := app.provider.Snapshot(app.ctx)
		return snapshotMsg{snapshot: snapshot, err: err}
	}
}

func (app *App) beginRefresh() tea.Cmd {
	if app.refreshing {
		return nil
	}
	app.tickGeneration++
	app.refreshing = true
	return app.refreshCmd()
}

func (app *App) tickCmd() tea.Cmd {
	if app.refreshInterval <= 0 {
		return nil
	}
	generation := app.tickGeneration
	return tea.Tick(app.refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg{generation: generation}
	})
}

func (app *App) focusCmd() tea.Cmd {
	row, ok := app.Selected()
	if !ok {
		app.statusMessage = "focus unavailable: no visible session selected"
		return nil
	}
	if !row.Focusable {
		app.statusMessage = fmt.Sprintf("focus unavailable: session %d is not live", row.Session.ID)
		return nil
	}
	return func() tea.Msg {
		if app.focuser == nil {
			return focusMsg{id: row.Session.ID, err: fmt.Errorf("monitor focus adapter is not configured")}
		}
		return focusMsg{id: row.Session.ID, err: app.focuser.Focus(app.ctx, row.Session.ID)}
	}
}

func (app *App) applyFocusResult(id int, err error) {
	if err != nil {
		app.statusMessage = fmt.Sprintf("focus failed for session %d: %v", id, err)
		return
	}
	app.statusMessage = fmt.Sprintf("focused session %d", id)
}

func isLiveActive(session statusbackend.Session) bool {
	return session.DashboardStatus == statusbackend.DashboardStatusActive &&
		session.LiveStatus == string(live.StatusLive)
}

func firstRecoveryHint(snapshot statusbackend.Snapshot) string {
	if len(snapshot.RecoveryHints) > 0 {
		return snapshot.RecoveryHints[0]
	}
	for _, session := range snapshot.Sessions {
		if session.RecoveryHint != "" {
			return session.RecoveryHint
		}
	}
	return ""
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func codexLabel(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return "codex:" + value
}

type snapshotMsg struct {
	snapshot statusbackend.Snapshot
	err      error
}

type focusMsg struct {
	id  int
	err error
}

type tickMsg struct {
	generation uint64
}
