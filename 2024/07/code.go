package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
}

func powInt(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

// iterator for function cartesian products
func cartesianProductPairs(a []func(a int, b int) int, b []func(a int, b int) int) (result [][]func(a int, b int) int) {
	for _, funcA := range a {
		for _, funcB := range b {
			temp := []func(a int, b int) int{funcA, funcB}
			result = append(result, temp)
		}
	}
	return result
}

func extendCartesianProductPairs(a [][]func(a int, b int) int, b []func(a int, b int) int) (result [][]func(a int, b int) int) {
	for _, funcA := range a {
		for _, funcB := range b {
			temp := append(make([]func(a int, b int) int, 0), funcA...)
			temp = append(temp, funcB)
			result = append(result, temp)
		}
	}
	return result
}

func genCartesianProducts(srcOpFuncArray []func(a int, b int) int, numInProd int) [][]func(a int, b int) int {
	// numToGenerate := powInt(len(srcOpFuncArray), numInProd)
	dstOpFuncArray := append(make([][]func(a int, b int) int, 0), srcOpFuncArray)
	for i := 1; i < numInProd; i++ {
		if i == 1 {
			dstOpFuncArray = cartesianProductPairs(srcOpFuncArray, srcOpFuncArray)
		} else {
			dstOpFuncArray = extendCartesianProductPairs(dstOpFuncArray, srcOpFuncArray)
		}
	}

	return dstOpFuncArray
}

// helper functions
type Problem struct {
	solution   int
	components []int
}

func convertStrArrToIntArr(strArr []string) []int {
	intArr := make([]int, 0, len(strArr))
	for _, str := range strArr {
		i, err := strconv.Atoi(str)
		if err != nil {
			return intArr
		}
		intArr = append(intArr, i)
	}
	return intArr
}

func parseInput(input string) []Problem {
	rows := strings.Split(input, "\n")

	elems := make([]Problem, 0)
	for _, row := range rows {
		splitBitz := strings.Split(row, ": ")
		sol, _ := strconv.Atoi(splitBitz[0])
		comps := convertStrArrToIntArr(strings.Split(splitBitz[1], " "))
		newProv := Problem{
			solution:   sol,
			components: comps,
		}
		elems = append(elems, newProv)
	}

	return elems
}

func canBeComputed(problem Problem, operatorFuncArray []func(a int, b int) int) int {
	operatorPerm := genCartesianProducts(operatorFuncArray, len(problem.components)-1)

	for _, operators := range operatorPerm {
		runningTotal := 0
		for i, operator := range operators {
			if i == 0 {
				runningTotal = operator(problem.components[i], problem.components[i+1])
			} else {
				runningTotal = operator(runningTotal, problem.components[i+1])
			}
		}
		if runningTotal == problem.solution {
			return runningTotal
		}
	}
	return 0
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	// Parse input
	problems := parseInput(input)

	// Variable setup
	multiply := func(a int, b int) int {
		return a * b
	}
	add := func(a int, b int) int {
		return a + b
	}
	operatorFuncArray := []func(a int, b int) int{
		add,
		multiply,
	}

	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		return "not implemented"
	}
	// solve part 1 here
	fmt.Println(problems)
	canBeComputed(problems[0], operatorFuncArray)
	return 42
}
