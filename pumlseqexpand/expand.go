package pumlseqexpand

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Kuniwak/puml-tools/lines"
	"github.com/Kuniwak/puml-tools/pumlseq"
	"github.com/Kuniwak/puml-tools/pumlseqast"
	"github.com/Kuniwak/puml-tools/sets"
)

func Expand(name string, src []byte, base string, read func(path string) ([]byte, error)) (*pumlseqast.Diagram, []StripWarning, error) {
	if read == nil {
		panic("pumlseqexpand: Expand requires a non-nil read function")
	}

	d, err := pumlseqast.Parse(string(src))
	if err != nil {
		return nil, nil, err
	}

	e := &expander{read: read, decls: &declSet{byName: map[string]declaration{}, boxes: map[string]string{}}}

	if err := e.decls.gather(d.Nodes, true, boxTracker{}); err != nil {
		return nil, nil, err
	}

	parts := make([]string, 0, len(d.Nodes)+1)
	insertAfter := -1
	parentBoxes := map[string]position{}
	var box boxTracker
	for _, n := range d.Nodes {
		text, err := e.renderNode(n, walk{src: source{name: name, dir: base}, keepDecls: true})
		if err != nil {
			return nil, nil, err
		}
		parts = append(parts, text)
		switch v := n.(type) {
		case *pumlseqast.Participant:

			if box.depth == 0 {
				insertAfter = len(parts) - 1
			}
		case *pumlseqast.Start:
			if insertAfter < 0 {
				insertAfter = len(parts) - 1
			}
		case *pumlseqast.Raw:

			offset := 0
			for _, l := range strings.SplitAfter(v.Text, "\n") {
				if box.depth == 1 && pumlseq.IsBoxEnd(l) {
					if _, ok := parentBoxes[box.key]; !ok {
						parentBoxes[box.key] = position{part: len(parts) - 1, offset: offset}
					}
				}
				box.feed(l)
				offset += len(l)
			}
		}
	}

	parts = insertAll(parts, e.decls.hoists(), position{part: insertAfter + 1}, parentBoxes)

	out, err := pumlseqast.Parse(strings.Join(parts, ""))
	if err != nil {
		return nil, nil, err
	}
	return out, e.warnings, nil
}

type position struct {
	part   int
	offset int
}

func insertAll(parts []string, hoists []hoist, plain position, parentBoxes map[string]position) []string {
	pending := map[position]string{}
	var order []position
	add := func(at position, text string) {
		if _, ok := pending[at]; !ok {
			order = append(order, at)
		}
		pending[at] += text
	}
	for _, h := range hoists {
		at, ok := parentBoxes[h.box]
		switch {
		case h.box == "":
			add(plain, h.text)
		case ok:
			add(at, h.text)
		default:

			add(plain, h.header+h.text+"end box"+lineTerm(h.header))
		}
	}

	slices.SortFunc(order, func(a, b position) int {
		if a.part != b.part {
			return b.part - a.part
		}
		return b.offset - a.offset
	})
	for _, at := range order {
		if at.part >= len(parts) {
			parts = append(parts, pending[at])
			continue
		}
		p := parts[at.part]
		parts[at.part] = p[:at.offset] + pending[at] + p[at.offset:]
	}
	return parts
}

