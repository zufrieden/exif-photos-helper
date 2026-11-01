// Package exif drives a long-running exiftool process in -stay_open mode.
package exif

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Apps launched from Finder don't inherit the shell PATH, so look in the
// usual package-manager locations first (MacPorts, Homebrew Intel, Homebrew ARM).
var searchDirs = []string{"/opt/local/bin", "/usr/local/bin", "/opt/homebrew/bin"}

// Find returns the path to the exiftool binary.
func Find() (string, error) {
	for _, d := range searchDirs {
		p := filepath.Join(d, "exiftool")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	if p, err := exec.LookPath("exiftool"); err == nil {
		return p, nil
	}
	return "", errors.New("exiftool not found (install it with: sudo port install p5-image-exiftool)")
}

// Tool is a single exiftool process. Commands are serialized.
type Tool struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr *bufio.Reader
	seq    int
}

// Start launches exiftool in -stay_open mode.
func Start(bin string) (*Tool, error) {
	cmd := exec.Command(bin, "-stay_open", "True", "-@", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &Tool{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReaderSize(stdout, 1<<20),
		stderr: bufio.NewReader(stderr),
	}, nil
}

// Run executes one exiftool command and returns its stdout and stderr.
// Arguments must not contain newlines (the -@ arg file is line based).
func (t *Tool) Run(args ...string) (stdout []byte, stderr string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.seq++
	marker := fmt.Sprintf("{ready%d}", t.seq)

	var b strings.Builder
	for _, a := range args {
		if strings.ContainsAny(a, "\r\n") {
			return nil, "", fmt.Errorf("argument contains a newline: %q", a)
		}
		b.WriteString(a)
		b.WriteByte('\n')
	}
	// -echo4 prints the marker to stderr once processing is done, so both
	// streams can be read to a known end.
	fmt.Fprintf(&b, "-echo4\n%s\n-execute%d\n", marker, t.seq)
	if _, err := io.WriteString(t.stdin, b.String()); err != nil {
		return nil, "", err
	}

	type result struct {
		data []byte
		err  error
	}
	errCh := make(chan result, 1)
	go func() {
		d, err := readUntil(t.stderr, []byte(marker+"\n"))
		errCh <- result{d, err}
	}()
	out, outErr := readUntil(t.stdout, []byte(marker+"\n"))
	errRes := <-errCh
	if outErr != nil {
		return nil, "", outErr
	}
	if errRes.err != nil {
		return nil, "", errRes.err
	}
	return out, strings.TrimSpace(string(errRes.data)), nil
}

// readUntil reads r until data ends with suffix, returning data without it.
func readUntil(r *bufio.Reader, suffix []byte) ([]byte, error) {
	var buf bytes.Buffer
	last := suffix[len(suffix)-1]
	for {
		chunk, err := r.ReadBytes(last)
		buf.Write(chunk)
		if bytes.HasSuffix(buf.Bytes(), suffix) {
			return buf.Bytes()[:buf.Len()-len(suffix)], nil
		}
		if err != nil {
			return nil, fmt.Errorf("exiftool: %w", err)
		}
	}
}

// Close stops the exiftool process.
func (t *Tool) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, _ = io.WriteString(t.stdin, "-stay_open\nFalse\n")
	_ = t.stdin.Close()
	return t.cmd.Wait()
}

// Pool is a fixed set of exiftool processes for parallel work.
type Pool struct {
	tools chan *Tool
	all   []*Tool
}

// NewPool starts n exiftool processes.
func NewPool(bin string, n int) (*Pool, error) {
	p := &Pool{tools: make(chan *Tool, n)}
	for i := 0; i < n; i++ {
		t, err := Start(bin)
		if err != nil {
			p.Close()
			return nil, err
		}
		p.all = append(p.all, t)
		p.tools <- t
	}
	return p, nil
}

// Run executes a command on the next free process.
func (p *Pool) Run(args ...string) ([]byte, string, error) {
	t := <-p.tools
	defer func() { p.tools <- t }()
	return t.Run(args...)
}

// Close stops all processes.
func (p *Pool) Close() {
	for _, t := range p.all {
		_ = t.Close()
	}
}
