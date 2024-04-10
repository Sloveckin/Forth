//go:build !solution

package main

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

type defenition struct {
	index int
	data  []string
}

func newDefenition(index int) *defenition {
	data := make([]string, 0)
	return &defenition{index, data}
}

type Evaluator struct {
	mapping  map[string]*defenition
	defCount int
}

// NewEvaluator creates evaluator.
func NewEvaluator() *Evaluator {
	mp := make(map[string]*defenition, 0)
	return &Evaluator{mp, 0}
}

// Why isn't it memeber of Evaluator?
// I don't know
var (
	operationTable = map[string]func(*[]int) error{
		"+":    add,
		"-":    sub,
		"*":    mul,
		"/":    div,
		"dup":  dup,
		"drop": drop,
		"swap": swap,
		"over": over,
	}
)

// If it was haskell, I could write just "return bin(stack, +, "add", false)"
func add(stack *[]int) error {
	f := func(a, b int) int {
		return a + b
	}
	return bin(stack, f, "add", false)
}

func mul(stack *[]int) error {
	f := func(a, b int) int {
		return a * b
	}
	return bin(stack, f, "mul", false)
}

func sub(stack *[]int) error {
	f := func(a, b int) int {
		return a - b
	}
	return bin(stack, f, "sub", false)
}

func div(stack *[]int) error {
	f := func(a, b int) int {
		return a / b
	}
	return bin(stack, f, "div", true)
}

func bin(stack *[]int, fun func(int, int) int, errMsg string, flag bool) error {
	if err := binaryCheker(*stack, errMsg); err != nil {
		return err
	}

	a, b := takeTwo(stack)

	/// FP flashback
	if flag && a == 0 {
		return errors.New("division by zero")
	}
	*stack = append(*stack, fun(b, a))
	return nil
}

func dup(stack *[]int) error {
	if err := unaryChecker(*stack, "dup"); err != nil {
		return err
	}
	top := (*stack)[len(*stack)-1]
	*stack = append(*stack, top)
	return nil
}

func over(stack *[]int) error {
	if err := binaryCheker(*stack, "over"); err != nil {
		return err
	}
	*stack = append(*stack, (*stack)[len(*stack)-2])
	return nil
}

func drop(stack *[]int) error {
	if err := unaryChecker(*stack, "drop"); err != nil {
		return err
	}
	*stack = (*stack)[:len(*stack)-1]
	return nil
}

func swap(stack *[]int) error {
	if err := binaryCheker(*stack, "swap"); err != nil {
		return err
	}
	a, b := takeTwo(stack)
	*stack = append(*stack, a)
	*stack = append(*stack, b)
	return nil
}

func binaryCheker(stack []int, errMsg string) error {
	if len(stack) == 0 {
		return errors.New("nothing to " + errMsg)
	} else if len(stack) == 1 {
		return errors.New(errMsg + " arity")
	}
	return nil
}

func unaryChecker(stack []int, errMsg string) error {
	if len(stack) == 0 {
		return errors.New("nothing to " + errMsg)
	}
	return nil
}

func takeTwo(stack *[]int) (int, int) {
	f := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	s := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return f, s
}

func (e *Evaluator) addDefenition(index int, key string, defention []string) error {
	_, err := strconv.Atoi(key)
	if err == nil {
		return errors.New("redefine numbers")
	}

	newDefenition := newDefenition(index)
	for _, el := range defention {
		if mpData, ok := e.mapping[el]; ok {
			newDefenition.data = append(newDefenition.data, mpData.data...)
		} else {
			newDefenition.data = append(newDefenition.data, el)
		}
	}
	e.mapping[key] = newDefenition
	return nil
}

func (e *Evaluator) wordHandle(word string, stack *[]int) error {
	val, err := strconv.Atoi(word)
	if err == nil {
		*stack = append(*stack, val)
		return nil
	}

	if def, ok := e.mapping[word]; ok {
		for _, val := range def.data {
			if el, ok := e.mapping[val]; ok {
				if el.index < def.index {
					e.wordHandle(strings.ToLower(val), stack)
				}
				err = operationTable[val](stack)
				if err != nil {
					return err
				}
				continue
			} else {
				e.wordHandle(strings.ToLower(val), stack)
			}
		}
		return nil
	}

	f, ok := operationTable[word]
	if !ok {
		return errors.New("non-existent word")
	}
	return f(stack)
}

func (e *Evaluator) sequenceHandler(parts []string) ([]int, error) {
	stack := make([]int, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		word := strings.ToLower(parts[i])
		err := e.wordHandle(word, &stack)
		if err != nil {
			return stack, err
		}
	}
	return stack, nil
}

// Process evaluates sequence of words or definition.
//
// Returns resulting stack state and an error.
func (e *Evaluator) Process(row string) ([]int, error) {
	parts := strings.Split(row, " ")
	firstRune, _ := utf8.DecodeRuneInString(parts[0])
	if firstRune == ':' {
		err := e.addDefenition(e.defCount, strings.ToLower(parts[1]), parts[2:len(parts)-1])
		e.defCount++
		return []int{}, err
	}

	return e.sequenceHandler(parts)
}
