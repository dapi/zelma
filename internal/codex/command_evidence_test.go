package codex

import "testing"

func TestFindCommandSessionEvidenceFromResumeArg(t *testing.T) {
	got := FindCommandSessionEvidence("codex resume 019F3D81-B070-7A91-9A6F-9F50F1CBA355")

	if got.Verdict != SessionEvidenceResolved {
		t.Fatalf("Verdict = %q, want resolved: %+v", got.Verdict, got)
	}
	if got.Ref == nil || got.Ref.SessionID != "019f3d81-b070-7a91-9a6f-9f50f1cba355" {
		t.Fatalf("Ref = %+v, want resume UUID", got.Ref)
	}
	if got.Ref.Source != CodexSessionRefSourceArgvResume {
		t.Fatalf("Source = %q, want %q", got.Ref.Source, CodexSessionRefSourceArgvResume)
	}
	if got.Ref.Confidence != MetadataConfidenceStrong {
		t.Fatalf("Confidence = %q, want strong", got.Ref.Confidence)
	}
}

func TestFindCommandSessionEvidenceFromResumeArgAfterGlobalOptions(t *testing.T) {
	command := "codex --dangerously-bypass-approvals-and-sandbox --search resume 019f3d81-b070-7a91-9a6f-9f50f1cba355"

	got := FindCommandSessionEvidence(command)

	if got.Verdict != SessionEvidenceResolved || got.Ref == nil || got.Ref.SessionID != "019f3d81-b070-7a91-9a6f-9f50f1cba355" {
		t.Fatalf("evidence = %+v, want resume UUID after global options", got)
	}
}

func TestFindCommandSessionEvidenceFromNodeCodexResumeArg(t *testing.T) {
	command := `node --require ./register.js "/opt/openai/bin/codex" resume 019f3d81-b070-7a91-9a6f-9f50f1cba355`

	got := FindCommandSessionEvidence(command)

	if got.Verdict != SessionEvidenceResolved || got.Ref == nil || got.Ref.SessionID != "019f3d81-b070-7a91-9a6f-9f50f1cba355" {
		t.Fatalf("evidence = %+v, want node Codex resume UUID", got)
	}
}

func TestFindCommandSessionEvidenceFromExternalSessionUUIDArg(t *testing.T) {
	command := `codex --search -c "developer_instructions='External session UUID: 019f3d81-b070-7a91-9a6f-9f50f1cba355. Metadata only.'"`

	got := FindCommandSessionEvidence(command)

	if got.Verdict != SessionEvidenceResolved {
		t.Fatalf("Verdict = %q, want resolved: %+v", got.Verdict, got)
	}
	if got.Ref == nil || got.Ref.SessionID != "019f3d81-b070-7a91-9a6f-9f50f1cba355" {
		t.Fatalf("Ref = %+v, want external UUID", got.Ref)
	}
	if got.Ref.Source != CodexSessionRefSourceArgvExternalSessionUUID {
		t.Fatalf("Source = %q, want %q", got.Ref.Source, CodexSessionRefSourceArgvExternalSessionUUID)
	}
}

func TestFindCommandSessionEvidenceFromEnvAssignment(t *testing.T) {
	command := `env CODEX_EXTERNAL_SESSION_UUID=019f3d81-b070-7a91-9a6f-9f50f1cba355 codex --search`

	got := FindCommandSessionEvidence(command)

	if got.Verdict != SessionEvidenceResolved || got.Ref == nil || got.Ref.SessionID != "019f3d81-b070-7a91-9a6f-9f50f1cba355" {
		t.Fatalf("evidence = %+v, want external UUID from env assignment", got)
	}
}

func TestFindCommandSessionEvidenceRejectsNonCodexCommand(t *testing.T) {
	got := FindCommandSessionEvidence("grep resume 019f3d81-b070-7a91-9a6f-9f50f1cba355 README.md")

	if got.Verdict != SessionEvidenceInsufficient {
		t.Fatalf("Verdict = %q, want insufficient: %+v", got.Verdict, got)
	}
	if got.Ref != nil {
		t.Fatalf("Ref = %+v, want nil for non-Codex command", got.Ref)
	}
}
