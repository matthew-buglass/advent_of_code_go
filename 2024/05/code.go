package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
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

func parseInput(input string) (rules []string, pageOrders [][]int) {
	problemHalves := strings.Split(input, "\n\n")
	rules = strings.Split(problemHalves[0], "\n")
	pageOrdersTemp := strings.Split(problemHalves[1], "\n")

	pageOrders = make([][]int, len(pageOrders))
	for _, order := range pageOrdersTemp {
		pageOrders = append(pageOrders, convertStrArrToIntArr(strings.Split(order, ",")))
	}

	return rules, pageOrders
}

func findMiddleValueIfValid(rules []string, pageOrder []int, channel chan int) {
	middleValue := 0

	if len(pageOrder)%2 == 0 {
		panic("Pages is even")
	}
	middleIndex := len(pageOrder) / 2 // 5 items, indexes 0, 1, 2, 3, 4. 5 / 2 = 2. 2 is middle index.

	for i, firstPage := range pageOrder {
		if i == middleIndex {
			middleValue = firstPage
		}
		for _, secondPage := range pageOrder[i:] {
			// Build the rule that we may be breaking
			breakingRule := fmt.Sprintf("%d|%d", secondPage, firstPage)
			// If we break a rule, send 0 and short circuit
			if slices.Contains(rules, breakingRule) {
				middleValue = 0
				channel <- middleValue
				return
			}
		}
	}

	// Send the middle value of a correct page order
	// fmt.Println(pageOrder, middleValue)
	channel <- middleValue
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	rules, pageOrders := parseInput(input)
	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		return "not implemented"
	}
	// solve part 1 here
	middleValueChannel := make(chan int)
	for _, order := range pageOrders {
		go findMiddleValueIfValid(rules, order, middleValueChannel)
	}

	sumOfMiddleValues := 0
	for i := 0; i < len(pageOrders); i++ {
		sumOfMiddleValues += <-middleValueChannel
	}
	return sumOfMiddleValues
}
