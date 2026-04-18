package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	statusPass = "PASS"
	statusFail = "FAIL"
	statusSkip = "SKIP"
)

type config struct {
	Coverage coverageConfig
	Security securityConfig
	Diff     diffConfig
}

type coverageConfig struct {
	GlobalMin  float64
	ChangedMin float64
	Exclusions []string
}

type securityConfig struct {
	FailSeverities []string
}

type diffConfig struct {
	BaseRef string
}

type stepResult struct {
	Name     string
	Status   string
	Summary  string
	Artifact string
	Duration time.Duration
	Output   string
}

type coverageResult struct {
	GlobalPercent       float64
	GlobalCovered       int
	GlobalStatements    int
	GlobalThreshold     float64
	GlobalStatus        string
	ChangedPercent      float64
	ChangedCoveredLines int
	ChangedLines        int
	ChangedThreshold    float64
	ChangedStatus       string
	ChangedNote         string
}

type finding struct {
	Tool     string
	Rule     string
	Severity string
	File     string
	Line     int
	Message  string
}

type reportData struct {
	GeneratedAt        string
	Mode               string
	OverallStatus      string
	BaseRef            string
	MergeBase          string
	Steps              []stepResult
	Coverage           coverageResult
	ChangedFiles       []string
	NewLintFindings    []finding
	NewSASTFindings    []finding
	ArtifactsDir       string
	ReportPath         string
	Warnings           []string
	GroupedLint        []findingGroup
	GroupedSAST        []findingGroup
	RawArtifactEntries []artifactEntry
}

type findingGroup struct {
	File     string
	Findings []finding
}

type artifactEntry struct {
	Label string
	Path  string
}

type runOptions struct {
	Mode          string
	ConfigPath    string
	ArtifactsDir  string
	ReportPath    string
	GenerateHTML  bool
	BaseRef       string
	GolangCILint  string
	Govulncheck   string
	Gosec         string
	RepoRoot      string
	UnitCoverPath string
}

type commandResult struct {
	ExitCode int
	Output   string
	Err      error
}

type coverBlock struct {
	File      string
	StartLine int
	EndLine   int
	Stmts     int
	Count     int
}

func main() {
	opts := runOptions{}
	flag.StringVar(&opts.Mode, "mode", "ci", "execution mode: ci or local")
	flag.StringVar(&opts.ConfigPath, "config", "guardrails.yaml", "guardrail config path")
	flag.StringVar(&opts.ArtifactsDir, "artifacts-dir", "tmp/guardrails", "artifact output directory")
	flag.StringVar(&opts.ReportPath, "report", "tmp/guardrails/report.html", "local HTML report path")
	flag.BoolVar(&opts.GenerateHTML, "html", false, "generate an HTML report")
	flag.StringVar(&opts.BaseRef, "base-ref", "", "git base ref for changed-code gates")
	flag.StringVar(&opts.GolangCILint, "golangci-lint", "golangci-lint", "golangci-lint binary")
	flag.StringVar(&opts.Govulncheck, "govulncheck", "govulncheck", "govulncheck binary")
	flag.StringVar(&opts.Gosec, "gosec", "gosec", "gosec binary")
	flag.Parse()

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(opts runOptions) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	opts.RepoRoot = root

	cfg, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return err
	}
	if opts.BaseRef == "" {
		opts.BaseRef = cfg.Diff.BaseRef
	}
	if opts.BaseRef == "" {
		opts.BaseRef = "origin/main"
	}

	if err := os.MkdirAll(opts.ArtifactsDir, 0o755); err != nil {
		return fmt.Errorf("create artifacts directory: %w", err)
	}
	opts.UnitCoverPath = filepath.Join(opts.ArtifactsDir, "unit-cover.out")

	report := reportData{
		GeneratedAt:   time.Now().Format(time.RFC3339),
		Mode:          opts.Mode,
		OverallStatus: statusPass,
		BaseRef:       opts.BaseRef,
		ArtifactsDir:  opts.ArtifactsDir,
		ReportPath:    opts.ReportPath,
	}

	modulePath, err := modulePath(opts.RepoRoot)
	if err != nil {
		return err
	}

	mergeBase, err := gitMergeBase(opts.RepoRoot, opts.BaseRef)
	if err != nil {
		return fmt.Errorf("resolve git merge-base against %s: %w", opts.BaseRef, err)
	}
	report.MergeBase = mergeBase

	changedLines, changedFiles, err := changedGoLines(opts.RepoRoot, mergeBase)
	if err != nil {
		return fmt.Errorf("collect changed lines: %w", err)
	}
	report.ChangedFiles = changedFiles

	report.Steps = append(report.Steps, runLint(opts, changedLines, &report)...)
	report.Steps = append(report.Steps, runUnitTests(opts))
	report.Steps = append(report.Steps, runIntegrationTests(opts))
	report.Steps = append(report.Steps, runGovulncheck(opts, mergeBase, changedLines))
	report.Steps = append(report.Steps, runGosec(opts, changedLines, cfg.Security.FailSeverities, &report)...)

	coverage, err := evaluateCoverage(opts.UnitCoverPath, modulePath, cfg, changedLines)
	if err != nil {
		report.Coverage = coverageResult{
			GlobalThreshold:  cfg.Coverage.GlobalMin,
			ChangedThreshold: cfg.Coverage.ChangedMin,
			GlobalStatus:     statusFail,
			ChangedStatus:    statusFail,
			ChangedNote:      err.Error(),
		}
		report.Steps = append(report.Steps, stepResult{
			Name:    "coverage",
			Status:  statusFail,
			Summary: err.Error(),
		})
	} else {
		report.Coverage = coverage
		report.Steps = append(report.Steps, coverageStep(coverage))
	}

	report.GroupedLint = groupFindings(report.NewLintFindings)
	report.GroupedSAST = groupFindings(report.NewSASTFindings)
	report.RawArtifactEntries = artifactEntries(opts.ArtifactsDir)
	report.OverallStatus = overallStatus(report)

	if opts.GenerateHTML {
		if err := writeHTMLReport(report); err != nil {
			return fmt.Errorf("write HTML report: %w", err)
		}
		fmt.Printf("guardrail report: %s\n", opts.ReportPath)
	}

	printSummary(report)
	if report.OverallStatus != statusPass {
		return errors.New("guardrails failed")
	}
	return nil
}

