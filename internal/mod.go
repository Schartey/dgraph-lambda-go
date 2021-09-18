package internal

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func findGoModFile(root string, maxDepth int) (string, error) {
	if root == "/" || maxDepth == 0 {
		return "", errors.New("could not find go.mod file")
	}
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if filepath.Base(file) == "go.mod" {
			return root, nil
		}
	}
	return findGoModFile(filepath.Dir(root), maxDepth-1)
}

func GetModuleName() (string, error) {

	root, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}

	root, err = findGoModFile(root, 3)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
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
