package formula

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
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
	TokenThen
	TokenIf
	TokenElse
	TokenOr
	TokenAnd
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
	runes []rune
}

func NewLexer(input string) *Lexer {
	// Don't remove ALL spaces - only trim and normalize
	cleanInput := strings.TrimSpace(input)
	// Replace multiple spaces with single space, then remove spaces around operators
	cleanInput = normalizeSpaces(cleanInput)
	return &Lexer{
		input: cleanInput,
		pos:   0,
		runes: []rune(cleanInput),
	}
}

// normalizeSpaces removes spaces around operators but keeps spaces between words and numbers
func normalizeSpaces(input string) string {
	// Keep spaces that separate letters from numbers
	result := make([]rune, 0, len(input))
	runes := []rune(input)

	for i, r := range runes {
		if r == ' ' {
			// Check if we should keep this space
			if i > 0 && i < len(runes)-1 {
				prev := runes[i-1]
				next := runes[i+1]

				// Keep space if it separates a letter from a number or vice versa
				if (unicode.IsLetter(prev) && unicode.IsDigit(next)) ||
					(unicode.IsDigit(prev) && unicode.IsLetter(next)) ||
					(unicode.IsLetter(prev) && unicode.IsLetter(next)) {
					result = append(result, r)
					continue
				}
			}
			// Skip spaces around operators
			continue
		}
		result = append(result, r)
	}

	return string(result)
}

func (l *Lexer) NextToken() Token {
	// Skip whitespace
	for l.pos < len(l.runes) && unicode.IsSpace(l.runes[l.pos]) {
		l.pos++
	}

	if l.pos >= len(l.runes) {
		return Token{TokenEOF, "", l.pos}
	}

	char := l.runes[l.pos]

	// Numbers (including decimals)
	if unicode.IsDigit(char) {
		return l.readNumber()
	}

	// Variables, functions, and keywords
	if unicode.IsLetter(char) {
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
	for l.pos < len(l.runes) && (unicode.IsDigit(l.runes[l.pos]) || l.runes[l.pos] == '.') {
		l.pos++
	}
	return Token{TokenNumber, string(l.runes[start:l.pos]), start}
}

func (l *Lexer) readIdentifier() Token {
	start := l.pos
	// Read only letters and underscores for identifiers - no digits
	for l.pos < len(l.runes) && (unicode.IsLetter(l.runes[l.pos]) || l.runes[l.pos] == '_') {
		l.pos++
	}

	value := string(l.runes[start:l.pos])
	upperValue := strings.ToUpper(value)

	// Check for Russian keywords
	switch upperValue {
	case "ЕСЛИ":
		return Token{TokenIf, value, start}
	case "ТОГДА":
		return Token{TokenThen, value, start}
	case "ИНАЧЕ":
		return Token{TokenElse, value, start}
	case "ИЛИ":
		return Token{TokenOr, value, start}
	case "И":
		return Token{TokenAnd, value, start}
	}

	// Check for English keywords
	switch upperValue {
	case "IF":
		return Token{TokenIf, value, start}
	case "THEN":
		return Token{TokenThen, value, start}
	case "ELSE":
		return Token{TokenElse, value, start}
	case "OR":
		return Token{TokenOr, value, start}
	case "AND":
		return Token{TokenAnd, value, start}
	}

	// Check if it's a function (followed by parenthesis)
	// Skip whitespace to check for opening parenthesis
	tempPos := l.pos
	for tempPos < len(l.runes) && unicode.IsSpace(l.runes[tempPos]) {
		tempPos++
	}
	if tempPos < len(l.runes) && l.runes[tempPos] == '(' {
		return Token{TokenFunction, value, start}
	}

	return Token{TokenVariable, value, start}
}

func (l *Lexer) readOperator() Token {
	start := l.pos

	// Handle multi-character operators
	if l.pos+1 < len(l.runes) {
		twoChar := string(l.runes[l.pos : l.pos+2])
		switch twoChar {
		case ">=", "<=", "==", "!=":
			l.pos += 2
			return Token{TokenOperator, twoChar, start}
		}
	}

	l.pos++
	return Token{TokenOperator, string(l.runes[start]), start}
}

// Removed isDigit and isLetter functions - using unicode package instead

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
	// Check for IF statement at the beginning
	if p.current.Type == TokenIf {
		return p.parseIfStatement()
	}
	return p.parseLogicalOr()
}