func runLint(opts runOptions, changed map[string]map[int]bool, report *reportData) []stepResult {
	jsonPath := filepath.Join(opts.ArtifactsDir, "golangci-lint.json")
	textPath := filepath.Join(opts.ArtifactsDir, "golangci-lint.txt")
	start := time.Now()
	res := runCommand(opts.RepoRoot, textPath, opts.GolangCILint,
		"run", "./...",
		"--issues-exit-code=0",
		"--output.json.path="+jsonPath,
		"--output.text.path=stdout",
		"--show-stats=false",
	)
	step := stepResult{
		Name:     "lint",
		Status:   statusPass,
		Artifact: jsonPath,
		Duration: time.Since(start),
		Output:   res.Output,
	}
	if res.Err != nil {
		step.Status = statusFail
		step.Summary = fmt.Sprintf("golangci-lint failed to run: %v", res.Err)
		return []stepResult{step}
	}
	findings, err := parseGolangCI(jsonPath, changed)
	if err != nil {
		step.Status = statusFail
		step.Summary = err.Error()
		return []stepResult{step}
	}
	report.NewLintFindings = findings
	if len(findings) > 0 {
		step.Status = statusFail
		step.Summary = fmt.Sprintf("%d changed-line lint finding(s)", len(findings))
	} else {
		step.Summary = "no changed-line lint findings"
	}
	return []stepResult{step}
}

func runUnitTests(opts runOptions) stepResult {
	textPath := filepath.Join(opts.ArtifactsDir, "unit-test.txt")
	start := time.Now()
	res := runCommand(opts.RepoRoot, textPath, "go",
		"test", "./internal/...",
		"-covermode=atomic",
		"-coverprofile="+opts.UnitCoverPath,
	)
	step := stepResult{
		Name:     "unit tests",
		Status:   statusPass,
		Artifact: opts.UnitCoverPath,
		Duration: time.Since(start),
		Output:   res.Output,
		Summary:  "unit tests passed",
	}
	if res.Err != nil {
		step.Status = statusFail
		step.Summary = fmt.Sprintf("unit tests failed: %v", res.Err)
	}
	return step
}

func runIntegrationTests(opts runOptions) stepResult {
	textPath := filepath.Join(opts.ArtifactsDir, "integration-test.txt")
	start := time.Now()
	res := runCommand(opts.RepoRoot, textPath, "go", "test", "-count=1", "./tests/integration/...")
	step := stepResult{
		Name:     "integration tests",
		Status:   statusPass,
		Artifact: textPath,
		Duration: time.Since(start),
		Output:   res.Output,
		Summary:  "integration tests passed",
	}
	if res.Err != nil {
		step.Status = statusFail
		step.Summary = fmt.Sprintf("integration tests failed: %v", res.Err)
	}
	return step
}

