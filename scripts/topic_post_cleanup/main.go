package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var defaultExcludeDirs = []string{
	".git",
	"node_modules",
	"dist",
	"tmp",
	"storage",
	"app/models/deprecatedmodels",
	"app/service/datamigration",
}

var defaultExtensions = []string{
	".go",
	".ts",
	".vue",
	".md",
	".gohtml",
	".css",
	".json",
	".mjs",
	".yaml",
	".yml",
}

var defaultKeywords = []string{
	"article",
	"articles",
	"reply",
	"replies",
}

type config struct {
	ExcludeDirs []string     `json:"exclude_dirs"`
	Scan        scanConfig   `json:"scan"`
	Rules       []ruleConfig `json:"rules"`
}

type scanConfig struct {
	Extensions []string `json:"extensions"`
	Keywords   []string `json:"keywords"`
}

type ruleConfig struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Paths        []string        `json:"paths"`
	Renames      []renameConfig  `json:"renames"`
	Replacements []replaceConfig `json:"replacements"`
}

type renameConfig struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type replaceConfig struct {
	Type    string `json:"type"`
	From    string `json:"from"`
	To      string `json:"to"`
	Pattern string `json:"pattern"`
}

type scanHit struct {
	File string
	Line int
	Text string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "scan":
		if err := runScan(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "list-rules":
		if err := runListRules(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "apply":
		if err := runApply(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		exitErr(fmt.Errorf("unknown command %q", os.Args[1]))
	}
}

func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	root := fs.String("root", ".", "project root")
	configPath := fs.String("config", "", "optional JSON config file")
	keywordsCSV := fs.String("keywords", "", "comma-separated keywords override")
	extensionsCSV := fs.String("extensions", "", "comma-separated file extensions override")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	keywords := pickList(splitCSV(*keywordsCSV), cfg.Scan.Keywords, defaultKeywords)
	extensions := pickList(splitCSV(*extensionsCSV), cfg.Scan.Extensions, defaultExtensions)
	hits, err := scanProject(*root, cfg.ExcludeDirs, extensions, keywords)
	if err != nil {
		return err
	}
	if len(hits) == 0 {
		fmt.Println("No matches found.")
		return nil
	}

	for _, hit := range hits {
		fmt.Printf("%s:%d: %s\n", hit.File, hit.Line, hit.Text)
	}
	fmt.Printf("\nFound %d matches across %d files.\n", len(hits), countFiles(hits))
	return nil
}

func runListRules(args []string) error {
	fs := flag.NewFlagSet("list-rules", flag.ContinueOnError)
	configPath := fs.String("config", "", "optional JSON config file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	if len(cfg.Rules) == 0 {
		fmt.Println("No rules configured.")
		return nil
	}
	for _, rule := range cfg.Rules {
		fmt.Printf("- %s\n", rule.Name)
		if rule.Description != "" {
			fmt.Printf("  %s\n", rule.Description)
		}
	}
	return nil
}

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	root := fs.String("root", ".", "project root")
	configPath := fs.String("config", "", "JSON config file")
	rulesCSV := fs.String("rules", "all", "comma-separated rule names or 'all'")
	dryRun := fs.Bool("dry-run", false, "print planned changes without writing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	if len(cfg.Rules) == 0 {
		return errors.New("no rules configured; add a config file or use list-rules to inspect available rules")
	}

	selected, err := selectRules(cfg.Rules, splitCSV(*rulesCSV))
	if err != nil {
		return err
	}

	totalFileChanges := 0
	totalRenames := 0
	for _, rule := range selected {
		fileChanges, renames, err := applyRule(*root, cfg.ExcludeDirs, rule, *dryRun)
		if err != nil {
			return fmt.Errorf("rule %q failed: %w", rule.Name, err)
		}
		totalFileChanges += fileChanges
		totalRenames += renames
	}

	mode := "Applied"
	if *dryRun {
		mode = "Planned"
	}
	fmt.Printf("%s %d content changes and %d renames.\n", mode, totalFileChanges, totalRenames)
	return nil
}

