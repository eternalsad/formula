package formula

import (
	"fmt"
	"strconv"
	"strings"
)

// Token types
type TokenType int

const (
	TokenNumber TokenType = iota
	TokenVariable
	TokenOperator
	TokenParenOpen
	TokenParenClose
	TokenComma
	TokenFunction
	TokenEOF
)

// Token represents a token in the formula
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// Lexer tokenizes the input formula
type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: strings.ReplaceAll(input, " ", ""), // Remove spaces
		pos:   0,
	}
}

func (l *Lexer) NextToken() Token {
	if l.pos >= len(l.input) {
		return Token{TokenEOF, "", l.pos}
	}

	char := l.input[l.pos]

	// Numbers (including decimals)
	if isDigit(char) {
		return l.readNumber()
	}

	// Variables and functions
	if isLetter(char) {
		return l.readIdentifier()
	}

	// Single character tokens
	switch char {
	case '+', '-', '*', '/', '>', '<', '=', '!':
		return l.readOperator()
	case '(':
		l.pos++
		return Token{TokenParenOpen, "(", l.pos - 1}
	case ')':
		l.pos++
		return Token{TokenParenClose, ")", l.pos - 1}
	case ',':
		l.pos++
		return Token{TokenComma, ",", l.pos - 1}
	}

	// Skip unknown characters
	l.pos++
	return l.NextToken()
}

func (l *Lexer) readNumber() Token {
	start := l.pos
	for l.pos < len(l.input) && (isDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
		l.pos++
	}
	return Token{TokenNumber, l.input[start:l.pos], start}
}

func (l *Lexer) readIdentifier() Token {
	start := l.pos
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
	}

	value := l.input[start:l.pos]

	// Check if it's a function (followed by parenthesis)
	if l.pos < len(l.input) && l.input[l.pos] == '(' {
		return Token{TokenFunction, value, start}
	}

	return Token{TokenVariable, value, start}
}

func (l *Lexer) readOperator() Token {
	start := l.pos

	// Handle multi-character operators
	if l.pos+1 < len(l.input) {
		twoChar := l.input[l.pos : l.pos+2]
		switch twoChar {
		case ">=", "<=", "==", "!=":
			l.pos += 2
			return Token{TokenOperator, twoChar, start}
		}
	}

	l.pos++
	return Token{TokenOperator, string(l.input[start]), start}
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// Parser converts tokens to AST
type Parser struct {
	lexer   *Lexer
	current Token
}

func NewParser(input string) *Parser {
	lexer := NewLexer(input)
	p := &Parser{lexer: lexer}
	p.nextToken() // Initialize current token
	return p
}

func (p *Parser) nextToken() {
	p.current = p.lexer.NextToken()
}

func (p *Parser) Parse() (ASTNode, error) {
	return p.parseExpression()
}

// parseExpression handles the top-level expression
func (p *Parser) parseExpression() (ASTNode, error) {
	return p.parseComparison()
}

// parseComparison handles comparison operators (>, <, ==, etc.)
func (p *Parser) parseComparison() (ASTNode, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenOperator && isComparisonOp(p.current.Value) {
		op := p.current.Value
		p.nextToken()

		right, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}

		left = &ComparisonNode{
			Operator: op,
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

// parseAddSub handles + and - operators
func (p *Parser) parseAddSub() (ASTNode, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenOperator && (p.current.Value == "+" || p.current.Value == "-") {
		op := p.current.Value
		p.nextToken()

		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}

		left = &OperationNode{
			Operator: op,
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

// parseMulDiv handles * and / operators
func (p *Parser) parseMulDiv() (ASTNode, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenOperator && (p.current.Value == "*" || p.current.Value == "/") {
		op := p.current.Value
		p.nextToken()

		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}

		left = &OperationNode{
			Operator: op,
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

// parseFactor handles numbers, variables, functions, and parenthesized expressions
func (p *Parser) parseFactor() (ASTNode, error) {
	switch p.current.Type {
	case TokenNumber:
		value, err := strconv.ParseFloat(p.current.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", p.current.Value)
		}
		p.nextToken()
		return &LiteralNode{Value: value}, nil

	case TokenVariable:
		name := p.current.Value
		p.nextToken()
		return &VariableNode{Name: name}, nil

	case TokenFunction:
		return p.parseFunction()

	case TokenParenOpen:
		p.nextToken() // consume '('
		node, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if p.current.Type != TokenParenClose {
			return nil, fmt.Errorf("expected ')' but got %s", p.current.Value)
		}
		p.nextToken() // consume ')'
		return node, nil

	default:
		return nil, fmt.Errorf("unexpected token: %s", p.current.Value)
	}
}

// parseFunction handles function calls like IF(condition, then, else)
func (p *Parser) parseFunction() (ASTNode, error) {
	funcName := p.current.Value
	p.nextToken() // consume function name

	if p.current.Type != TokenParenOpen {
		return nil, fmt.Errorf("expected '(' after function name")
	}
	p.nextToken() // consume '('

	// Handle specific functions
	switch strings.ToUpper(funcName) {
	case "IF":
		return p.parseIfFunction()
	default:
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}
}

// parseIfFunction handles IF(condition, then, else) function
func (p *Parser) parseIfFunction() (ASTNode, error) {
	// Parse condition
	condition, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("error parsing IF condition: %v", err)
	}

	if p.current.Type != TokenComma {
		return nil, fmt.Errorf("expected ',' after IF condition")
	}
	p.nextToken() // consume ','

	// Parse then branch
	thenNode, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("error parsing IF then branch: %v", err)
	}

	var elseNode ASTNode
	if p.current.Type == TokenComma {
		p.nextToken() // consume ','
		elseNode, err = p.parseExpression()
		if err != nil {
			return nil, fmt.Errorf("error parsing IF else branch: %v", err)
		}
	}

	if p.current.Type != TokenParenClose {
		return nil, fmt.Errorf("expected ')' to close IF function")
	}
	p.nextToken() // consume ')'

	return &ConditionalNode{
		Condition: condition,
		Then:      thenNode,
		Else:      elseNode,
	}, nil
}

// Helper function to check if operator is a comparison operator
func isComparisonOp(op string) bool {
	switch op {
	case ">", "<", ">=", "<=", "=", "!=":
		return true
	default:
		return false
	}
}

// SimpleFormulaParser is the main interface for parsing formulas
type SimpleFormulaParser struct{}

func NewSimpleParser() *SimpleFormulaParser {
	return &SimpleFormulaParser{}
}

// ParseString parses a formula string into an AST
func (sfp *SimpleFormulaParser) ParseString(formula string) (ASTNode, error) {
	// Clean the input
	formula = strings.TrimSpace(formula)
	if formula == "" {
		return nil, fmt.Errorf("empty formula")
	}

	parser := NewParser(formula)
	return parser.Parse()
}
