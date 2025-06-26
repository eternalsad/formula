package main

import (
	"fmt"
	"github.com/eternalsad/formula"
)

func main() {
	validator := formula.NewFormulaValidator()

	// Тестовые случаи - невалидные формулы
	invalidFormulas := []string{
		"asdasdasdas",          // Неизвестные переменные (но синтаксически корректно)
		"A !&?@#$'\"}{[}",      // Недопустимые символы
		"ц !&?@#$'\"}{[}",      // Недопустимые символы с кириллицей
		"A*A!!!@@)((*",         // Недопустимые символы и операторы
		"",                     // Пустая формула
		"A + + B",              // Двойные операторы
		"A + B)",               // Лишняя скобка
		"(A + B",               // Незакрытая скобка
		"A +",                  // Оператор в конце
		"* A + B",              // Оператор в начале (кроме унарного минуса)
		"A ++ B",               // Недопустимая последовательность операторов
		"A === B",              // Тройное равенство
		"A & B",                // Недопустимый оператор &
		"A | B",                // Недопустимый оператор |
		"A @ B",                // Недопустимый символ @
		"ЕСЛИ A > B ТОГДА",     // Незавершенная конструкция
		"IF A > B THEN C ELSE", // Незавершенная конструкция
		"((A + B) * C",         // Несбалансированные скобки
		"A + B) * C)",          // Лишние скобки
		"функция(A, B, C)",     // Кириллические имена функций
	}

	// Валидные формулы для сравнения
	validFormulas := []string{
		"A + B",
		"A * B - C",
		"(A + B) * C",
		"-A + B", // Унарный минус
		"ЕСЛИ A > B ТОГДА C ИНАЧЕ D",
		"IF A > B THEN C ELSE D",
		"A = 5 ИЛИ B = 10",
		"A >= B AND C <= D",
		"функция(A, B, C)", // Кириллические имена функций
	}

	fmt.Println("=== ТЕСТ НЕВАЛИДНЫХ ФОРМУЛ ===\n")

	for i, formula := range invalidFormulas {
		fmt.Printf("Тест %d: \"%s\"\n", i+1, formula)

		result := validator.ValidateFormula(formula)

		if result.IsValid {
			fmt.Printf("  ⚠️  НЕОЖИДАННО: формула прошла валидацию!\n")
		} else {
			fmt.Printf("  ❌ Невалидна (как ожидалось)\n")
			for j, err := range result.Errors {
				fmt.Printf("     %d. %s\n", j+1, err.Error())
			}
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("  ⚠️  Предупреждения:\n")
			for j, warning := range result.Warnings {
				fmt.Printf("     %d. %s\n", j+1, warning)
			}
		}

		fmt.Println()
	}

	fmt.Println("=== ТЕСТ ВАЛИДНЫХ ФОРМУЛ ===\n")

	for i, formula := range validFormulas {
		fmt.Printf("Тест %d: \"%s\"\n", i+1, formula)

		result := validator.ValidateFormula(formula)

		if result.IsValid {
			fmt.Printf("  ✅ Валидна\n")
		} else {
			fmt.Printf("  ❌ НЕОЖИДАННО: формула не прошла валидацию!\n")
			for j, err := range result.Errors {
				fmt.Printf("     %d. %s\n", j+1, err.Error())
			}
		}

		if len(result.Warnings) > 0 {
			fmt.Printf("  ⚠️  Предупреждения:\n")
			for j, warning := range result.Warnings {
				fmt.Printf("     %d. %s\n", j+1, warning)
			}
		}

		fmt.Println()
	}

	// Демонстрация быстрой валидации
	fmt.Println("=== БЫСТРАЯ ВАЛИДАЦИЯ ===\n")

	testQuick := []string{
		"A + B",
		"A !@# B",
		"((A + B))",
		"A + + B",
	}

	for _, testFormula := range testQuick {
		isValid := formula.QuickValidate(testFormula)
		status := "❌"
		if isValid {
			status = "✅"
		}
		fmt.Printf("%s \"%s\"\n", status, testFormula)
	}

	// Демонстрация валидации с получением ошибок
	fmt.Println("\n=== ВАЛИДАЦИЯ С ПОЛУЧЕНИЕМ ОШИБОК ===\n")

	testFormula := "A !@# + + B"
	isValid, errors := formula.ValidateAndGetErrors(testFormula)

	fmt.Printf("Формула: \"%s\"\n", testFormula)
	fmt.Printf("Валидна: %t\n", isValid)
	if !isValid {
		fmt.Println("Ошибки:")
		for i, err := range errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
	}
}
