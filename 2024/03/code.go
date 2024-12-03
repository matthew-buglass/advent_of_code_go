package main

import (
	"regexp"
	"strconv"

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

func multiplyFromString(instructionString string, channel chan int) {
	numberRe := regexp.MustCompile("[0-9]{1,3}")
	numbers := convertStrArrToIntArr(numberRe.FindAllString(instructionString, -1)) // Will only be 2 numbers
	// fmt.Println(instructionString, numbers[0]*numbers[1])
	channel <- numbers[0] * numbers[1]
}

func getLeadingIndices(matchPairs [][]int) []int {
	firstIndices := make([]int, 0, len(matchPairs))
	for _, matchPair := range matchPairs {
		firstIndices = append(firstIndices, matchPair[0])
	}
	return firstIndices
}

func filterInactiveInstructions(instruction string, instructionIndex int, doIndices []int, dontIndices []int, channel chan string) {
	recentDoIndex := 0
	recentDontIndex := -1 // because do > don't the instructions are active

	// the first time we hit a do/don't instruction with a greater index than our multiply, get the instruction before it and break
	for _, doIdx := range doIndices {
		if doIdx > instructionIndex {
			break
		} else {
			recentDoIndex = doIdx
		}
	}
	for _, dontIdx := range dontIndices {
		if dontIdx > instructionIndex {
			break
		} else {
			recentDontIndex = dontIdx
		}
	}

	if recentDoIndex > recentDontIndex { // if we should do the instruction (do is more recent than don't) add the instruction to the channel
		channel <- instruction
	} else { // Other-wise add a no-op (we need to know how many instructions to pull so we need something in the channel)
		channel <- ""
	}
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		multRe := regexp.MustCompile(`mul\([0-9]{1,3},[0-9]{1,3}\)`)
		doRe := regexp.MustCompile(`do\(\)`)
		dontRe := regexp.MustCompile(`don\'t\(\)`)

		instructions := multRe.FindAllString(input, -1)
		instructionsIdx := getLeadingIndices(multRe.FindAllStringSubmatchIndex(input, -1))
		doIdx := getLeadingIndices(doRe.FindAllStringSubmatchIndex(input, -1))
		dontIdx := getLeadingIndices(dontRe.FindAllStringSubmatchIndex(input, -1))

		instructionChannel := make(chan string)
		for i, instruction := range instructions {
			go filterInactiveInstructions(instruction, instructionsIdx[i], doIdx, dontIdx, instructionChannel)
		}

		validInstructionCount := 0
		multiplyChannel := make(chan int)
		for i := 0; i < len(instructions); i++ {
			instruction := <-instructionChannel
			if !(instruction == "") {
				validInstructionCount += 1
				go multiplyFromString(instruction, multiplyChannel)
			}
		}

		sumProd := 0
		for i := 0; i < validInstructionCount; i++ {
			sumProd += <-multiplyChannel
		}
		return sumProd
	}
	// solve part 1 here
	// get the matching patterns
	re := regexp.MustCompile(`mul\([0-9]{1,3},[0-9]{1,3}\)`)
	instructions := re.FindAllString(input, -1)

	multiplyChannel := make(chan int)
	for _, instruction := range instructions {
		go multiplyFromString(instruction, multiplyChannel)
	}

	sumProd := 0
	for i := 0; i < len(instructions); i++ {
		sumProd += <-multiplyChannel
	}
	return sumProd
}
