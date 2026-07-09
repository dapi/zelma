package supervisor

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dapi/zelma/internal/config"
	"github.com/dapi/zelma/internal/zellij"
)

const (
	ResultVersion = 1

	StatusMergedSimulated = "merged_simulated"

	PhaseLaunched               = "launched"
	PhaseImplementationComplete = "implementation_complete"
	PhaseReviewFindings         = "review_findings"
	PhaseFixComplete            = "fix_complete"
	PhaseReviewClean            = "review_clean"
	PhaseMergeSimulated         = "merge_simulated"
	PhaseWorking                = "working"

	MarkerPrefix                 = "ZELMA_SUPERVISOR:"
	MarkerImplementationComplete = "implementation_complete"
	MarkerReviewFindings         = "review_findings"
	MarkerFixComplete            = "fix_complete"
	MarkerReviewClean            = "review_clean"
	MarkerMergeSimulated         = "merge_simulated"

	DefaultPollInterval    = time.Minute
	DefaultMaxPolls        = 20
	DefaultMaxReviewCycles = 5
)

type Runtime interface {
	zellij.PaneRunner
	zellij.TabRunner
	zellij.PaneLister
	zellij.PaneDumper
	zellij.PaneWriter
	zellij.PaneCloser
}

type Sleeper func(context.Context, time.Duration) error

type Request struct {
	Issue         int
	Repository    string
	Base          string
	Agent         string
	PromptFile    string
	RepoRoot      string
	ZellijSession string
	Surface       config.StartIssueSurfaceResolution
	PollInterval  time.Duration
	MaxPolls      int
	MaxReviews    int
	Runtime       Runtime
	Sleep         Sleeper
}

type Result struct {
	Version    int          `json:"version"`
	Issue      int          `json:"issue"`
	Repository string       `json:"repository"`
	Base       string       `json:"base"`
	Status     string       `json:"status"`
	Launch     LaunchState  `json:"launch"`
	Polling    PollingState `json:"polling"`
	Review     ReviewState  `json:"review"`
	Cleanup    CleanupState `json:"cleanup"`
}

type LaunchState struct {
	Surface       string   `json:"surface"`
	SurfaceSource string   `json:"surface_source"`
	ZellijSession string   `json:"zellij_session"`
	ZellijTab     string   `json:"zellij_tab,omitempty"`
	ZellijPane    string   `json:"zellij_pane"`
	Name          string   `json:"name"`
	CWD           string   `json:"cwd"`
	Command       []string `json:"command"`
	CommandLine   string   `json:"command_line"`
	PromptFile    string   `json:"prompt_file,omitempty"`
}

type PollingState struct {
	IntervalSeconds int        `json:"interval_seconds"`
	Snapshots       []Snapshot `json:"snapshots"`
}

type Snapshot struct {
	Sequence       int    `json:"sequence"`
	Phase          string `json:"phase"`
	Marker         string `json:"marker,omitempty"`
	ElapsedSeconds int    `json:"elapsed_seconds"`
}

type ReviewState struct {
	Cycles        int  `json:"cycles"`
	FindingsFixed int  `json:"findings_fixed"`
	Clean         bool `json:"clean"`
}

type CleanupState struct {
	PaneClosed bool   `json:"pane_closed"`
	Registry   string `json:"registry"`
}

type ErrorCode string

const (
	ErrorCodeInvalidInput   ErrorCode = "supervisor_invalid_input"
	ErrorCodeMaxPolls       ErrorCode = "supervisor_max_polls_reached"
	ErrorCodeMaxReviews     ErrorCode = "supervisor_max_review_cycles_reached"
	ErrorCodeMissingTabPane ErrorCode = "supervisor_tab_pane_not_found"
)

type Diagnostic struct {
	Code         ErrorCode
	Message      string
	RecoveryHint string
}

type DiagnosticError struct {
	Diagnostic Diagnostic
	Err        error
}

