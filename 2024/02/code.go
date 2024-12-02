package main

import (
	"regexp"
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

func parseInput(input string) [][]int {
	// split the readings into arrays of levels
	re := regexp.MustCompile("[0-9]+")
	readings := strings.Split(input, "\n")
	allLevels := make([][]int, 0, len(readings))
	for _, reading := range readings {
		levels := re.FindAllString(reading, -1)
		allLevels = append(allLevels, convertStrArrToIntArr(levels))
	}
	return allLevels
}

func areLevelsSafeAsync(reading []int, channel chan int) {
	isDecreasing := true
	isIncreasing := true
	lastVal := reading[0]

	for i := 1; i < len(reading); i++ {
		currVal := reading[i]
		valDiff := lastVal - currVal
		isDecreasing = isDecreasing && valDiff > 0
		isIncreasing = isIncreasing && valDiff < 0
		inRange := (valDiff >= 1 && valDiff <= 3) || (valDiff >= -3 && valDiff <= -1)

		if (isDecreasing || isIncreasing) && inRange {
			lastVal = currVal
		} else {
			// short circuit
			channel <- 0
			return
		}
	}
	// if every row passed its check
	channel <- 1
}

func areLevelsSafePart2(reading []int, channel chan int) {
	isDecreasing := true
	isIncreasing := true
	lastVal := reading[0]

	for i := 1; i < len(reading); i++ {
		currVal := reading[i]
		valDiff := lastVal - currVal
		isDecreasing = isDecreasing && valDiff > 0
		isIncreasing = isIncreasing && valDiff < 0
		inRange := (valDiff >= 1 && valDiff <= 3) || (valDiff >= -3 && valDiff <= -1)

		if (isDecreasing || isIncreasing) && inRange {
			lastVal = currVal
		} else {
			// if we would fail here, try the combinations of dropping this and previous numbers
			subChannel := make(chan int)
			for j := 0; j < i+1; j++ {
				subReading := append(make([]int, 0, len(reading)), reading...)
				subReading = append(subReading[:j], subReading[j+1:]...)
				go areLevelsSafeAsync(subReading, subChannel)
			}
			// get the results
			numSafe := 0
			for j := 0; j < i+1; j++ {
				numSafe += <-subChannel
			}

			if numSafe > 0 {
				channel <- 1
				return
			}
			channel <- 0
			return
		}
	}
	// if every row passed its check
	channel <- 1
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	levelReadings := parseInput(input)

	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		// send to channels
		safetyChannel := make(chan int)
		for _, reading := range levelReadings {
			go areLevelsSafePart2(reading, safetyChannel)
		}

		// get the results
		numSafe := 0
		for i := 0; i < len(levelReadings); i++ {
			numSafe += <-safetyChannel
		}
		return numSafe
	}
	// solve part 1 here
	// send to channels
	safetyChannel := make(chan int)
	for _, reading := range levelReadings {
		go areLevelsSafeAsync(reading, safetyChannel)
	}

	// get the results
	numSafe := 0
	for i := 0; i < len(levelReadings); i++ {
		numSafe += <-safetyChannel
	}
	return numSafe

}
