// Package scan runs `deep` over every candidate from a pattern (or list),
// rate-limit-aware so messenger lookups don't trigger bans. Filters early via
// libphonenumber to drop invalid candidates before any HTTP/IQ traffic, then
// processes the survivors sequentially (or with --workers N for offline-only
// scans), checkpointing each completed result so a crash mid-scan never loses
// work.
package scan

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/AnshumanAtrey/clank/internal/deep"
	"github.com/AnshumanAtrey/clank/internal/local"
	"github.com/AnshumanAtrey/clank/internal/pattern"
)

const (
	DefaultMax     = 30
	DefaultSleep   = 8 * time.Second
	DefaultWorkers = 1
	DefaultTopN    = 20
	HardCapDeep    = 60 * time.Second
)

type Sample string

const (
	SampleHead   Sample = "head"
	SampleStride Sample = "stride"
	SampleRandom Sample = "random"
)

type Options struct {
	Pattern   string        // pattern with x placeholders OR single phone
	Phones    []string      // explicit phone list (alternative to Pattern)
	Region    string        // libphonenumber default region (e.g. "IN", "US")
	Max       int           // cap on candidates after dedup; 0 = unlimited
	Sample    Sample        // selection strategy when len(candidates) > Max
	Quick     bool          // skip messengers (telegram/whatsapp/ignorant)
	SkipAPIs  bool          // skip Tier 2 API providers
	SkipEdgar bool          // skip SEC EDGAR
	Sleep     time.Duration // between candidates (rate-limit floor)
	Workers   int           // concurrent candidates; 1 = sequential
	Force     bool          // allow patterns producing >100k combinations
	Out       string        // checkpoint JSONL path; auto-named if empty
	Resume    bool          // when Out exists, skip already-processed phones
	Progress  io.Writer     // progress lines (typically os.Stderr)
}

type Result struct {
	Phone string       `json:"phone"`
	Score int          `json:"score"`
	Took  string       `json:"took"`
	Deep  *deep.Result `json:"deep"`
}

type Summary struct {
	Pattern         string        `json:"pattern,omitempty"`
	TotalCandidates int           `json:"total_candidates"`
	ValidAfterParse int           `json:"valid_after_parse"`
	SampledTo       int           `json:"sampled_to"`
	Processed       int           `json:"processed"`
	SkippedResume   int           `json:"skipped_resume,omitempty"`
	StartedAt       time.Time     `json:"started_at"`
	Took            time.Duration `json:"-"`
	TookHuman       string        `json:"took"`
	OutFile         string        `json:"out_file,omitempty"`
	Results         []*Result     `json:"results"`
}

func Run(ctx context.Context, opts Options) (*Summary, error) {
	start := time.Now()
	if opts.Sample == "" {
		opts.Sample = SampleHead
	}
	if opts.Workers < 1 {
		opts.Workers = DefaultWorkers
	}
	if opts.Progress == nil {
		opts.Progress = io.Discard
	}

	candidates, err := buildCandidates(opts)
	if err != nil {
		return nil, err
	}
	total := len(candidates)
	if total == 0 {
		return nil, errors.New("no candidates produced from pattern / input")
	}

	valid := filterValid(candidates, opts.Region)
	if len(valid) == 0 {
		return nil, fmt.Errorf("no valid phone numbers after libphonenumber filter (try --region <ISO>; %d candidates parsed but none valid)", total)
	}

	sampled := sampleCandidates(valid, opts.Max, opts.Sample)

	// Open / resume checkpoint.
	outPath, ckpt, err := openCheckpoint(opts.Out, opts.Resume)
	if err != nil {
		return nil, err
	}
	defer ckpt.Close()

	processed := []*Result{}
	skippedResume := 0

	already := map[string]struct{}{}
	if opts.Resume {
		already, _ = readCheckpoint(outPath)
	}

	queue := make([]string, 0, len(sampled))
	for _, p := range sampled {
		if _, done := already[p]; done {
			skippedResume++
			continue
		}
		queue = append(queue, p)
	}

	if opts.Workers == 1 {
		for i, phone := range queue {
			if ctx.Err() != nil {
				break
			}
			r := runOne(ctx, phone, opts, i+1, len(queue))
			processed = append(processed, r)
			ckpt.write(r)
			if i < len(queue)-1 && opts.Sleep > 0 {
				select {
				case <-time.After(opts.Sleep):
				case <-ctx.Done():
					break
				}
			}
		}
	} else {
		// Parallel: only sensible for --quick (no messenger ban risk).
		var wg sync.WaitGroup
		var mu sync.Mutex
		sem := make(chan struct{}, opts.Workers)
		results := make([]*Result, len(queue))
		for i, phone := range queue {
			i, phone := i, phone
			wg.Add(1)
			sem <- struct{}{}
			go func() {
				defer wg.Done()
				defer func() { <-sem }()
				if ctx.Err() != nil {
					return
				}
				r := runOne(ctx, phone, opts, i+1, len(queue))
				results[i] = r
				mu.Lock()
				ckpt.write(r)
				mu.Unlock()
			}()
		}
		wg.Wait()
		for _, r := range results {
			if r != nil {
				processed = append(processed, r)
			}
		}
	}

	// Sort by score desc, then phone asc for stability.
	sort.SliceStable(processed, func(i, j int) bool {
		if processed[i].Score != processed[j].Score {
			return processed[i].Score > processed[j].Score
		}
		return processed[i].Phone < processed[j].Phone
	})

	took := time.Since(start)
	return &Summary{
		Pattern:         opts.Pattern,
		TotalCandidates: total,
		ValidAfterParse: len(valid),
		SampledTo:       len(sampled),
		Processed:       len(processed),
		SkippedResume:   skippedResume,
		StartedAt:       start,
		Took:            took,
		TookHuman:       took.Round(time.Second).String(),
		OutFile:         outPath,
		Results:         processed,
	}, nil
}

