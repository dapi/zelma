package codex

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var externalSessionUUIDPattern = regexp.MustCompile(`(?i)(?:\bCODEX_EXTERNAL_SESSION_UUID=|External session UUID:\s*)([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

func FindCommandSessionEvidence(command string) SessionEvidenceResult {
	entrypoint, args, ok := CodexCommandEntrypointAndArgs(command)
	if !ok {
		return insufficient("command does not identify Codex")
	}

	if sessionID := resumeSessionID(args); sessionID != "" {
		return commandEvidence(sessionID, CodexSessionRefSourceArgvResume)
	}
	if sessionID := externalSessionUUID(command); sessionID != "" {
		return commandEvidence(sessionID, CodexSessionRefSourceArgvExternalSessionUUID)
	}
	if entrypoint == "" {
		return insufficient("command does not identify Codex")
	}
	return insufficient("Codex command does not contain resume UUID or external session UUID")
}

func CodexCommandEntrypoint(command string) string {
	entrypoint, _, ok := CodexCommandEntrypointAndArgs(command)
	if !ok {
		return ""
	}
	return entrypoint
}

func CodexCommandEntrypointAndArgs(command string) (string, []string, bool) {
	tokens := commandTokens(command)
	if len(tokens) == 0 {
		executable := CommandExecutable(command)
		if isCodexExecutableToken(executable) {
			return executable, nil, true
		}
		return "", nil, false
	}

	tokens = stripEnvCommandPrefix(tokens)
	if len(tokens) == 0 {
		return "", nil, false
	}
	if isCodexExecutableToken(tokens[0]) {
		return tokens[0], tokens[1:], true
	}
	if isNodeExecutableToken(tokens[0]) {
		entrypointIndex := nodeScriptEntrypointIndex(tokens[1:])
		if entrypointIndex == -1 {
			return "", nil, false
		}
		entrypointIndex++
		entrypoint := tokens[entrypointIndex]
		if isCodexExecutableToken(entrypoint) {
			return entrypoint, tokens[entrypointIndex+1:], true
		}
	}
	return "", nil, false
}

func CommandExecutable(command string) string {
	command = strings.TrimLeftFunc(command, unicode.IsSpace)
	if command == "" {
		return ""
	}

	if command[0] == '\'' || command[0] == '"' {
		quote := command[0]
		for i := 1; i < len(command); i++ {
			if command[i] == quote {
				return command[1:i]
			}
		}
		return ""
	}

	var builder strings.Builder
	escaped := false
	for _, r := range command {
		if escaped {
			builder.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if unicode.IsSpace(r) {
			break
		}
		builder.WriteRune(r)
	}
	if escaped {
		builder.WriteRune('\\')
	}
	return builder.String()
}

func commandTokens(command string) []string {
	var tokens []string
	var builder strings.Builder
	var quote rune
	escaped := false

	flush := func() {
		if builder.Len() == 0 {
			return
		}
		tokens = append(tokens, builder.String())
		builder.Reset()
	}

	for _, r := range command {
		if escaped {
			builder.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			builder.WriteRune(r)
			continue
		}
		if r == '\'' || r == '"' {
			quote = r
			continue
		}
		if unicode.IsSpace(r) {
			flush()
			continue
		}
		builder.WriteRune(r)
	}
	if escaped {
		builder.WriteRune('\\')
	}
	flush()
	return tokens
}

func stripEnvCommandPrefix(tokens []string) []string {
	if len(tokens) == 0 {
		return tokens
	}
	if !isEnvExecutableToken(tokens[0]) {
		return stripLeadingAssignments(tokens)
	}

	i := 1
	for i < len(tokens) {
		token := tokens[i]
		if token == "--" {
			i++
			break
		}
		if token == "-u" || token == "--unset" || token == "-C" || token == "--chdir" {
			i += 2
			continue
		}
		if strings.HasPrefix(token, "-") {
			i++
			continue
		}
		if strings.Contains(token, "=") {
			i++
			continue
		}
		break
	}
	return tokens[i:]
}

func stripLeadingAssignments(tokens []string) []string {
	i := 0
	for i < len(tokens) && strings.Contains(tokens[i], "=") && !strings.HasPrefix(tokens[i], "-") {
		i++
	}
	return tokens[i:]
}

func resumeSessionID(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] != "resume" {
			continue
		}
		if i+1 >= len(args) {
			return ""
		}
		return normalizeSessionID(args[i+1])
	}
	return ""
}

func externalSessionUUID(command string) string {
	match := externalSessionUUIDPattern.FindStringSubmatch(command)
	if len(match) != 2 {
		return ""
	}
	return normalizeSessionID(match[1])
}

func commandEvidence(sessionID string, source CodexSessionRefSource) SessionEvidenceResult {
	return SessionEvidenceResult{
		Verdict: SessionEvidenceResolved,
		Ref: &CodexSessionRef{
			SessionID:  sessionID,
			Source:     source,
			Confidence: MetadataConfidenceStrong,
		},
	}
}

func normalizeSessionID(sessionID string) string {
	sessionID = strings.ToLower(strings.TrimSpace(sessionID))
	if !uuidPattern.MatchString(sessionID) {
		return ""
	}
	return sessionID
}

func isCodexExecutableToken(token string) bool {
	if token == "" {
		return false
	}
	base := strings.ToLower(filepath.Base(token))
	return base == "codex" || base == "codex.exe"
}

func isNodeExecutableToken(token string) bool {
	base := strings.ToLower(filepath.Base(token))
	return base == "node" || base == "node.exe"
}

func isEnvExecutableToken(token string) bool {
	base := strings.ToLower(filepath.Base(token))
	return base == "env" || base == "env.exe"
}

func nodeScriptEntrypointIndex(tokens []string) int {
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token == "--" {
			if i+1 < len(tokens) {
				return i + 1
			}
			return -1
		}
		if token == "-e" || token == "--eval" || token == "-p" || token == "--print" {
			return -1
		}
		if nodeOptionConsumesValue(token) {
			i++
			continue
		}
		if strings.HasPrefix(token, "-") {
			continue
		}
		return i
	}
	return -1
}

func nodeOptionConsumesValue(token string) bool {
	switch token {
	case "-r", "--require", "--import", "--loader", "--experimental-loader", "--require-module":
		return true
	default:
		return false
	}
}
