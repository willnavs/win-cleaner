package sysutil

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// swShowNormal é o valor de nCmdShow usado em ShellExecuteW.
const swShowNormal = 1

// ShellExecuteW retorna um valor > 32 em caso de sucesso.
const shellExecuteSuccessThreshold = 32

var (
	procGetConsoleProcessList = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleProcessList")
	procShellExecuteW         = windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteW")
)

// IsDoubleClickLaunch detecta inicialização por duplo clique no Explorer.
// Quando o Windows cria um console dedicado ao processo, apenas ele está
// anexado à janela (n == 1). Execuções a partir de cmd/PowerShell compartilham
// o console com o shell e retornam n > 1.
func IsDoubleClickLaunch() bool {
	if err := procGetConsoleProcessList.Find(); err != nil {
		return false
	}

	var pids [16]uint32
	n, _, _ := procGetConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids)),
	)
	return n == 1
}

// WaitForEnter bloqueia até o usuário pressionar ENTER, evitando que a janela
// do console feche imediatamente após a execução por duplo clique.
func WaitForEnter() {
	fmt.Print("\nPressione ENTER para sair...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

// RelaunchElevated reinicia o executável atual com elevação (UAC) via
// ShellExecuteW com o verbo "runas". Retorna erro se o usuário cancelar.
func RelaunchElevated(args []string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("localizar executável: %w", err)
	}

	verb, err := windows.UTF16PtrFromString("runas")
	if err != nil {
		return fmt.Errorf("converter verbo: %w", err)
	}
	file, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return fmt.Errorf("converter caminho: %w", err)
	}
	dir, err := windows.UTF16PtrFromString(filepath.Dir(exePath))
	if err != nil {
		return fmt.Errorf("converter diretório: %w", err)
	}

	var params uintptr
	if len(args) > 0 {
		paramPtr, err := windows.UTF16PtrFromString(joinQuotedArgs(args))
		if err != nil {
			return fmt.Errorf("converter argumentos: %w", err)
		}
		params = uintptr(unsafe.Pointer(paramPtr))
	}

	ret, _, lastErr := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		params,
		uintptr(unsafe.Pointer(dir)),
		uintptr(swShowNormal),
	)
	if ret <= shellExecuteSuccessThreshold {
		if lastErr != nil && lastErr != windows.ERROR_SUCCESS {
			return fmt.Errorf("elevação negada ou cancelada: %w", lastErr)
		}
		return fmt.Errorf("elevação negada ou cancelada (código %d)", ret)
	}

	return nil
}

// joinQuotedArgs monta uma linha de comando preservando argumentos com espaços.
func joinQuotedArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		if strings.ContainsAny(arg, " \t\"") {
			quoted = append(quoted, `"`+strings.ReplaceAll(arg, `"`, `\"`)+`"`)
			continue
		}
		quoted = append(quoted, arg)
	}
	return strings.Join(quoted, " ")
}
