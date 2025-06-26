package formula

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Message  string
	Position int
	Code     string
}

func (e *ValidationError) Error() string {
	if e.Position >= 0 {
		return fmt.Sprintf("validation error at position %d: %s", e.Position, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ValidationResult содержит результат валидации
type ValidationResult struct {
	IsValid  bool
	Errors   []ValidationError
	Warnings []string
}

// FormulaValidator валидирует формулы
type FormulaValidator struct {
	allowedOperators map[rune]bool
	keywords         map[string]bool
}

// NewFormulaValidator создает новый валидатор
func NewFormulaValidator() *FormulaValidator {
	return &FormulaValidator{
		allowedOperators: map[rune]bool{
			'+': true, '-': true, '*': true, '/': true,
			'=': true, '!': true, '>': true, '<': true,
			'(': true, ')': true, ',': true, '.': true,
		},
		keywords: map[string]bool{
			// Русские ключевые слова
			"ЕСЛИ": true, "ИЛИ": true, "И": true,
			"ТОГДА": true, "ИНАЧЕ": true,
			// Английские ключевые слова
			"IF": true, "THEN": true, "ELSE": true,
			"OR": true, "AND": true,
		},
	}
}

// ValidateFormula выполняет комплексную валидацию формулы
func (v *FormulaValidator) ValidateFormula(formula string) ValidationResult {
	result := ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Базовые проверки
	if err := v.validateBasicStructure(formula); err != nil {
		result.Errors = append(result.Errors, *err)
		result.IsValid = false
	}

	// Проверка недопустимых символов
	if errors := v.validateCharacters(formula); len(errors) > 0 {
		result.Errors = append(result.Errors, errors...)
		result.IsValid = false
	}

	// Проверка использования кириллицы
	if errors := v.validateCyrillicUsage(formula); len(errors) > 0 {
		result.Errors = append(result.Errors, errors...)
		result.IsValid = false
	}

	// Проверка скобок
	if err := v.validateParentheses(formula); err != nil {
		result.Errors = append(result.Errors, *err)
		result.IsValid = false
	}

	// Проверка операторов
	if errors := v.validateOperators(formula); len(errors) > 0 {
		result.Errors = append(result.Errors, errors...)
		result.IsValid = false
	}

	// Проверка синтаксиса через токенизацию
	if result.IsValid {
		if err := v.validateSyntax(formula); err != nil {
			result.Errors = append(result.Errors, *err)
			result.IsValid = false
		}
	}

	// Предупреждения
	warnings := v.generateWarnings(formula)
	result.Warnings = append(result.Warnings, warnings...)

	return result
}

// validateBasicStructure проверяет базовую структуру формулы
func (v *FormulaValidator) validateBasicStructure(formula string) *ValidationError {
	trimmed := strings.TrimSpace(formula)

	if len(trimmed) == 0 {
		return &ValidationError{
			Message: "формула не может быть пустой",
			Code:    "EMPTY_FORMULA",
		}
	}

	if len(trimmed) > 1000 {
		return &ValidationError{
			Message: "формула слишком длинная (максимум 1000 символов)",
			Code:    "FORMULA_TOO_LONG",
		}
	}

	return nil
}

// validateCharacters проверяет недопустимые символы
func (v *FormulaValidator) validateCharacters(formula string) []ValidationError {
	var errors []ValidationError
	runes := []rune(formula)

	for i, r := range runes {
		if !v.isValidCharacter(r) {
			errors = append(errors, ValidationError{
				Message:  fmt.Sprintf("недопустимый символ '%c'", r),
				Position: i,
				Code:     "INVALID_CHARACTER",
			})
		}
	}

	return errors
}

// isValidCharacter проверяет, является ли символ допустимым
func (v *FormulaValidator) isValidCharacter(r rune) bool {
	// Цифры
	if unicode.IsDigit(r) {
		return true
	}

	// Пробелы
	if unicode.IsSpace(r) {
		return true
	}

	// Разрешенные операторы и символы
	if v.allowedOperators[r] {
		return true
	}

	// Подчеркивание для переменных
	if r == '_' {
		return true
	}

	// Только латинские буквы
	if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
		return true
	}

	// Кириллица разрешена только в составе ключевых слов
	// Проверяем это отдельно в validateCyrillicUsage
	if unicode.In(r, unicode.Cyrillic) {
		return true // Временно разрешаем, проверим контекст позже
	}

	return false
}

// validateCyrillicUsage проверяет использование кириллицы
func (v *FormulaValidator) validateCyrillicUsage(formula string) []ValidationError {
	var errors []ValidationError

	// Находим все кириллические слова
	cyrillicWords := v.extractCyrillicWords(formula)

	for word, positions := range cyrillicWords {
		upperWord := strings.ToUpper(word)
		if !v.keywords[upperWord] {
			// Кириллическое слово не является ключевым словом
			for _, pos := range positions {
				errors = append(errors, ValidationError{
					Message:  fmt.Sprintf("кириллическое слово '%s' не является допустимым ключевым словом. Разрешены только: ЕСЛИ, ИЛИ, И, ТОГДА, ИНАЧЕ", word),
					Position: pos,
					Code:     "INVALID_CYRILLIC_WORD",
				})
			}
		}
	}

	return errors
}

// extractCyrillicWords извлекает все кириллические слова и их позиции
func (v *FormulaValidator) extractCyrillicWords(formula string) map[string][]int {
	words := make(map[string][]int)
	runes := []rune(formula)

	for i := 0; i < len(runes); {
		if unicode.In(runes[i], unicode.Cyrillic) {
			// Начало кириллического слова
			start := i
			wordRunes := []rune{}

			// Читаем все кириллические символы подряд
			for i < len(runes) && (unicode.In(runes[i], unicode.Cyrillic) || runes[i] == '_') {
				wordRunes = append(wordRunes, runes[i])
				i++
			}

			word := string(wordRunes)
			if _, exists := words[word]; !exists {
				words[word] = []int{}
			}
			words[word] = append(words[word], start)
		} else {
			i++
		}
	}

	return words
}

// validateParentheses проверяет правильность расстановки скобок
func (v *FormulaValidator) validateParentheses(formula string) *ValidationError {
	stack := 0
	runes := []rune(formula)

	for i, r := range runes {
		switch r {
		case '(':
			stack++
		case ')':
			stack--
			if stack < 0 {
				return &ValidationError{
					Message:  "лишняя закрывающая скобка",
					Position: i,
					Code:     "EXTRA_CLOSING_PAREN",
				}
			}
		}
	}

	if stack > 0 {
		return &ValidationError{
			Message: fmt.Sprintf("не хватает %d закрывающих скобок", stack),
			Code:    "MISSING_CLOSING_PAREN",
		}
	}

	return nil
}

// validateOperators проверяет операторы
func (v *FormulaValidator) validateOperators(formula string) []ValidationError {
	var errors []ValidationError

	// Проверка на подряд идущие операторы
	operatorPattern := regexp.MustCompile(`[+\-*/=!><]{3,}`)
	matches := operatorPattern.FindAllStringIndex(formula, -1)

	for _, match := range matches {
		errors = append(errors, ValidationError{
			Message:  "недопустимая последовательность операторов",
			Position: match[0],
			Code:     "INVALID_OPERATOR_SEQUENCE",
		})
	}

	// Проверка на операторы в начале/конце (кроме унарного минуса)
	trimmed := strings.TrimSpace(formula)
	if len(trimmed) > 0 {
		lastChar := rune(trimmed[len(trimmed)-1])
		if strings.ContainsRune("*/=!><", lastChar) {
			errors = append(errors, ValidationError{
				Message:  "формула не может заканчиваться оператором",
				Position: len(formula) - 1,
				Code:     "FORMULA_ENDS_WITH_OPERATOR",
			})
		}
	}

	return errors
}

// validateSyntax проверяет синтаксис через токенизацию
func (v *FormulaValidator) validateSyntax(formula string) *ValidationError {
	lexer := NewLexer(formula)

	// Пытаемся токенизировать всю формулу
	for {
		token := lexer.NextToken()
		if token.Type == TokenEOF {
			break
		}

		// Проверяем на неожиданные токены
		if token.Value == "" && token.Type != TokenEOF {
			return &ValidationError{
				Message:  "неожиданный токен в формуле",
				Position: token.Pos,
				Code:     "UNEXPECTED_TOKEN",
			}
		}
	}

	// Пытаемся распарсить формулу
	parser := NewParser(formula)
	_, err := parser.Parse()
	if err != nil {
		return &ValidationError{
			Message: fmt.Sprintf("ошибка синтаксиса: %v", err),
			Code:    "SYNTAX_ERROR",
		}
	}

	return nil
}

// generateWarnings генерирует предупреждения
func (v *FormulaValidator) generateWarnings(formula string) []string {
	var warnings []string

	// Предупреждение о смешении языков
	hasRussian := regexp.MustCompile(`[а-яё]`).MatchString(strings.ToLower(formula))
	hasEnglish := regexp.MustCompile(`[a-z]`).MatchString(strings.ToLower(formula))

	if hasRussian && hasEnglish {
		warnings = append(warnings, "формула содержит смешение русских и английских ключевых слов")
	}

	// Предупреждение о сложности
	if strings.Count(formula, "(") > 5 {
		warnings = append(warnings, "формула может быть слишком сложной для понимания")
	}

	// Предупреждение о длинных именах переменных
	variablePattern := regexp.MustCompile(`[a-zA-Zа-яёА-ЯЁ_][a-zA-Zа-яёА-ЯЁ0-9_]*`)
	variables := variablePattern.FindAllString(formula, -1)

	for _, variable := range variables {
		if !v.keywords[strings.ToUpper(variable)] && len(variable) > 20 {
			warnings = append(warnings, fmt.Sprintf("переменная '%s' имеет очень длинное имя", variable))
		}
	}

	return warnings
}

// QuickValidate быстрая валидация для простых случаев
func QuickValidate(formula string) bool {
	validator := NewFormulaValidator()
	result := validator.ValidateFormula(formula)
	return result.IsValid
}

// ValidateAndGetErrors валидация с возвратом всех ошибок
func ValidateAndGetErrors(formula string) (bool, []string) {
	validator := NewFormulaValidator()
	result := validator.ValidateFormula(formula)

	var errorMessages []string
	for _, err := range result.Errors {
		errorMessages = append(errorMessages, err.Error())
	}

	return result.IsValid, errorMessages
}
