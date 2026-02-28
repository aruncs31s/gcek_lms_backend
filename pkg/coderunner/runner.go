// Package coderunner provides sandboxed code execution for Python and JavaScript.
//
// How it works:
//  1. User code is written to a temporary file.
//  2. The appropriate interpreter (python3 / node) is invoked via os/exec.
//  3. stdin is fed from the provided Input string.
//  4. stdout / stderr are captured and returned.
//  5. A hard timeout (default 10 s) kills the process if it hangs.
//  6. The temp file is always cleaned up regardless of outcome.
//
// Supported languages: "python", "javascript"
package coderunner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const defaultTimeout = 10 * time.Second

// Result holds the output of a single code execution.
type Result struct {
	Stdout          string
	Stderr          string
	Error           string
	ExecutionTimeMs int64
}

// RunCode executes code in the specified language with the given stdin input.
// language must be "python" or "javascript".
func RunCode(language, code, input string) (*Result, error) {
	return RunCodeWithTimeout(language, code, input, defaultTimeout)
}

// RunCodeWithTimeout is like RunCode but with a configurable timeout.
func RunCodeWithTimeout(language, code, input string, timeout time.Duration) (*Result, error) {
	ext, interpreter, err := resolveLanguage(language)
	if err != nil {
		return nil, err
	}

	// Write code to a temp file
	tmpFile, err := os.CreateTemp("", "lms_code_*"+ext)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.WriteString(code); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write code: %w", err)
	}
	tmpFile.Close()

	// Run with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, interpreter, tmpFile.Name())
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	res := &Result{
		Stdout:          strings.TrimRight(stdout.String(), "\n\r"),
		Stderr:          strings.TrimRight(stderr.String(), "\n\r"),
		ExecutionTimeMs: elapsed,
	}

	if ctx.Err() == context.DeadlineExceeded {
		res.Error = fmt.Sprintf("execution timed out after %s", timeout)
		return res, nil
	}

	if runErr != nil {
		// Non-zero exit code is not a framework error – propagate via res.Error
		res.Error = runErr.Error()
	}

	return res, nil
}

// TestCase defines a single stdin/stdout test case.
type TestCase struct {
	ID             string
	Description    string
	Input          string
	ExpectedOutput string
	IsHidden       bool
}

// TestResult is the result for a single test case run.
type TestResult struct {
	TestCaseID      string
	Description     string
	Input           string
	Expected        string
	Actual          string
	Passed          bool
	Error           string
	ExecutionTimeMs int64
}

// RunTests executes the code against all provided test cases and returns individual results.
// Hidden test-case details (input / expected) are preserved here; callers decide what to redact.
func RunTests(language, code string, cases []TestCase) []TestResult {
	results := make([]TestResult, 0, len(cases))
	for _, tc := range cases {
		res, err := RunCode(language, code, tc.Input)

		tr := TestResult{
			TestCaseID:  tc.ID,
			Description: tc.Description,
			Input:       tc.Input,
			Expected:    tc.ExpectedOutput,
		}

		if err != nil {
			tr.Error = err.Error()
			tr.Passed = false
			results = append(results, tr)
			continue
		}

		tr.ExecutionTimeMs = res.ExecutionTimeMs
		tr.Error = res.Error

		if res.Error != "" {
			// Execution error (timeout, runtime error)
			tr.Actual = res.Stderr
			tr.Passed = false
		} else {
			tr.Actual = res.Stdout
			tr.Passed = strings.TrimSpace(res.Stdout) == strings.TrimSpace(tc.ExpectedOutput)
		}

		results = append(results, tr)
	}
	return results
}

// resolveLanguage returns the file extension and interpreter binary for a given language name.
func resolveLanguage(lang string) (ext, interpreter string, err error) {
	switch strings.ToLower(lang) {
	case "python":
		return ".py", "python3", nil
	case "javascript":
		return ".js", "node", nil
	default:
		return "", "", fmt.Errorf("unsupported language: %q (supported: python, javascript)", lang)
	}
}
