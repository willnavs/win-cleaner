package cleaner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Options struct {
	DryRun        bool
	KillProcesses bool
	Extras        bool
	AutoYes       bool
}

type CategoryReport struct {
	Name         string
	FilesRemoved int64
	BytesFreed   int64
	Errors       int64
}

type Report struct {
	Categories  []CategoryReport
	TotalFiles  int64
	TotalBytes  int64
	TotalErrors int64
	Duration    time.Duration
}

func (r *Report) Print() {
	fmt.Println("\n╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    RELATÓRIO DE LIMPEZA                         ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-30s %10s %12s %10s ║\n", "Categoria", "Arquivos", "Espaço", "Erros")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════╣")

	for _, cat := range r.Categories {
		sizeStr := formatBytes(cat.BytesFreed)
		fmt.Printf("║ %-30s %10d %12s %10d ║\n", cat.Name, cat.FilesRemoved, sizeStr, cat.Errors)
	}

	fmt.Println("╠═══════════════════════════════════════════════════════════════════╣")
	totalSize := formatBytes(r.TotalBytes)
	fmt.Printf("║ %-30s %10d %12s %10d ║\n", "TOTAL", r.TotalFiles, totalSize, r.TotalErrors)
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Printf("\nTempo total: %v\n", r.Duration.Round(time.Millisecond))
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

type Engine struct {
	opts    Options
	targets []Target
	report  Report
	mu      sync.Mutex
	sem     *semaphore.Weighted
}

func NewEngine(opts Options) *Engine {
	e := &Engine{
		opts: opts,
		sem:  semaphore.NewWeighted(int64(runtime.NumCPU() * 2)),
	}
	e.targets = e.buildTargets()
	return e
}

func (e *Engine) buildTargets() []Target {
	var targets []Target

	// Windows system targets
	targets = append(targets, NewWindowsTargets()...)

	// Browser targets
	targets = append(targets, NewBrowserTargets(e.opts.Extras)...)

	// App targets
	targets = append(targets, NewAppTargets(e.opts.Extras)...)

	return targets
}

func (e *Engine) Run() (*Report, error) {
	start := time.Now()

	if !e.opts.AutoYes && !e.opts.DryRun {
		fmt.Print("Esta operação irá remover arquivos temporários e caches. Continuar? [s/N]: ")
		var resp string
		fmt.Scanln(&resp)
		if resp != "s" && resp != "S" {
			fmt.Println("Operação cancelada.")
			return &e.report, nil
		}
	}

	if e.opts.KillProcesses {
		e.killBrowserAndOneDriveProcesses()
	}

	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.NumCPU() * 2)

	for _, target := range e.targets {
		target := target
		g.Go(func() error {
			return e.cleanTarget(ctx, target)
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	e.report.Duration = time.Since(start)
	return &e.report, nil
}

func (e *Engine) cleanTarget(ctx context.Context, target Target) error {
	paths := target.Paths()
	var wg sync.WaitGroup
	errCh := make(chan error, len(paths))

	for _, basePath := range paths {
		basePath := basePath
		wg.Add(1)
		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(runtime.NumCPU())

		g.Go(func() error {
			defer wg.Done()
			return filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					if os.IsPermission(err) || isSharingViolation(err) {
						e.addError(target.Category())
						return nil // Silenciosamente ignora
					}
					return err
				}

				if d.IsDir() {
					return nil
				}

				if !target.ShouldClean(path) {
					return nil
				}

				if err := e.acquireSemaphore(ctx); err != nil {
					return err
				}
				defer e.releaseSemaphore()

				info, err := d.Info()
				if err != nil {
					return nil
				}

				size := info.Size()
				if e.opts.DryRun {
					e.addResult(target.Category(), 1, size, 0)
					return nil
				}

				_, err = target.Clean(ctx, path, e.opts.DryRun)
				if err != nil {
					if isSharingViolation(err) || os.IsPermission(err) {
						e.addError(target.Category())
						return nil
					}
					return err
				}

				e.addResult(target.Category(), 1, size, 0)
				return nil
			})
		})
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	// Remove empty directories after cleaning
	if !e.opts.DryRun {
		e.removeEmptyDirs(paths)
	}

	return nil
}

func (e *Engine) acquireSemaphore(ctx context.Context) error {
	return e.sem.Acquire(ctx, 1)
}

func (e *Engine) releaseSemaphore() {
	e.sem.Release(1)
}

func (e *Engine) addResult(category string, files, bytes, errors int64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i := range e.report.Categories {
		if e.report.Categories[i].Name == category {
			e.report.Categories[i].FilesRemoved += files
			e.report.Categories[i].BytesFreed += bytes
			e.report.Categories[i].Errors += errors
			e.report.TotalFiles += files
			e.report.TotalBytes += bytes
			e.report.TotalErrors += errors
			return
		}
	}

	e.report.Categories = append(e.report.Categories, CategoryReport{
		Name:         category,
		FilesRemoved: files,
		BytesFreed:   bytes,
		Errors:       errors,
	})
	e.report.TotalFiles += files
	e.report.TotalBytes += bytes
	e.report.TotalErrors += errors
}

func (e *Engine) addError(category string) {
	e.addResult(category, 0, 0, 1)
}

func (e *Engine) removeEmptyDirs(paths []string) {
	for _, basePath := range paths {
		filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			entries, _ := os.ReadDir(path)
			if len(entries) == 0 {
				os.Remove(path) // Ignora erro
			}
			return nil
		})
	}
}

func (e *Engine) killBrowserAndOneDriveProcesses() {
	processNames := []string{
		"chrome.exe", "msedge.exe", "brave.exe", "vivaldi.exe",
		"opera.exe", "opera_gx.exe", "firefox.exe",
		"OneDrive.exe", "OneDriveSetup.exe",
	}

	procs, _ := process.Processes()
	for _, p := range procs {
		name, _ := p.Name()
		for _, target := range processNames {
			if equalFold(name, target) {
				p.Kill()
				break
			}
		}
	}
	// Pequena pausa para processos finalizarem
	time.Sleep(500 * time.Millisecond)
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func isSharingViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsSubstring(errStr, "sharing violation") ||
		containsSubstring(errStr, "being used by another process") ||
		containsSubstring(errStr, "access denied") ||
		containsSubstring(errStr, "Acesso negado") ||
		containsSubstring(errStr, "usado por outro processo")
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}