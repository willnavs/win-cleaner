package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

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

	// Duplo clique no Explorer: sem argumentos e console dedicado ao processo.
	// Nesse caso o programa guia o usuário e pausa ao final antes de fechar.
	if len(os.Args) == 1 && sysutil.IsDoubleClickLaunch() {
		runInteractive()
		return
	}

	runFlags()
}

// runFlags processa a interface CLI tradicional baseada em flags.
func runFlags() {
	dryRun := flag.Bool("dry-run", false, "Simula a limpeza e mostra quanto espaço seria liberado sem apagar nada")
	killProcesses := flag.Bool("kill", false, "Encerra processos de navegadores e OneDrive antes de limpar (padrão: false)")
	extras := flag.Bool("extras", false, "Inclui caches pesados modernos (Discord, Steam, Opera/Opera GX, DirectX Shader Cache)")
	yes := flag.Bool("y", false, "Executa direto sem pedir confirmação interativa")
	yesLong := flag.Bool("yes", false, "Executa direto sem pedir confirmação interativa")
	interactive := flag.Bool("interactive", false, "Força o assistente interativo (mesmo fluxo do duplo clique)")
	showVersion := flag.Bool("version", false, "Mostra versão e sai")
	showHelp := flag.Bool("help", false, "Mostra ajuda e sai")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "win-cleaner v%s (%s, %s)\n\n", version, commit, date)
		fmt.Fprintf(flag.CommandLine.Output(), "Uso: win-cleaner [opções]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Limpa arquivos temporários, caches e logs do Windows de forma segura.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Sem argumentos, o duplo clique abre um assistente interativo.\n\n")
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

	if *interactive {
		runInteractive()
		return
	}

	if !sysutil.IsAdmin() {
		printAdminWarning()
	}

	opts := cleaner.Options{
		DryRun:        *dryRun,
		KillProcesses: *killProcesses,
		Extras:        *extras,
		AutoYes:       *yes || *yesLong,
	}

	runEngine(opts)
}

// runInteractive conduz a limpeza quando o executável é aberto por duplo clique.
func runInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("win-cleaner v%s\n", version)
	fmt.Println("Limpeza segura de temporários, caches e logs do Windows")
	fmt.Println(strings.Repeat("─", 62))
	fmt.Println()

	if !sysutil.IsAdmin() {
		printAdminWarning()
		if askYesNo(reader, "Reiniciar como Administrador agora? [s/N]: ") {
			if err := sysutil.RelaunchElevated(nil); err != nil {
				fmt.Printf("\n✖ Não foi possível elevar: %v\n", err)
				sysutil.WaitForEnter()
			}
			// Em caso de sucesso, a janela elevada assume o trabalho.
			return
		}
		fmt.Println()
	}

	fmt.Println("Configuração (pressione ENTER para manter o padrão):")
	extras := askYesNo(reader, "  Incluir caches extras (Discord, Steam, Opera/Opera GX, DirectX)? [s/N]: ")
	kill := askYesNo(reader, "  Encerrar navegadores e OneDrive antes da limpeza? [s/N]: ")
	dry := askYesNo(reader, "  Apenas simular, sem apagar nada? [s/N]: ")

	if !dry && !askYesNo(reader, "\nIniciar limpeza? [s/N]: ") {
		fmt.Println("Operação cancelada.")
		sysutil.WaitForEnter()
		return
	}

	opts := cleaner.Options{
		DryRun:        dry,
		KillProcesses: kill,
		Extras:        extras,
		AutoYes:       true, // confirmação já coletada acima
	}

	report, err := cleaner.NewEngine(opts).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nErro: %v\n", err)
		sysutil.WaitForEnter()
		return
	}

	report.Print()

	if opts.DryRun {
		fmt.Println("\n[DRY-RUN] Nenhum arquivo foi removido. Execute novamente respondendo \"n\" à simulação para limpar de verdade.")
	}

	sysutil.WaitForEnter()
}

// runEngine executa a limpeza, imprime o relatório e encerra com código adequado.
func runEngine(opts cleaner.Options) {
	report, err := cleaner.NewEngine(opts).Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}

	report.Print()

	if opts.DryRun {
		fmt.Println("\n[DRY-RUN] Nenhum arquivo foi removido. Execute sem --dry-run para limpar de verdade.")
	}
}

func printAdminWarning() {
	fmt.Println("⚠ AVISO: Executando sem privilégios de Administrador.")
	fmt.Println("   Pastas de sistema (C:\\Windows\\Temp, C:\\Windows\\Logs, outros usuários) serão ignoradas.")
	fmt.Println("   Para limpeza completa, execute como Administrador.")
}

// askYesNo lê uma resposta do usuário e interpreta como confirmação.
// ENTER (resposta vazia) mantém o padrão negativo, preservando a segurança.
func askYesNo(reader *bufio.Reader, prompt string) bool {
	fmt.Print(prompt)

	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "s", "sim", "y", "yes":
		return true
	default:
		return false
	}
}