// parseIfStatement handles ЕСЛИ...ТОГДА...ИНАЧЕ construction
func (p *Parser) parseIfStatement() (ASTNode, error) {
	if p.current.Type != TokenIf {
		return nil, fmt.Errorf("expected IF/ЕСЛИ")
	}
	p.nextToken() // consume IF/ЕСЛИ

	// Parse condition
	condition, err := p.parseLogicalOr()
	if err != nil {
		return nil, fmt.Errorf("error parsing IF condition: %v", err)
	}

	if p.current.Type != TokenThen {
		return nil, fmt.Errorf("expected THEN/ТОГДА after IF condition")
	}
	p.nextToken() // consume THEN/ТОГДА

	// Parse then branch
	thenNode, err := p.parseLogicalOr()
	if err != nil {
		return nil, fmt.Errorf("error parsing IF then branch: %v", err)
	}

	var elseNode ASTNode
	if p.current.Type == TokenElse {
		p.nextToken() // consume ELSE/ИНАЧЕ
		elseNode, err = p.parseLogicalOr()
		if err != nil {
			return nil, fmt.Errorf("error parsing IF else branch: %v", err)
		}
	}

	return &ConditionalNode{
		Condition: condition,
		Then:      thenNode,
		Else:      elseNode,
	}, nil
}

// parseLogicalOr handles OR/ИЛИ operators
func (p *Parser) parseLogicalOr() (ASTNode, error) {
	left, err := p.parseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenOr {
		p.nextToken() // consume OR/ИЛИ

		right, err := p.parseLogicalAnd()
		if err != nil {
			return nil, err
		}

		left = &LogicalNode{
			Operator: "OR",
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

// parseLogicalAnd handles AND/И operators
func (p *Parser) parseLogicalAnd() (ASTNode, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}

	for p.current.Type == TokenAnd {
		p.nextToken() // consume AND/И

		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}

		left = &LogicalNode{
			Operator: "AND",
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
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

// parseFactor handles numbers, variables, functions, unary operators, and parenthesized expressions
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

	case TokenOperator:
		// Handle unary operators (+ and -)
		if p.current.Value == "+" || p.current.Value == "-" {
			op := p.current.Value
			p.nextToken()

			operand, err := p.parseFactor()
			if err != nil {
				return nil, err
			}

			return &UnaryNode{
				Operator: op,
				Operand:  operand,
			}, nil
		}
		return nil, fmt.Errorf("unexpected operator: %s", p.current.Value)

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
	case "IF", "ЕСЛИ":
		return p.parseIfFunction()
	default:
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}
}

// parseIfFunction handles IF(condition, then, else) function
func (p *Parser) parseIfFunction() (ASTNode, error) {
	// Parse condition
	condition, err := p.parseLogicalOr()
	if err != nil {
		return nil, fmt.Errorf("error parsing IF condition: %v", err)
	}

	if p.current.Type != TokenComma {
		return nil, fmt.Errorf("expected ',' after IF condition")
	}
	p.nextToken() // consume ','

	// Parse then branch
	thenNode, err := p.parseLogicalOr()
	if err != nil {
		return nil, fmt.Errorf("error parsing IF then branch: %v", err)
	}

	var elseNode ASTNode
	if p.current.Type == TokenComma {
		p.nextToken() // consume ','
		elseNode, err = p.parseLogicalOr()
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