func lineTerm(line string) string {
	if strings.HasSuffix(line, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

type boxTracker struct {
	depth  int
	key    string
	header string
}

func (t *boxTracker) feed(line string) {
	switch {
	case pumlseq.IsBoxEnd(line):
		t.depth = max(0, t.depth-1)
		if t.depth == 0 {
			t.key, t.header = "", ""
		}
	case pumlseq.IsBoxStart(line):
		t.depth++
		if t.depth == 1 {
			t.header = line
			if !strings.HasSuffix(t.header, "\n") {
				t.header += "\n"
			}
			label, ok := pumlseq.BoxLabel(line)
			if !ok {
				label = strings.TrimSpace(line)
			}
			t.key = label
		}
	}
}

func (t *boxTracker) feedText(text string) {
	for _, l := range strings.SplitAfter(text, "\n") {
		t.feed(l)
	}
}

type expander struct {
	read     func(path string) ([]byte, error)
	decls    *declSet
	warnings []StripWarning
}

type StripWarning struct {
	Name    string
	Line    int
	Path    string
	Keyword string
}

func (w StripWarning) Reason() string {
	return fmt.Sprintf("%s: dropped diagram-wide directive %q instead of inlining it (the parent diagram's appearance is unchanged)",
		w.Path, w.Keyword)
}

func (e *expander) warn(w StripWarning) {
	if slices.Contains(e.warnings, w) {
		return
	}
	e.warnings = append(e.warnings, w)
}

type walk struct {
	src       source
	ref       refSite
	stack     []string
	keepDecls bool
	box       boxTracker
}

type refSite struct {
	name string
	line int
}

type source struct {
	name string
	dir  string
}

func (e *expander) renderNode(n pumlseqast.Node, w walk) (string, error) {
	switch v := n.(type) {
	case *pumlseqast.Ref:
		text, err := e.expandRef(v, w)
		if err != nil {

			return "", fmt.Errorf("%s:%d: %w", w.src.name, v.Line, err)
		}
		return text, nil
	case *pumlseqast.Participant:
		if w.keepDecls {

			return v.Render(), nil
		}
		if err := e.decls.register(v, false, w.box.key, w.box.header); err != nil {
			return "", err
		}
		return "", nil
	case *pumlseqast.Raw:
		if w.keepDecls {
			return v.Render(), nil
		}
		text, stripped := stripComposable(v.Text)
		for _, keyword := range stripped {
			e.warn(StripWarning{Name: w.ref.name, Line: w.ref.line, Path: w.src.name, Keyword: keyword})
		}
		return text, nil
	case *pumlseqast.Block:
		var b strings.Builder
		b.WriteString(v.HeaderText)
		b.WriteString(v.HeaderTerm)
		body, err := e.renderNodes(v.Body, w)
		if err != nil {
			return "", err
		}
		b.WriteString(body)
		for _, c := range v.Elses {
			b.WriteString(c.SepText)
			b.WriteString(c.SepTerm)
			clause, err := e.renderNodes(c.Body, w)
			if err != nil {
				return "", err
			}
			b.WriteString(clause)
		}
		b.WriteString(v.EndText)
		b.WriteString(v.EndTerm)
		return b.String(), nil
	default:
		return n.Render(), nil
	}
}

func (e *expander) renderNodes(nodes []pumlseqast.Node, w walk) (string, error) {
	var b strings.Builder
	for _, n := range nodes {
		text, err := e.renderNode(n, w)
		if err != nil {
			return "", err
		}
		b.WriteString(text)
		if raw, ok := n.(*pumlseqast.Raw); ok && !w.keepDecls {

			w.box.feedText(raw.Text)
		}
	}
	return b.String(), nil
}

func (e *expander) expandRef(r *pumlseqast.Ref, w walk) (string, error) {
	rel, err := ParsePath(r.Label())
	if err != nil {
		return "", err
	}
	resolved := filepath.Join(w.src.dir, rel)
	if slices.Contains(w.stack, resolved) {
		cycle := append(append([]string{}, w.stack...), resolved)
		return "", fmt.Errorf("ref reference cycle detected: %s", strings.Join(cycle, " -> "))
	}

	subSrc, err := e.read(resolved)
	if err != nil {
		return "", fmt.Errorf("cannot read ref target: %s (a ref label is a relative path from the referencing file): %w", resolved, err)
	}
	sub, err := pumlseqast.Parse(string(subSrc))
	if err != nil {
		return "", err
	}
	if err := validateScope(resolved, r, sub); err != nil {
		return "", err
	}

	nodes, err := bodyNodes(resolved, sub.Nodes)
	if err != nil {
		return "", err
	}
	body, err := e.renderNodes(nodes, walk{
		src:   source{name: resolved, dir: filepath.Dir(resolved)},
		ref:   refSite{name: w.src.name, line: r.Line},
		stack: append(w.stack, resolved),
	})
	if err != nil {
		return "", err
	}
	return lines.Indent(body, lines.LeadingIndent(r.HeaderPrefix)), nil
}

func bodyNodes(path string, nodes []pumlseqast.Node) ([]pumlseqast.Node, error) {
	starts, ends := 0, 0
	from, to := -1, -1
	for i, n := range nodes {
		switch n.(type) {
		case *pumlseqast.Start:
			if starts == 0 {
				from = i + 1
			}
			starts++
		case *pumlseqast.End:
			if ends == 0 {
				to = i
			}
			ends++
		}
	}
	if starts != 1 || ends != 1 || from > to {
		return nil, fmt.Errorf("%s is not a single diagram enclosed by @startuml and @enduml (%d @startuml, %d @enduml); a ref target must be exactly one diagram",
			path, starts, ends)
	}
	return nodes[from:to], nil
}

var strippedKeywords = sets.New(
	"title",
	"caption",
	"header",
	"footer",
	"autonumber",
	"skinparam",
	"skin",
	"scale",
	"hide",
	"show",
	"newpage",
	"legend",
)

var blockDirectives = sets.New(
	"title",
	"header",
	"footer",
	"legend",
)

func stripComposable(text string) (string, []string) {
	if strings.HasPrefix(strings.TrimSpace(text), "/'") {
		return text, nil
	}

	var b strings.Builder
	var stripped []string
	closing := ""
	for _, l := range strings.SplitAfter(text, "\n") {
		if closing != "" {
			if isDirectiveEnd(l, closing) {
				closing = ""
			}
			continue
		}
		if pumlseq.IsBoxStart(l) || pumlseq.IsBoxEnd(l) {
			continue
		}
		keyword, ok := documentLevelKeyword(l)
		if !ok {
			b.WriteString(l)
			continue
		}
		if !slices.Contains(stripped, keyword) {
			stripped = append(stripped, keyword)
		}
		if blockDirectives.Contains(keyword) && len(strings.Fields(strings.TrimSpace(l))) == 1 {
			closing = keyword
		}
	}
	return b.String(), stripped
}

func isDirectiveEnd(line, keyword string) bool {
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "end " + keyword, "end" + keyword:
		return true
	}
	return false
}

func documentLevelKeyword(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "'") || strings.HasPrefix(trimmed, "/'") {
		return "", false
	}
	fields := strings.Fields(strings.ToLower(trimmed))
	if len(fields) == 0 {
		return "", false
	}
	if strippedKeywords.Contains(fields[0]) {
		return fields[0], true
	}
	return "", false
}

