package outfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type File struct {
	Name    string
	Content string
}

func WriteFiles(dir string, files []File) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = filepath.Join(dir, f.Name)
		if info, err := os.Stat(paths[i]); err == nil && info.IsDir() {
			return nil, fmt.Errorf("cannot write %s: it is a directory", paths[i])
		}
	}

	var temps []string
	removeTemps := func() {
		for _, tmp := range temps {
			os.Remove(tmp)
		}
	}

	for _, f := range files {
		tmp, err := os.CreateTemp(dir, ".outfiles-*")
		if err != nil {
			removeTemps()
			return nil, err
		}
		temps = append(temps, tmp.Name())

		_, writeErr := tmp.WriteString(f.Content)

		if err := errors.Join(writeErr, tmp.Close(), os.Chmod(tmp.Name(), 0o644)); err != nil {
			removeTemps()
			return nil, err
		}
	}

	for i, tmp := range temps {
		if err := os.Rename(tmp, paths[i]); err != nil {
			temps = temps[i:]
			removeTemps()
			return nil, err
		}
	}
	return paths, nil
}
