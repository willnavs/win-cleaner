package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Target interface para todos os alvos de limpeza
type Target interface {
	Name() string
	Category() string
	Paths() []string
	ShouldClean(path string) bool
	Clean(ctx context.Context, path string, dryRun bool) (int64, error)
}

// matchPattern faz matching simples de glob (* e ?)
func matchPattern(name, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(name, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(name, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	}
	return name == pattern
}

// contains verifica se slice contém item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// baseTarget implementação base para targets simples
type baseTarget struct {
	name     string
	category string
	paths    []string
	patterns []string
}

func (b *baseTarget) Name() string        { return b.name }
func (b *baseTarget) Category() string    { return b.category }
func (b *baseTarget) Paths() []string     { return b.paths }
func (b *baseTarget) ShouldClean(path string) bool {
	base := filepath.Base(path)
	for _, pat := range b.patterns {
		if matchPattern(base, pat) {
			return true
		}
	}
	return false
}
func (b *baseTarget) Clean(ctx context.Context, path string, dryRun bool) (int64, error) {
	if dryRun {
		info, err := os.Stat(path)
		if err != nil {
			return 0, err
		}
		return info.Size(), nil
	}
	return 0, os.Remove(path)
}

// filterExistingPaths filtra apenas caminhos que existem no filesystem
func filterExistingPaths(paths []string) []string {
	var valid []string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			valid = append(valid, p)
		}
	}
	return valid
}