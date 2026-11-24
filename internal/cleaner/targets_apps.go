package cleaner

import (
	"os"
	"path/filepath"
	"strings"
)

func NewAppTargets(extras bool) []Target {
	var targets []Target

	localAppData := os.Getenv("LocalAppData")
	appData := os.Getenv("AppData")
	programData := os.Getenv("ProgramData")

	// OneDrive logs and cache
	if localAppData != "" {
		oneDriveSetupLogs := filepath.Join(localAppData, "Microsoft", "OneDrive", "setup", "logs")
		if _, err := os.Stat(oneDriveSetupLogs); err == nil {
			targets = append(targets, &baseTarget{
				name:     "OneDrive_Logs",
				category: "Aplicativos",
				paths:    []string{oneDriveSetupLogs},
				patterns: []string{"*.log"},
			})
		}

		oneDriveRoot := filepath.Join(localAppData, "Microsoft", "OneDrive")
		if _, err := os.Stat(oneDriveRoot); err == nil {
			targets = append(targets, &baseTarget{
				name:     "OneDrive_Cache",
				category: "Aplicativos",
				paths:    []string{oneDriveRoot},
				patterns: []string{"*.odl", "*.aodl", "*.otc", "*.qmlc"},
			})
		}
	}

	// Spotify
	if localAppData != "" {
		spotifyPath := filepath.Join(localAppData, "Spotify")
		if _, err := os.Stat(spotifyPath); err == nil {
			targets = append(targets, &baseTarget{
				name:     "Spotify_Data",
				category: "Aplicativos",
				paths:    filterExistingPaths([]string{filepath.Join(spotifyPath, "Data")}),
				patterns: []string{"*.file"},
			})
			targets = append(targets, &baseTarget{
				name:     "Spotify_BrowserCache",
				category: "Aplicativos",
				paths:    filterExistingPaths([]string{filepath.Join(spotifyPath, "Browser", "Cache")}),
				patterns: []string{"*"},
			})
		}
	}

	// Adobe Media Cache
	if appData != "" {
		adobeCache := filepath.Join(appData, "Adobe", "Common", "Media Cache Files")
		if _, err := os.Stat(adobeCache); err == nil {
			targets = append(targets, &baseTarget{
				name:     "Adobe_MediaCache_Files",
				category: "Aplicativos",
				paths:    []string{adobeCache},
				patterns: []string{"*"},
			})
		}

		adobeLogs := filepath.Join(appData, "Adobe", "Logs")
		if _, err := os.Stat(adobeLogs); err == nil {
			targets = append(targets, &baseTarget{
				name:     "Adobe_Logs",
				category: "Aplicativos",
				paths:    []string{adobeLogs},
				patterns: []string{"*.log"},
			})
		}
	}

	// VMware
	if programData != "" {
		vmwareLogs := filepath.Join(programData, "VMware", "logs")
		if _, err := os.Stat(vmwareLogs); err == nil {
			targets = append(targets, &baseTarget{
				name:     "VMware_Logs",
				category: "Aplicativos",
				paths:    []string{vmwareLogs},
				patterns: []string{"*.log"},
			})
		}
	}

	// TeamViewer
	if programData != "" {
		teamViewerPath := filepath.Join(programData, "TeamViewer")
		if _, err := os.Stat(teamViewerPath); err == nil {
			targets = append(targets, &baseTarget{
				name:     "TeamViewer_EdgeBrowserControl",
				category: "Aplicativos",
				paths:    filterExistingPaths([]string{filepath.Join(teamViewerPath, "EdgeBrowserControl")}),
				patterns: []string{"*"},
			})
		}
	}

	// Extras: Discord, Steam, DirectX Shader Cache
	if extras {
		// Discord
		if localAppData != "" {
			discordPath := filepath.Join(localAppData, "Discord")
			if _, err := os.Stat(discordPath); err == nil {
				// Current version caches
				targets = append(targets, &baseTarget{
					name:     "Discord_Cache",
					category: "Aplicativos (Extras)",
					paths:    filterExistingPaths([]string{filepath.Join(discordPath, "Cache"), filepath.Join(discordPath, "Code Cache"), filepath.Join(discordPath, "GPUCache")}),
					patterns: []string{"*"},
				})
				// Discord versioned folders
				entries, _ := os.ReadDir(discordPath)
				for _, entry := range entries {
					if entry.IsDir() && (strings.HasPrefix(entry.Name(), "app-") || strings.HasPrefix(entry.Name(), "update-")) {
						verPath := filepath.Join(discordPath, entry.Name())
						targets = append(targets, &baseTarget{
							name:     "Discord_" + entry.Name() + "_Cache",
							category: "Aplicativos (Extras)",
							paths:    filterExistingPaths([]string{filepath.Join(verPath, "Cache"), filepath.Join(verPath, "Code Cache"), filepath.Join(verPath, "GPUCache")}),
							patterns: []string{"*"},
						})
					}
				}
			}
		}

		// Steam htmlcache
		if programData != "" {
			steamPath := filepath.Join(programData, "Steam")
			if _, err := os.Stat(steamPath); err == nil {
				targets = append(targets, &baseTarget{
					name:     "Steam_htmlcache",
					category: "Aplicativos (Extras)",
					paths:    filterExistingPaths([]string{filepath.Join(steamPath, "htmlcache"), filepath.Join(steamPath, "appcache", "httpcache")}),
					patterns: []string{"*"},
				})
				// Steam userdata cache
				userdataPath := filepath.Join(steamPath, "userdata")
				if entries, err := os.ReadDir(userdataPath); err == nil {
					for _, entry := range entries {
						if entry.IsDir() {
							userCache := filepath.Join(userdataPath, entry.Name(), "cache")
							if _, err := os.Stat(userCache); err == nil {
								targets = append(targets, &baseTarget{
									name:     "Steam_Userdata_" + entry.Name() + "_cache",
									category: "Aplicativos (Extras)",
									paths:    []string{userCache},
									patterns: []string{"*"},
								})
							}
						}
					}
				}
			}
		}

		// DirectX Shader Cache
		if localAppData != "" {
			dxCache := filepath.Join(localAppData, "D3DSCache")
			if _, err := os.Stat(dxCache); err == nil {
				targets = append(targets, &baseTarget{
					name:     "DirectX_ShaderCache",
					category: "Aplicativos (Extras)",
					paths:    []string{dxCache},
					patterns: []string{"*"},
				})
			}
		}
		// Also check Windows DirectX cache
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot != "" {
			dxCache2 := filepath.Join(systemRoot, "Temp", "D3DSCache")
			if _, err := os.Stat(dxCache2); err == nil {
				targets = append(targets, &baseTarget{
					name:     "DirectX_ShaderCache_System",
					category: "Aplicativos (Extras)",
					paths:    []string{dxCache2},
					patterns: []string{"*"},
				})
			}
		}
	}

	return targets
}