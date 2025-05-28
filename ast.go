package formula

import (
	"errors"
	"fmt"
	"math"
)

// NodeType определяет тип узла AST
type NodeType string

const (
	NodeTypeLiteral     NodeType = "literal"
	NodeTypeVariable    NodeType = "variable"
	NodeTypeOperation   NodeType = "operation"
	NodeTypeConditional NodeType = "conditional"
	NodeTypeComparison  NodeType = "comparison"
	NodeTypeFunction    NodeType = "function"
)

// ASTNode базовый интерфейс для всех узлов AST
type ASTNode interface {
	Evaluate(ctx *Context) (float64, error)
	GetType() NodeType
}

// Context содержит переменные и функции для вычисления
type Context struct {
	Variables map[string]float64
	Functions map[string]func([]float64) (float64, error)
}

// LiteralNode представляет числовое значение
type LiteralNode struct {
	Value float64 `json:"value"`
}

func (n *LiteralNode) Evaluate(ctx *Context) (float64, error) {
	return n.Value, nil
}

func (n *LiteralNode) GetType() NodeType {
	return NodeTypeLiteral
}

// VariableNode представляет переменную
type VariableNode struct {
	Name string `json:"name"`
}

func (n *VariableNode) Evaluate(ctx *Context) (float64, error) {
	if value, exists := ctx.Variables[n.Name]; exists {
		return value, nil
	}
	return 0, fmt.Errorf("variable '%s' not found", n.Name)
}

func (n *VariableNode) GetType() NodeType {
	return NodeTypeVariable
}

// OperationNode представляет математическую операцию
type OperationNode struct {
	Operator string  `json:"operator"`
	Left     ASTNode `json:"left"`
	Right    ASTNode `json:"right"`
}

func (n *OperationNode) Evaluate(ctx *Context) (float64, error) {
	left, err := n.Left.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	right, err := n.Right.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	switch n.Operator {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, errors.New("division by zero")
		}
		return left / right, nil
	case "^", "**":
		return math.Pow(left, right), nil
	case "%":
		if right == 0 {
			return 0, errors.New("modulo by zero")
		}
		return math.Mod(left, right), nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", n.Operator)
	}
}

func (n *OperationNode) GetType() NodeType {
	return NodeTypeOperation
}

// ComparisonNode представляет операцию сравнения
type ComparisonNode struct {
	Operator string  `json:"operator"`
	Left     ASTNode `json:"left"`
	Right    ASTNode `json:"right"`
}

func (n *ComparisonNode) Evaluate(ctx *Context) (float64, error) {
	left, err := n.Left.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	right, err := n.Right.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	var result bool
	switch n.Operator {
	case "=":
		result = left == right
	case "!=":
		result = left != right
	case ">":
		result = left > right
	case "<":
		result = left < right
	case ">=":
		result = left >= right
	case "<=":
		result = left <= right
	default:
		return 0, fmt.Errorf("unknown comparison operator: %s", n.Operator)
	}

	if result {
		return 1, nil
	}
	return 0, nil
}

func (n *ComparisonNode) GetType() NodeType {
	return NodeTypeComparison
}

// ConditionalNode представляет условное выражение IF-THEN-ELSE
type ConditionalNode struct {
	Condition ASTNode `json:"condition"`
	Then      ASTNode `json:"then"`
	Else      ASTNode `json:"else"`
}

func (n *ConditionalNode) Evaluate(ctx *Context) (float64, error) {
	condition, err := n.Condition.Evaluate(ctx)
	if err != nil {
		return 0, err
	}

	if condition != 0 { // 0 считается false, все остальное true
		return n.Then.Evaluate(ctx)
	} else if n.Else != nil {
		return n.Else.Evaluate(ctx)
	}

	return 0, nil
}

func (n *ConditionalNode) GetType() NodeType {
	return NodeTypeConditional
}

// FunctionNode представляет вызов функции
type FunctionNode struct {
	Name string    `json:"name"`
	Args []ASTNode `json:"args"`
}

func (n *FunctionNode) Evaluate(ctx *Context) (float64, error) {
	fn, exists := ctx.Functions[n.Name]
	if !exists {
		return 0, fmt.Errorf("function '%s' not found", n.Name)
	}

	args := make([]float64, len(n.Args))
	for i, arg := range n.Args {
		value, err := arg.Evaluate(ctx)
		if err != nil {
			return 0, err
		}
		args[i] = value
	}

	return fn(args)
}

func (n *FunctionNode) GetType() NodeType {
	return NodeTypeFunction
}
