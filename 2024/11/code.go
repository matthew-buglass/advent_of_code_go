package main

import (
	"cmp"
	"math"
	"regexp"
	"slices"
	"strconv"
	"sync"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
}

func waitAndClose(channel chan []int, wg *sync.WaitGroup) {
	defer close(channel)
	wg.Wait()
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

func parseInput(input string) []int {
	numberRe := regexp.MustCompile("[0-9]+")
	return convertStrArrToIntArr(numberRe.FindAllString(input, -1))
}

func isEvenDigits(number int) bool {
	return (int(math.Floor(math.Log10(float64(number))))+1)%2 == 0
}

func blinkStone(stoneNumber int, stoneIndex int, resultChannel chan []int, wg *sync.WaitGroup) {
	// Returns an array where the first number is the index order of the stones and the subsequent
	// elements are the results of blinking the stone.
	defer wg.Done()

	resultArray := []int{stoneIndex}
	if stoneNumber == 0 {
		resultArray = append(resultArray, 1)
	} else if isEvenDigits(stoneNumber) {
		stoneStringNumber := strconv.Itoa(stoneNumber)
		halfIdx := len(stoneStringNumber) / 2
		num1, _ := strconv.Atoi(stoneStringNumber[:halfIdx])
		num2, _ := strconv.Atoi(stoneStringNumber[halfIdx:])
		resultArray = append(resultArray, num1, num2)
	} else {
		resultArray = append(resultArray, stoneNumber*2024)
	}
	resultChannel <- resultArray
}

func blinkStoneMemoized(stoneNumber int, stoneIndex int, memoMap *sync.Map, resultChannel chan []int, wg *sync.WaitGroup) {
	// Returns an array where the first number is the index order of the stones and the subsequent
	// elements are the results of blinking the stone.
	defer wg.Done()

	resultArray := []int{stoneIndex}
	newNumbers := make([]int, 0, 2)
	memo, ok := memoMap.Load(stoneNumber)
	if ok {
		newNumbers = memo.([]int)
	} else {
		if stoneNumber == 0 {
			newNumbers = append(newNumbers, 1)
		} else if isEvenDigits(stoneNumber) {
			stoneStringNumber := strconv.Itoa(stoneNumber)
			halfIdx := len(stoneStringNumber) / 2
			num1, _ := strconv.Atoi(stoneStringNumber[:halfIdx])
			num2, _ := strconv.Atoi(stoneStringNumber[halfIdx:])
			newNumbers = append(newNumbers, num1, num2)
		} else {
			newNumbers = append(newNumbers, stoneNumber*2024)
		}
		memoMap.Store(stoneNumber, newNumbers)
	}
	resultArray = append(resultArray, newNumbers...)
	resultChannel <- resultArray
}

// func blinkStoneRecursive(stoneNumber int, stoneIndex int, currentSplits int, totalSplits int, resultChannel chan []int, wg *sync.WaitGroup) {
// 	// Returns an array where the first number is the index order of the stones and the subsequent
// 	// elements are the results of blinking the stone.
// 	defer wg.Done()

// 	resultArray := []int{stoneIndex}
// 	stoneStringNumber := strconv.Itoa(stoneNumber)
// 	if stoneNumber == 0 {
// 		resultArray = append(resultArray, 1)
// 	} else if len(stoneStringNumber)%2 == 0 {
// 		halfIdx := len(stoneStringNumber) / 2
// 		num1, _ := strconv.Atoi(stoneStringNumber[:halfIdx])
// 		num2, _ := strconv.Atoi(stoneStringNumber[halfIdx:])
// 		resultArray = append(resultArray, num1, num2)
// 	} else {
// 		resultArray = append(resultArray, stoneNumber*2024)
// 	}

// 	if currentSplits == totalSplits {
// 		resultChannel <- resultArray
// 	}

// }

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	stoneNumbers := parseInput(input)

	numBlinks := 25

	if part2 {
		numBlinks = 75
		blinkMemo := sync.Map{}
		for i := 0; i < numBlinks; i++ {
			var wg sync.WaitGroup
			resultChannel := make(chan []int)

			// queue sub tasks
			for j, stone := range stoneNumbers {
				wg.Add(1)
				go blinkStoneMemoized(stone, j, &blinkMemo, resultChannel, &wg)
			}

			go waitAndClose(resultChannel, &wg)

			// read results
			results := make([][]int, 0, len(stoneNumbers))
			for result := range resultChannel {
				results = append(results, result)
			}

			// sort by the index
			slices.SortStableFunc(results, func(a []int, b []int) int {
				return cmp.Compare(a[0], b[0])
			})

			// Build the results
			newStones := make([]int, 0, len(stoneNumbers)) // Allocate at least enough space for the last batch
			for _, stones := range results {
				newStones = append(newStones, stones[1:]...)
			}
			stoneNumbers = newStones
		}

		return len(stoneNumbers)
	}

	for i := 0; i < numBlinks; i++ {
		var wg sync.WaitGroup
		resultChannel := make(chan []int)

		// queue sub tasks
		for j, stone := range stoneNumbers {
			wg.Add(1)
			go blinkStone(stone, j, resultChannel, &wg)
		}

		go waitAndClose(resultChannel, &wg)

		// read results
		results := make([][]int, 0, len(stoneNumbers))
		for result := range resultChannel {
			results = append(results, result)
		}

		// sort by the index
		slices.SortStableFunc(results, func(a []int, b []int) int {
			return cmp.Compare(a[0], b[0])
		})

		// Build the results
		newStones := make([]int, 0)
		for _, stones := range results {
			newStones = append(newStones, stones[1:]...)
		}
		stoneNumbers = newStones
	}

	return len(stoneNumbers)
}
