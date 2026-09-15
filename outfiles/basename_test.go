package outfiles

import (
	"errors"
	"testing"
)

func TestValidateBasename(t *testing.T) {
	tests := map[string]struct {
		Basename    string
		ExpectedErr error
	}{
		"an ordinary name passes":                           {Basename: "seq"},
		"a name with spaces also passes":                    {Basename: "my seq"},
		"a name with a dot also passes":                     {Basename: "seq.v2"},
		"a name pointing to a parent directory is rejected": {Basename: "../escaped", ExpectedErr: ErrBasenameHasPathSeparator},
		"a name with a slash is rejected":                   {Basename: "sub/seq", ExpectedErr: ErrBasenameHasPathSeparator},
		"a name with a backslash is rejected":               {Basename: `sub\seq`, ExpectedErr: ErrBasenameHasPathSeparator},
		"an empty name is rejected":                         {Basename: "", ExpectedErr: ErrBasenameEmpty},
		".. alone, with no path separator, passes":          {Basename: ".."},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := ValidateBasename(test.Basename); !errors.Is(err, test.ExpectedErr) {
				t.Errorf("ValidateBasename(%q) returned error %v, want %v", test.Basename, err, test.ExpectedErr)
			}
		})
	}
}

func TestResolveBasename(t *testing.T) {
	tests := map[string]struct {
		FlagValue   string
		FilePath    string
		Expected    string
		ExpectedErr error
	}{
		"the flag takes priority":                                    {FlagValue: "flow", FilePath: "docs/seq.puml", Expected: "flow"},
		"an empty flag falls back to the input path":                 {FlagValue: "", FilePath: "docs/seq.puml", Expected: "seq"},
		"cannot be determined when reading stdin with an empty flag": {FlagValue: "", FilePath: "", ExpectedErr: ErrBasenameEmpty},
		"the flag still wins even when reading stdin":                {FlagValue: "flow", FilePath: "", Expected: "flow"},
		"a flag with a path separator is rejected":                   {FlagValue: "../escaped", FilePath: "docs/seq.puml", ExpectedErr: ErrBasenameHasPathSeparator},
		"a name derived from the input path also passes validation":  {FlagValue: "", FilePath: "docs/my seq.puml", Expected: "my seq"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ResolveBasename(test.FlagValue, test.FilePath)
			if !errors.Is(err, test.ExpectedErr) {
				t.Fatalf("ResolveBasename(%q, %q) returned error %v, want %v", test.FlagValue, test.FilePath, err, test.ExpectedErr)
			}
			if test.ExpectedErr != nil {
				return
			}
			if got != test.Expected {
				t.Errorf("ResolveBasename(%q, %q) = %q, want %q", test.FlagValue, test.FilePath, got, test.Expected)
			}
		})
	}
}
