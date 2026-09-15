package tools

import (
	"flag"
	"io"
)

var HelpOption CommonOptions = CommonOptions{Help: true}

type CommonRawOptions struct {
	Help         bool
	VersionShort bool
	VersionLong  bool
}

func DeclareCommonFlags(flags *flag.FlagSet, commonRawOptions *CommonRawOptions) {
	flags.BoolVar(&commonRawOptions.VersionShort, "v", false, "print version and exit")
	flags.BoolVar(&commonRawOptions.VersionLong, "version", false, "print version and exit")
}

type CommonOptions struct {
	Help    bool
	Version bool
}

func ValidateCommonOptions(commonRawOptions *CommonRawOptions) (CommonOptions, error) {
	if commonRawOptions.Help {
		return HelpOption, nil
	}
	if commonRawOptions.VersionShort || commonRawOptions.VersionLong {
		return CommonOptions{Version: true}, nil
	}
	return CommonOptions{}, nil
}

type NamedInput struct {
	Name   string
	Reader io.ReadCloser
}
