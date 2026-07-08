package zellij

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type OutputKind string

const (
	OutputKindPanes  OutputKind = "panes"
	OutputKindPaneID OutputKind = "pane_id"
)

type ParseErrorCode string

const (
	ParseErrorInvalidJSON  ParseErrorCode = "zellij_invalid_json"
	ParseErrorTrailingData ParseErrorCode = "zellij_trailing_data"
	ParseErrorMissingField ParseErrorCode = "zellij_missing_required_field"
	ParseErrorInvalidField ParseErrorCode = "zellij_invalid_field"
)

type ParseError struct {
	Kind    OutputKind
	Code    ParseErrorCode
	Path    string
	Message string
	Err     error
}

func (err *ParseError) Error() string {
	if err == nil {
		return ""
	}
	if err.Path == "" {
		return fmt.Sprintf("parse zellij %s output: %s: %s", err.Kind, err.Code, err.Message)
	}
	return fmt.Sprintf("parse zellij %s output: %s at %s: %s", err.Kind, err.Code, err.Path, err.Message)
}

func (err *ParseError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

type PaneKind string

const (
	PaneKindTerminal PaneKind = "terminal"
	PaneKindPlugin   PaneKind = "plugin"
)

type PaneID struct {
	Kind   PaneKind
	Number int
}

func (id PaneID) String() string {
	return string(id.Kind) + "_" + strconv.Itoa(id.Number)
}

func ParsePaneIDOutput(data []byte) (PaneID, error) {
	if len(data) == 0 {
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id output is empty", nil)
	}
	if !utf8.Valid(data) {
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id output is not valid UTF-8", nil)
	}
	if bytes.Contains(data, []byte{0}) {
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id output contains NUL byte", nil)
	}

	value := strings.TrimSpace(string(data))
	if value == "" {
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id output is blank", nil)
	}
	if containsControl(value) {
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id output contains control character", nil)
	}

	paneID, err := ParsePaneID(value)
	if err != nil {
		return PaneID{}, err
	}
	return paneID, nil
}

func ParsePaneID(value string) (PaneID, error) {
	kindValue, numberValue, ok := strings.Cut(value, "_")
	if !ok {
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id must use <kind>_<id> format", nil)
	}

	var kind PaneKind
	switch PaneKind(kindValue) {
	case PaneKindTerminal:
		kind = PaneKindTerminal
	case PaneKindPlugin:
		kind = PaneKindPlugin
	default:
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id kind must be terminal or plugin", nil)
	}

	number, err := strconv.Atoi(numberValue)
	if err != nil {
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id number must be an integer", err)
	}
	if number < 0 {
		return PaneID{}, parseError(OutputKindPaneID, ParseErrorInvalidField, "stdout", "pane id number must be non-negative", nil)
	}
	return PaneID{Kind: kind, Number: number}, nil
}

type Pane struct {
	ID              PaneID
	ProcessID       *int
	Title           string
	IsFocused       bool
	IsFloating      bool
	IsSuppressed    bool
	Exited          bool
	ExitStatus      *int
	TabID           int
	TabPosition     int
	TabName         string
	TerminalCommand *string
	PaneCommand     *string
	PaneCWD         *string
	PluginURL       *string
}

func ParseListPanesJSON(data []byte) ([]Pane, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))

	var raw []paneJSON
	if err := decoder.Decode(&raw); err != nil {
		return nil, parseError(OutputKindPanes, ParseErrorInvalidJSON, "", fmt.Sprintf("parse JSON array: %v", err), err)
	}
	if raw == nil {
		return nil, parseError(OutputKindPanes, ParseErrorInvalidJSON, "", "pane JSON must be an array; use [] for an empty pane list", nil)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, parseError(OutputKindPanes, ParseErrorTrailingData, "", "pane JSON contains trailing data after the top-level array", err)
	}

	panes := make([]Pane, 0, len(raw))
	seenPaneIDs := map[string]int{}
	for index, rawPane := range raw {
		pane, err := rawPane.pane(index)
		if err != nil {
			return nil, err
		}
		paneID := pane.ID.String()
		if firstIndex, ok := seenPaneIDs[paneID]; ok {
			return nil, parseError(OutputKindPanes, ParseErrorInvalidField, "panes["+strconv.Itoa(index)+"].id", "duplicates typed pane id from panes["+strconv.Itoa(firstIndex)+"]", nil)
		}
		seenPaneIDs[paneID] = index
		panes = append(panes, pane)
	}

	return panes, nil
}

