package pumlpng

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"strings"
)

var ErrNoMetadata = errors.New("no plantuml metadata in PNG")
var ErrNotPNG = errors.New("not a PNG file")

var pngSignature = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

func IsPNG(b []byte) bool {
	return len(b) >= len(pngSignature) && bytes.Equal(b[:len(pngSignature)], pngSignature)
}

func Source(b []byte) ([]byte, error) {
	if !IsPNG(b) {
		return b, nil
	}
	return Extract(b)
}

func Extract(b []byte) ([]byte, error) {
	if !IsPNG(b) {
		return nil, ErrNotPNG
	}

	r := bytes.NewReader(b[len(pngSignature):])
	for {
		var length uint32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		chunkType := make([]byte, 4)
		if _, err := io.ReadFull(r, chunkType); err != nil {
			return nil, err
		}

		data := make([]byte, length)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}

		if _, err := r.Seek(4, io.SeekCurrent); err != nil {
			return nil, err
		}

		switch string(chunkType) {
		case "tEXt":
			if src, ok := extractPUMLFromTEXt(data); ok {
				return src, nil
			}
		case "zTXt":
			if src, ok, err := extractPUMLFromZTXt(data); ok {
				return src, err
			} else if err != nil {
				return nil, err
			}
		case "iTXt":
			if src, ok, err := extractPUMLFromITXt(data); ok {
				return src, err
			} else if err != nil {
				return nil, err
			}
		case "IEND":
			return nil, ErrNoMetadata
		}
	}

	return nil, ErrNoMetadata
}

func NewPNG() []byte {
	return assemblePNG(nil)
}

func NewPUMLPNG(pumlSource string) []byte {
	textData := append([]byte("Plantuml\x00"), []byte(pumlSource)...)
	return assemblePNG(textData)
}

func assemblePNG(pumlTEXtData []byte) []byte {
	ihdrData := []byte{0, 0, 0, 1, 0, 0, 0, 1, 8, 0, 0, 0, 0}
	var imgBuf bytes.Buffer
	w := zlib.NewWriter(&imgBuf)
	w.Write([]byte{0, 0})
	w.Close()

	var buf bytes.Buffer
	buf.Write(pngSignature)
	buf.Write(pngChunk("IHDR", ihdrData))
	if pumlTEXtData != nil {
		buf.Write(pngChunk("tEXt", pumlTEXtData))
	}
	buf.Write(pngChunk("IDAT", imgBuf.Bytes()))
	buf.Write(pngChunk("IEND", nil))
	return buf.Bytes()
}

func pngChunk(chunkType string, data []byte) []byte {
	typeBytes := []byte(chunkType)
	crc := crc32.ChecksumIEEE(append(typeBytes, data...))

	buf := make([]byte, 4+4+len(data)+4)
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(data)))
	copy(buf[4:8], typeBytes)
	copy(buf[8:8+len(data)], data)
	binary.BigEndian.PutUint32(buf[8+len(data):], crc)
	return buf
}

func extractPUMLFromITXt(data []byte) ([]byte, bool, error) {
	sep := bytes.IndexByte(data, 0x00)
	if sep < 0 || !strings.EqualFold(string(data[:sep]), "plantuml") {
		return nil, false, nil
	}
	rest := data[sep+1:]
	if len(rest) < 2 {
		return nil, false, nil
	}
	compressionFlag := rest[0]
	rest = rest[2:]
	for i := 0; i < 2; i++ {
		idx := bytes.IndexByte(rest, 0x00)
		if idx < 0 {
			return nil, false, nil
		}
		rest = rest[idx+1:]
	}
	if compressionFlag == 0 {
		return rest, true, nil
	}
	r, err := zlib.NewReader(bytes.NewReader(rest))
	if err != nil {
		return nil, false, err
	}
	defer r.Close()
	text, err := io.ReadAll(r)
	return text, true, err
}

func extractPUMLFromZTXt(data []byte) ([]byte, bool, error) {
	sep := bytes.IndexByte(data, 0x00)
	if sep < 0 || !strings.EqualFold(string(data[:sep]), "plantuml") {
		return nil, false, nil
	}
	if len(data) < sep+2 || data[sep+1] != 0 {
		return nil, false, nil
	}
	r, err := zlib.NewReader(bytes.NewReader(data[sep+2:]))
	if err != nil {
		return nil, false, err
	}
	defer r.Close()
	text, err := io.ReadAll(r)
	return text, true, err
}

func extractPUMLFromTEXt(data []byte) ([]byte, bool) {
	sep := bytes.IndexByte(data, 0x00)
	if sep < 0 {
		return nil, false
	}
	keyword := string(data[:sep])
	if !strings.EqualFold(keyword, "plantuml") {
		return nil, false
	}
	return data[sep+1:], true
}