func printUsage() {
	fmt.Println(`topic_post_cleanup

Usage:
  go run ./scripts/topic_post_cleanup scan [--config file] [--root dir]
  go run ./scripts/topic_post_cleanup list-rules [--config file]
  go run ./scripts/topic_post_cleanup apply --config file [--rules name1,name2|all] [--dry-run]

Commands:
  scan       Search the repo for article/reply leftovers.
  list-rules Print configured cleanup rules.
  apply      Apply configured cleanup rules.

Notes:
  - The script skips deprecated models and migration directories by default.
  - Use --dry-run before apply to preview writes and renames.
  - Rules come from a JSON config file; see scripts/topic_post_cleanup/rules.example.json.`)
}

func loadConfig(configPath string) (config, error) {
	cfg := config{
		ExcludeDirs: append([]string(nil), defaultExcludeDirs...),
		Scan: scanConfig{
			Extensions: append([]string(nil), defaultExtensions...),
			Keywords:   append([]string(nil), defaultKeywords...),
		},
	}
	if configPath == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config{}, err
	}
	var loaded config
	if err := json.Unmarshal(data, &loaded); err != nil {
		return config{}, err
	}

	if len(loaded.ExcludeDirs) > 0 {
		cfg.ExcludeDirs = loaded.ExcludeDirs
	}
	if len(loaded.Scan.Extensions) > 0 {
		cfg.Scan.Extensions = loaded.Scan.Extensions
	}
	if len(loaded.Scan.Keywords) > 0 {
		cfg.Scan.Keywords = loaded.Scan.Keywords
	}
	cfg.Rules = loaded.Rules
	return cfg, nil
}

func scanProject(root string, excludeDirs, extensions, keywords []string) ([]scanHit, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	re, err := makeKeywordRegexp(keywords)
	if err != nil {
		return nil, err
	}

	hits := make([]scanHit, 0)
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if shouldSkipPath(rel, d.IsDir(), excludeDirs) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || !hasAllowedExtension(rel, extensions) {
			return nil
		}
		fileHits, err := scanFile(path, rel, re)
		if err != nil {
			return err
		}
		hits = append(hits, fileHits...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].File == hits[j].File {
			return hits[i].Line < hits[j].Line
		}
		return hits[i].File < hits[j].File
	})
	return hits, nil
}

func scanFile(absPath, relPath string, re *regexp.Regexp) ([]scanHit, error) {
	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hits := make([]scanHit, 0)
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		if re.MatchString(line) {
			hits = append(hits, scanHit{
				File: relPath,
				Line: lineNo,
				Text: strings.TrimSpace(line),
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return hits, nil
}

func applyRule(root string, excludeDirs []string, rule ruleConfig, dryRun bool) (int, int, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return 0, 0, err
	}

	paths := normalizePaths(rule.Paths)
	fileChanges := 0
	renames := 0

	if len(rule.Replacements) > 0 {
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if shouldSkipPath(rel, d.IsDir(), excludeDirs) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.IsDir() || !matchesTargets(rel, paths) {
				return nil
			}

			changed, err := applyReplacements(path, rel, rule.Replacements, dryRun)
			if err != nil {
				return err
			}
			if changed {
				fileChanges++
			}
			return nil
		})
		if err != nil {
			return 0, 0, err
		}
	}

	if len(rule.Renames) > 0 {
		for _, rename := range rule.Renames {
			from := filepath.Join(root, filepath.FromSlash(rename.From))
			to := filepath.Join(root, filepath.FromSlash(rename.To))
			if _, err := os.Stat(from); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return fileChanges, renames, err
			}
			if dryRun {
				fmt.Printf("[dry-run] rename %s -> %s\n", rename.From, rename.To)
				renames++
				continue
			}
			if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
				return fileChanges, renames, err
			}
			if err := os.Rename(from, to); err != nil {
				return fileChanges, renames, err
			}
			fmt.Printf("renamed %s -> %s\n", rename.From, rename.To)
			renames++
		}
	}

	return fileChanges, renames, nil
}

