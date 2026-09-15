package c4ast

import (
	"fmt"
	"strings"
)

func Parse(src string) (*Diagram, error) {
	return NewParser(src).Parse()
}

type Parser struct {
	input         string
	pos           int
	line          int
	col           int
	boundaryStack []ElementID
}

func NewParser(input string) *Parser {
	return &Parser{input: input, pos: 0, line: 1, col: 1}
}

var boundaryMacros = map[string]bool{
	"Enterprise_Boundary": true,
	"System_Boundary":     true,
	"Container_Boundary":  true,
	"Boundary":            true,
}

var elementBaseKinds = map[string]bool{
	"Person": true, "System": true, "SystemDb": true, "SystemQueue": true,
	"Container": true, "ContainerDb": true, "ContainerQueue": true,
	"Component": true, "ComponentDb": true, "ComponentQueue": true,
}

func (p *Parser) Parse() (*Diagram, error) {
	d := &Diagram{
		Elements:   make(map[ElementID]Element),
		Boundaries: make(map[ElementID]Boundary),
	}

	if !p.expectString("@startuml") {
		return nil, fmt.Errorf("c4ast.Parser.Parse: expected @startuml at line %d, col %d", p.line, p.col)
	}
	p.skipInlineSpace()
	if name := p.restOfLine(); name != "" {
		d.Name = name
	}
	p.skipLine()

	for !p.isAtEnd() {
		p.skipTrivia()
		if p.isAtEnd() || p.peekString("@enduml") {
			break
		}
		if p.peekKeyword("note") {
			p.skipNoteBlock()
			continue
		}
		if p.peek() == '}' {
			if n := len(p.boundaryStack); n > 0 {
				p.boundaryStack = p.boundaryStack[:n-1]
			}
			p.skipLine()
			continue
		}
		if isIDChar(p.peek()) {
			if name, args, ok := p.parseMacroCall(); ok {
				p.classifyMacro(d, name, args)
				p.skipLine()
				continue
			}
		}
		p.skipLine()
	}

	if !p.expectString("@enduml") {
		return nil, fmt.Errorf("c4ast.Parser.Parse: expected @enduml at line %d, col %d", p.line, p.col)
	}
	return d, nil
}

func (p *Parser) classifyMacro(d *Diagram, name string, args []string) {
	if boundaryMacros[name] && len(args) >= 2 {
		id := ElementID(unquote(args[0]))
		d.addBoundary(Boundary{
			ID:     id,
			Kind:   name,
			Label:  unquote(args[1]),
			Parent: p.currentBoundary(),
		})
		p.boundaryStack = append(p.boundaryStack, id)
		return
	}

	base, external := name, false
	if strings.HasSuffix(name, "_Ext") {
		base, external = strings.TrimSuffix(name, "_Ext"), true
	}
	if elementBaseKinds[base] && len(args) >= 2 {
		e := Element{
			ID:       ElementID(unquote(args[0])),
			Kind:     base,
			External: external,
			Label:    unquote(args[1]),
			Boundary: p.currentBoundary(),
		}
		if len(args) >= 3 {
			e.Descr = unquote(args[2])
		}
		d.addElement(e)
		return
	}

	if isRelMacro(name) && len(args) >= 2 {
		from := ElementID(unquote(args[0]))
		to := ElementID(unquote(args[1]))
		var label string
		if len(args) >= 3 {
			label = unquote(args[2])
		}
		switch {
		case name == "Rel_Back":
			d.Rels = append(d.Rels, Rel{From: to, To: from, Label: label})
		case name == "BiRel" || strings.HasPrefix(name, "BiRel_"):
			d.Rels = append(d.Rels, Rel{From: from, To: to, Label: label})
			d.Rels = append(d.Rels, Rel{From: to, To: from, Label: label})
		default:
			d.Rels = append(d.Rels, Rel{From: from, To: to, Label: label})
		}
	}
}

func isRelMacro(name string) bool {
	return name == "Rel" || strings.HasPrefix(name, "Rel_") ||
		name == "BiRel" || strings.HasPrefix(name, "BiRel_")
}

func (d *Diagram) addElement(e Element) {
	if _, ok := d.Elements[e.ID]; ok {
		return
	}
	d.Elements[e.ID] = e
	d.Order = append(d.Order, e.ID)
}

func (d *Diagram) addBoundary(b Boundary) {
	if _, ok := d.Boundaries[b.ID]; ok {
		return
	}
	d.Boundaries[b.ID] = b
}

