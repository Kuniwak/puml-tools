package outfiles

import (
	"errors"
	"strings"

	"github.com/Kuniwak/puml-tools/tools"
)

var ErrBasenameEmpty = errors.New("basename must not be empty")

var ErrBasenameHasPathSeparator = errors.New("basename must not contain a path separator")

func ValidateBasename(basename string) error {
	if basename == "" {
		return ErrBasenameEmpty
	}
	if strings.ContainsAny(basename, `/\`) {
		return ErrBasenameHasPathSeparator
	}
	return nil
}

func ResolveBasename(flagValue, filePath string) (string, error) {
	basename := flagValue
	if basename == "" && filePath != "" {
		basename = tools.Basename(filePath)
	}
	if err := ValidateBasename(basename); err != nil {
		return "", err
	}
	return basename, nil
}
