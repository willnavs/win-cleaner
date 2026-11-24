package cleaner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

var (
	shell32                 = syscall.NewLazyDLL("shell32.dll")
	procSHEmptyRecycleBinW  = shell32.NewProc("SHEmptyRecycleBinW")
)

const (
	SHERB_NOCONFIRMATION = 0x00000001
	SHERB_NOPROGRESSUI   = 0x00000002
	SHERB_NOSOUND        = 0x00000004
)

type recycleBinTarget struct{}

func (r *recycleBinTarget) Name() string     { return "Lixeira" }
func (r *recycleBinTarget) Category() string { return "Lixeira" }
func (r *recycleBinTarget) Paths() []string  { return []string{} }
func (r *recycleBinTarget) ShouldClean(path string) bool { return false }
func (r *recycleBinTarget) Clean(ctx context.Context, path string, dryRun bool) (int64, error) {
	if dryRun {
		return estimateRecycleBinSize(), nil
	}
	return 0, emptyRecycleBin()
}

func emptyRecycleBin() error {
	ret, _, err := procSHEmptyRecycleBinW.Call(
		0, // hwnd
		0, // pszRootPath (NULL = all drives)
		uintptr(SHERB_NOCONFIRMATION|SHERB_NOPROGRESSUI|SHERB_NOSOUND),
	)
	if ret != 0 {
		return err
	}
	return nil
}

func estimateRecycleBinSize() int64 {
	// Estimativa conservadora - em dry-run mostramos ~50MB como referência
	return 50 * 1024 * 1024
}

func getUserTempPaths() []string {
	var paths []string
	usersDir := filepath.Join(os.Getenv("SystemDrive"), "Users")
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return paths
	}

	skipDirs := map[string]bool{
		"Public": true, "Default": true, "Default User": true, "All Users": true,
		"Public.Desktop": true, "Default.migrated": true,
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if skipDirs[name] {
			continue
		}
		tempPath := filepath.Join(usersDir, name, "AppData", "Local", "Temp")
		if info, err := os.Stat(tempPath); err == nil && info.IsDir() {
			paths = append(paths, tempPath)
		}
	}

	if temp := os.Getenv("TEMP"); temp != "" {
		paths = append(paths, temp)
	}
	if tmp := os.Getenv("TMP"); tmp != "" && tmp != os.Getenv("TEMP") {
		paths = append(paths, tmp)
	}

	return paths
}

func NewWindowsTargets() []Target {
	var targets []Target

	// Lixeira
	targets = append(targets, &recycleBinTarget{})

	// %TEMP% de todos os perfis
	targets = append(targets, &baseTarget{
		name:     "Temp_Usuarios",
		category: "Windows/Temp",
		paths:    getUserTempPaths(),
		patterns: []string{"*"},
	})

	// C:\Windows\Temp
	targets = append(targets, &baseTarget{
		name:     "Windows_Temp",
		category: "Windows/Temp",
		paths:    []string{filepath.Join(os.Getenv("SystemRoot"), "Temp")},
		patterns: []string{"*"},
	})

	// Logs seguros do Windows
	systemRoot := os.Getenv("SystemRoot")
	programData := os.Getenv("ProgramData")
	localAppData := os.Getenv("LocalAppData")

	logTargets := []struct {
		name     string
		category string
		paths    []string
		patterns []string
	}{
		{"Windows_Logs_CBS", "Windows/Logs", []string{filepath.Join(systemRoot, "Logs", "CBS")}, []string{"*.log"}},
		{"Windows_Logs_setupact", "Windows/Logs", []string{filepath.Join(systemRoot, "Logs")}, []string{"setupact.log", "setuperr.log"}},
		{"Windows_Logs_MeasuredBoot", "Windows/Logs", []string{filepath.Join(systemRoot, "Logs", "MeasuredBoot")}, []string{"*.log"}},
		{"Windows_Logs_MoSetup", "Windows/Logs", []string{filepath.Join(systemRoot, "Panther")}, []string{"*.log"}},
		{"Windows_Logs_Panther", "Windows/Logs", []string{filepath.Join(systemRoot, "Panther")}, []string{"*.log"}},
		{"Windows_Logs_WinSAT", "Windows/Logs", []string{filepath.Join(systemRoot, "Performance", "WinSAT")}, []string{"winsat.log"}},
		{"Windows_Logs_inf", "Windows/Logs", []string{filepath.Join(systemRoot, "inf")}, []string{"*.log"}},
		{"Windows_Logs_SoftwareDistribution", "Windows/Logs", []string{filepath.Join(systemRoot, "SoftwareDistribution")}, []string{"*.log"}},
		{"Windows_Logs_NET", "Windows/Logs", []string{filepath.Join(systemRoot, "Microsoft.NET")}, []string{"*.log"}},
		{"Windows_Logs_MpCmdRun", "Windows/Logs", []string{filepath.Join(programData, "Microsoft", "Windows Defender", "Support")}, []string{"MpCmdRun.log"}},
		{"CrashDumps", "Windows/Logs", []string{filepath.Join(systemRoot, "Minidump"), filepath.Join(systemRoot, "MEMORY.DMP")}, []string{"*.dmp", "MEMORY.DMP"}},
		{"Explorer_ThumbCache", "Windows/Explorer", []string{filepath.Join(localAppData, "Microsoft", "Windows", "Explorer")}, []string{"thumbcache_*.db", "iconcache_*.db"}},
		{"INetCache_IE", "Windows/Explorer", []string{filepath.Join(localAppData, "Microsoft", "Windows", "INetCache", "IE")}, []string{"*"}},
		{"CryptnetUrlCache", "Windows/Explorer", []string{filepath.Join(localAppData, "Microsoft", "Windows", "CryptnetUrlCache")}, []string{"*"}},
	}

	for _, lt := range logTargets {
		// Filter out paths that don't exist on this system
		var validPaths []string
		for _, p := range lt.paths {
			if _, err := os.Stat(p); err == nil {
				validPaths = append(validPaths, p)
			}
		}
		if len(validPaths) > 0 {
			targets = append(targets, &baseTarget{
				name:     lt.name,
				category: lt.category,
				paths:    validPaths,
				patterns: lt.patterns,
			})
		}
	}

	// Force GOMAXPROCS to be available for runtime
	_ = runtime.NumCPU()

	return targets
}