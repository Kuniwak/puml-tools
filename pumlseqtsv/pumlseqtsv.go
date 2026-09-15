package pumlseqtsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

func ParseMessageKeys(r io.Reader) (map[pumlseqast.MessageKey]bool, error) {
	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.FieldsPerRecord = -1
	keys := map[pumlseqast.MessageKey]bool{}
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return keys, nil
		}
		if err != nil {
			return nil, err
		}
		var k pumlseqast.MessageKey
		if len(rec) > 0 {
			k.Sender = rec[0]
		}
		if len(rec) > 1 {
			k.Receiver = rec[1]
		}
		if len(rec) > 2 {
			k.Label = rec[2]
		}
		if k.Sender == "" && k.Receiver == "" && k.Label == "" {
			continue
		}
		keys[k] = true
	}
}

func ParseFolding(r io.Reader) (pumlseqast.Folding, error) {
	cr := csv.NewReader(r)
	cr.Comma = '\t'
	cr.FieldsPerRecord = -1
	phi := pumlseqast.Folding{}
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			return phi, nil
		}
		if err != nil {
			return nil, err
		}
		det, abs := "", ""
		if len(rec) > 0 {
			det = rec[0]
		}
		if len(rec) > 1 {
			abs = rec[1]
		}
		if det == "" && abs == "" {
			continue
		}
		if det == "" || abs == "" {
			return nil, fmt.Errorf("invalid map row %q: need <detailID><TAB><abstractID>", strings.Join(rec, "\t"))
		}
		phi[det] = abs
	}
}
