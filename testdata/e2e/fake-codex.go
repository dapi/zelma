package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	openedPath, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--cd" && i+1 < len(os.Args) {
			openedPath = os.Args[i+1]
			i++
		}
	}

	codexHome := os.Getenv("CODEX_HOME")
	if codexHome == "" {
		fatal(fmt.Errorf("CODEX_HOME is required"))
	}

	sessionID, err := uuidV4()
	if err != nil {
		fatal(err)
	}

	sessionDir := filepath.Join(codexHome, "sessions", "2026", "07", "08")
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		fatal(err)
	}

	file, err := os.Create(filepath.Join(sessionDir, "session-"+sessionID+".jsonl"))
	if err != nil {
		fatal(err)
	}
	defer file.Close()

	record := map[string]any{
		"type": "session_meta",
		"payload": map[string]any{
			"session_id":  sessionID,
			"cwd":         filepath.Clean(openedPath),
			"cli_version": "fake-codex 0.0.0",
			"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
		},
	}
	if err := json.NewEncoder(file).Encode(record); err != nil {
		fatal(err)
	}

	for {
		time.Sleep(time.Hour)
	}
}

func uuidV4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
