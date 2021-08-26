package internal

import (
	"os"
	"os/exec"

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

func Tidy() error {
	cmd := exec.Command("go", "mod", "tidy")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