func runGovulncheck(opts runOptions, mergeBase string, changed map[string]map[int]bool) stepResult {
	textPath := filepath.Join(opts.ArtifactsDir, "govulncheck.txt")
	start := time.Now()
	res := runCommand(opts.RepoRoot, textPath, opts.Govulncheck, "./...")
	step := stepResult{
		Name:     "vulnerability check",
		Status:   statusPass,
		Artifact: textPath,
		Duration: time.Since(start),
		Output:   res.Output,
		Summary:  "govulncheck found no reachable vulnerabilities",
	}
	if res.Err != nil {
		changedModules := changedDependencyModules(opts.RepoRoot, mergeBase)
		changedTrace := govulncheckTouchesChangedLines(res.Output, changed)
		changedDependency := govulncheckTouchesChangedDependencies(res.Output, changedModules)
		if changedDependency || changedTrace {
			step.Status = statusFail
			step.Summary = "govulncheck found vulnerabilities connected to changed code or dependency files"
			return step
		}
		if !strings.Contains(res.Output, "Vulnerability #") {
			step.Status = statusFail
			step.Summary = fmt.Sprintf("govulncheck failed to run: %v", res.Err)
			return step
		}
		step.Summary = "govulncheck found existing vulnerabilities outside changed code"
	}
	return step
}

func runGosec(opts runOptions, changed map[string]map[int]bool, failSeverities []string, report *reportData) []stepResult {
	jsonPath := filepath.Join(opts.ArtifactsDir, "gosec.json")
	textPath := filepath.Join(opts.ArtifactsDir, "gosec.txt")
	start := time.Now()
	res := runCommand(opts.RepoRoot, textPath, opts.Gosec, "-fmt=json", "-out="+jsonPath, "-no-fail", "./...")
	step := stepResult{
		Name:     "sast",
		Status:   statusPass,
		Artifact: jsonPath,
		Duration: time.Since(start),
		Output:   res.Output,
	}
	if res.Err != nil {
		step.Status = statusFail
		step.Summary = fmt.Sprintf("gosec failed to run: %v", res.Err)
		return []stepResult{step}
	}
	findings, err := parseGosec(jsonPath, changed, failSeverities)
	if err != nil {
		step.Status = statusFail
		step.Summary = err.Error()
		return []stepResult{step}
	}
	report.NewSASTFindings = findings
	if len(findings) > 0 {
		step.Status = statusFail
		step.Summary = fmt.Sprintf("%d changed-line high-or-worse SAST finding(s)", len(findings))
	} else {
		step.Summary = "no changed-line high-or-worse SAST findings"
	}
	return []stepResult{step}
}

func runCommand(repoRoot, artifactPath, name string, args ...string) commandResult {
	cmd := exec.Command(name, args...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(),
		"GOCACHE="+filepath.Join(repoRoot, "tmp", "go-build-cache"),
		"GOLANGCI_LINT_CACHE="+filepath.Join(repoRoot, "tmp", "golangci-lint-cache"),
	)

	var buf bytes.Buffer
	writers := []io.Writer{&buf, os.Stdout}
	if artifactPath != "" {
		if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err == nil {
			if f, err := os.Create(artifactPath); err == nil {
				defer func() { _ = f.Close() }()
				writers = append(writers, f)
			}
		}
	}
	cmd.Stdout = io.MultiWriter(writers...)
	cmd.Stderr = io.MultiWriter(writers...)
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	return commandResult{ExitCode: exitCode, Output: buf.String(), Err: err}
}

