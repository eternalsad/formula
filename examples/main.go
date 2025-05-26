package main

import (
	"encoding/json"
	"fmt"
	"github.com/eternalsad/formula"
	"log"
	"net/http"
)

func main() {
	// Пример 1: Простая арифметика (a + b * 2)
	example1 := `{
		"type": "operation",
		"operator": "+",
		"left": {
			"type": "variable",
			"name": "a"
		},
		"right": {
			"type": "operation",
			"operator": "*",
			"left": {
				"type": "variable",
				"name": "b"
			},
			"right": {
				"type": "literal",
				"value": 2
			}
		}
	}`

	// Пример 2: Условное выражение IF(age > 18, salary * 1.2, salary)
	example2 := `{
		"type": "conditional",
		"condition": {
			"type": "comparison",
			"operator": ">",
			"left": {
				"type": "variable",
				"name": "age"
			},
			"right": {
				"type": "literal",
				"value": 18
			}
		},
		"then": {
			"type": "operation",
			"operator": "*",
			"left": {
				"type": "variable",
				"name": "salary"
			},
			"right": {
				"type": "literal",
				"value": 1.2
			}
		},
		"else": {
			"type": "variable",
			"name": "salary"
		}
	}`

	// Пример 3: Функция MAX(a, b) + sqrt(c)
	example3 := `{
		"type": "operation",
		"operator": "+",
		"left": {
			"type": "function",
			"name": "max",
			"args": [
				{
					"type": "variable",
					"name": "a"
				},
				{
					"type": "variable",
					"name": "b"
				}
			]
		},
		"right": {
			"type": "function",
			"name": "sqrt",
			"args": [
				{
					"type": "variable",
					"name": "c"
				}
			]
		}
	}`

	// Тестируем примеры
	runExample("Пример 1: a + b * 2", example1, map[string]float64{
		"a": 10,
		"b": 5,
	})

	runExample("Пример 2: IF(age > 18, salary * 1.2, salary)", example2, map[string]float64{
		"age":    25,
		"salary": 50000,
	})

	runExample("Пример 3: MAX(a, b) + sqrt(c)", example3, map[string]float64{
		"a": 15,
		"b": 20,
		"c": 16,
	})

	// Пример сложной формулы с несколькими условиями
	complexExample := `{
		"type": "conditional",
		"condition": {
			"type": "comparison",
			"operator": ">=",
			"left": {
				"type": "variable",
				"name": "score"
			},
			"right": {
				"type": "literal",
				"value": 90
			}
		},
		"then": {
			"type": "literal",
			"value": 5
		},
		"else": {
			"type": "conditional",
			"condition": {
				"type": "comparison",
				"operator": ">=",
				"left": {
					"type": "variable",
					"name": "score"
				},
				"right": {
					"type": "literal",
					"value": 80
				}
			},
			"then": {
				"type": "literal",
				"value": 4
			},
			"else": {
				"type": "conditional",
				"condition": {
					"type": "comparison",
					"operator": ">=",
					"left": {
						"type": "variable",
						"name": "score"
					},
					"right": {
						"type": "literal",
						"value": 70
					}
				},
				"then": {
					"type": "literal",
					"value": 3
				},
				"else": {
					"type": "literal",
					"value": 2
				}
			}
		}
	}`

	runExample("Пример 4: Система оценок", complexExample, map[string]float64{
		"score": 85,
	})
}

func runExample(title, jsonStr string, variables map[string]float64) {
	fmt.Printf("\n=== %s ===\n", title)

	// Парсим AST из JSON
	node, err := formula.UnmarshalASTNode([]byte(jsonStr))
	if err != nil {
		log.Printf("Ошибка парсинга: %v", err)
		return
	}

	// Создаем контекст с переменными
	ctx := formula.NewContext()
	ctx.Variables = variables

	// Вычисляем результат
	result, err := node.Evaluate(ctx)
	if err != nil {
		log.Printf("Ошибка вычисления: %v", err)
		return
	}

	fmt.Printf("Переменные: %+v\n", variables)
	fmt.Printf("Результат: %.2f\n", result)
}

// HTTP Handler для API
func FormulaHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Formula   json.RawMessage    `json:"formula"`
		Variables map[string]float64 `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим AST
	node, err := formula.UnmarshalASTNode(request.Formula)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid formula: %v", err), http.StatusBadRequest)
		return
	}

	// Создаем контекст
	ctx := formula.NewContext()
	ctx.Variables = request.Variables

	// Вычисляем
	result, err := node.Evaluate(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Evaluation error: %v", err), http.StatusBadRequest)
		return
	}

	// Возвращаем результат
	response := map[string]interface{}{
		"result": result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
