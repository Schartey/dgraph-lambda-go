package internal

import (
	"os"
	"path/filepath"
	"strings"
)

func CreateFile(p string) (*os.File, error) {
	path := p
	file := ""
	if strings.Contains(filepath.Base(p), ".") {
		path = filepath.Dir(p)
		file = filepath.Base(p)
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		print(err.Error())
		return nil, err
	}
	if file == "" {
		return nil, nil
	}
	return os.Create(p)
}
