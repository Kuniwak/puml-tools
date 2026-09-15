package pumlpng

import (
	"bytes"
	"compress/zlib"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func makePNGWith(extraChunks ...[]byte) []byte {
	ihdrData := []byte{0, 0, 0, 1, 0, 0, 0, 1, 8, 0, 0, 0, 0}
	var imgBuf bytes.Buffer
	w := zlib.NewWriter(&imgBuf)
	w.Write([]byte{0, 0})
	w.Close()

	var buf bytes.Buffer
	buf.Write(pngSignature)
	buf.Write(pngChunk("IHDR", ihdrData))
	for _, c := range extraChunks {
		buf.Write(c)
	}
	buf.Write(pngChunk("IDAT", imgBuf.Bytes()))
	buf.Write(pngChunk("IEND", nil))
	return buf.Bytes()
}

func makeTEXtChunk(keyword, text string) []byte {
	data := append([]byte(keyword+"\x00"), []byte(text)...)
	return pngChunk("tEXt", data)
}

func makeZTXtChunk(keyword, text string) []byte {
	var compressed bytes.Buffer
	w := zlib.NewWriter(&compressed)
	w.Write([]byte(text))
	w.Close()
	data := append([]byte(keyword+"\x00\x00"), compressed.Bytes()...)
	return pngChunk("zTXt", data)
}

func makeITXtChunk(keyword, text string, compress bool) []byte {
	var textBytes []byte
	compressionFlag := byte(0)
	if compress {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		w.Write([]byte(text))
		w.Close()
		textBytes = buf.Bytes()
		compressionFlag = 1
	} else {
		textBytes = []byte(text)
	}
	data := append([]byte(keyword+"\x00"), compressionFlag, 0)
	data = append(data, 0)
	data = append(data, 0)
	data = append(data, textBytes...)
	return pngChunk("iTXt", data)
}

func TestIsPNG(t *testing.T) {
	tests := map[string]struct {
		Input    []byte
		Expected bool
	}{
		"PNG": {
			Input:    NewPNG(),
			Expected: true,
		},
		"plain text": {
			Input:    []byte("@startuml\n@enduml\n"),
			Expected: false,
		},
		"empty": {
			Input:    []byte{},
			Expected: false,
		},
		"too short": {
			Input:    []byte{0x89, 0x50},
			Expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsPNG(test.Input)
			if got != test.Expected {
				t.Errorf("IsPNG mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestExtract(t *testing.T) {
	tests := map[string]struct {
		Input         []byte
		Expected      []byte
		ExpectedError error
	}{
		"plantuml tEXt chunk": {
			Input:    NewPUMLPNG("@startuml\n@enduml\n"),
			Expected: []byte("@startuml\n@enduml\n"),
		},
		"Plantuml tEXt chunk (capital P)": {
			Input:    makePNGWith(makeTEXtChunk("Plantuml", "@startuml\n@enduml\n")),
			Expected: []byte("@startuml\n@enduml\n"),
		},
		"Plantuml zTXt chunk": {
			Input:    makePNGWith(makeZTXtChunk("Plantuml", "@startuml\n@enduml\n")),
			Expected: []byte("@startuml\n@enduml\n"),
		},
		"Plantuml iTXt chunk uncompressed": {
			Input:    makePNGWith(makeITXtChunk("Plantuml", "@startuml\n@enduml\n", false)),
			Expected: []byte("@startuml\n@enduml\n"),
		},
		"Plantuml iTXt chunk compressed": {
			Input:    makePNGWith(makeITXtChunk("Plantuml", "@startuml\n@enduml\n", true)),
			Expected: []byte("@startuml\n@enduml\n"),
		},
		"no plantuml chunk": {
			Input:         NewPNG(),
			ExpectedError: ErrNoMetadata,
		},
		"not PNG": {
			Input:         []byte("@startuml\n@enduml\n"),
			ExpectedError: ErrNotPNG,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Extract(test.Input)
			if test.ExpectedError != nil {
				if !errors.Is(err, test.ExpectedError) {
					t.Errorf("Extract error = %v, want %v", err, test.ExpectedError)
				}
				return
			}
			if err != nil {
				t.Fatalf("Extract returned unexpected error: %v", err)
			}
			if !bytes.Equal(test.Expected, got) {
				t.Errorf("Extract mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}

func TestSource(t *testing.T) {
	tests := map[string]struct {
		Input         []byte
		Expected      []byte
		ExpectedError error
	}{
		"a PNG returns the PlantUML from its metadata": {
			Input:    NewPUMLPNG("@startuml\n@enduml\n"),
			Expected: []byte("@startuml\n@enduml\n"),
		},
		"a non-PNG is returned as-is": {
			Input:    []byte("@startuml\n@enduml\n"),
			Expected: []byte("@startuml\n@enduml\n"),
		},
		"a PNG without PlantUML metadata is an error": {
			Input:         NewPNG(),
			ExpectedError: ErrNoMetadata,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := Source(test.Input)
			if test.ExpectedError != nil {
				if !errors.Is(err, test.ExpectedError) {
					t.Errorf("Source error = %v, want %v", err, test.ExpectedError)
				}
				return
			}
			if err != nil {
				t.Fatalf("Source returned unexpected error: %v", err)
			}
			if !bytes.Equal(test.Expected, got) {
				t.Errorf("Source mismatch (-want +got):\n%s", cmp.Diff(test.Expected, got))
			}
		})
	}
}
