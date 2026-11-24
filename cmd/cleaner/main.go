package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/willnavs/win-cleaner/internal/cleaner"
	"github.com/willnavs/win-cleaner/internal/sysutil"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	dryRun := flag.Bool("dry-run", false, "Simula a limpeza e mostra quanto espaço seria liberado sem apagar nada")
	killProcesses := flag.Bool("kill", false, "Encerra processos de navegadores e OneDrive antes de limpar (padrão: false)")
	extras := flag.Bool("extras", false, "Inclui caches pesados modernos (Discord, Steam, Opera/Opera GX, DirectX Shader Cache)")
	yes := flag.Bool("y", false, "Executa direto sem pedir confirmação interativa")
	yesLong := flag.Bool("yes", false, "Executa direto sem pedir confirmação interativa")
	showVersion := flag.Bool("version", false, "Mostra versão e sai")
	showHelp := flag.Bool("help", false, "Mostra ajuda e sai")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "win-cleaner v%s (%s, %s)\n\n", version, commit, date)
		fmt.Fprintf(flag.CommandLine.Output(), "Uso: win-cleaner [opções]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Limpa arquivos temporários, caches e logs do Windows de forma segura.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Opções:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExemplos:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  win-cleaner --dry-run          # Simula limpeza\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  win-cleaner --kill --extras    # Encerra processos e limpa tudo incluindo extras\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  win-cleaner -y                 # Limpa sem confirmação\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("win-cleaner v%s (%s, %s)\n", version, commit, date)
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	autoYes := *yes || *yesLong

	if !sysutil.IsAdmin() {
		fmt.Println("⚠ AVISO: Executando sem privilégios de Administrador.")
		fmt.Println("   Pastas de sistema (C:\\Windows\\Temp, C:\\Windows\\Logs, outros usuários) serão ignoradas.")
		fmt.Println("   Para limpeza completa, execute como Administrador.")
	}

	opts := cleaner.Options{
		DryRun:       *dryRun,
		KillProcesses: *killProcesses,
		Extras:       *extras,
		AutoYes:      autoYes,
	}

	engine := cleaner.NewEngine(opts)
	report, err := engine.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}

	report.Print()

	if opts.DryRun {
		fmt.Println("\n[DRY-RUN] Nenhum arquivo foi removido. Execute sem --dry-run para limpar de verdade.")
	}
}