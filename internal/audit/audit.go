// Package audit appends a JSONL line per clank invocation to ~/.clank/history.jsonl
// so users can `clank history` to see what they looked up. Phone-only — never
// logs API keys, query strings beyond the first positional, or output content.
package audit

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const HistoryFile = "history.jsonl"

type Entry struct {
	Time  time.Time `json:"ts"`
	Cmd   string    `json:"cmd"`
	Phone string    `json:"phone,omitempty"`
	Code  int       `json:"exit"`
	Took  string    `json:"took,omitempty"`
}

func historyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".clank")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, HistoryFile), nil
}

// Log appends one entry. Best-effort — failures are silently swallowed so
// audit logging never breaks the user's primary command.
func Log(e Entry) {
	if os.Getenv("CLANK_NO_AUDIT") != "" {
		return
	}
	path, err := historyPath()
	if err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(e)
}

// FirstPhoneArg returns the first arg that looks phone-shaped (mostly digits
// with optional + prefix, length 6-15). Used to redact API keys and other
// secrets while still capturing the most interesting positional.
func FirstPhoneArg(args []string) string {
	for _, a := range args {
		if looksLikePhone(a) {
			return a
		}
	}
	return ""
}

func looksLikePhone(s string) bool {
	s = strings.TrimSpace(s)
	body := strings.TrimPrefix(s, "+")
	if len(body) < 6 || len(body) > 15 {
		return false
	}
	digits := 0
	for _, c := range body {
		if c >= '0' && c <= '9' {
			digits++
		} else if c != ' ' && c != '-' && c != '(' && c != ')' && c != 'x' && c != 'X' && c != '*' {
			return false
		}
	}
	return digits >= 6
}

// Read returns the last `n` entries from the history file, newest first. If
// `grep` is non-empty, only entries whose Phone or Cmd contains the substring
// (case-insensitive) are returned.
func Read(n int, grep string) ([]Entry, error) {
	path, err := historyPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var all []Entry
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		var e Entry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			continue
		}
		all = append(all, e)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	if grep != "" {
		needle := strings.ToLower(grep)
		filtered := all[:0]
		for _, e := range all {
			if strings.Contains(strings.ToLower(e.Phone), needle) ||
				strings.Contains(strings.ToLower(e.Cmd), needle) {
				filtered = append(filtered, e)
			}
		}
		all = filtered
	}

	// newest first
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	if n > 0 && len(all) > n {
		all = all[:n]
	}
	return all, nil
}

// Path returns the absolute history file path (creating the parent dir if needed).
func Path() string {
	p, _ := historyPath()
	return p
}

// Wrap runs fn(args) while timing it and emitting a history entry. Returns
// the exit code so the caller can os.Exit with it.
func Wrap(name string, args []string, fn func([]string) int) int {
	start := time.Now()
	code := fn(args)
	Log(Entry{
		Time:  start,
		Cmd:   name,
		Phone: FirstPhoneArg(args),
		Code:  code,
		Took:  time.Since(start).Round(time.Millisecond).String(),
	})
	return code
}

// LogManual is for cases where the caller wants to log without using Wrap
// (e.g., the default pattern flow that doesn't have a sub-Command function).
func LogManual(cmd, phone string, start time.Time, code int) {
	Log(Entry{
		Time:  start,
		Cmd:   cmd,
		Phone: phone,
		Code:  code,
		Took:  time.Since(start).Round(time.Millisecond).String(),
	})
}

func formatTime(t time.Time) string {
	return t.Local().Format("2006-01-02 15:04:05")
}

// FormatLine renders one history line for the human-readable history view.
func FormatLine(e Entry) string {
	phone := e.Phone
	if phone == "" {
		phone = "-"
	}
	took := e.Took
	if took == "" {
		took = "-"
	}
	status := "ok"
	if e.Code != 0 {
		status = fmt.Sprintf("exit=%d", e.Code)
	}
	return fmt.Sprintf("%s  %-10s  %-25s  %-7s  %s",
		formatTime(e.Time), e.Cmd, phone, took, status)
}
