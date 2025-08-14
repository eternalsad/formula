package main

import (
	"fmt"
	"github.com/eternalsad/formula"
)

func main() {
	parser := formula.NewSimpleParser()

	// Test cases
	testCases := []string{
		"A + B И 123",
		"A + B * C",
		"A + B / C - D",
		"(A + B) * C",
		"IF(age > 18, salary * 1.2, salary)",
		"IF(A + B > 1000, A * 2, B * 3)",
		"A >= 100",
		"price * (1 + tax)",
		"IF(score >= 90, 5, IF(score >= 80, 4, 3))",
	}

	fmt.Println("=== Тестирование парсера формул ===\n")

	for i, testCase := range testCases {
		fmt.Printf("Тест %d: %s\n", i+1, testCase)

		ast, err := parser.ParseString(testCase)
		if err != nil {
			fmt.Printf("❌ Ошибка: %v\n", err)
		} else {
			//fmt.Printf("✅ AST: %s\n", ast.String())

			// Тестируем вычисление с примерными значениями
			variables := map[string]float64{
				"A":      10,
				"B":      5,
				"C":      2,
				"D":      1,
				"age":    25,
				"salary": 50000,
				"score":  85,
				"price":  100,
				"tax":    0.1,
			}

			result, evalErr := ast.Evaluate(&formula.Context{
				Variables: variables,
				Functions: nil,
			})
			if evalErr != nil {
				fmt.Printf("⚠️  Ошибка вычисления: %v\n", evalErr)
			} else {
				fmt.Printf("📊 Результат: %.2f\n", result)
			}
		}
		fmt.Println()
	}

	// Демонстрация пошагового разбора
	fmt.Println("=== Пошаговый разбор формулы 'IF(age > 18, salary * 1.2, salary)' ===")
	demonstrateTokenization("IF(age > 18, salary * 1.2, salary)")
}

// demonstrateTokenization shows how the lexer tokenizes input
func demonstrateTokenization(input string) {
	lexer := formula.NewLexer(input)
	fmt.Printf("Исходная строка: %s\n", input)
	fmt.Println("Токены:")

	for {
		token := lexer.NextToken()
		if token.Type == formula.TokenEOF {
			break
		}

		var tokenTypeName string
		switch token.Type {
		case formula.TokenNumber:
			tokenTypeName = "NUMBER"
		case formula.TokenVariable:
			tokenTypeName = "VARIABLE"
		case formula.TokenOperator:
			tokenTypeName = "OPERATOR"
		case formula.TokenParenOpen:
			tokenTypeName = "PAREN_OPEN"
		case formula.TokenParenClose:
			tokenTypeName = "PAREN_CLOSE"
		case formula.TokenComma:
			tokenTypeName = "COMMA"
		case formula.TokenFunction:
			tokenTypeName = "FUNCTION"
		}

		fmt.Printf("  %s: '%s' (позиция %d)\n", tokenTypeName, token.Value, token.Pos)
	}
}
