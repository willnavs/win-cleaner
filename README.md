# win-cleaner

> Ferramenta CLI em Go para limpeza segura e rápida de arquivos temporários, caches e logs do Windows.

## Características

- **Rápido**: Paralelismo controlado com worker pools (`errgroup` + `semaphore`)
- **Seguro**: Ignora silenciosamente arquivos em uso/protegidos (Access Denied, Sharing Violation)
- **Idiomático**: Código Go limpo, tipado, sem dependências desnecessárias
- **Standalone**: Binário único (~3-5 MB), sem runtime externo
- **Windows nativo**: Usa APIs do Windows (SHEmptyRecycleBinW) para Lixeira

## Instalação

### Binário pré-compilado
Baixe o `win-cleaner.exe` mais recente em [Releases](https://github.com/willnavs/win-cleaner/releases).

### Compilar da fonte
```bash
git clone https://github.com/willnavs/win-cleaner
cd win-cleaner
go build -ldflags="-s -w" -o win-cleaner.exe ./cmd/cleaner
```

### Via Go install
```bash
go install github.com/willnavs/win-cleaner/cmd/cleaner@latest
```

## Uso

```bash
win-cleaner [opções]
```

### Opções

| Flag | Curto | Descrição |
|------|-------|-----------|
| `--dry-run` | | Simula a limpeza e mostra exatamente quanto espaço (MB/GB) seria liberado **sem apagar nada** |
| `--kill` | | Encerra processos de navegadores (Chrome, Edge, Firefox, Brave, Vivaldi, Opera) e OneDrive **antes** de limpar. Padrão: `false` (apenas pula arquivos em uso) |
| `--extras` | | Inclui caches pesados modernos: **Discord** (Cache/Code Cache), **Steam** (htmlcache), **Opera/Opera GX**, **DirectX Shader Cache** |
| `--yes` / `-y` | `-y` | Executa direto sem pedir confirmação interativa no início |
| `--version` | | Mostra versão e sai |
| `--help` | | Mostra ajuda e sai |

### Exemplos

```bash
# Simular limpeza (recomendado na primeira execução)
win-cleaner --dry-run

# Limpeza padrão (pede confirmação)
win-cleaner

# Limpeza automática sem confirmação
win-cleaner -y

# Limpeza completa: encerra navegadores + extras
win-cleaner --kill --extras -y

# Apenas limpar Lixeira, Temp e logs do Windows (sem navegadores/apps)
win-cleaner --dry-run
```

## O que é limpo

### Lixeira
- Esvaziamento via `SHEmptyRecycleBinW` (API nativa do Windows) com flags `SHERB_NOCONFIRMATION | SHERB_NOPROGRESSUI | SHERB_NOSOUND`

### Windows / Temp
- `%TEMP%` de **todos os perfis** em `C:\Users\*` (ignora `Public`, `Default`, `Default User`, `All Users`)
- `C:\Windows\Temp` (requer Administrador)

### Windows / Logs (seguros)
- `C:\Windows\Logs\CBS\*.log`
- `C:\Windows\Logs\setupact.log`, `setuperr.log`
- `C:\Windows\Logs\MeasuredBoot\*.log`
- `C:\Windows\Panther\*.log`
- `C:\Windows\Performance\WinSAT\winsat.log`
- `C:\Windows\inf\*.log`
- `C:\Windows\SoftwareDistribution\*.log`
- `C:\Windows\Microsoft.NET\*.log`
- `C:\ProgramData\Microsoft\Windows Defender\Support\MpCmdRun.log`
- `C:\Windows\Minidump\*.dmp`, `C:\Windows\MEMORY.DMP`

### Windows / Explorer
- `C:\Users\<user>\AppData\Local\Microsoft\Windows\Explorer\thumbcache_*.db`, `iconcache_*.db`
- `C:\Users\<user>\AppData\Local\Microsoft\Windows\INetCache\IE\*`
- `C:\Users\<user>\AppData\Local\Microsoft\Windows\CryptnetUrlCache\*`

### Navegadores (auto-descoberta de perfis)
**Edge, Chrome, Brave, Vivaldi** (perfis: `Default`, `Guest Profile`, `Profile 1`, `Profile 2`, ...):
- `Cache\*`, `Cache_Data\*`
- `Code Cache\js\*`, `Code Cache\wasm\*`, `Code Cache\webui_js\*`
- `GPUCache\*`
- `Service Worker\CacheStorage\*`, `Service Worker\ScriptCache\*`
- `JumpListIconsRecentClosed\*`
- `BrowserMetrics*.pma`

**Firefox** (corrigido: **apenas** `AppData\Local\Mozilla\Firefox\Profiles\*\*`):
- `cache2\*`
- `startupCache\*`
- `jumpListCache\*`
- **NUNCA** toca em `AppData\Roaming\Mozilla` nem arquivos sem extensão na raiz do perfil

**Opera / Opera GX** (apenas com `--extras`):
- Mesmos diretórios dos navegadores Chromium acima

### Aplicativos
- **OneDrive**: `setup\logs\*.log`, `*.odl`, `*.aodl`, `*.otc`, `*.qmlc`
- **Spotify**: `Data\*.file`, `Browser\Cache\*`
- **Adobe**: `Common\Media Cache Files\*`, `Logs\*.log`
- **VMware**: `C:\ProgramData\VMware\logs\*.log`
- **TeamViewer**: `EdgeBrowserControl\*`

### Extras (apenas com `--extras`)
- **Discord**: `Cache\*`, `Code Cache\*`, `GPUCache\*` (todas versões `app-*` e `update-*`)
- **Steam**: `htmlcache\*`, `appcache\httpcache\*`, `userdata\<id>\cache\*`
- **DirectX Shader Cache**: `%LocalAppData%\D3DSCache\*`, `C:\Windows\Temp\D3DSCache\*`

## Privilégios de Administrador

| Cenário | Comportamento |
|---------|---------------|
| **Como Administrador** | Limpeza completa: `C:\Windows\Temp`, logs do sistema, perfis de outros usuários |
| **Como Usuário Comum** | ⚠ Aviso exibido. Pastas de sistema e outros usuários são **puladas**. Limpa apenas o perfil atual. |

> **Recomendação**: Execute como Administrador para limpeza máxima.

## Relatório de saída

```
╔═══════════════════════════════════════════════════════════════════╗
║                    RELATÓRIO DE LIMPEZA                         ║
╠═══════════════════════════════════════════════════════════════════╣
║ Categoria                        Arquivos       Espaço   Erros ║
╠═══════════════════════════════════════════════════════════════════╣
║ Lixeira                                0       50.00 MiB      0 ║
║ Windows/Temp                        1234      342.15 MiB      3 ║
║ Windows/Logs                         567       89.42 MiB      0 ║
║ Windows/Explorer                      89       12.30 MiB      1 ║
║ Navegadores                         4321      1.23 GiB       5 ║
║ Aplicativos                          234       210.55 MiB      0 ║
╠═══════════════════════════════════════════════════════════════════╣
║ TOTAL                               6445      1.90 GiB       9 ║
╚═══════════════════════════════════════════════════════════════════╝

Tempo total: 2.34s
```

- **Arquivos**: Quantidade de arquivos removidos
- **Espaço**: Tamanho total liberado (formato legível: B, KiB, MiB, GiB)
- **Erros**: Arquivos ignorados por estarem em uso/protegidos (Access Denied, Sharing Violation)

## Requisitos

- Windows 10/11 (x64 ou ARM64)
- Go 1.22+ (para compilar)

## Desenvolvimento

```bash
# Dependências
go mod tidy

# Análise estática
go vet ./...

# Testes com race detector
go test -race ./...

# Build otimizado
go build -ldflags="-s -w" -o win-cleaner.exe ./cmd/cleaner

# Cross-compile
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win-cleaner_windows_amd64.exe ./cmd/cleaner
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o win-cleaner_windows_arm64.exe ./cmd/cleaner
```

## Estrutura do projeto

```
win-cleaner/
├── cmd/cleaner/main.go              # Entry point CLI (flag stdlib)
├── internal/
│   ├── sysutil/admin_windows.go     # Verificação de Admin (Windows API)
│   └── cleaner/
│       ├── engine.go                # Motor de limpeza paralela + relatório
│       ├── targets_windows.go       # Alvos Windows (Lixeira, Temp, Logs, Explorer)
│       ├── targets_browsers.go      # Alvos Navegadores (auto-descoberta perfis)
│       └── targets_apps.go          # Alvos Apps (OneDrive, Spotify, Adobe, etc.)
├── go.mod
└── README.md
```

## Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.

---

**Desenvolvido com ❤️ em Go para Windows**