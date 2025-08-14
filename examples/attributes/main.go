package main

import (
	"fmt"
	"github.com/eternalsad/formula"
	"math"
	"unicode"
)

func main() {
	parser := formula.NewSimpleParser()

	// Test cases
	testCases := []string{
		"A/DC-1",
		"A asdsadasd",
		"A + B",
		"A + B * C",
		"(A + B) * C",
		"ЕСЛИ(age = 18 И 1 = 1) ТОГДА salary * 1.2 ИНАЧЕ salary",
		"ЕСЛИ 1 -(B/A)/B=0 ИЛИ 1 -(B/A)/B >1 ТОГДА 1 ИНАЧЕ (1-(B-A)/B*(-1))",
	}

	fmt.Println("=== Тестирование парсера формул ===\n")

	for i, testCase := range testCases {
		fmt.Printf("Тест %d: %s\n", i+1, testCase)

		validator := formula.NewFormulaValidator()

		res := validator.ValidateFormula(testCase)
		if res.Errors != nil {
			fmt.Printf("validate formula err: %v\n", res.Errors)
			return
		}

		ast, err := parser.ParseString(toUpperLatinOnlyUnicode(testCase))
		if err != nil {
			fmt.Printf("❌ Ошибка: %v\n", err)
		} else {
			//fmt.Printf("✅ AST: %s\n", ast.String())

			// Тестируем вычисление с примерными значениями
			// подставляем переменные нужно сохранить просто в таблице
			// добавим поле attributes
			// {
			//.  "A": {
			//.   "letter":"A",
			//.   "id": "12"
			//.   }
			// }
			// затем постепенно из json достать в репоизтории значения нужные по айди
			// и вот сюда вставить в виде переменных
			variables := map[string]float64{
				"A":      69456935.000000,
				"B":      69456939.000000,
				"C":      2,
				"age":    18,
				"salary": 50000,
			}

			result, evalErr := ast.Evaluate(&formula.Context{
				Variables: variables,
				Functions: nil,
			})

			result = roundFloat(result, 2)
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

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
func toUpperLatinOnlyUnicode(s string) string {
	runes := []rune(s)
	result := make([]rune, len(runes))

	for i, r := range runes {
		// Проверяем, является ли символ латинской буквой нижнего регистра
		if r >= 'a' && r <= 'z' && unicode.In(r, unicode.Latin) {
			result[i] = unicode.ToUpper(r)
		} else {
			result[i] = r // Оставляем без изменений
		}
	}

	return string(result)
}