func validateScope(path string, r *pumlseqast.Ref, sub *pumlseqast.Diagram) error {
	allowed := sets.New(r.Targets...)
	for _, p := range pumlseqast.Participants(sub) {
		if !allowed.Contains(p) {
			return fmt.Errorf("%s uses lifeline %q, which is not covered by ref over %s; add it to the over clause or fix the ref target",
				path, p, strings.Join(r.Targets, ", "))
		}
	}
	return nil
}

type declaration struct {
	text    string
	keyword string
	label   string
	box     string
}

type declSet struct {
	byName map[string]declaration
	added  []string
	boxes  map[string]string
}

func (s *declSet) gather(nodes []pumlseqast.Node, root bool, box boxTracker) error {
	for _, n := range nodes {
		switch v := n.(type) {
		case *pumlseqast.Participant:
			if err := s.register(v, root, box.key, box.header); err != nil {
				return err
			}
		case *pumlseqast.Raw:
			box.feedText(v.Text)
		case *pumlseqast.Block:
			if err := s.gather(v.Body, root, box); err != nil {
				return err
			}
			for _, c := range v.Elses {
				if err := s.gather(c.Body, root, box); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *declSet) register(p *pumlseqast.Participant, root bool, box, boxHeader string) error {
	cur := declaration{text: p.Render(), keyword: p.Keyword, label: p.Label(), box: box}
	if prev, ok := s.byName[p.Name]; ok {
		if prev.keyword != cur.keyword || prev.label != cur.label {
			return fmt.Errorf("participant %q declared inconsistently: %s %q vs %s %q",
				p.Name, prev.keyword, prev.label, cur.keyword, cur.label)
		}

		if prev.box != cur.box {
			return fmt.Errorf("participant %q's box differs: %s vs %s; diagrams joined by ref must place it in the same box",
				p.Name, boxPhrase(prev.box), boxPhrase(cur.box))
		}
		return nil
	}
	s.byName[p.Name] = cur
	if !root {
		s.added = append(s.added, p.Name)
		if _, ok := s.boxes[box]; box != "" && !ok {
			s.boxes[box] = boxHeader
		}
	}
	return nil
}

func boxPhrase(box string) string {
	if box == "" {
		return "outside any box"
	}
	return fmt.Sprintf("box %q", box)
}

type hoist struct {
	box    string
	header string
	text   string
}

func (s *declSet) hoists() []hoist {
	var out []hoist
	index := map[string]int{}
	for _, name := range s.added {
		d := s.byName[name]
		i, ok := index[d.box]
		if !ok {
			i = len(out)
			index[d.box] = i
			out = append(out, hoist{box: d.box, header: s.boxes[d.box]})
		}
		line := strings.TrimLeft(d.text, " \t")
		if !strings.HasSuffix(line, "\n") {
			line += "\n"
		}
		out[i].text += line
	}
	return out
}
