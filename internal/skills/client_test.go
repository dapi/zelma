package skills

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListInstancesInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "version": 1,
  "instances": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/zelma",
      "state": "active"
    }
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.ListInstances(context.Background(), ListOptions{})

	if err != nil {
		t.Fatalf("ListInstances() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "list", "--json")
	if got.Version != 1 || len(got.Instances) != 1 {
		t.Fatalf("ListInstances() = %+v, want schema v1 with one instance", got)
	}
	if got.Instances[0].ZellijSession != "zelma-main" || got.Instances[0].State != "active" {
		t.Fatalf("session = %+v, want parsed CLI instance", got.Instances[0])
	}
}

func TestListInstancesLiveInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "version": 1,
  "instances": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_1",
      "codex_session": "",
      "opened_path": "/workspace/zelma",
      "state": "candidate",
      "live_status": "live"
    }
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.ListInstances(context.Background(), ListOptions{Live: true})

	if err != nil {
		t.Fatalf("ListInstances() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "list", "--live", "--json")
	if got.Instances[0].LiveStatus != "live" {
		t.Fatalf("live_status = %q, want live", got.Instances[0].LiveStatus)
	}
}

func TestCreateInstanceInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "created": 1,
  "registered": 1,
  "skipped": 0
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.CreateInstance(context.Background(), "nested/path")

	if err != nil {
		t.Fatalf("CreateInstance() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "create", "nested/path", "--json")
	if got.Created != 1 || got.Registered != 1 || got.Skipped != 0 {
		t.Fatalf("CreateInstance() = %+v, want created=1 registered=1 skipped=0", got)
	}
}

func TestCreateInstanceParsesDuplicateGuardSkip(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "created": 0,
  "registered": 0,
  "skipped": 1,
  "instance": {
    "id": 3,
    "zellij_session": "zelma-main",
    "zellij_pane": "terminal_1",
    "codex_session": "11111111-1111-4111-8111-111111111111",
    "opened_path": "/workspace/zelma",
    "state": "active"
  }
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.CreateInstance(context.Background(), "")

	if err != nil {
		t.Fatalf("CreateInstance() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "create", "--json")
	if got.Created != 0 || got.Registered != 0 || got.Skipped != 1 {
		t.Fatalf("CreateInstance() = %+v, want duplicate guard skipped result", got)
	}
	if got.Instance.ID != 3 || got.Instance.State != "active" || got.Instance.OpenedPath != "/workspace/zelma" {
		t.Fatalf("Instance = %+v, want existing active instance", got.Instance)
	}
}

func TestPreviewCreateInstanceInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "opened_path": "/workspace/zelma",
  "working_directory": "/workspace/zelma",
  "binary": "/usr/local/bin/codex",
  "args": [
    "--cd",
    "/workspace/zelma"
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.PreviewCreateInstance(context.Background(), "")

	if err != nil {
		t.Fatalf("PreviewCreateInstance() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "create", "--dry-run", "--json")
	if got.OpenedPath != "/workspace/zelma" || got.Binary == "" || len(got.Args) != 2 {
		t.Fatalf("PreviewCreateInstance() = %+v, want launch contract", got)
	}
}

func TestDetectInstancesInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "added": 0,
  "unchanged": 0,
  "skipped": 1,
  "active": 0,
  "candidate": 0,
  "stale": 1,
  "stale_candidates": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_9",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/zelma",
      "previous_state": "active",
      "reason": "missing_pane"
    }
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.DetectInstances(context.Background())

	if err != nil {
		t.Fatalf("DetectInstances() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "detect", "--json")
	if got.Stale != 1 || len(got.StaleCandidates) != 1 || got.StaleCandidates[0].Reason != "missing_pane" {
		t.Fatalf("DetectInstances() = %+v, want one stale candidate", got)
	}
}

func TestFocusInstanceInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "id": 2,
  "zellij_session": "zelma-main",
  "zellij_tab": "tab_6",
  "zellij_pane": "terminal_75",
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "opened_path": "/workspace/zelma",
  "state": "active"
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.FocusInstance(context.Background(), 2)

	if err != nil {
		t.Fatalf("FocusInstance() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "focus", "2", "--json")
	if got.ID != 2 || got.ZellijPane != "terminal_75" || got.State != "active" {
		t.Fatalf("FocusInstance() = %+v, want focused active instance 2", got)
	}
}

func TestSendMessageInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "id": 2,
  "zellij_session": "zelma-main",
  "zellij_tab": "tab_6",
  "zellij_pane": "terminal_75",
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "opened_path": "/workspace/zelma",
  "state": "active",
  "message": {
    "source": "argument",
    "byte_count": 18,
    "line_count": 1,
    "submitted": true
  }
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.SendMessage(context.Background(), 2, "continue carefully")

	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "send", "2", "--json", "--", "continue carefully")
	if got.ID != 2 || got.ZellijPane != "terminal_75" || got.Message.Source != "argument" || !got.Message.Submitted {
		t.Fatalf("SendMessage() = %+v, want sent active instance metadata", got)
	}
}

func TestSendMessageFromStdinInvokesZelmaCLIWithStdin(t *testing.T) {
	root := t.TempDir()
	runner := &recordingRunner{
		result: CommandResult{
			Stdout: []byte(`{
  "id": 2,
  "zellij_session": "zelma-main",
  "zellij_pane": "terminal_75",
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "opened_path": "/workspace/zelma",
  "state": "active",
  "message": {
    "source": "stdin",
    "byte_count": 17,
    "line_count": 2,
    "submitted": true
  }
}
`),
		},
	}
	client := Client{Binary: "zelma-test", WorkDir: root, Runner: runner}

	got, err := client.SendMessageFromStdin(context.Background(), 2, []byte("line one\nline two"))

	if err != nil {
		t.Fatalf("SendMessageFromStdin() error = %v", err)
	}
	if runner.calls != 1 {
		t.Fatalf("runner calls = %d, want 1", runner.calls)
	}
	wantArgs := []string{"instances", "send", "2", "--stdin", "--json"}
	if strings.Join(runner.request.Args, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %#v, want %#v", runner.request.Args, wantArgs)
	}
	if runner.request.Binary != "zelma-test" || runner.request.WorkDir != root {
		t.Fatalf("request = %+v, want configured binary/workdir", runner.request)
	}
	if !runner.request.HasStdin || string(runner.request.Stdin) != "line one\nline two" {
		t.Fatalf("stdin = has:%t %q, want exact message bytes", runner.request.HasStdin, string(runner.request.Stdin))
	}
	if got.Message.Source != "stdin" || got.Message.LineCount != 2 || !got.Message.Submitted {
		t.Fatalf("SendMessageFromStdin() = %+v, want stdin metadata", got)
	}
}

func TestObserveInstanceBufferInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "version": 1,
  "instance_id": 2,
  "source": "zellij_buffer",
  "captured_at": "2026-07-10T00:00:00Z",
  "truncated": true,
  "limit": 2,
  "items": [
    {
      "line": 3,
      "text": "synthetic pane line"
    }
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.ObserveInstanceBuffer(context.Background(), 2, 2)

	if err != nil {
		t.Fatalf("ObserveInstanceBuffer() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "buffer", "2", "--json", "--tail", "2")
	if got.Version != 1 || got.Source != "zellij_buffer" || got.InstanceID != 2 || len(got.Items) != 1 {
		t.Fatalf("ObserveInstanceBuffer() = %+v, want parsed buffer observation", got)
	}
	if got.Items[0].Line != 3 || got.Items[0].Text != "synthetic pane line" {
		t.Fatalf("Buffer item = %+v, want parsed line", got.Items[0])
	}
}

func TestObserveInstanceTranscriptInvokesZelmaCLI(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "version": 1,
  "instance_id": 2,
  "source": "codex_transcript",
  "captured_at": "2026-07-10T00:00:00Z",
  "truncated": false,
  "limit": 1,
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "items": [
    {
      "index": 4,
      "type": "assistant_message",
      "payload": {
        "text": "synthetic answer"
      }
    }
  ]
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	got, err := client.ObserveInstanceTranscript(context.Background(), 2, 1)

	if err != nil {
		t.Fatalf("ObserveInstanceTranscript() error = %v", err)
	}
	assertCall(t, calls, root, "instances", "transcript", "2", "--json", "--tail", "1")
	if got.Version != 1 || got.Source != "codex_transcript" || got.CodexSession == "" || len(got.Items) != 1 {
		t.Fatalf("ObserveInstanceTranscript() = %+v, want parsed transcript observation", got)
	}
	if got.Items[0].Index != 4 || got.Items[0].Type != "assistant_message" || !strings.Contains(string(got.Items[0].Payload), "synthetic answer") {
		t.Fatalf("Transcript item = %+v, want parsed event", got.Items[0])
	}
}

func TestCommandErrorPreservesDiagnosticsAndRecovery(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma instances list: registry_unsupported_version: unsupported schema version 2\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.ListInstances(context.Background(), ListOptions{})

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("ListInstances() error = %T, want CommandError", err)
	}
	if commandErr.ExitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", commandErr.ExitCode)
	}
	if !strings.Contains(commandErr.Stderr, "registry_unsupported_version") {
		t.Fatalf("Stderr = %q, want preserved CLI diagnostic", commandErr.Stderr)
	}
	if !strings.Contains(commandErr.Recovery.Message, "schema v1") {
		t.Fatalf("Recovery = %+v, want schema recovery", commandErr.Recovery)
	}
	if commandErr.Recovery.Action != RecoveryActionStop || commandErr.Recovery.ReasonCode != "registry_unsupported_version" {
		t.Fatalf("Recovery = %+v, want stop for registry_unsupported_version", commandErr.Recovery)
	}
	assertCall(t, calls, root, "instances", "list", "--json")
}

func TestListInstancesRejectsUnsupportedSchemaVersion(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "version": 2,
  "instances": []
}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	_, err := client.ListInstances(context.Background(), ListOptions{})

	var contractErr *ContractError
	if !errors.As(err, &contractErr) {
		t.Fatalf("ListInstances() error = %T, want ContractError", err)
	}
	if !strings.Contains(contractErr.Error(), "schema version 2") {
		t.Fatalf("ContractError = %v, want schema version diagnostic", contractErr)
	}
	assertCall(t, calls, root, "instances", "list", "--json")
}

func TestCreatePartialFailureSuggestsDetectRecovery(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma instances create: create instance: create_pane_unconfirmed: created pane could not be confirmed; recovery: run zelma instances detect --json\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.CreateInstance(context.Background(), "")

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("CreateInstance() error = %T, want CommandError", err)
	}
	if commandErr.Recovery.Action != RecoveryActionDetect || commandErr.Recovery.ReasonCode != "create_pane_unconfirmed" {
		t.Fatalf("Recovery = %+v, want detect for create_pane_unconfirmed", commandErr.Recovery)
	}
	assertRecoveryCommand(t, commandErr.Recovery, DefaultZelmaBinary, "instances", "detect", "--json")
}

func TestRecoveryJSONUsesAgentContractFields(t *testing.T) {
	data, err := json.Marshal(Recovery{
		Action:      RecoveryActionDetect,
		ReasonCode:  "create_pane_unconfirmed",
		Message:     "reconcile through detect",
		NextCommand: []string{DefaultZelmaBinary, "instances", "detect", "--json"},
	})
	if err != nil {
		t.Fatalf("Marshal(Recovery) error = %v", err)
	}

	got := string(data)
	for _, want := range []string{`"action"`, `"reason_code"`, `"message"`, `"next_command"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("Recovery JSON = %s, want field %s", got, want)
		}
	}
	for _, unwanted := range []string{`"Action"`, `"ReasonCode"`, `"NextCommand"`} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("Recovery JSON = %s, must not expose Go field %s", got, unwanted)
		}
	}
}

func TestRepoNotReadyErrorSuggestsSetup(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma instances list: unsupported repo: no Git worktree found from /tmp/outside\nhint: run zelma instances list from inside a Git repository\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.ListInstances(context.Background(), ListOptions{})

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("ListInstances() error = %T, want CommandError", err)
	}
	if commandErr.Recovery.Action != RecoveryActionSetup || commandErr.Recovery.ReasonCode != ReasonUnsupportedRepo {
		t.Fatalf("Recovery = %+v, want setup for unsupported repo", commandErr.Recovery)
	}
	assertRecoveryCommand(t, commandErr.Recovery, DefaultZelmaBinary, "setup")
}

func TestZellijUnavailableErrorStopsForEnvironmentFix(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma instances detect: zellij adapter: zellij_missing_binary: zellij binary was not found; recovery: install zellij or configure the adapter binary path\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.DetectInstances(context.Background())

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("DetectInstances() error = %T, want CommandError", err)
	}
	if commandErr.Recovery.Action != RecoveryActionStop || commandErr.Recovery.ReasonCode != "zellij_missing_binary" {
		t.Fatalf("Recovery = %+v, want stop for zellij_missing_binary", commandErr.Recovery)
	}
	if len(commandErr.Recovery.NextCommand) != 0 {
		t.Fatalf("NextCommand = %#v, want no automatic retry command", commandErr.Recovery.NextCommand)
	}
}

func TestSendNotReadySuggestsLiveListOnly(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma instances send: send message: codex_runtime_missing: pane command evidence does not indicate Codex; recovery: run zelma instances list --live --json\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.SendMessage(context.Background(), 2, "SECRET_PROMPT_BODY")

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("SendMessage() error = %T, want CommandError", err)
	}
	if commandErr.Recovery.Action != RecoveryActionInspect || commandErr.Recovery.ReasonCode != "send_target_not_ready" {
		t.Fatalf("Recovery = %+v, want inspect for send_target_not_ready", commandErr.Recovery)
	}
	assertRecoveryCommand(t, commandErr.Recovery, DefaultZelmaBinary, "instances", "list", "--live", "--json")
	if strings.Contains(commandErr.Error(), "SECRET_PROMPT_BODY") || strings.Contains(strings.Join(commandErr.Command, " "), "SECRET_PROMPT_BODY") {
		t.Fatalf("CommandError leaked message body: command=%#v error=%q", commandErr.Command, commandErr.Error())
	}
}

func TestSendMessageWithDashPrefixedPromptUsesSeparatorAndRedactsDiagnostics(t *testing.T) {
	root := t.TempDir()
	stderr := writeFile(t, root, "stderr.txt", "zelma instances send: send message: codex_runtime_missing: pane command evidence does not indicate Codex\n")
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, "", stderr, "1", calls)

	_, err := client.SendMessage(context.Background(), 2, "-SECRET_PROMPT_BODY")

	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("SendMessage() error = %T, want CommandError", err)
	}
	assertCall(t, calls, root, "instances", "send", "2", "--json", "--", "-SECRET_PROMPT_BODY")
	if strings.Contains(commandErr.Error(), "-SECRET_PROMPT_BODY") || strings.Contains(strings.Join(commandErr.Command, " "), "-SECRET_PROMPT_BODY") {
		t.Fatalf("CommandError leaked dash-prefixed message body: command=%#v error=%q", commandErr.Command, commandErr.Error())
	}
	if !strings.Contains(strings.Join(commandErr.Command, " "), "<redacted message>") {
		t.Fatalf("CommandError command = %#v, want redacted message placeholder", commandErr.Command)
	}
}

func TestEmptyRegistryWithLikelyLivePanesSuggestsDetect(t *testing.T) {
	recovery := RecoveryForListResult(InstanceList{Version: InstanceSchemaVersion}, ListRecoveryOptions{
		LivePanesLikely: true,
	})

	if recovery.Action != RecoveryActionDetect || recovery.ReasonCode != ReasonEmptyRegistryPanesLikely {
		t.Fatalf("Recovery = %+v, want detect for likely live panes", recovery)
	}
	assertRecoveryCommand(t, recovery, DefaultZelmaBinary, "instances", "detect", "--json")
}

func TestEmptyRegistryWithoutLikelyLivePanesHasNoRecovery(t *testing.T) {
	recovery := RecoveryForListResult(InstanceList{Version: InstanceSchemaVersion}, ListRecoveryOptions{})

	if !isZeroRecovery(recovery) {
		t.Fatalf("Recovery = %+v, want zero recovery", recovery)
	}
}

func TestDetectStaleResultSuggestsCleanupPreviewOnly(t *testing.T) {
	recovery := RecoveryForDetectResult(DetectSummary{
		Stale: 1,
		StaleCandidates: []StaleCandidate{
			{
				ZellijSession: "zelma-main",
				ZellijPane:    "terminal_9",
				PreviousState: "active",
				Reason:        "missing_pane",
			},
		},
	})

	if recovery.Action != RecoveryActionInspect || recovery.ReasonCode != ReasonStaleSessionsDetected {
		t.Fatalf("Recovery = %+v, want inspect for stale instances", recovery)
	}
	assertRecoveryCommand(t, recovery, DefaultZelmaBinary, "instances", "cleanup", "--json")
}

func TestRecoveryCommandsStayInsideSafeZelmaSurface(t *testing.T) {
	recoveries := []Recovery{
		recoveryFor("zelma instances list: unsupported repo: no Git worktree found"),
		recoveryFor("zelma instances create: create instance: create_pane_unconfirmed: created pane could not be confirmed"),
		recoveryFor("zelma instances create: create instance: create_registry_write_failed: write instances registry failed"),
		recoveryFor("zelma instances send: send message: codex_runtime_missing: pane command evidence does not indicate Codex"),
		RecoveryForListResult(InstanceList{Version: InstanceSchemaVersion}, ListRecoveryOptions{LivePanesLikely: true}),
		RecoveryForDetectResult(DetectSummary{Stale: 1}),
	}

	for _, recovery := range recoveries {
		command := recovery.NextCommand
		if len(command) == 0 {
			continue
		}
		if command[0] != DefaultZelmaBinary {
			t.Fatalf("Recovery = %+v, want next command to use zelma CLI", recovery)
		}
		joined := strings.Join(command, " ")
		for _, forbidden := range []string{"zellij", ".zelma", "instances.json", "--confirm"} {
			if strings.Contains(joined, forbidden) {
				t.Fatalf("Recovery = %+v, command must not contain %q", recovery, forbidden)
			}
		}
	}
}

func TestDecodeRejectsTrailingData(t *testing.T) {
	root := t.TempDir()
	stdout := writeFile(t, root, "stdout.json", `{
  "created": 1,
  "registered": 1,
  "skipped": 0
}
{}
`)
	calls := filepath.Join(root, "calls.txt")
	client := fakeCLIClient(t, root, stdout, "", "0", calls)

	_, err := client.CreateInstance(context.Background(), "")

	var decodeErr *DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("CreateInstance() error = %T, want DecodeError", err)
	}
	if !strings.Contains(decodeErr.Stdout, `"created": 1`) {
		t.Fatalf("DecodeError stdout = %q, want preserved stdout", decodeErr.Stdout)
	}
}

type recordingRunner struct {
	request CommandRequest
	result  CommandResult
	err     error
	calls   int
}

func (runner *recordingRunner) Run(ctx context.Context, request CommandRequest) (CommandResult, error) {
	runner.calls++
	runner.request = request
	return runner.result, runner.err
}

func fakeCLIClient(t *testing.T, workDir, stdoutPath, stderrPath, exitCode, callPath string) Client {
	t.Helper()

	binary := writeFakeZelma(t, t.TempDir())
	return Client{
		Binary:  binary,
		WorkDir: workDir,
		Env: []string{
			"ZELMA_FAKE_STDOUT=" + stdoutPath,
			"ZELMA_FAKE_STDERR=" + stderrPath,
			"ZELMA_FAKE_EXIT=" + exitCode,
			"ZELMA_FAKE_CALLS=" + callPath,
		},
	}
}

func writeFakeZelma(t *testing.T, dir string) string {
	t.Helper()

	path := filepath.Join(dir, "zelma")
	content := `#!/bin/sh
{
  printf 'pwd=%s\n' "$PWD"
  printf 'args='
  for arg in "$@"; do
    printf '[%s]' "$arg"
  done
  printf '\n'
} >> "$ZELMA_FAKE_CALLS"
if [ -n "$ZELMA_FAKE_STDOUT" ]; then
  cat "$ZELMA_FAKE_STDOUT"
fi
if [ -n "$ZELMA_FAKE_STDERR" ]; then
  cat "$ZELMA_FAKE_STDERR" >&2
fi
exit "${ZELMA_FAKE_EXIT:-0}"
`
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake zelma: %v", err)
	}
	return path
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func assertCall(t *testing.T, path, wantWorkDir string, wantArgs ...string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fake CLI calls: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 || !strings.HasPrefix(lines[0], "pwd=") || !strings.HasPrefix(lines[1], "args=") {
		t.Fatalf("fake CLI call = %q, want pwd and args lines", string(data))
	}
	gotWorkDir := strings.TrimPrefix(lines[0], "pwd=")
	if resolvedPath(t, gotWorkDir) != resolvedPath(t, wantWorkDir) {
		t.Fatalf("workdir = %q, want %q", gotWorkDir, wantWorkDir)
	}
	wantArgsLine := "args=" + bracketArgs(wantArgs)
	if lines[1] != wantArgsLine {
		t.Fatalf("args line = %q, want %q", lines[1], wantArgsLine)
	}
}

func resolvedPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve path %q: %v", path, err)
	}
	return resolved
}

func bracketArgs(args []string) string {
	var b strings.Builder
	for _, arg := range args {
		b.WriteString("[")
		b.WriteString(arg)
		b.WriteString("]")
	}
	return b.String()
}

func assertRecoveryCommand(t *testing.T, recovery Recovery, want ...string) {
	t.Helper()

	if strings.Join(recovery.NextCommand, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("NextCommand = %#v, want %#v", recovery.NextCommand, want)
	}
}

func isZeroRecovery(recovery Recovery) bool {
	return recovery.Action == "" &&
		recovery.ReasonCode == "" &&
		recovery.Message == "" &&
		len(recovery.NextCommand) == 0
}
