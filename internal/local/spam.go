package local

import (
	"bufio"
	"bytes"
	_ "embed"
	"strings"
)

//go:embed data/spam-oros42.csv
var spamOros42 []byte

//go:embed data/spam-us.csv
var spamUS []byte

var spamMap map[string]string

func init() {
	spamMap = make(map[string]string, 1024)
	loadSpamCSV(spamOros42, "Oros42")
	loadSpamCSV(spamUS, "US-blocked")
}

func loadSpamCSV(blob []byte, source string) {
	sc := bufio.NewScanner(bytes.NewReader(blob))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		comma := strings.Index(line, ",")
		var num, tag string
		if comma == -1 {
			num = line
			tag = source
		} else {
			num = strings.TrimSpace(line[:comma])
			tag = strings.TrimSpace(line[comma+1:])
		}
		num = normalizeForLookup(num)
		if num == "" {
			continue
		}
		if tag == "" {
			tag = source
		} else {
			tag = source + ":" + tag
		}
		if _, exists := spamMap[num]; !exists {
			spamMap[num] = tag
		}
	}
}

func normalizeForLookup(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "tel:")
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "+") {
		return s
	}
	allDigits := true
	for _, c := range s {
		if c < '0' || c > '9' {
			allDigits = false
			break
		}
	}
	if !allDigits {
		return ""
	}
	if len(s) >= 10 {
		return "+" + s
	}
	return s
}

func lookupSpam(e164 string) (string, bool) {
	if reason, ok := spamMap[e164]; ok {
		return reason, true
	}
	return "", false
}

func SpamCount() int { return len(spamMap) }
