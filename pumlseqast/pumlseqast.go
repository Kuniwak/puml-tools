package pumlseqast

import (
	"strings"
	"unicode"

	"github.com/Kuniwak/puml-tools/lex"
	"github.com/Kuniwak/puml-tools/lines"
	"github.com/Kuniwak/puml-tools/pumlseq"
)

type Node interface {
	Render() string
}

type Diagram struct {
	Nodes []Node
}

func (d *Diagram) String() string {
	var b strings.Builder
	for _, n := range d.Nodes {
		b.WriteString(n.Render())
	}
	return b.String()
}

type Start struct {
	Text string
	Name string
	Term string
}

func (n *Start) Render() string { return n.Text + n.Term }

type End struct {
	Text string
	Term string
}

func (n *End) Render() string { return n.Text + n.Term }

type Participant struct {
	Text    string
	Keyword string
	Name    string
	Term    string
}

func (n *Participant) Render() string { return n.Text + n.Term }

func (n *Participant) Label() string {
	toks := lex.TokenizeQuoted(strings.TrimSpace(n.Text))
	if len(toks) >= 2 {
		return lex.Unquote(toks[1])
	}
	return n.Name
}

type Message struct {
	Indent   string
	LeftRaw  string
	Mid      string
	RightRaw string
	Rest     string
	Term     string
	Sender   string
	Receiver string
	Line     int
}

func (n *Message) Render() string {
	return n.Indent + n.LeftRaw + n.Mid + n.RightRaw + n.Rest + n.Term
}

type Note struct {
	HeaderPrefix string
	TargetsRaw   string
	HeaderSuffix string
	HeaderTerm   string
	Targets      []string
	Body         string
	EndText      string
	EndTerm      string
	Single       bool
}

func (n *Note) Render() string {
	head := n.HeaderPrefix + n.TargetsRaw + n.HeaderSuffix + n.HeaderTerm
	if n.Single {
		return head
	}
	return head + n.Body + n.EndText + n.EndTerm
}

type Ref struct {
	HeaderPrefix string
	TargetsRaw   string
	HeaderSuffix string
	HeaderTerm   string
	Targets      []string
	Body         string
	EndText      string
	EndTerm      string
	Single       bool
	Line         int
}

func (n *Ref) Render() string {
	head := n.HeaderPrefix + n.TargetsRaw + n.HeaderSuffix + n.HeaderTerm
	if n.Single {
		return head
	}
	return head + n.Body + n.EndText + n.EndTerm
}

