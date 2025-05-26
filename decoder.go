package formula

import (
	"encoding/json"
	"fmt"
	"math"
)

// NodeData используется для десериализации JSON
type NodeData struct {
	Type      NodeType          `json:"type"`
	Value     *float64          `json:"value,omitempty"`
	Name      *string           `json:"name,omitempty"`
	Operator  *string           `json:"operator,omitempty"`
	Left      json.RawMessage   `json:"left,omitempty"`
	Right     json.RawMessage   `json:"right,omitempty"`
	Condition json.RawMessage   `json:"condition,omitempty"`
	Then      json.RawMessage   `json:"then,omitempty"`
	Else      json.RawMessage   `json:"else,omitempty"`
	Args      []json.RawMessage `json:"args,omitempty"`
}

// UnmarshalJSON десериализует JSON в ASTNode
func UnmarshalASTNode(data []byte) (ASTNode, error) {
	var nodeData NodeData
	if err := json.Unmarshal(data, &nodeData); err != nil {
		return nil, err
	}

	switch nodeData.Type {
	case NodeTypeLiteral:
		if nodeData.Value == nil {
			return nil, fmt.Errorf("literal node missing value")
		}
		return &LiteralNode{Value: *nodeData.Value}, nil

	case NodeTypeVariable:
		if nodeData.Name == nil {
			return nil, fmt.Errorf("variable node missing name")
		}
		return &VariableNode{Name: *nodeData.Name}, nil

	case NodeTypeOperation:
		if nodeData.Operator == nil {
			return nil, fmt.Errorf("operation node missing operator")
		}

		left, err := UnmarshalASTNode(nodeData.Left)
		if err != nil {
			return nil, fmt.Errorf("error parsing left operand: %v", err)
		}

		right, err := UnmarshalASTNode(nodeData.Right)
		if err != nil {
			return nil, fmt.Errorf("error parsing right operand: %v", err)
		}

		return &OperationNode{
			Operator: *nodeData.Operator,
			Left:     left,
			Right:    right,
		}, nil

	case NodeTypeComparison:
		if nodeData.Operator == nil {
			return nil, fmt.Errorf("comparison node missing operator")
		}

		left, err := UnmarshalASTNode(nodeData.Left)
		if err != nil {
			return nil, fmt.Errorf("error parsing left operand: %v", err)
		}

		right, err := UnmarshalASTNode(nodeData.Right)
		if err != nil {
			return nil, fmt.Errorf("error parsing right operand: %v", err)
		}

		return &ComparisonNode{
			Operator: *nodeData.Operator,
			Left:     left,
			Right:    right,
		}, nil

	case NodeTypeConditional:
		condition, err := UnmarshalASTNode(nodeData.Condition)
		if err != nil {
			return nil, fmt.Errorf("error parsing condition: %v", err)
		}

		then, err := UnmarshalASTNode(nodeData.Then)
		if err != nil {
			return nil, fmt.Errorf("error parsing then branch: %v", err)
		}

		node := &ConditionalNode{
			Condition: condition,
			Then:      then,
		}

		if len(nodeData.Else) > 0 {
			elseNode, err := UnmarshalASTNode(nodeData.Else)
			if err != nil {
				return nil, fmt.Errorf("error parsing else branch: %v", err)
			}
			node.Else = elseNode
		}

		return node, nil

	case NodeTypeFunction:
		if nodeData.Name == nil {
			return nil, fmt.Errorf("function node missing name")
		}

		args := make([]ASTNode, len(nodeData.Args))
		for i, argData := range nodeData.Args {
			arg, err := UnmarshalASTNode(argData)
			if err != nil {
				return nil, fmt.Errorf("error parsing function argument %d: %v", i, err)
			}
			args[i] = arg
		}

		return &FunctionNode{
			Name: *nodeData.Name,
			Args: args,
		}, nil

	default:
		return nil, fmt.Errorf("unknown node type: %s", nodeData.Type)
	}
}

// Helper функция для создания контекста
func NewContext() *Context {
	ctx := &Context{
		Variables: make(map[string]float64),
		Functions: make(map[string]func([]float64) (float64, error)),
	}

	// Добавляем базовые математические функции
	ctx.Functions["abs"] = func(args []float64) (float64, error) {
		if len(args) != 1 {
			return 0, fmt.Errorf("abs requires exactly 1 argument")
		}
		return math.Abs(args[0]), nil
	}

	ctx.Functions["sqrt"] = func(args []float64) (float64, error) {
		if len(args) != 1 {
			return 0, fmt.Errorf("sqrt requires exactly 1 argument")
		}
		if args[0] < 0 {
			return 0, fmt.Errorf("sqrt of negative number")
		}
		return math.Sqrt(args[0]), nil
	}

	ctx.Functions["max"] = func(args []float64) (float64, error) {
		if len(args) == 0 {
			return 0, fmt.Errorf("max requires at least 1 argument")
		}
		max := args[0]
		for _, arg := range args[1:] {
			if arg > max {
				max = arg
			}
		}
		return max, nil
	}

	ctx.Functions["min"] = func(args []float64) (float64, error) {
		if len(args) == 0 {
			return 0, fmt.Errorf("min requires at least 1 argument")
		}
		min := args[0]
		for _, arg := range args[1:] {
			if arg < min {
				min = arg
			}
		}
		return min, nil
	}

	ctx.Functions["sum"] = func(args []float64) (float64, error) {
		sum := 0.0
		for _, arg := range args {
			sum += arg
		}
		return sum, nil
	}

	return ctx
}
