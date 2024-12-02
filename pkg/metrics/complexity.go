package metrics

import (
	ast "go/ast"
	"go/token"

	myAst "github.com/BelehovEgor/go-fuzz-targets-search-engine/pkg/ast"
)

type Complexity struct {
	// Func info
	Name string

	// Functions
	cyclomatic int

	// Loops
	number_of_loops                int
	number_of_nested_loops         int
	maximum_nesting_level_of_loops int
}

func CalculateComplexities(code string) ([]*Complexity, error) {
	f, err := myAst.ParseFile(code)
	if err != nil {
		return nil, err
	}

	var complexity = make([]*Complexity, 0)
	for _, target := range myAst.FindFuncDecls(f) {
		dimension, err := calculateComplexity(target)
		if err != nil {
			return nil, err
		}

		complexity = append(complexity, dimension)
	}

	return complexity, nil
}

func CalculateComplexity(code string, funcName string) (*Complexity, error) {
	f, err := myAst.ParseFile(code)
	if err != nil {
		return nil, err
	}

	targetFunc, err := myAst.FindFuncDeclByName(f, funcName)
	if err != nil {
		return nil, err
	}

	return calculateComplexity(targetFunc)
}

func calculateComplexity(targetFunc *ast.FuncDecl) (*Complexity, error) {
	var complexity *Complexity = &Complexity{}

	complexity.Name = targetFunc.Name.Name

	complexity.cyclomatic = calculateCyclomaticComplexity(targetFunc)
	complexity.number_of_loops = countCycles(targetFunc)

	countNestedLoops, maxDepth := countNestedLoops(targetFunc)
	complexity.number_of_nested_loops = countNestedLoops
	complexity.maximum_nesting_level_of_loops = maxDepth

	return complexity, nil
}

func calculateCyclomaticComplexity(f *ast.FuncDecl) int {
	var complexity int = 1

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt:
			complexity++
		case *ast.RangeStmt:
			complexity++
		case *ast.SwitchStmt:
			complexity += len(x.Body.List)
		case *ast.TypeSwitchStmt:
			complexity += len(x.Body.List)
		case *ast.BinaryExpr:
			if x.Op == token.LAND || x.Op == token.LOR {
				complexity++
			}
		}
		return true
	})

	return complexity
}

func countCycles(f *ast.FuncDecl) int {
	var cycleCount int

	ast.Inspect(f, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			cycleCount++
		}
		return true
	})

	return cycleCount
}

func countNestedLoops(f *ast.FuncDecl) (int, int) {
	nestedLoopCount := 0
	maxDepth := 0
	currentDepth := 0

	myAst.Inspect(f,
		func(node ast.Node) {
			switch node.(type) {
			case *ast.RangeStmt, *ast.ForStmt:
				nestedLoopCount += currentDepth

				currentDepth++
			}
		},
		func(node ast.Node) bool {
			return true
		},
		func(node ast.Node) {
			switch node.(type) {
			case *ast.RangeStmt, *ast.ForStmt:
				maxDepth = max(maxDepth, currentDepth)
				currentDepth--
			}
		})

	return nestedLoopCount, maxDepth
}