func (n *Ref) Label() string {
	if n.Single {
		if i := strings.Index(n.HeaderSuffix, ":"); i >= 0 {
			return strings.TrimSpace(n.HeaderSuffix[i+1:])
		}
		return ""
	}
	var out []string
	for _, l := range strings.Split(n.Body, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return strings.Join(out, "\n")
}

type Activation struct {
	Before    string
	TargetRaw string
	After     string
	Term      string
	Target    string
}

func (n *Activation) Render() string { return n.Before + n.TargetRaw + n.After + n.Term }

type Block struct {
	HeaderText string
	HeaderTerm string
	Keyword    string
	Label      string
	Body       []Node
	Elses      []ElseClause
	EndText    string
	EndTerm    string
}

func (n *Block) Render() string {
	var b strings.Builder
	b.WriteString(n.HeaderText)
	b.WriteString(n.HeaderTerm)
	for _, c := range n.Body {
		b.WriteString(c.Render())
	}
	for _, e := range n.Elses {
		b.WriteString(e.SepText)
		b.WriteString(e.SepTerm)
		for _, c := range e.Body {
			b.WriteString(c.Render())
		}
	}
	b.WriteString(n.EndText)
	b.WriteString(n.EndTerm)
	return b.String()
}

type ElseClause struct {
	SepText string
	SepTerm string
	Label   string
	Body    []Node
}

type Delay struct {
	Text string
	Term string
}

func (n *Delay) Render() string { return n.Text + n.Term }

type Raw struct {
	Text string
}

func (n *Raw) Render() string { return n.Text }

func Parse(src string) (*Diagram, error) {
	p := &parser{pl: splitPhysLines(src)}
	return &Diagram{Nodes: p.parseSequence(false)}, nil
}

type physLine struct {
	content string
	term    string
}

func splitPhysLines(src string) []physLine {
	ls, nl := lines.Split(src)
	out := make([]physLine, len(ls))
	for i, c := range ls {
		term := "\n"
		if i == len(ls)-1 && !nl {
			term = ""
		}
		out[i] = physLine{content: c, term: term}
	}
	return out
}

type parser struct {
	pl  []physLine
	pos int
}

func (p *parser) eof() bool { return p.pos >= len(p.pl) }

func (p *parser) parseSequence(insideBlock bool) []Node {
	var nodes []Node
	var rawBuf strings.Builder
	flushRaw := func() {
		if rawBuf.Len() > 0 {
			nodes = append(nodes, &Raw{Text: rawBuf.String()})
			rawBuf.Reset()
		}
	}
	for !p.eof() {
		line := p.pl[p.pos]
		c := strings.TrimSpace(line.content)
		if insideBlock {
			if pumlseq.IsBlockEnd(c) {
				flushRaw()
				return nodes
			}
			if _, ok := pumlseq.ElseLabel(c); ok {
				flushRaw()
				return nodes
			}
		}
		switch {
		case strings.HasPrefix(c, "/'"):
			flushRaw()
			nodes = append(nodes, p.parseBlockComment())
		case strings.HasPrefix(c, "'"):
			rawBuf.WriteString(line.content + line.term)
			p.pos++
		case strings.HasPrefix(c, "..."):
			flushRaw()
			nodes = append(nodes, &Delay{Text: line.content, Term: line.term})
			p.pos++
		case strings.HasPrefix(c, "@startuml"):
			flushRaw()
			nodes = append(nodes, p.parseStart(line))
			p.pos++
		case strings.HasPrefix(c, "@enduml"):
			flushRaw()
			nodes = append(nodes, &End{Text: line.content, Term: line.term})
			p.pos++
		case pumlseq.IsNoteOpener(c):
			flushRaw()
			nodes = append(nodes, p.parseNote())
		case pumlseq.IsRefOpener(c):
			flushRaw()
			nodes = append(nodes, p.parseRef())
		case pumlseq.IsBlockStart(c):
			flushRaw()
			nodes = append(nodes, p.parseBlock())
		case isDeclaration(c):
			flushRaw()
			nodes = append(nodes, p.parseParticipant(line))
			p.pos++
		case isActivation(c):
			flushRaw()
			nodes = append(nodes, p.parseActivation(line))
			p.pos++
		case isMessage(line.content):
			flushRaw()
			nodes = append(nodes, p.parseMessage(line))
			p.pos++
		default:
			rawBuf.WriteString(line.content + line.term)
			p.pos++
		}
	}
	flushRaw()
	return nodes
}

func (p *parser) parseBlockComment() *Raw {
	var b strings.Builder
	opener := p.pl[p.pos]
	b.WriteString(opener.content + opener.term)
	p.pos++
	if rest := strings.TrimSpace(opener.content)[len("/'"):]; strings.Contains(rest, "'/") {
		return &Raw{Text: b.String()}
	}
	for !p.eof() {
		l := p.pl[p.pos]
		b.WriteString(l.content + l.term)
		p.pos++
		if strings.Contains(l.content, "'/") {
			break
		}
	}
	return &Raw{Text: b.String()}
}

func (p *parser) parseStart(line physLine) *Start {
	name := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line.content), "@startuml"))
	return &Start{Text: line.content, Name: name, Term: line.term}
}

func (p *parser) parseParticipant(line physLine) *Participant {
	c := strings.TrimSpace(line.content)
	name, _ := pumlseq.DeclaredParticipant(c)
	keyword := ""
	if fields := strings.Fields(c); len(fields) > 0 {
		keyword = fields[0]
	}
	return &Participant{Text: line.content, Keyword: keyword, Name: name, Term: line.term}
}

func (p *parser) parseActivation(line physLine) *Activation {
	toks := tokenizeSpans(line.content)
	if len(toks) < 2 {
		return &Activation{Before: line.content, Term: line.term}
	}
	target := activationTargetName(line.content, toks)

	ref := toks[1]
	for _, t := range toks[1:] {
		if lex.Unquote(t.text) == target {
			ref = t
		}
	}
	return &Activation{
		Before:    line.content[:ref.start],
		TargetRaw: line.content[ref.start:ref.end],
		After:     line.content[ref.end:],
		Term:      line.term,
		Target:    target,
	}
}

func (p *parser) parseMessage(line physLine) *Message {
	parts, _ := pumlseq.ParseMessageParts(line.content)
	return &Message{
		Indent:   parts.Indent,
		LeftRaw:  parts.LeftRaw,
		Mid:      parts.Mid,
		RightRaw: parts.RightRaw,
		Rest:     parts.Rest,
		Term:     line.term,
		Sender:   parts.Sender,
		Receiver: parts.Receiver,
		Line:     p.pos + 1,
	}
}

