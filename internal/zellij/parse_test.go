package zellij

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseSessionListFT012Fixtures(t *testing.T) {
	tests := []struct {
		name string
		want []Session
	}{
		{
			name: "short-multiple.txt",
			want: []Session{
				{Name: "zelma"},
				{Name: "feature-issue-24"},
				{Name: "research"},
			},
		},
		{
			name: "short-empty.txt",
			want: []Session{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSessionList(readZellijFixture(t, "sessions", tt.name))
			if err != nil {
				t.Fatalf("parseSessionList(%s) error = %v, want nil", tt.name, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("sessions = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseSessionListRejectsFormattedFT012Fixture(t *testing.T) {
	_, err := parseSessionList(readZellijFixture(t, "sessions", "formatted-not-short.txt"))
	if err == nil {
		t.Fatal("parseSessionList() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "formatted metadata") {
		t.Fatalf("error = %q, want formatted metadata detail", err.Error())
	}
}

func TestParseListPanesValidFixtures(t *testing.T) {
	tests := []struct {
		name      string
		wantPanes int
	}{
		{name: "list-panes-all-0.44.3.json", wantPanes: 3},
		{name: "list-panes-missing-command-metadata-0.44.3.json", wantPanes: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			panes, err := ParseListPanesJSON(readZellijFixture(t, "panes", tt.name))
			if err != nil {
				t.Fatalf("ParseListPanesJSON(%s) error = %v, want nil", tt.name, err)
			}
			if len(panes) != tt.wantPanes {
				t.Fatalf("len(panes) = %d, want %d", len(panes), tt.wantPanes)
			}
		})
	}
}

func TestParseListPanesPreservesTypedPaneIDs(t *testing.T) {
	panes, err := ParseListPanesJSON(readZellijFixture(t, "panes", "list-panes-all-0.44.3.json"))
	if err != nil {
		t.Fatalf("ParseListPanesJSON() error = %v, want nil", err)
	}

	got := []string{panes[0].ID.String(), panes[1].ID.String(), panes[2].ID.String()}
	want := []string{"plugin_0", "terminal_0", "terminal_2"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pane ID %d = %q, want %q", i, got[i], want[i])
		}
	}
	if panes[0].ID.String() == panes[1].ID.String() {
		t.Fatalf("plugin and terminal panes with raw id 0 must not collapse: %+v %+v", panes[0].ID, panes[1].ID)
	}
}

func TestParseListPanesPreservesCommandAndTabMetadata(t *testing.T) {
	panes, err := ParseListPanesJSON(readZellijFixture(t, "panes", "list-panes-all-0.44.3.json"))
	if err != nil {
		t.Fatalf("ParseListPanesJSON() error = %v, want nil", err)
	}

	if panes[1].Title != "codex" {
		t.Fatalf("Title = %q, want %q", panes[1].Title, "codex")
	}
	if panes[1].PaneCommand == nil || *panes[1].PaneCommand != "/usr/local/bin/codex --cd /workspace/zelma" {
		t.Fatalf("PaneCommand = %v, want codex command", panes[1].PaneCommand)
	}
	if panes[1].PaneCWD == nil || *panes[1].PaneCWD != "/workspace/zelma" {
		t.Fatalf("PaneCWD = %v, want /workspace/zelma", panes[1].PaneCWD)
	}
	if panes[1].TabID != 1 || panes[1].TabPosition != 0 || panes[1].TabName != "work" {
		t.Fatalf("tab metadata = id:%d position:%d name:%q, want id:1 position:0 name:work", panes[1].TabID, panes[1].TabPosition, panes[1].TabName)
	}
	if panes[2].ExitStatus == nil || *panes[2].ExitStatus != 0 {
		t.Fatalf("ExitStatus = %v, want 0", panes[2].ExitStatus)
	}
}

func TestParseListPanesAllowsMissingOptionalCommandMetadata(t *testing.T) {
	panes, err := ParseListPanesJSON(readZellijFixture(t, "panes", "list-panes-missing-command-metadata-0.44.3.json"))
	if err != nil {
		t.Fatalf("ParseListPanesJSON() error = %v, want nil", err)
	}
	if panes[0].PaneCommand != nil {
		t.Fatalf("PaneCommand = %q, want nil", *panes[0].PaneCommand)
	}
	if panes[0].PaneCWD != nil {
		t.Fatalf("PaneCWD = %q, want nil", *panes[0].PaneCWD)
	}
}

func TestParseListPanesRejectsInvalidFixtures(t *testing.T) {
	tests := []struct {
		name     string
		wantCode ParseErrorCode
		wantPath string
	}{
		{
			name:     "list-panes-top-level-object.json",
			wantCode: ParseErrorInvalidJSON,
		},
		{
			name:     "list-panes-null.json",
			wantCode: ParseErrorInvalidJSON,
		},
		{
			name:     "list-panes-id-string.json",
			wantCode: ParseErrorInvalidJSON,
		},
		{
			name:     "list-panes-missing-id.json",
			wantCode: ParseErrorMissingField,
			wantPath: "panes[0].id",
		},
		{
			name:     "list-panes-missing-title.json",
			wantCode: ParseErrorMissingField,
			wantPath: "panes[0].title",
		},
		{
			name:     "list-panes-negative-id.json",
			wantCode: ParseErrorInvalidField,
			wantPath: "panes[0].id",
		},
		{
			name:     "list-panes-duplicate-typed-id.json",
			wantCode: ParseErrorInvalidField,
			wantPath: "panes[1].id",
		},
		{
			name:     "list-panes-trailing-data.json",
			wantCode: ParseErrorTrailingData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseListPanesJSON(readZellijFixture(t, "panes", tt.name))
			if err == nil {
				t.Fatal("ParseListPanesJSON() error = nil, want error")
			}

			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("ParseListPanesJSON() error = %T, want *ParseError", err)
			}
			if parseErr.Kind != OutputKindPanes {
				t.Fatalf("Kind = %q, want %q", parseErr.Kind, OutputKindPanes)
			}
			if parseErr.Code != tt.wantCode {
				t.Fatalf("Code = %q, want %q", parseErr.Code, tt.wantCode)
			}
			if parseErr.Path != tt.wantPath {
				t.Fatalf("Path = %q, want %q", parseErr.Path, tt.wantPath)
			}
		})
	}
}

func TestParsePaneIDOutput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want PaneID
	}{
		{
			name: "terminal",
			data: []byte("terminal_7\n"),
			want: PaneID{Kind: PaneKindTerminal, Number: 7},
		},
		{
			name: "plugin",
			data: []byte(" plugin_3 "),
			want: PaneID{Kind: PaneKindPlugin, Number: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePaneIDOutput(tt.data)
			if err != nil {
				t.Fatalf("ParsePaneIDOutput() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Fatalf("pane id = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParsePaneIDOutputRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: nil},
		{name: "blank", data: []byte("\n")},
		{name: "unknown kind", data: []byte("command_1\n")},
		{name: "missing separator", data: []byte("terminal1\n")},
		{name: "negative", data: []byte("terminal_-1\n")},
		{name: "control character", data: []byte("terminal_1\nplugin_2\n")},
		{name: "invalid utf8", data: []byte{0xff}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePaneIDOutput(tt.data)
			if err == nil {
				t.Fatal("ParsePaneIDOutput() error = nil, want error")
			}

			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("ParsePaneIDOutput() error = %T, want *ParseError", err)
			}
			if parseErr.Kind != OutputKindPaneID {
				t.Fatalf("Kind = %q, want %q", parseErr.Kind, OutputKindPaneID)
			}
		})
	}
}

func readZellijFixture(t *testing.T, parts ...string) []byte {
	t.Helper()

	pathParts := append([]string{"testdata"}, parts...)
	data, err := os.ReadFile(filepath.Join(pathParts...))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