func loadConfig(path string) (config, error) {
	cfg := config{
		Coverage: coverageConfig{
			GlobalMin:  70,
			ChangedMin: 80,
			Exclusions: []string{
				"cmd/**",
				"tools/**",
				"**/*_mock.go",
				"mocks/**",
				"**/mocks/**",
				"internal/**/domain/**",
				"internal/**/port/**",
				"internal/config/**",
				"internal/database/**",
				"internal/httpserver/**",
				"internal/logger/**",
				"internal/shared/health/**",
				"internal/agent/adapters/**",
				"internal/health/adapters/**",
				"internal/petcare/adapters/secondary/**",
				"internal/petcare/adapters/primary/http/*.go",
				"internal/gcalendar/**",
				"internal/telegram/**",
			},
		},
		Security: securityConfig{FailSeverities: []string{"HIGH", "CRITICAL"}},
		Diff:     diffConfig{BaseRef: "origin/main"},
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	section := ""
	listKey := ""
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(raw, " ") && strings.HasSuffix(line, ":") {
			section = strings.TrimSuffix(line, ":")
			listKey = ""
			continue
		}
		if strings.HasPrefix(line, "- ") {
			value := trimYAMLValue(strings.TrimPrefix(line, "- "))
			switch section + "." + listKey {
			case "coverage.exclusions":
				cfg.Coverage.Exclusions = appendUnique(cfg.Coverage.Exclusions, value)
			case "security.fail_severities":
				cfg.Security.FailSeverities = appendUnique(cfg.Security.FailSeverities, strings.ToUpper(value))
			}
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = trimYAMLValue(value)
		if value == "" {
			listKey = key
			continue
		}
		switch section + "." + key {
		case "coverage.global_min":
			cfg.Coverage.GlobalMin = parseFloat(value, cfg.Coverage.GlobalMin)
		case "coverage.changed_min":
			cfg.Coverage.ChangedMin = parseFloat(value, cfg.Coverage.ChangedMin)
		case "diff.base_ref":
			cfg.Diff.BaseRef = value
		}
	}
	return cfg, nil
}

func trimYAMLValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func parseFloat(value string, fallback float64) float64 {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func modulePath(repoRoot string) (string, error) {
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve module path: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func gitMergeBase(repoRoot, baseRef string) (string, error) {
	cmd := exec.Command("git", "merge-base", baseRef, "HEAD")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func changedGoLines(repoRoot, base string) (map[string]map[int]bool, []string, error) {
	cmd := exec.Command("git", "diff", "--unified=0", "--no-ext-diff", base, "--", "*.go")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, err
	}
	changed := map[string]map[int]bool{}
	currentFile := ""
	hunkRE := regexp.MustCompile(`@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "+++ b/") {
			currentFile = strings.TrimPrefix(line, "+++ b/")
			continue
		}
		if strings.HasPrefix(line, "+++ /dev/null") {
			currentFile = ""
			continue
		}
		if currentFile == "" || !strings.HasPrefix(line, "@@ ") {
			continue
		}
		matches := hunkRE.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}
		start, _ := strconv.Atoi(matches[1])
		count := 1
		if matches[2] != "" {
			count, _ = strconv.Atoi(matches[2])
		}
		if count == 0 {
			continue
		}
		if changed[currentFile] == nil {
			changed[currentFile] = map[int]bool{}
		}
		for i := 0; i < count; i++ {
			changed[currentFile][start+i] = true
		}
	}
	files := make([]string, 0, len(changed))
	for file := range changed {
		files = append(files, file)
	}
	sort.Strings(files)
	return changed, files, nil
}

func changedDependencyModules(repoRoot, base string) map[string]bool {
	modules := map[string]bool{}
	cmd := exec.Command("git", "diff", "--unified=0", "--no-ext-diff", base, "--", "go.mod", "go.sum")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		modules["*"] = true
		return modules
	}
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, "+"))
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "require" && len(fields) > 1 {
			modules[fields[1]] = true
			continue
		}
		if strings.Contains(fields[0], ".") && !strings.HasPrefix(fields[0], "go") {
			modules[fields[0]] = true
		}
	}
	return modules
}

func govulncheckTouchesChangedLines(output string, changed map[string]map[int]bool) bool {
	traceRE := regexp.MustCompile(`([A-Za-z0-9_./-]+\.go):(\d+):\d+`)
	for _, match := range traceRE.FindAllStringSubmatch(output, -1) {
		line, err := strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		if isChangedLine(changed, normalizeRepoPath(match[1]), line) {
			return true
		}
	}
	return false
}

func govulncheckTouchesChangedDependencies(output string, changedModules map[string]bool) bool {
	if len(changedModules) == 0 {
		return false
	}
	if changedModules["*"] {
		return true
	}
	foundRE := regexp.MustCompile(`Found in: ([^\s@]+)@`)
	for _, match := range foundRE.FindAllStringSubmatch(output, -1) {
		if changedModules[match[1]] {
			return true
		}
	}
	return false
}

func evaluateCoverage(path, modulePath string, cfg config, changed map[string]map[int]bool) (coverageResult, error) {
	blocks, err := parseCoverageProfile(path, modulePath)
	result := coverageResult{
		GlobalThreshold:  cfg.Coverage.GlobalMin,
		ChangedThreshold: cfg.Coverage.ChangedMin,
	}
	if err != nil {
		return result, err
	}
	for _, block := range blocks {
		if excluded(block.File, cfg.Coverage.Exclusions) {
			continue
		}
		result.GlobalStatements += block.Stmts
		if block.Count > 0 {
			result.GlobalCovered += block.Stmts
		}
	}
	if result.GlobalStatements > 0 {
		result.GlobalPercent = percent(result.GlobalCovered, result.GlobalStatements)
	}
	if result.GlobalPercent > cfg.Coverage.GlobalMin {
		result.GlobalStatus = statusPass
	} else {
		result.GlobalStatus = statusFail
	}

	type lineCoverage struct {
		executable bool
		covered    bool
	}
	changedCoverage := map[string]map[int]*lineCoverage{}
	for file, lines := range changed {
		if excluded(file, cfg.Coverage.Exclusions) || strings.HasSuffix(file, "_test.go") {
			continue
		}
		changedCoverage[file] = map[int]*lineCoverage{}
		for line := range lines {
			changedCoverage[file][line] = &lineCoverage{}
		}
	}
	for _, block := range blocks {
		lines, ok := changedCoverage[block.File]
		if !ok {
			continue
		}
		for line, cov := range lines {
			if line >= block.StartLine && line <= block.EndLine {
				cov.executable = true
				if block.Count > 0 {
					cov.covered = true
				}
			}
		}
	}
	for _, lines := range changedCoverage {
		for _, cov := range lines {
			if !cov.executable {
				continue
			}
			result.ChangedLines++
			if cov.covered {
				result.ChangedCoveredLines++
			}
		}
	}
	if result.ChangedLines == 0 {
		result.ChangedPercent = 100
		result.ChangedStatus = statusPass
		result.ChangedNote = "no changed executable production lines"
	} else {
		result.ChangedPercent = percent(result.ChangedCoveredLines, result.ChangedLines)
		if result.ChangedPercent >= cfg.Coverage.ChangedMin {
			result.ChangedStatus = statusPass
		} else {
			result.ChangedStatus = statusFail
		}
	}
	return result, nil
}

func parseCoverageProfile(path, modulePath string) ([]coverBlock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read coverage profile: %w", err)
	}
	var blocks []coverBlock
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		location, rest, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		file, ranges, ok := strings.Cut(location, ":")
		if !ok {
			continue
		}
		start, end, ok := strings.Cut(ranges, ",")
		if !ok {
			continue
		}
		startLine := lineNumber(start)
		endLine := lineNumber(end)
		fields := strings.Fields(rest)
		if len(fields) != 2 {
			continue
		}
		stmts, _ := strconv.Atoi(fields[0])
		count, _ := strconv.Atoi(fields[1])
		file = strings.TrimPrefix(file, modulePath+"/")
		file = filepath.ToSlash(file)
		blocks = append(blocks, coverBlock{File: file, StartLine: startLine, EndLine: endLine, Stmts: stmts, Count: count})
	}
	return blocks, nil
}

func lineNumber(position string) int {
	line, _, _ := strings.Cut(position, ".")
	n, _ := strconv.Atoi(line)
	return n
}

func percent(covered, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round((float64(covered)/float64(total))*1000) / 10
}

func coverageStep(result coverageResult) stepResult {
	status := statusPass
	summary := fmt.Sprintf("global %.1f%% > %.1f%%, changed %.1f%% >= %.1f%%", result.GlobalPercent, result.GlobalThreshold, result.ChangedPercent, result.ChangedThreshold)
	if result.GlobalStatus != statusPass || result.ChangedStatus != statusPass {
		status = statusFail
	}
	return stepResult{Name: "coverage gates", Status: status, Summary: summary}
}

func excluded(path string, patterns []string) bool {
	path = filepath.ToSlash(path)
	for _, pattern := range patterns {
		if globMatch(filepath.ToSlash(pattern), path) {
			return true
		}
	}
	return false
}

func globMatch(pattern, path string) bool {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				b.WriteString(".*")
				i++
			} else {
				b.WriteString(`[^/]*`)
			}
		case '?':
			b.WriteByte('.')
		default:
			b.WriteString(regexp.QuoteMeta(string(pattern[i])))
		}
	}
	b.WriteString("$")
	ok, _ := regexp.MatchString(b.String(), path)
	return ok
}

func parseGolangCI(path string, changed map[string]map[int]bool) ([]finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read golangci-lint JSON: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	var payload struct {
		Issues []struct {
			FromLinter string `json:"FromLinter"`
			Text       string `json:"Text"`
			Severity   string `json:"Severity"`
			Pos        struct {
				Filename string `json:"Filename"`
				Line     int    `json:"Line"`
			} `json:"Pos"`
		} `json:"Issues"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse golangci-lint JSON: %w", err)
	}
	findings := make([]finding, 0, len(payload.Issues))
	for _, issue := range payload.Issues {
		file := normalizeRepoPath(issue.Pos.Filename)
		if !isChangedLine(changed, file, issue.Pos.Line) {
			continue
		}
		findings = append(findings, finding{
			Tool:     "golangci-lint",
			Rule:     issue.FromLinter,
			Severity: issue.Severity,
			File:     file,
			Line:     issue.Pos.Line,
			Message:  issue.Text,
		})
	}
	sortFindings(findings)
	return findings, nil
}

func parseGosec(path string, changed map[string]map[int]bool, failSeverities []string) ([]finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read gosec JSON: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	var payload struct {
		Issues []struct {
			Severity string `json:"severity"`
			RuleID   string `json:"rule_id"`
			Details  string `json:"details"`
			File     string `json:"file"`
			Line     string `json:"line"`
		} `json:"Issues"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse gosec JSON: %w", err)
	}
	var findings []finding
	failSeverity := map[string]bool{}
	for _, severity := range failSeverities {
		failSeverity[strings.ToUpper(severity)] = true
	}
	for _, issue := range payload.Issues {
		severity := strings.ToUpper(issue.Severity)
		if !failSeverity[severity] {
			continue
		}
		line, _ := strconv.Atoi(issue.Line)
		file := normalizeRepoPath(issue.File)
		if !isChangedLine(changed, file, line) {
			continue
		}
		findings = append(findings, finding{
			Tool:     "gosec",
			Rule:     issue.RuleID,
			Severity: severity,
			File:     file,
			Line:     line,
			Message:  issue.Details,
		})
	}
	sortFindings(findings)
	return findings, nil
}

func normalizeRepoPath(path string) string {
	path = filepath.ToSlash(path)
	if wd, err := os.Getwd(); err == nil {
		wd = filepath.ToSlash(wd)
		path = strings.TrimPrefix(path, wd+"/")
	}
	return strings.TrimPrefix(path, "./")
}

func isChangedLine(changed map[string]map[int]bool, file string, line int) bool {
	if line <= 0 {
		return false
	}
	lines, ok := changed[file]
	return ok && lines[line]
}

func sortFindings(findings []finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].File == findings[j].File {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].File < findings[j].File
	})
}

func groupFindings(findings []finding) []findingGroup {
	byFile := map[string][]finding{}
	for _, f := range findings {
		byFile[f.File] = append(byFile[f.File], f)
	}
	files := make([]string, 0, len(byFile))
	for file := range byFile {
		files = append(files, file)
	}
	sort.Strings(files)
	groups := make([]findingGroup, 0, len(files))
	for _, file := range files {
		groups = append(groups, findingGroup{File: file, Findings: byFile[file]})
	}
	return groups
}

func artifactEntries(dir string) []artifactEntry {
	entries := []artifactEntry{
		{Label: "Unit coverage profile", Path: filepath.Join(dir, "unit-cover.out")},
		{Label: "Unit test output", Path: filepath.Join(dir, "unit-test.txt")},
		{Label: "Integration test output", Path: filepath.Join(dir, "integration-test.txt")},
		{Label: "golangci-lint JSON", Path: filepath.Join(dir, "golangci-lint.json")},
		{Label: "golangci-lint text", Path: filepath.Join(dir, "golangci-lint.txt")},
		{Label: "govulncheck output", Path: filepath.Join(dir, "govulncheck.txt")},
		{Label: "gosec JSON", Path: filepath.Join(dir, "gosec.json")},
		{Label: "gosec text", Path: filepath.Join(dir, "gosec.txt")},
	}
	return entries
}

func overallStatus(report reportData) string {
	if report.Coverage.GlobalStatus == statusFail || report.Coverage.ChangedStatus == statusFail {
		return statusFail
	}
	for _, step := range report.Steps {
		if step.Status == statusFail {
			return statusFail
		}
	}
	return statusPass
}

func printSummary(report reportData) {
	fmt.Printf("\nGuardrails: %s\n", report.OverallStatus)
	for _, step := range report.Steps {
		fmt.Printf("- %s: %s (%s)\n", step.Name, step.Status, step.Summary)
	}
}

func writeHTMLReport(report reportData) error {
	if err := os.MkdirAll(filepath.Dir(report.ReportPath), 0o755); err != nil {
		return err
	}
	tpl, err := template.New("report").Funcs(template.FuncMap{
		"statusClass": statusClass,
		"duration":    formatDuration,
		"printf":      fmt.Sprintf,
	}).Parse(reportTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, report); err != nil {
		return err
	}
	return os.WriteFile(report.ReportPath, buf.Bytes(), 0o644)
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(100 * time.Millisecond).String()
}

func statusClass(status string) string {
	switch status {
	case statusPass:
		return "pass"
	case statusFail:
		return "fail"
	default:
		return "skip"
	}
}

const reportTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Alfredo Guardrails</title>
  <style>
    :root {
      color-scheme: light;
      --ink: #17211f;
      --muted: #5a6662;
      --line: #d8dfdc;
      --soft: #f5f7f4;
      --pass: #167245;
      --pass-bg: #e7f5ee;
      --fail: #b42318;
      --fail-bg: #fff0ed;
      --warn: #875a00;
      --panel: #ffffff;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }
    body {
      margin: 0;
      background: #eef2ef;
      color: var(--ink);
    }
    main {
      max-width: 1180px;
      margin: 0 auto;
      padding: 28px;
    }
    header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 24px;
      border-bottom: 1px solid var(--line);
      padding-bottom: 22px;
      margin-bottom: 24px;
    }
    h1, h2, h3, p {
      margin-top: 0;
    }
    h1 {
      font-size: 32px;
      line-height: 1.1;
      margin-bottom: 8px;
    }
    h2 {
      font-size: 19px;
      margin-bottom: 14px;
    }
    h3 {
      font-size: 15px;
      margin-bottom: 8px;
    }
    .meta, .muted {
      color: var(--muted);
    }
    .status {
      display: inline-flex;
      align-items: center;
      border-radius: 8px;
      padding: 8px 12px;
      font-weight: 800;
      letter-spacing: 0;
      border: 1px solid var(--line);
      white-space: nowrap;
    }
    .status.pass { color: var(--pass); background: var(--pass-bg); border-color: #a7d8bf; }
    .status.fail { color: var(--fail); background: var(--fail-bg); border-color: #f0b7ae; }
    .status.skip { color: var(--warn); background: #fff7e5; border-color: #f2d08a; }
    .grid {
      display: grid;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 14px;
      margin-bottom: 24px;
    }
    .panel {
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 16px;
      box-shadow: 0 8px 24px rgba(23, 33, 31, 0.06);
    }
    .metric {
      font-size: 30px;
      font-weight: 850;
      margin-bottom: 6px;
    }
    .steps {
      display: grid;
      gap: 10px;
    }
    .step {
      display: grid;
      grid-template-columns: 170px 88px minmax(0, 1fr) 92px;
      gap: 12px;
      align-items: center;
      border-top: 1px solid var(--line);
      padding-top: 10px;
    }
    code {
      background: var(--soft);
      border: 1px solid var(--line);
      border-radius: 6px;
      padding: 2px 5px;
      overflow-wrap: anywhere;
    }
    ul {
      padding-left: 20px;
    }
    li {
      margin-bottom: 6px;
    }
    .two-col {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 14px;
      margin-top: 14px;
    }
    .finding {
      border-top: 1px solid var(--line);
      padding: 10px 0;
    }
    .finding:last-child {
      padding-bottom: 0;
    }
    .rule {
      font-weight: 800;
    }
    .artifacts {
      columns: 2;
    }
    @media (max-width: 900px) {
      main { padding: 18px; }
      header, .two-col { display: block; }
      .grid { grid-template-columns: 1fr 1fr; }
      .step { grid-template-columns: 1fr; }
      .artifacts { columns: 1; }
    }
  </style>
</head>
<body>
<main>
  <header>
    <div>
      <h1>Alfredo Guardrails</h1>
      <p class="meta">Generated {{ .GeneratedAt }} · mode {{ .Mode }} · base <code>{{ .BaseRef }}</code> · merge base <code>{{ .MergeBase }}</code></p>
    </div>
    <div class="status {{ statusClass .OverallStatus }}">{{ .OverallStatus }}</div>
  </header>

  <section class="grid">
    <div class="panel">
      <h2>Global Coverage</h2>
      <div class="metric">{{ printf "%.1f" .Coverage.GlobalPercent }}%</div>
      <p class="muted">{{ .Coverage.GlobalCovered }}/{{ .Coverage.GlobalStatements }} statements · must be &gt; {{ printf "%.1f" .Coverage.GlobalThreshold }}%</p>
      <div class="status {{ statusClass .Coverage.GlobalStatus }}">{{ .Coverage.GlobalStatus }}</div>
    </div>
    <div class="panel">
      <h2>Changed Coverage</h2>
      <div class="metric">{{ printf "%.1f" .Coverage.ChangedPercent }}%</div>
      <p class="muted">{{ .Coverage.ChangedCoveredLines }}/{{ .Coverage.ChangedLines }} executable lines · must be ≥ {{ printf "%.1f" .Coverage.ChangedThreshold }}%</p>
      {{ if .Coverage.ChangedNote }}<p class="muted">{{ .Coverage.ChangedNote }}</p>{{ end }}
      <div class="status {{ statusClass .Coverage.ChangedStatus }}">{{ .Coverage.ChangedStatus }}</div>
    </div>
    <div class="panel">
      <h2>Lint Findings</h2>
      <div class="metric">{{ len .NewLintFindings }}</div>
      <p class="muted">changed-line code smells</p>
      <div class="status {{ if .NewLintFindings }}fail{{ else }}pass{{ end }}">{{ if .NewLintFindings }}FAIL{{ else }}PASS{{ end }}</div>
    </div>
    <div class="panel">
      <h2>SAST Findings</h2>
      <div class="metric">{{ len .NewSASTFindings }}</div>
      <p class="muted">changed-line HIGH or CRITICAL issues</p>
      <div class="status {{ if .NewSASTFindings }}fail{{ else }}pass{{ end }}">{{ if .NewSASTFindings }}FAIL{{ else }}PASS{{ end }}</div>
    </div>
  </section>

  <section class="panel">
    <h2>Gate Status</h2>
    <div class="steps">
      {{ range .Steps }}
      <div class="step">
        <strong>{{ .Name }}</strong>
        <span class="status {{ statusClass .Status }}">{{ .Status }}</span>
        <span>{{ .Summary }}</span>
        <span class="muted">{{ duration .Duration }}</span>
      </div>
      {{ end }}
    </div>
  </section>

  <section class="two-col">
    <div class="panel">
      <h2>Changed Files</h2>
      {{ if .ChangedFiles }}
      <ul>
        {{ range .ChangedFiles }}<li><code>{{ . }}</code></li>{{ end }}
      </ul>
      {{ else }}
      <p class="muted">No changed Go files detected.</p>
      {{ end }}
    </div>
    <div class="panel">
      <h2>Raw Artifacts</h2>
      <ul class="artifacts">
        {{ range .RawArtifactEntries }}<li>{{ .Label }}: <code>{{ .Path }}</code></li>{{ end }}
      </ul>
    </div>
  </section>

  <section class="two-col">
    <div class="panel">
      <h2>Lint Details</h2>
      {{ if .GroupedLint }}
        {{ range .GroupedLint }}
        <h3><code>{{ .File }}</code></h3>
        {{ range .Findings }}
        <div class="finding">
          <div><span class="rule">{{ .Rule }}</span> line {{ .Line }}</div>
          <div>{{ .Message }}</div>
        </div>
        {{ end }}
        {{ end }}
      {{ else }}
      <p class="muted">No changed-line lint findings.</p>
      {{ end }}
    </div>
    <div class="panel">
      <h2>SAST Details</h2>
      {{ if .GroupedSAST }}
        {{ range .GroupedSAST }}
        <h3><code>{{ .File }}</code></h3>
        {{ range .Findings }}
        <div class="finding">
          <div><span class="rule">{{ .Rule }}</span> line {{ .Line }} · {{ .Severity }}</div>
          <div>{{ .Message }}</div>
        </div>
        {{ end }}
        {{ end }}
      {{ else }}
      <p class="muted">No changed-line HIGH or CRITICAL SAST findings.</p>
      {{ end }}
    </div>
  </section>
</main>
</body>
</html>
`