func (p *parser) parseNote() *Note {
	opener := p.pl[p.pos]
	p.pos++
	prefix, targetsRaw, suffix := splitNoteHeader(opener.content)
	n := &Note{
		HeaderPrefix: prefix,
		TargetsRaw:   targetsRaw,
		HeaderSuffix: suffix,
		HeaderTerm:   opener.term,
		Targets:      pumlseq.NoteTargets(strings.TrimRight(opener.content, "\r")),
	}
	if lex.IndexOutsideQuotes(opener.content, ':') >= 0 {
		n.Single = true
		return n
	}
	var body strings.Builder
	for !p.eof() {
		l := p.pl[p.pos]
		if pumlseq.IsNoteTerminator(l.content) {
			n.EndText = l.content
			n.EndTerm = l.term
			p.pos++
			break
		}
		body.WriteString(l.content + l.term)
		p.pos++
	}
	n.Body = body.String()
	return n
}

func (p *parser) parseRef() *Ref {
	opener := p.pl[p.pos]
	line := p.pos + 1
	p.pos++
	prefix, targetsRaw, suffix := splitNoteHeader(opener.content)
	n := &Ref{
		HeaderPrefix: prefix,
		TargetsRaw:   targetsRaw,
		HeaderSuffix: suffix,
		HeaderTerm:   opener.term,
		Targets:      pumlseq.NoteTargets(strings.TrimRight(opener.content, "\r")),
		Line:         line,
	}
	if lex.IndexOutsideQuotes(opener.content, ':') >= 0 {
		n.Single = true
		return n
	}
	var body strings.Builder
	for !p.eof() {
		l := p.pl[p.pos]
		if pumlseq.IsRefTerminator(l.content) {
			n.EndText = l.content
			n.EndTerm = l.term
			p.pos++
			break
		}
		body.WriteString(l.content + l.term)
		p.pos++
	}
	n.Body = body.String()
	return n
}

func (p *parser) parseBlock() *Block {
	opener := p.pl[p.pos]
	keyword, label, _ := pumlseq.GroupStartLabel(strings.TrimSpace(opener.content))
	p.pos++
	b := &Block{HeaderText: opener.content, HeaderTerm: opener.term, Keyword: keyword, Label: label}
	b.Body = p.parseSequence(true)
	for !p.eof() {
		if label2, ok := pumlseq.ElseLabel(strings.TrimSpace(p.pl[p.pos].content)); ok {
			sep := p.pl[p.pos]
			p.pos++
			b.Elses = append(b.Elses, ElseClause{
				SepText: sep.content,
				SepTerm: sep.term,
				Label:   label2,
				Body:    p.parseSequence(true),
			})
			continue
		}
		break
	}
	if !p.eof() && pumlseq.IsBlockEnd(strings.TrimSpace(p.pl[p.pos].content)) {
		end := p.pl[p.pos]
		b.EndText = end.content
		b.EndTerm = end.term
		p.pos++
	}
	return b
}

func splitNoteHeader(header string) (prefix, targetsRaw, suffix string) {
	region := header
	if i := lex.IndexOutsideQuotes(header, ':'); i >= 0 {
		region = header[:i]
		suffix = header[i:]
	}
	for _, sep := range []string{" over ", " of "} {
		idx := strings.Index(region, sep)
		if idx < 0 {
			continue
		}
		at := idx + len(sep)
		targets := strings.TrimRight(region[at:], " \t\r")
		return header[:at], targets, region[at+len(targets):] + suffix
	}
	return header, "", ""
}

var activationKeywords = map[string]bool{
	"activate":   true,
	"deactivate": true,
	"destroy":    true,
	"create":     true,
}

func isActivation(trimmed string) bool {
	fields := strings.Fields(trimmed)
	return len(fields) > 0 && activationKeywords[fields[0]]
}

func isDeclaration(trimmed string) bool {
	_, ok := pumlseq.DeclaredParticipant(trimmed)
	return ok
}

func isMessage(content string) bool {
	_, ok := pumlseq.ParseMessageParts(content)
	return ok
}

func activationTargetName(content string, toks []spanTok) string {
	rest := content[toks[1].start:]
	if name, ok := pumlseq.DeclaredParticipant(rest); ok {
		return name
	}
	return lex.Unquote(toks[1].text)
}

type spanTok struct {
	start, end int
	text       string
}

func tokenizeSpans(s string) []spanTok {
	var toks []spanTok
	inQuote := false
	start := -1
	flush := func(end int) {
		if start >= 0 {
			toks = append(toks, spanTok{start: start, end: end, text: s[start:end]})
			start = -1
		}
	}
	for i, r := range s {
		switch {
		case r == '"':
			if start < 0 {
				start = i
			}
			if inQuote {
				inQuote = false
				flush(i + 1)
			} else {
				inQuote = true
			}
		case !inQuote && unicode.IsSpace(r):
			flush(i)
		default:
			if start < 0 {
				start = i
			}
		}
	}
	flush(len(s))
	return toks
}