func (p *Parser) currentBoundary() ElementID {
	if n := len(p.boundaryStack); n > 0 {
		return p.boundaryStack[n-1]
	}
	return ""
}

func (p *Parser) parseMacroCall() (string, []string, bool) {
	name := p.parseID()
	if name == "" || p.peek() != '(' {
		return name, nil, false
	}
	p.advance()

	var inner strings.Builder
	depth := 1
	inString := false
	escaped := false
	for !p.isAtEnd() && depth > 0 {
		c := p.peek()
		switch {
		case escaped:
			escaped = false
		case inString && c == '\\':
			escaped = true
		case c == '"':
			inString = !inString
		case !inString && c == '(':
			depth++
		case !inString && c == ')':
			depth--
			if depth == 0 {
				p.advance()
				return name, splitArgs(inner.String()), true
			}
		}
		inner.WriteByte(c)
		p.advance()
	}
	return name, splitArgs(inner.String()), true
}

func (p *Parser) parseID() string {
	var sb strings.Builder
	for !p.isAtEnd() && isIDChar(p.peek()) {
		sb.WriteByte(p.peek())
		p.advance()
	}
	return sb.String()
}

func (p *Parser) restOfLine() string {
	end := p.pos
	for end < len(p.input) && p.input[end] != '\n' {
		end++
	}
	return strings.TrimSpace(p.input[p.pos:end])
}

func (p *Parser) skipTrivia() {
	for {
		for !p.isAtEnd() && (p.peek() == ' ' || p.peek() == '\t' || p.peek() == '\n' || p.peek() == '\r') {
			p.advance()
		}
		if p.peekString("/'") {
			p.skipBlockComment()
			continue
		}
		if p.peek() == '\'' {
			p.skipLine()
			continue
		}
		return
	}
}

func (p *Parser) skipBlockComment() {
	p.expectString("/'")
	for !p.isAtEnd() && !p.peekString("'/") {
		p.advance()
	}
	p.expectString("'/")
}

func (p *Parser) skipNoteBlock() {
	p.skipLine()
	for !p.isAtEnd() {
		switch p.restOfLine() {
		case "end note":
			p.skipLine()
			return
		case "@enduml":
			return
		}
		p.skipLine()
	}
}

func (p *Parser) peekKeyword(kw string) bool {
	if !p.peekString(kw) {
		return false
	}
	next := p.pos + len(kw)
	return next >= len(p.input) || !isIDChar(p.input[next])
}

func (p *Parser) skipInlineSpace() {
	for !p.isAtEnd() && (p.peek() == ' ' || p.peek() == '\t') {
		p.advance()
	}
}

func (p *Parser) skipLine() {
	for !p.isAtEnd() && p.peek() != '\n' {
		p.advance()
	}
	if !p.isAtEnd() {
		p.advance()
	}
}

func (p *Parser) peek() byte {
	if p.isAtEnd() {
		return 0
	}
	return p.input[p.pos]
}

func (p *Parser) advance() byte {
	if p.isAtEnd() {
		return 0
	}
	c := p.input[p.pos]
	p.pos++
	if c == '\n' {
		p.line++
		p.col = 1
	} else {
		p.col++
	}
	return c
}

func (p *Parser) isAtEnd() bool { return p.pos >= len(p.input) }

func (p *Parser) expectString(expected string) bool {
	if p.pos+len(expected) > len(p.input) {
		return false
	}
	if p.input[p.pos:p.pos+len(expected)] != expected {
		return false
	}
	for i := 0; i < len(expected); i++ {
		p.advance()
	}
	return true
}

func (p *Parser) peekString(expected string) bool {
	if p.pos+len(expected) > len(p.input) {
		return false
	}
	return p.input[p.pos:p.pos+len(expected)] == expected
}

func isIDChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
}

func splitArgs(s string) []string {
	var args []string
	var cur strings.Builder
	inString := false
	escaped := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case escaped:
			cur.WriteByte(c)
			escaped = false
		case inString && c == '\\':
			cur.WriteByte(c)
			escaped = true
		case c == '"':
			inString = !inString
			cur.WriteByte(c)
		case c == ',' && !inString:
			args = append(args, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	last := strings.TrimSpace(cur.String())
	if last != "" || len(args) > 0 {
		args = append(args, last)
	}
	return args
}

func unquote(arg string) string {
	arg = strings.TrimSpace(arg)
	if len(arg) >= 2 && arg[0] == '"' && arg[len(arg)-1] == '"' {
		inner := arg[1 : len(arg)-1]
		return strings.ReplaceAll(inner, `\"`, `"`)
	}
	return arg
}
