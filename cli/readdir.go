package cli

import (
	"os"
	"path/filepath"
)

func ReadDirContents(dir string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var out map[string]string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		if out == nil {
			out = make(map[string]string, len(entries))
		}
		out[e.Name()] = string(b)
	}
	return out, nil
}