func runOne(ctx context.Context, phone string, opts Options, idx, total int) *Result {
	t0 := time.Now()
	dctx, cancel := context.WithTimeout(ctx, HardCapDeep)
	defer cancel()
	d := deep.Run(dctx, phone, deep.Options{
		Region:         opts.Region,
		SkipMessengers: opts.Quick,
		SkipAPIs:       opts.SkipAPIs,
		SkipEdgar:      opts.SkipEdgar,
		Timeout:        HardCapDeep,
	})
	r := &Result{
		Phone: phone,
		Deep:  d,
		Score: scoreResult(d),
		Took:  time.Since(t0).Round(time.Millisecond).String(),
	}
	emitProgress(opts.Progress, idx, total, r)
	return r
}

func emitProgress(w io.Writer, idx, total int, r *Result) {
	if w == io.Discard {
		return
	}
	tags := []string{}
	if r.Deep != nil {
		if r.Deep.Local != nil && r.Deep.Local.Valid {
			tags = append(tags, "valid")
		}
		if r.Deep.WhatsApp != nil && r.Deep.WhatsApp.Result != nil && r.Deep.WhatsApp.Result.Registered {
			tags = append(tags, "wa")
		}
		if r.Deep.Telegram != nil && r.Deep.Telegram.Result != nil && r.Deep.Telegram.Result.Found {
			tags = append(tags, "tg")
		}
		if r.Deep.Edgar != nil && r.Deep.Edgar.Result != nil && r.Deep.Edgar.Result.Total > 0 {
			tags = append(tags, fmt.Sprintf("edgar=%d", r.Deep.Edgar.Result.Total))
		}
		if r.Deep.Ignorant != nil {
			for _, ig := range r.Deep.Ignorant.Results {
				if ig.Exists {
					tags = append(tags, ig.Name)
				}
			}
		}
	}
	tagStr := "-"
	if len(tags) > 0 {
		tagStr = strings.Join(tags, ",")
	}
	fmt.Fprintf(w, "[%d/%d] %-18s score=%d took=%s  %s\n",
		idx, total, r.Phone, r.Score, r.Took, tagStr)
}