func (err *DiagnosticError) Error() string {
	if err == nil {
		return ""
	}
	message := fmt.Sprintf("supervisor: %s: %s", err.Diagnostic.Code, err.Diagnostic.Message)
	if err.Diagnostic.RecoveryHint != "" {
		message += fmt.Sprintf("; recovery: %s", err.Diagnostic.RecoveryHint)
	}
	return message
}

func (err *DiagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func StartIssue(ctx context.Context, request Request) (Result, error) {
	if err := validateRequest(request); err != nil {
		return Result{}, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	pollInterval := request.PollInterval
	if pollInterval <= 0 {
		pollInterval = DefaultPollInterval
	}
	maxPolls := request.MaxPolls
	if maxPolls <= 0 {
		maxPolls = DefaultMaxPolls
	}
	maxReviews := request.MaxReviews
	if maxReviews <= 0 {
		maxReviews = DefaultMaxReviewCycles
	}
	sleep := request.Sleep
	if sleep == nil {
		sleep = sleepContext
	}

	launch, err := launchStartIssue(ctx, request)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		Version:    ResultVersion,
		Issue:      request.Issue,
		Repository: request.Repository,
		Base:       request.Base,
		Status:     PhaseLaunched,
		Launch:     launch,
		Polling: PollingState{
			IntervalSeconds: int(pollInterval / time.Second),
			Snapshots:       []Snapshot{},
		},
		Cleanup: CleanupState{Registry: "simulated_no_registry_records"},
	}

	var awaitingFix bool
	var lastPhase string
	var observedMarkers []string
	nextSnapshotSequence := 1
	for poll := 1; poll <= maxPolls; poll++ {
		screen, err := request.Runtime.DumpScreen(ctx, zellij.DumpScreenRequest{
			Session: launch.ZellijSession,
			PaneID:  launch.ZellijPane,
			Full:    true,
		})
		if err != nil {
			return Result{}, err
		}

		markers := screenMarkers(screen)
		newMarkers := newMarkerEvents(observedMarkers, markers)
		observedMarkers = append(observedMarkers[:0], markers...)
		if len(newMarkers) == 0 {
			newMarkers = []string{""}
		}
		actionTaken := false
		for _, marker := range newMarkers {
			phase := phaseForMarker(marker)
			result.Polling.Snapshots = append(result.Polling.Snapshots, Snapshot{
				Sequence:       nextSnapshotSequence,
				Phase:          phase,
				Marker:         marker,
				ElapsedSeconds: (poll - 1) * result.Polling.IntervalSeconds,
			})
			nextSnapshotSequence++

			switch phase {
			case PhaseImplementationComplete:
				if result.Review.Cycles == 0 {
					if result.Review.Cycles >= maxReviews {
						return Result{}, maxReviewsFailure(maxReviews)
					}
					if err := request.Runtime.WriteChars(ctx, zellij.WriteCharsRequest{
						Session: launch.ZellijSession,
						PaneID:  launch.ZellijPane,
						Chars:   "/review\n",
					}); err != nil {
						return Result{}, err
					}
					result.Review.Cycles++
					actionTaken = true
				}
			case PhaseReviewFindings:
				if result.Review.Cycles > 0 && !awaitingFix {
					if err := request.Runtime.WriteChars(ctx, zellij.WriteCharsRequest{
						Session: launch.ZellijSession,
						PaneID:  launch.ZellijPane,
						Chars:   fixInstruction(request.Issue),
					}); err != nil {
						return Result{}, err
					}
					result.Review.FindingsFixed++
					result.Review.Clean = false
					awaitingFix = true
					actionTaken = true
				}
			case PhaseFixComplete:
				if awaitingFix {
					if result.Review.Cycles >= maxReviews {
						return Result{}, maxReviewsFailure(maxReviews)
					}
					if err := request.Runtime.WriteChars(ctx, zellij.WriteCharsRequest{
						Session: launch.ZellijSession,
						PaneID:  launch.ZellijPane,
						Chars:   "/review\n",
					}); err != nil {
						return Result{}, err
					}
					result.Review.Cycles++
					awaitingFix = false
					actionTaken = true
				}
			case PhaseReviewClean:
				if result.Review.Cycles > 0 && !awaitingFix {
					result.Review.Clean = true
				}
			case PhaseMergeSimulated:
				if !result.Review.Clean || result.Review.Cycles < 1 {
					return Result{}, invalidInputFailure("merge simulation appeared before a clean review")
				}
				if err := request.Runtime.ClosePane(ctx, zellij.ClosePaneRequest{
					Session: launch.ZellijSession,
					PaneID:  launch.ZellijPane,
				}); err != nil {
					return Result{}, err
				}
				result.Cleanup.PaneClosed = true
				result.Status = StatusMergedSimulated
				return result, nil
			}
		}

		phase := phaseForMarker(newMarkers[len(newMarkers)-1])
		if !actionTaken && phase == lastPhase {
			if err := sleep(ctx, pollInterval); err != nil {
				return Result{}, err
			}
		}
		lastPhase = phase
	}

	return Result{}, maxPollsFailure(maxPolls)
}

func newMarkerEvents(previous, current []string) []string {
	if len(current) == 0 {
		return nil
	}

	sharedPrefix := 0
	for sharedPrefix < len(previous) &&
		sharedPrefix < len(current) &&
		previous[sharedPrefix] == current[sharedPrefix] {
		sharedPrefix++
	}
	return current[sharedPrefix:]
}

func validateRequest(request Request) error {
	if request.Runtime == nil {
		return invalidInputFailure("runtime is required")
	}
	if request.Issue <= 0 {
		return invalidInputFailure("issue must be a positive integer")
	}
	if strings.TrimSpace(request.Repository) == "" {
		return invalidInputFailure("repository is required")
	}
	if strings.TrimSpace(request.Base) == "" {
		return invalidInputFailure("base branch is required")
	}
	if strings.TrimSpace(request.RepoRoot) == "" {
		return invalidInputFailure("repo root is required")
	}
	if strings.TrimSpace(request.ZellijSession) == "" {
		return invalidInputFailure("zellij session is required")
	}
	if request.Surface.Surface != config.StartIssueSurfacePane && request.Surface.Surface != config.StartIssueSurfaceTab {
		return invalidInputFailure("surface must be pane or tab")
	}
	if request.PollInterval > time.Minute {
		return invalidInputFailure("poll interval must be one minute or less")
	}
	return nil
}

func launchStartIssue(ctx context.Context, request Request) (LaunchState, error) {
	name := fmt.Sprintf("issue-%d", request.Issue)
	command := startIssueCommand(request)
	state := LaunchState{
		Surface:       request.Surface.Surface,
		SurfaceSource: request.Surface.Source,
		ZellijSession: request.ZellijSession,
		Name:          name,
		CWD:           filepath.Clean(request.RepoRoot),
		Command:       command,
		CommandLine:   strings.Join(command, " "),
		PromptFile:    request.PromptFile,
	}

	switch request.Surface.Surface {
	case config.StartIssueSurfacePane:
		ref, err := request.Runtime.RunPane(ctx, zellij.RunPaneRequest{
			Session: request.ZellijSession,
			CWD:     state.CWD,
			Name:    name,
			Command: command,
		})
		if err != nil {
			return LaunchState{}, err
		}
		state.ZellijPane = ref.PaneID.String()
		return state, nil
	case config.StartIssueSurfaceTab:
		ref, err := request.Runtime.RunTab(ctx, zellij.RunTabRequest{
			Session: request.ZellijSession,
			CWD:     state.CWD,
			Name:    name,
			Command: command,
		})
		if err != nil {
			return LaunchState{}, err
		}
		state.ZellijTab = "tab_" + strconv.Itoa(ref.TabID)
		pane, err := findStartedPane(ctx, request.Runtime, ref.Session, ref.TabID, state.CWD, name)
		if err != nil {
			return LaunchState{}, err
		}
		state.ZellijPane = pane.ID.String()
		return state, nil
	default:
		return LaunchState{}, invalidInputFailure("surface must be pane or tab")
	}
}

func startIssueCommand(request Request) []string {
	command := []string{"start-issue", strconv.Itoa(request.Issue), "--repo", request.Repository, "--base", request.Base}
	if strings.TrimSpace(request.Agent) != "" {
		command = append(command, "--agent", strings.TrimSpace(request.Agent))
	}
	if strings.TrimSpace(request.PromptFile) != "" {
		command = append(command, "--prompt-file", strings.TrimSpace(request.PromptFile))
	}
	return command
}

func findStartedPane(ctx context.Context, runtime Runtime, session string, tabID int, cwd, name string) (zellij.Pane, error) {
	panes, err := runtime.ListPanes(ctx, session)
	if err != nil {
		return zellij.Pane{}, err
	}
	for _, pane := range panes {
		if pane.ID.Kind != zellij.PaneKindTerminal || pane.Exited || pane.TabID != tabID {
			continue
		}
		if pane.TabName == name {
			return pane, nil
		}
		if pane.PaneCWD != nil && filepath.Clean(*pane.PaneCWD) == cwd {
			return pane, nil
		}
	}
	return zellij.Pane{}, &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeMissingTabPane,
			Message:      fmt.Sprintf("new tab %d did not expose a terminal pane for observation", tabID),
			RecoveryHint: "inspect zellij list-panes output for the created issue tab and retry after zellij reports the pane",
		},
	}
}

