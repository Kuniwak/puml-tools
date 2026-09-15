package pumlseqgroupsplit

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/Kuniwak/puml-tools/outfiles"
	"github.com/Kuniwak/puml-tools/pumlseqast"
)

var ErrNoGroup = errors.New("no group found")

var ErrMessageOutsideGroup = errors.New("message outside a group")

func Split(d *pumlseqast.Diagram, basename string) ([]outfiles.File, error) {
	if err := outfiles.ValidateBasename(basename); err != nil {
		return nil, err
	}

	groups := pumlseqast.OutermostGroups(d)
	if len(groups) == 0 {
		return nil, ErrNoGroup
	}

	if outside := pumlseqast.MessagesOutsideGroups(d); len(outside) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrMessageOutsideGroup, describeOutside(outside))
	}

	files := make([]outfiles.File, 0, len(groups))
	for i, b := range groups {
		cut := pumlseqast.CutBlocks(d, []*pumlseqast.Block{b}, "")
		files = append(files, outfiles.File{
			Name:    FileName(basename, i+1, len(groups), b.Label),
			Content: cut.String(),
		})
	}
	return files, nil
}

func FileName(basename string, idx, total int, title string) string {
	num := fmt.Sprintf("%0*d", len(fmt.Sprint(total)), idx)
	sanitized := SanitizeTitle(title)
	if sanitized == "" {
		return fmt.Sprintf("%s_%s.puml", basename, num)
	}
	return fmt.Sprintf("%s_%s_%s.puml", basename, num, sanitized)
}

func SanitizeTitle(title string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || strings.ContainsRune(`/\:*?"<>|`, r) {
			return '_'
		}
		return r
	}, title)
}

func describeOutside(outside []*pumlseqast.Message) string {
	first := outside[0]
	head := fmt.Sprintf("line %d: %s", first.Line, strings.TrimSpace(first.Render()))
	if len(outside) == 1 {
		return head
	}
	return fmt.Sprintf("%s (and %d more)", head, len(outside)-1)
}