func buildCandidates(opts Options) ([]string, error) {
	if len(opts.Phones) > 0 {
		seen := map[string]struct{}{}
		out := make([]string, 0, len(opts.Phones))
		for _, p := range opts.Phones {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if _, dup := seen[p]; dup {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
		return out, nil
	}
	if opts.Pattern == "" {
		return nil, errors.New("no pattern or phones provided")
	}
	return pattern.Generate(opts.Pattern, opts.Force)
}

// filterValid runs libphonenumber on every candidate (offline, cheap,
// trivially parallelisable) and returns only those it considers valid.
// Drops the noise from broad patterns before any expensive enrichment.
func filterValid(candidates []string, region string) []string {
	type item struct {
		idx     int
		phone   string
		isValid bool
	}
	results := make([]item, len(candidates))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 16)
	for i, c := range candidates {
		i, c := i, c
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			l := local.Inspect(c, region)
			results[i] = item{idx: i, phone: l.E164, isValid: l.Valid}
			if l.E164 == "" {
				results[i].phone = c
			}
		}()
	}
	wg.Wait()

	out := make([]string, 0, len(results)/2)
	seen := map[string]struct{}{}
	for _, r := range results {
		if !r.isValid {
			continue
		}
		if _, dup := seen[r.phone]; dup {
			continue
		}
		seen[r.phone] = struct{}{}
		out = append(out, r.phone)
	}
	return out
}

func sampleCandidates(all []string, max int, strategy Sample) []string {
	if max <= 0 || max >= len(all) {
		return all
	}
	switch strategy {
	case SampleStride:
		step := len(all) / max
		if step < 1 {
			step = 1
		}
		out := make([]string, 0, max)
		for i := 0; i < len(all) && len(out) < max; i += step {
			out = append(out, all[i])
		}
		return out
	case SampleRandom:
		idxs := rand.Perm(len(all))[:max]
		sort.Ints(idxs)
		out := make([]string, max)
		for i, idx := range idxs {
			out[i] = all[idx]
		}
		return out
	default:
		return all[:max]
	}
}

func scoreResult(d *deep.Result) int {
	if d == nil {
		return 0
	}
	s := 0
	if d.Local != nil && d.Local.Valid {
		s++
	}
	if d.Local != nil && d.Local.Spam {
		s -= 2
	}
	for _, a := range d.APIs {
		if a != nil && a.Result != nil && a.Result.Valid {
			s++
		}
	}
	if d.Telegram != nil && d.Telegram.Result != nil && d.Telegram.Result.Found {
		s += 2
		if d.Telegram.Result.Username != "" {
			s++
		}
	}
	if d.WhatsApp != nil && d.WhatsApp.Result != nil && d.WhatsApp.Result.Registered {
		s += 2
		if d.WhatsApp.Result.About != "" || d.WhatsApp.Result.VerifiedBusinessName != "" {
			s++
		}
	}
	if d.Ignorant != nil {
		for _, r := range d.Ignorant.Results {
			if r.Exists {
				s++
			}
		}
	}
	if d.Edgar != nil && d.Edgar.Result != nil && d.Edgar.Result.Total > 0 {
		s += 2
	}
	return s
}

// ----- checkpoint I/O -----

type checkpoint struct {
	f *os.File
	w *bufio.Writer
}

func (c *checkpoint) write(r *Result) {
	if c == nil || c.w == nil {
		return
	}
	b, err := json.Marshal(r)
	if err != nil {
		return
	}
	_, _ = c.w.Write(b)
	_, _ = c.w.WriteString("\n")
	_ = c.w.Flush()
}

func (c *checkpoint) Close() {
	if c == nil {
		return
	}
	if c.w != nil {
		_ = c.w.Flush()
	}
	if c.f != nil {
		_ = c.f.Close()
	}
}

func openCheckpoint(path string, resume bool) (string, *checkpoint, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", nil, err
		}
		dir := filepath.Join(home, ".clank")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", nil, err
		}
		path = filepath.Join(dir, fmt.Sprintf("scan-%s.jsonl", time.Now().Format("2006-01-02T15-04-05")))
	}
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if !resume {
		flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	f, err := os.OpenFile(path, flags, 0o600)
	if err != nil {
		return path, nil, fmt.Errorf("open checkpoint %s: %w", path, err)
	}
	return path, &checkpoint{f: f, w: bufio.NewWriter(f)}, nil
}

func readCheckpoint(path string) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	for sc.Scan() {
		var r Result
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			continue
		}
		if r.Phone != "" {
			out[r.Phone] = struct{}{}
		}
	}
	return out, sc.Err()
}

// ReadInput reads phones from a path (one per line) or from stdin if path == "-".
func ReadInput(path string) ([]string, error) {
	var r io.Reader
	if path == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open input %s: %w", path, err)
		}
		defer f.Close()
		r = f
	}
	out := []string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out, sc.Err()
}