func screenMarkers(screen string) []string {
	var markers []string
	for _, line := range strings.Split(screen, "\n") {
		value := strings.TrimSpace(line)
		if !strings.HasPrefix(value, MarkerPrefix) {
			continue
		}
		candidate := strings.TrimSpace(strings.TrimPrefix(value, MarkerPrefix))
		switch candidate {
		case MarkerImplementationComplete, MarkerReviewFindings, MarkerFixComplete, MarkerReviewClean, MarkerMergeSimulated:
			markers = append(markers, candidate)
		}
	}
	return markers
}

func classifyScreen(screen string) (phase, marker string) {
	markers := screenMarkers(screen)
	if len(markers) == 0 {
		return PhaseWorking, ""
	}
	marker = markers[len(markers)-1]
	return phaseForMarker(marker), marker
}

func phaseForMarker(marker string) string {
	switch marker {
	case MarkerImplementationComplete:
		return PhaseImplementationComplete
	case MarkerReviewFindings:
		return PhaseReviewFindings
	case MarkerFixComplete:
		return PhaseFixComplete
	case MarkerReviewClean:
		return PhaseReviewClean
	case MarkerMergeSimulated:
		return PhaseMergeSimulated
	default:
		return PhaseWorking
	}
}

func fixInstruction(issue int) string {
	return fmt.Sprintf("Fix all critical/high/important review findings in scope for issue %d. Run relevant checks, commit, and push fixes. Do not fix unrelated findings.\n", issue)
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func invalidInputFailure(message string) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeInvalidInput,
			Message:      message,
			RecoveryHint: "fix the supervisor start-issue request and retry before launching zellij work",
		},
	}
}

func maxPollsFailure(maxPolls int) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeMaxPolls,
			Message:      fmt.Sprintf("task pane did not reach terminal merge simulation after %d polls", maxPolls),
			RecoveryHint: "inspect the task pane screen, then retry or continue supervision manually",
		},
	}
}

func maxReviewsFailure(maxReviews int) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ErrorCodeMaxReviews,
			Message:      fmt.Sprintf("review cycle limit %d reached before clean re-review", maxReviews),
			RecoveryHint: "inspect review findings and either raise the cycle limit or stop with a blocker",
		},
	}
}
