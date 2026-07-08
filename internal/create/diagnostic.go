package create

import (
	"errors"
	"fmt"

	"github.com/dapi/zelma/internal/codex"
	"github.com/dapi/zelma/internal/registry"
	"github.com/dapi/zelma/internal/zellij"
)

type ReasonCode string

const (
	ReasonInvalidRequest      ReasonCode = "create_invalid_request"
	ReasonCodexInvalidInput   ReasonCode = "create_codex_invalid_input"
	ReasonCodexMissingBinary  ReasonCode = "create_codex_missing_binary"
	ReasonPaneLaunchFailed    ReasonCode = "create_pane_launch_failed"
	ReasonPaneUnconfirmed     ReasonCode = "create_pane_unconfirmed"
	ReasonConfirmationFailed  ReasonCode = "create_confirmation_failed"
	ReasonRegistryWriteFailed ReasonCode = "create_registry_write_failed"
	causeRegistryLocked                  = "registry_locked"
)

type Diagnostic struct {
	Code         ReasonCode
	CauseCode    string
	Retryable    bool
	Message      string
	RecoveryHint string
	Summary      Summary
}

type DiagnosticError struct {
	Diagnostic Diagnostic
	Err        error
}

func (err *DiagnosticError) Error() string {
	if err == nil {
		return ""
	}

	diagnostic := err.Diagnostic
	message := fmt.Sprintf("create session: %s: %s; retryable=%t", diagnostic.Code, diagnostic.Message, diagnostic.Retryable)
	if diagnostic.CauseCode != "" {
		message += fmt.Sprintf("; cause=%s", diagnostic.CauseCode)
	}
	if err.Err != nil {
		message += fmt.Sprintf("; detail: %v", err.Err)
	}
	if !diagnostic.Summary.IsZero() {
		message += fmt.Sprintf(
			"; summary: created=%d registered=%d skipped=%d",
			diagnostic.Summary.Created,
			diagnostic.Summary.Registered,
			diagnostic.Summary.Skipped,
		)
	}
	if diagnostic.RecoveryHint != "" {
		message += fmt.Sprintf("; recovery: %s", diagnostic.RecoveryHint)
	}
	return message
}

func (err *DiagnosticError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

func (summary Summary) IsZero() bool {
	return summary.Created == 0 && summary.Registered == 0 && summary.Skipped == 0
}

func PreflightFailure(err error) error {
	if err == nil {
		return nil
	}

	diagnostic := Diagnostic{
		Code:         ReasonInvalidRequest,
		CauseCode:    diagnosticCauseCode(err),
		Message:      "create preflight failed",
		RecoveryHint: "fix the create input or environment, then retry",
	}

	var codexErr *codex.DiagnosticError
	if errors.As(err, &codexErr) {
		diagnostic.Message = codexErr.Diagnostic.Message
		switch codexErr.Diagnostic.Code {
		case codex.ErrorCodeMissingBinary:
			diagnostic.Code = ReasonCodexMissingBinary
			diagnostic.RecoveryHint = "fix environment: " + codexErr.Diagnostic.RecoveryHint
		case codex.ErrorCodeInvalidInput:
			diagnostic.Code = ReasonCodexInvalidInput
			diagnostic.RecoveryHint = "fix input: " + codexErr.Diagnostic.RecoveryHint
		}
	}

	return &DiagnosticError{Diagnostic: diagnostic, Err: err}
}

func RegistryWriteFailure(summary Summary, path string, err error) error {
	if err == nil {
		return nil
	}

	diagnostic := Diagnostic{
		Code:      ReasonRegistryWriteFailed,
		CauseCode: diagnosticCauseCode(err),
		Message:   "write sessions registry failed",
		Summary:   summary,
		RecoveryHint: "inspect the registry path and filesystem permissions; run \"zelma sessions detect\" " +
			"to reconcile the created pane before retrying create",
	}
	if path != "" {
		diagnostic.Message = fmt.Sprintf("write sessions registry failed at %s", path)
	}
	if errors.Is(err, registry.ErrRegistryLocked) {
		diagnostic.Retryable = true
		diagnostic.RecoveryHint = "retry after the other registry writer finishes; run \"zelma sessions detect\" " +
			"first if a Codex pane may already have been created"
	}

	return &DiagnosticError{Diagnostic: diagnostic, Err: err}
}

func invalidRequestFailure(message string, err error) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ReasonInvalidRequest,
			Message:      message,
			RecoveryHint: "fix the create request and retry; zelma did not write registry state",
		},
		Err: err,
	}
}

func paneLaunchFailure(err error) error {
	diagnostic := Diagnostic{
		Code:         ReasonPaneLaunchFailed,
		CauseCode:    diagnosticCauseCode(err),
		Message:      "zellij pane launch failed",
		RecoveryHint: "inspect zellij session and command availability, then retry; zelma did not write registry state",
		Retryable:    true,
	}

	var zellijErr *zellij.DiagnosticError
	if errors.As(err, &zellijErr) {
		diagnostic.Message = zellijErr.Diagnostic.Message
		switch zellijErr.Diagnostic.Code {
		case zellij.ErrorCodeMissingBinary:
			diagnostic.Retryable = false
			diagnostic.RecoveryHint = "fix environment: " + zellijErr.Diagnostic.RecoveryHint
		case zellij.ErrorCodeInvalidInput:
			diagnostic.Retryable = false
			diagnostic.RecoveryHint = "fix the create request and retry; zelma did not write registry state"
		case zellij.ErrorCodeInvalidOutput:
			diagnostic.Retryable = false
			diagnostic.RecoveryHint = "inspect zellij run output and adapter compatibility; zelma did not write registry state"
		case zellij.ErrorCodeCommandFailed:
			if zellijErr.Diagnostic.RecoveryHint != "" {
				diagnostic.RecoveryHint = zellijErr.Diagnostic.RecoveryHint
			}
		}
	}

	return &DiagnosticError{Diagnostic: diagnostic, Err: err}
}

func confirmationFailure(summary Summary, err error) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:         ReasonConfirmationFailed,
			CauseCode:    diagnosticCauseCode(err),
			Message:      "confirm created pane failed",
			RecoveryHint: "run \"zelma sessions detect\" to reconcile any live Codex panes, inspect zellij list-panes for the target session, then retry only after resolving the confirmation failure",
			Summary:      summary,
		},
		Err: err,
	}
}

func paneUnconfirmedFailure(summary Summary, ref zellij.PaneRef) error {
	return &DiagnosticError{
		Diagnostic: Diagnostic{
			Code:    ReasonPaneUnconfirmed,
			Message: fmt.Sprintf("created pane %s in zellij session %q could not be confirmed as Codex", ref.PaneID.String(), ref.Session),
			RecoveryHint: fmt.Sprintf(
				"run \"zelma sessions detect\" to reconcile live Codex panes, inspect zellij pane %s in session %q, then fix the environment before retrying create",
				ref.PaneID.String(),
				ref.Session,
			),
			Summary: summary,
		},
	}
}

func diagnosticCauseCode(err error) string {
	var codexErr *codex.DiagnosticError
	if errors.As(err, &codexErr) {
		return string(codexErr.Diagnostic.Code)
	}

	var zellijErr *zellij.DiagnosticError
	if errors.As(err, &zellijErr) {
		return string(zellijErr.Diagnostic.Code)
	}

	var registryErr *registry.DiagnosticError
	if errors.As(err, &registryErr) {
		return string(registryErr.Diagnostic.Code)
	}
	if errors.Is(err, registry.ErrRegistryLocked) {
		return causeRegistryLocked
	}
	return ""
}
