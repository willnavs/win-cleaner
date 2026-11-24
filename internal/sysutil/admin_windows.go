package sysutil

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// IsAdmin verifica se o processo atual possui privilégios de Administrador no Windows.
// Retorna true se estiver rodando como Administrador, false caso contrário.
func IsAdmin() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	type tokenElevation struct {
		TokenIsElevated uint32
	}
	var elevation tokenElevation
	var returned uint32
	err = windows.GetTokenInformation(token, windows.TokenElevation, (*byte)(unsafe.Pointer(&elevation)), uint32(unsafe.Sizeof(elevation)), &returned)
	if err != nil {
		return false
	}

	return elevation.TokenIsElevated != 0
}