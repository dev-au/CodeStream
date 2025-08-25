package resources

import (
	"CodeStream/src"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RunnerConfig struct {
	Image     string
	Cmd       string
	Memory    string
	CPUs      string
	ExtraArgs []string
}

var runners = map[string]RunnerConfig{
	"python": {
		Image:  "runner-python:latest",
		Cmd:    "python3 %s",
		Memory: "50m",
		CPUs:   "0.5",
	},
	"javascript": {
		Image:  "runner-node:latest",
		Cmd:    "node %s",
		Memory: "50m",
		CPUs:   "0.5",
	},
	"go": {
		Image:  "runner-go:latest",
		Cmd:    "go run %s",
		Memory: "50m",
		CPUs:   "1",
		ExtraArgs: []string{
			"--tmpfs", "/tmp:rw,exec,nosuid,nodev,size=50m",
			"-v", "/var/go-cache:/root/.cache:rw",
			"-v", "/var/go-cache:/go-cache:rw",
			"-v", "go-build-cache:/root/.cache/go-build",
		},
	},
	"cpp": {
		Image:  "runner-cpp:latest",
		Cmd:    "bash -lc 'g++ %s -O2 -std=c++17 -o /tmp/a && /tmp/a'",
		Memory: "50m",
		CPUs:   "1",
		ExtraArgs: []string{
			"--tmpfs", "/tmp:rw,exec,nosuid,nodev,size=50m",
		},
	},
}

type RunRequest struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type RunResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration"`
}

type LimitedWriter struct {
	Limit int
	Buf   strings.Builder
	Hit   bool
}

func (lw *LimitedWriter) Write(p []byte) (int, error) {
	if lw.Buf.Len()+len(p) > lw.Limit {
		lw.Hit = true
		remaining := lw.Limit - lw.Buf.Len()
		if remaining > 0 {
			lw.Buf.Write(p[:remaining])
		}
		return len(p), errors.New("output limit exceeded")
	}
	return lw.Buf.Write(p)
}

func RunUserCode(ctx context.Context, baseWorkdir string, req RunRequest) (*RunResponse, error) {
	lang, ok := runners[req.Language]
	if !ok {
		return &RunResponse{Error: "unsupported language"}, errors.New("unsupported language")
	}

	jobID := randString(12)
	jobDir := filepath.Join(baseWorkdir, jobID)
	if err := os.MkdirAll(jobDir, 0o700); err != nil {
		return &RunResponse{Error: "failed to create job dir"}, err
	}
	defer os.RemoveAll(jobDir)

	fname := filenameForLang(req.Language)
	hostPath := filepath.Join(jobDir, fname)
	if err := os.WriteFile(hostPath, []byte(req.Code), 0o600); err != nil {
		return &RunResponse{Error: "failed to write code file"}, err
	}

	containerPath := "/app/" + fname
	containerName := "job-" + jobID

	dockerArgs := []string{
		"run", "--rm", "--name", containerName,
		"--network=none",
		"--pids-limit=64",
		"--memory=" + lang.Memory,
		"--cpus=" + lang.CPUs,
		"--read-only",
		"--security-opt", "no-new-privileges",
		"-v", fmt.Sprintf("%s:/app/%s:ro", hostPath, fname),
		"-w", "/app",
	}

	dockerArgs = append(dockerArgs, lang.ExtraArgs...)
	dockerArgs = append(dockerArgs, lang.Image, "sh", "-c", fmt.Sprintf(lang.Cmd, containerPath))
	ctxTimeout, cancel := context.WithTimeout(ctx, time.Duration(src.Config.RunTimeoutSecond)*time.Second)
	defer cancel()

	cmd := exec.Command("docker", dockerArgs...)
	cmd.Stdin = nil

	const outputLimit = 8 * 1024
	stdoutLimit := &LimitedWriter{Limit: outputLimit}
	stderrLimit := &LimitedWriter{Limit: outputLimit}
	cmd.Stdout = stdoutLimit
	cmd.Stderr = stderrLimit

	errCh := make(chan error, 1)
	start := time.Now()

	go func() {
		errCh <- cmd.Run()
	}()

	select {
	case <-ctxTimeout.Done():
		_ = exec.Command("docker", "kill", containerName).Run()
		return &RunResponse{
			Error:    "Time Limit Error",
			ExitCode: -1,
			Duration: time.Since(start).String(),
		}, nil

	case err := <-errCh:
		_ = exec.Command("docker", "kill", containerName).Run()
		duration := time.Since(start)
		res := &RunResponse{
			Stdout:   stdoutLimit.Buf.String(),
			Stderr:   stderrLimit.Buf.String(),
			Duration: duration.String(),
			ExitCode: 0,
		}

		if stdoutLimit.Hit || stderrLimit.Hit {
			res.Error = "Output Limit Error"
			res.ExitCode = -1
			res.Stdout = ""
			return res, nil
		}

		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				res.ExitCode = exitErr.ExitCode()
			}
		}
		if res.ExitCode == 137 {
			res.Error = "Memory Limit Error"
		}
		return res, nil
	}
}

func filenameForLang(lang string) string {
	switch lang {
	case "python":
		return "main.py"
	case "javascript":
		return "main.js"
	case "go":
		return "main.go"
	case "cpp":
		return "main.cpp"
	default:
		return "code.txt"
	}
}

func randString(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