type paneJSON struct {
	ID              *int    `json:"id"`
	ProcessID       *int    `json:"pid"`
	PaneProcessID   *int    `json:"pane_pid"`
	IsPlugin        *bool   `json:"is_plugin"`
	Title           *string `json:"title"`
	IsFocused       *bool   `json:"is_focused"`
	IsFloating      *bool   `json:"is_floating"`
	IsSuppressed    *bool   `json:"is_suppressed"`
	Exited          *bool   `json:"exited"`
	ExitStatus      *int    `json:"exit_status"`
	TabID           *int    `json:"tab_id"`
	TabPosition     *int    `json:"tab_position"`
	TabName         *string `json:"tab_name"`
	TerminalCommand *string `json:"terminal_command"`
	PaneCommand     *string `json:"pane_command"`
	PaneCWD         *string `json:"pane_cwd"`
	PluginURL       *string `json:"plugin_url"`
}

func (raw paneJSON) pane(index int) (Pane, error) {
	path := "panes[" + strconv.Itoa(index) + "]"
	if raw.ID == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".id", "id is required", nil)
	}
	if *raw.ID < 0 {
		return Pane{}, parseError(OutputKindPanes, ParseErrorInvalidField, path+".id", "id must be non-negative", nil)
	}
	if raw.IsPlugin == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".is_plugin", "is_plugin is required", nil)
	}
	if raw.Title == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".title", "title is required", nil)
	}
	if raw.IsFocused == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".is_focused", "is_focused is required", nil)
	}
	if raw.IsFloating == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".is_floating", "is_floating is required", nil)
	}
	if raw.IsSuppressed == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".is_suppressed", "is_suppressed is required", nil)
	}
	if raw.Exited == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".exited", "exited is required", nil)
	}
	if raw.TabID == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".tab_id", "tab_id is required", nil)
	}
	if *raw.TabID < 0 {
		return Pane{}, parseError(OutputKindPanes, ParseErrorInvalidField, path+".tab_id", "tab_id must be non-negative", nil)
	}
	if raw.TabPosition == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".tab_position", "tab_position is required", nil)
	}
	if *raw.TabPosition < 0 {
		return Pane{}, parseError(OutputKindPanes, ParseErrorInvalidField, path+".tab_position", "tab_position must be non-negative", nil)
	}
	if raw.TabName == nil {
		return Pane{}, parseError(OutputKindPanes, ParseErrorMissingField, path+".tab_name", "tab_name is required", nil)
	}

	kind := PaneKindTerminal
	if *raw.IsPlugin {
		kind = PaneKindPlugin
	}
	processID := raw.ProcessID
	if processID == nil {
		processID = raw.PaneProcessID
	}
	if processID != nil && *processID <= 0 {
		return Pane{}, parseError(OutputKindPanes, ParseErrorInvalidField, path+".pid", "pid must be positive when present", nil)
	}

	return Pane{
		ID:              PaneID{Kind: kind, Number: *raw.ID},
		ProcessID:       processID,
		Title:           *raw.Title,
		IsFocused:       *raw.IsFocused,
		IsFloating:      *raw.IsFloating,
		IsSuppressed:    *raw.IsSuppressed,
		Exited:          *raw.Exited,
		ExitStatus:      raw.ExitStatus,
		TabID:           *raw.TabID,
		TabPosition:     *raw.TabPosition,
		TabName:         *raw.TabName,
		TerminalCommand: raw.TerminalCommand,
		PaneCommand:     raw.PaneCommand,
		PaneCWD:         raw.PaneCWD,
		PluginURL:       raw.PluginURL,
	}, nil
}

func parseError(kind OutputKind, code ParseErrorCode, path, message string, err error) error {
	return &ParseError{
		Kind:    kind,
		Code:    code,
		Path:    path,
		Message: message,
		Err:     err,
	}
}
