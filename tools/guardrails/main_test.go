package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCoverageStepPassesAtThreshold(t *testing.T) {
	t.Parallel()

	step := coverageStep(coverageResult{
		GlobalPercent:   70,
		GlobalThreshold: 70,
		GlobalStatus:    statusPass,
		ChangedPercent:  80,
		ChangedThreshold: 80,
		ChangedStatus:   statusPass,
	})

	if step.Status != statusPass {
		t.Fatalf("expected PASS at the threshold, got %s", step.Status)
	}
	if step.Summary != "global 70.0% >= 70.0%, changed 80.0% >= 80.0%" {
		t.Fatalf("unexpected summary: %q", step.Summary)
	}
}

func TestChangedGoLinesUsesCachedDiffForLocalMode(t *testing.T) {
	repo := t.TempDir()

	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "codex@example.com")
	runGit(t, repo, "config", "user.name", "Codex")

	base := "package main\n\nfunc one() {}\n"
	if err := os.WriteFile(filepath.Join(repo, "sample.go"), []byte(base), 0o644); err != nil {
		t.Fatalf("write base file: %v", err)
	}
	runGit(t, repo, "add", "sample.go")
	runGit(t, repo, "commit", "-m", "base")

	staged := base + "// staged\n"
	if err := os.WriteFile(filepath.Join(repo, "sample.go"), []byte(staged), 0o644); err != nil {
		t.Fatalf("write staged file: %v", err)
	}
	runGit(t, repo, "add", "sample.go")

	withUnstaged := staged + "// unstaged\n"
	if err := os.WriteFile(filepath.Join(repo, "sample.go"), []byte(withUnstaged), 0o644); err != nil {
		t.Fatalf("write unstaged file: %v", err)
	}

	workingTree, files, err := changedGoLines(repo, "HEAD", false)
	if err != nil {
		t.Fatalf("changedGoLines working tree: %v", err)
	}
	if len(files) != 1 || files[0] != "sample.go" {
		t.Fatalf("unexpected changed files for working tree diff: %v", files)
	}
	if !workingTree["sample.go"][4] || !workingTree["sample.go"][5] {
		t.Fatalf("working tree diff should include both appended lines, got %#v", workingTree["sample.go"])
	}

	stagedDiff, stagedFiles, err := changedGoLines(repo, "HEAD", true)
	if err != nil {
		t.Fatalf("changedGoLines cached diff: %v", err)
	}
	if len(stagedFiles) != 1 || stagedFiles[0] != "sample.go" {
		t.Fatalf("unexpected changed files for cached diff: %v", stagedFiles)
	}
	if !stagedDiff["sample.go"][4] {
		t.Fatalf("cached diff should include the staged line, got %#v", stagedDiff["sample.go"])
	}
	if stagedDiff["sample.go"][5] {
		t.Fatalf("cached diff should ignore the unstaged line, got %#v", stagedDiff["sample.go"])
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}