func applyReplacements(absPath, relPath string, replacements []replaceConfig, dryRun bool) (bool, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return false, err
	}
	original := string(data)
	updated := original

	for _, repl := range replacements {
		switch repl.Type {
		case "", "literal":
			if repl.From == "" {
				return false, fmt.Errorf("literal replacement in %s is missing 'from'", relPath)
			}
			updated = strings.ReplaceAll(updated, repl.From, repl.To)
		case "regex":
			if repl.Pattern == "" {
				return false, fmt.Errorf("regex replacement in %s is missing 'pattern'", relPath)
			}
			re, err := regexp.Compile(repl.Pattern)
			if err != nil {
				return false, err
			}
			updated = re.ReplaceAllString(updated, repl.To)
		default:
			return false, fmt.Errorf("unsupported replacement type %q", repl.Type)
		}
	}

	if updated == original {
		return false, nil
	}
	if dryRun {
		fmt.Printf("[dry-run] update %s\n", relPath)
		return true, nil
	}
	if err := os.WriteFile(absPath, []byte(updated), 0o644); err != nil {
		return false, err
	}
	fmt.Printf("updated %s\n", relPath)
	return true, nil
}

func selectRules(all []ruleConfig, selectedNames []string) ([]ruleConfig, error) {
	if len(selectedNames) == 0 || (len(selectedNames) == 1 && selectedNames[0] == "all") {
		return all, nil
	}

	index := make(map[string]ruleConfig, len(all))
	for _, rule := range all {
		index[rule.Name] = rule
	}

	selected := make([]ruleConfig, 0, len(selectedNames))
	for _, name := range selectedNames {
		rule, ok := index[name]
		if !ok {
			return nil, fmt.Errorf("unknown rule %q", name)
		}
		selected = append(selected, rule)
	}
	return selected, nil
}

func makeKeywordRegexp(keywords []string) (*regexp.Regexp, error) {
	if len(keywords) == 0 {
		return nil, errors.New("no keywords configured")
	}
	parts := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}
		parts = append(parts, regexp.QuoteMeta(keyword))
	}
	if len(parts) == 0 {
		return nil, errors.New("no valid keywords configured")
	}
	return regexp.Compile("(?i)(" + strings.Join(parts, "|") + ")")
}

func shouldSkipPath(rel string, isDir bool, excludeDirs []string) bool {
	rel = filepath.ToSlash(rel)
	if rel == "." || rel == "" {
		return false
	}
	for _, exclude := range excludeDirs {
		exclude = strings.Trim(filepath.ToSlash(exclude), "/")
		if exclude == "" {
			continue
		}
		if strings.Contains(exclude, "/") {
			if rel == exclude || strings.HasPrefix(rel, exclude+"/") {
				return true
			}
			continue
		}
		parts := strings.Split(rel, "/")
		limit := len(parts)
		if !isDir {
			limit--
		}
		for i := 0; i < limit; i++ {
			if parts[i] == exclude {
				return true
			}
		}
	}
	return false
}

func hasAllowedExtension(path string, extensions []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowed := range extensions {
		if strings.EqualFold(ext, allowed) {
			return true
		}
	}
	return false
}

func matchesTargets(rel string, targets []string) bool {
	if len(targets) == 0 {
		return true
	}
	rel = filepath.ToSlash(rel)
	for _, target := range targets {
		if rel == target || strings.HasPrefix(rel, target+"/") {
			return true
		}
	}
	return false
}

func normalizePaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.Trim(filepath.ToSlash(path), "/")
		if path != "" {
			out = append(out, path)
		}
	}
	return out
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func pickList(primary, fallback, defaults []string) []string {
	switch {
	case len(primary) > 0:
		return primary
	case len(fallback) > 0:
		return fallback
	default:
		return defaults
	}
}

func countFiles(hits []scanHit) int {
	files := make(map[string]struct{}, len(hits))
	for _, hit := range hits {
		files[hit.File] = struct{}{}
	}
	return len(files)
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
