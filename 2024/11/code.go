package main

import (
	"cmp"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"sync"

	"github.com/jpillora/puzzler/harness/aoc"
	"golang.org/x/sync/errgroup"
)

func main() {
	aoc.Harness(run)
}

func waitAndClose(channel chan []int, wg *sync.WaitGroup) {
	defer close(channel)
	wg.Wait()
}

func waitAndCloseInt(channel chan int, eg *errgroup.Group) {
	defer close(channel)
	eg.Wait()
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

func blinkStoneMemoizedRecursive(stoneNumber int, stoneIndex int, currentSteps int, desiredSteps int, memoMap *sync.Map, countChannel chan int, eg *errgroup.Group) error {
	// Returns an array where the first number is the index order of the stones and the subsequent
	// elements are the results of blinking the stone.
	// defer wg.Done()

	newNumbers := append(make([]int, 0, 1), stoneNumber)

	// if we haven't spilt and we still need to convert numbers
	for len(newNumbers) < 2 && currentSteps < desiredSteps {
		numToBranch := newNumbers[0]
		memo, ok := memoMap.Load(numToBranch)
		if ok {
			newNumbers = memo.([]int)
		} else {
			if numToBranch == 0 {
				newNumbers = []int{1}
			} else if isEvenDigits(numToBranch) {
				stoneStringNumber := strconv.Itoa(numToBranch)
				halfIdx := len(stoneStringNumber) / 2
				num1, _ := strconv.Atoi(stoneStringNumber[:halfIdx])
				num2, _ := strconv.Atoi(stoneStringNumber[halfIdx:])
				newNumbers = []int{num1, num2}
			} else {
				newNumbers = []int{numToBranch * 2024}
			}
			memoMap.Store(numToBranch, newNumbers) // associate the memo with the most up to date version of the data
		}
		currentSteps++
	}
	// fmt.Println(currentSteps, stoneNumber, newNumbers)
	if currentSteps == desiredSteps {
		countChannel <- len(newNumbers)
	} else {
		// wg.Add(len(newNumbers))
		for _, num := range newNumbers {
			// Try asynchronously
			started := eg.TryGo(func() error {
				return blinkStoneMemoizedRecursive(num, stoneIndex, currentSteps, desiredSteps, memoMap, countChannel, eg)
			})
			// Otherwise go synchronous
			if !started {
				blinkStoneMemoizedRecursive(num, stoneIndex, currentSteps, desiredSteps, memoMap, countChannel, eg)
			}
			// go blinkStoneMemoizedRecursive(num, stoneIndex, currentSteps, desiredSteps, memoMap, countChannel, wg)
		}
	}
	return nil
}

type MemoIndex struct {
	src        int
	stepsToDST int
}

func getBlinkResult(stoneNumber int) []int {
	// Returns an array where the first number is the index order of the stones and the subsequent
	// elements are the results of blinking the stone.
	resultArray := []int{}
	if stoneNumber == 0 {
		resultArray = []int{1}
	} else if isEvenDigits(stoneNumber) {
		stoneStringNumber := strconv.Itoa(stoneNumber)
		halfIdx := len(stoneStringNumber) / 2
		num1, _ := strconv.Atoi(stoneStringNumber[:halfIdx])
		num2, _ := strconv.Atoi(stoneStringNumber[halfIdx:])
		resultArray = []int{num1, num2}
	} else {
		resultArray = []int{stoneNumber * 2024}
	}
	return resultArray
}

func blinkStoneMemoizedRecursive2(stoneNumber int, stepsRemaining int, memoMap *sync.Map, countChannel chan int, eg *errgroup.Group) error {
	// Returns an array where the first number is the index order of the stones and the subsequent
	// elements are the results of blinking the stone.

	// Find the furthest step that we have recorded
	stepsFound := stepsRemaining
	memo, ok := memoMap.Load(MemoIndex{src: stoneNumber, stepsToDST: stepsFound})
	for !ok && stepsFound > 0 {
		stepsFound--
		memo, ok = memoMap.Load(MemoIndex{src: stoneNumber, stepsToDST: stepsFound})
	}

	jumpTo := []int{stoneNumber}
	if ok {
		jumpTo = memo.([]int)
	}

	// if we found a jump to the end, send it to the channel and return
	if stepsFound == stepsRemaining {
		// fmt.Println(stoneNumber, stepsFound, jumpTo)
		countChannel <- len(jumpTo)
		return nil
	}

	// Current State:
	//		jumpTo: all the results that we have stored
	//		stepsFound: the number of steps it took to get there

	// find the next jump and memoize it
	stepsFound++
	nextJumps := []int{}
	for _, jump := range jumpTo {
		m, ok := memoMap.Load(MemoIndex{src: jump, stepsToDST: 1})
		if ok {
			nextMemo := m.([]int)
			nextJumps = append(nextJumps, nextMemo...)
		} else {
			newJumps := getBlinkResult(jump)
			memoMap.Store(MemoIndex{src: jump, stepsToDST: 1}, newJumps)
			nextJumps = append(nextJumps, newJumps...)
		}
	}
	memoMap.Store(MemoIndex{src: stoneNumber, stepsToDST: stepsFound}, nextJumps)

	// for each of the new jumps that we have calculated, find their descendants
	for _, j := range nextJumps {
		// blinkStoneMemoizedRecursive2(j, stepsRemaining-stepsFound, memoMap, countChannel, eg)
		// try asynchronous
		started := eg.TryGo(func() error {
			return blinkStoneMemoizedRecursive2(j, stepsRemaining-stepsFound, memoMap, countChannel, eg)
		})
		// Otherwise go synchronous
		if !started {
			blinkStoneMemoizedRecursive2(j, stepsRemaining-stepsFound, memoMap, countChannel, eg)
		}
	}
	return nil
}

func getMemoIndex(src int, stepsToDST int) string {
	return fmt.Sprintf("%d,%d", src, stepsToDST)
}

func blinkStoneMemoizedRecursiveSync(stoneNumber int, stepsRemaining int, memoMap *map[MemoIndex][]int) int {
	// Returns an array where the first number is the index order of the stones and the subsequent
	// elements are the results of blinking the stone.

	// Find the furthest step that we have recorded
	stepsFound := stepsRemaining
	memo, ok := (*memoMap)[MemoIndex{src: stoneNumber, stepsToDST: stepsFound}]
	for !ok && stepsFound > 0 {
		stepsFound--
		memo, ok = (*memoMap)[MemoIndex{src: stoneNumber, stepsToDST: stepsFound}]
	}

	jumpTo := []int{stoneNumber}
	if ok {
		jumpTo = memo
	}

	// if we found a jump to the end, send it to the channel and return
	if stepsFound == stepsRemaining {
		// fmt.Println(len(*memoMap))
		return len(jumpTo)
	}

	// Current State:
	//		jumpTo: all the results that we have stored
	//		stepsFound: the number of steps it took to get there

	// find the next jump and memoize it
	stepsFound++
	nextJumps := []int{}
	for _, jump := range jumpTo {
		m, ok := (*memoMap)[MemoIndex{src: jump, stepsToDST: 1}]
		if ok {
			nextJumps = append(nextJumps, m...)
		} else {
			newJumps := getBlinkResult(jump)
			(*memoMap)[MemoIndex{src: jump, stepsToDST: 1}] = newJumps
			nextJumps = append(nextJumps, newJumps...)
		}
	}
	(*memoMap)[MemoIndex{src: stoneNumber, stepsToDST: stepsFound}] = nextJumps

	// count the number of next jumps
	jumpsToCount := make(map[int]int)
	for _, j := range nextJumps {
		jumpsToCount[j] = jumpsToCount[j] + 1
	}

	// for each of the new jumps that we have calculated, find their descendants
	total := 0
	for j, count := range jumpsToCount {
		total += (count * blinkStoneMemoizedRecursiveSync(j, stepsRemaining-stepsFound, memoMap))
	}
	return total
}

func blinkStoneMemoizedRecursiveSync2(stoneNumber int, stepsRemaining int, memoMap *map[MemoIndex]int) int {
	// I got this from the reddit, but why is it so much faster? Is it because of the array allocation?

	// Find the furthest step that we have recorded
	memo, ok := (*memoMap)[MemoIndex{src: stoneNumber, stepsToDST: stepsRemaining}]
	value := 0
	if stepsRemaining == 0 {
		return 1
	} else if ok {
		return memo
	} else if stoneNumber == 0 {
		value = blinkStoneMemoizedRecursiveSync2(1, stepsRemaining-1, memoMap)
	} else if isEvenDigits(stoneNumber) {
		stoneStringNumber := strconv.Itoa(stoneNumber)
		halfIdx := len(stoneStringNumber) / 2
		num1, _ := strconv.Atoi(stoneStringNumber[:halfIdx])
		num2, _ := strconv.Atoi(stoneStringNumber[halfIdx:])
		value = blinkStoneMemoizedRecursiveSync2(num1, stepsRemaining-1, memoMap) + blinkStoneMemoizedRecursiveSync2(num2, stepsRemaining-1, memoMap)
	} else {
		value = blinkStoneMemoizedRecursiveSync2(stoneNumber*2024, stepsRemaining-1, memoMap)
	}
	(*memoMap)[MemoIndex{src: stoneNumber, stepsToDST: stepsRemaining}] = value
	return value
}

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
		// blinkMemo := sync.Map{}
		// memoMap := make(map[MemoIndex][]int, 0)
		memoMap := make(map[MemoIndex]int, 0)
		// var wg sync.WaitGroup
		// countLeafsChannel := make(chan int, 300)

		// eg := new(errgroup.Group)
		// eg.SetLimit(1000) // limit the number of go routines

		// queue sub tasks
		total := 0
		for _, stone := range stoneNumbers {

			// startTime := time.Now()
			// wg.Add(1)
			// eg.Go(func() error {
			// 	return blinkStoneMemoizedRecursive2(stone, numBlinks, &blinkMemo, countLeafsChannel, eg)
			// })
			// numStones := blinkStoneMemoizedRecursiveSync(stone, numBlinks, &memoMap)
			numStones := blinkStoneMemoizedRecursiveSync2(stone, numBlinks, &memoMap)
			total += numStones
			// fmt.Printf("Stone %v makes %v in %v\n", stone, numStones, time.Since(startTime))
		}
		return total

		// go waitAndCloseInt(countLeafsChannel, eg)

		// // read results
		// totalLeafs := 0
		// for result := range countLeafsChannel {
		// 	totalLeafs += result
		// }
		// return totalLeafs
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
