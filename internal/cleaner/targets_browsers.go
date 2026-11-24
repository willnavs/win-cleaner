package cleaner

import (
	"os"
	"path/filepath"
	"strings"
)

func NewBrowserTargets(extras bool) []Target {
	var targets []Target

	// Chrome-based browsers (Edge, Chrome, Brave, Vivaldi)
	chromeBrowsers := []struct {
		name       string
		localAppData string
	}{
		{"Edge", "Microsoft/Edge/User Data"},
		{"Chrome", "Google/Chrome/User Data"},
		{"Brave", "BraveSoftware/Brave-Browser/User Data"},
		{"Vivaldi", "Vivaldi/User Data"},
	}

	for _, br := range chromeBrowsers {
		targets = append(targets, discoverChromeProfiles(br.name, br.localAppData)...)
	}

	// Firefox
	targets = append(targets, discoverFirefoxProfiles()...)

	// Opera/Opera GX (only with --extras)
	if extras {
		operaBrowsers := []struct {
			name       string
			localAppData string
		}{
			{"Opera", "Opera Software/Opera Stable"},
			{"Opera GX", "Opera Software/Opera GX Stable"},
		}
		for _, br := range operaBrowsers {
			targets = append(targets, discoverOperaProfiles(br.name, br.localAppData)...)
		}
	}

	return targets
}

func discoverChromeProfiles(browserName, basePath string) []Target {
	var targets []Target
	localAppData := os.Getenv("LocalAppData")
	if localAppData == "" {
		return targets
	}

	userDataPath := filepath.Join(localAppData, basePath)
	entries, err := os.ReadDir(userDataPath)
	if err != nil {
		return targets
	}

	profileNames := []string{"Default"}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "Profile ") {
			profileNames = append(profileNames, entry.Name())
		}
	}
	// Guest Profile
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == "Guest Profile" {
			profileNames = append(profileNames, "Guest Profile")
		}
	}

	for _, profile := range profileNames {
		profilePath := filepath.Join(userDataPath, profile)

		// Cache/Cache_Data
		cachePaths := []string{
			filepath.Join(profilePath, "Cache"),
			filepath.Join(profilePath, "Cache_Data"),
		}
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_Cache",
			category: "Navegadores",
			paths:    filterExistingPaths(cachePaths),
			patterns: []string{"*"},
		})

		// Code Cache/js
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_CodeCache_js",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "Code Cache", "js")}),
			patterns: []string{"*"},
		})

		// Code Cache/wasm
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_CodeCache_wasm",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "Code Cache", "wasm")}),
			patterns: []string{"*"},
		})

		// Code Cache/webui_js
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_CodeCache_webui_js",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "Code Cache", "webui_js")}),
			patterns: []string{"*"},
		})

		// GPUCache
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_GPUCache",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "GPUCache")}),
			patterns: []string{"*"},
		})

		// Service Worker/CacheStorage
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_ServiceWorker_CacheStorage",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "Service Worker", "CacheStorage")}),
			patterns: []string{"*"},
		})

		// Service Worker/ScriptCache
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_ServiceWorker_ScriptCache",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "Service Worker", "ScriptCache")}),
			patterns: []string{"*"},
		})

		// JumpListIconsRecentClosed
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_JumpListIcons",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "JumpListIconsRecentClosed")}),
			patterns: []string{"*"},
		})

		// BrowserMetrics*.pma
		targets = append(targets, &baseTarget{
			name:     browserName + "_" + profile + "_BrowserMetrics",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{profilePath}),
			patterns: []string{"BrowserMetrics*.pma"},
		})
	}

	return targets
}

func discoverFirefoxProfiles() []Target {
	var targets []Target
	localAppData := os.Getenv("LocalAppData")
	if localAppData == "" {
		return targets
	}

	// Firefox cache está em LocalAppData, NÃO em Roaming
	profilesPath := filepath.Join(localAppData, "Mozilla", "Firefox", "Profiles")
	entries, err := os.ReadDir(profilesPath)
	if err != nil {
		return targets
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		profileName := entry.Name()
		profilePath := filepath.Join(profilesPath, profileName)

		// cache2
		targets = append(targets, &baseTarget{
			name:     "Firefox_" + profileName + "_cache2",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "cache2")}),
			patterns: []string{"*"},
		})

		// startupCache
		targets = append(targets, &baseTarget{
			name:     "Firefox_" + profileName + "_startupCache",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "startupCache")}),
			patterns: []string{"*"},
		})

		// jumpListCache
		targets = append(targets, &baseTarget{
			name:     "Firefox_" + profileName + "_jumpListCache",
			category: "Navegadores",
			paths:    filterExistingPaths([]string{filepath.Join(profilePath, "jumpListCache")}),
			patterns: []string{"*"},
		})
	}

	return targets
}

func discoverOperaProfiles(browserName, basePath string) []Target {
	var targets []Target
	localAppData := os.Getenv("LocalAppData")
	if localAppData == "" {
		return targets
	}

	userDataPath := filepath.Join(localAppData, basePath)
	entries, err := os.ReadDir(userDataPath)
	if err != nil {
		return targets
	}

	profileNames := []string{"Default"}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "Profile ") {
			profileNames = append(profileNames, entry.Name())
		}
	}

	cacheDirs := []string{
		"Cache",
		"Cache_Data",
		"Code Cache",
		"GPUCache",
		"Service Worker",
	}

	for _, profile := range profileNames {
		profilePath := filepath.Join(userDataPath, profile)

		for _, cacheDir := range cacheDirs {
			targets = append(targets, &baseTarget{
				name:     browserName + "_" + profile + "_" + cacheDir,
				category: "Navegadores",
				paths:    filterExistingPaths([]string{filepath.Join(profilePath, cacheDir)}),
				patterns: []string{"*"},
			})
		}
	}

	return targets
}