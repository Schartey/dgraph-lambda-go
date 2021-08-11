package internal

import (
	"os"

	"golang.org/x/mod/modfile"
)

func GetModuleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	file, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", err
	}

	return file.Module.Mod.Path, nil
}
