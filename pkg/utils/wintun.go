package utils

import (
	_ "embed"
	"os"
)

//go:embed wintun.dll
var wintunDll []byte

func ExtractWintun() error {
	return os.WriteFile("wintun.dll", wintunDll, 0666)
}
