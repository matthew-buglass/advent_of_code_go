package main

import (
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
}

func getLeadingIndices(matchPairs [][]int) []int {
	if matchPairs == nil {
		return nil
	}
	firstIndices := make([]int, 0, len(matchPairs))
	for _, matchPair := range matchPairs {
		firstIndices = append(firstIndices, matchPair[0])
	}
	return firstIndices
}

func convertStrArrToIntArr(strArr []string) []int {
	intArr := make([]int, 0, len(strArr))
	for _, str := range strArr {
		i, err := strconv.Atoi(str)
		if err != nil {
			intArr = append(intArr, -1)
			// return intArr
		} else {
			intArr = append(intArr, i)
		}

	}
	return intArr
}

func getCoordinateFromIdx(idx int, rowLength int) []int {
	return []int{idx / rowLength, idx % rowLength}
}

func parseInput(input string) (problemArr [][]int, trailHeads [][]int, spaceBounds []int) {
	rows := strings.Split(input, "\n")
	spaceBounds = []int{len(rows) - 1, len(rows[0]) - 1}
	strippedInput := strings.ReplaceAll(input, "\n", "")

	// Get the trail head indices
	trailHeadRe := regexp.MustCompile(`0`)
	trailHeadIdx := getLeadingIndices(trailHeadRe.FindAllStringSubmatchIndex(strippedInput, -1))
	for _, idx := range trailHeadIdx {
		trailHeads = append(trailHeads, getCoordinateFromIdx(idx, spaceBounds[1]+1))
	}

	// Convert the puzzel input into a 2D slice
	for _, row := range rows {
		elems := convertStrArrToIntArr(strings.Split(row, ""))
		problemArr = append(problemArr, elems)
	}

	return problemArr, trailHeads, spaceBounds
}

func getTranslation(src []int, transVect []int) []int {
	return []int{src[0] + transVect[0], src[1] + transVect[1]}
}

func getValueFromCoords(puzzleMap *[][]int, position []int, inBoundsFunc func([]int) bool) (int, error) {
	if inBoundsFunc(position) {
		return (*puzzleMap)[position[0]][position[1]], nil
	} else {
		return -1, errors.New("position not in bounds")
	}
}

type TrailResult struct {
	startI int
	startJ int
	endI   int
	endJ   int
}

func (r TrailResult) deepCopy() TrailResult {
	return TrailResult{
		startI: r.startI,
		startJ: r.startJ,
		endI:   r.endI,
		endJ:   r.endJ,
	}
}

func findTrailScore(startingPos []int, puzzleMap *[][]int, countChannel chan TrailResult, waitGroup *sync.WaitGroup, inBoundsFunc func([]int) bool, resultTemplate *TrailResult, nilResult TrailResult) {
	defer waitGroup.Done()

	currentPos := append(make([]int, 0, 2), startingPos...)
	trailValue, err := getValueFromCoords(puzzleMap, currentPos, inBoundsFunc)

	if err != nil {
		countChannel <- nilResult
		return
	}

	for trailValue < 9 {
		// Get hte possible forks in the road
		trailsForward := make([][]int, 0, 3)

		leftPos := getTranslation(currentPos, []int{0, -1})
		rightPos := getTranslation(currentPos, []int{0, 1})
		upPos := getTranslation(currentPos, []int{-1, 0})
		downPos := getTranslation(currentPos, []int{1, 0})

		leftVal, leftErr := getValueFromCoords(puzzleMap, leftPos, inBoundsFunc)
		rightVal, rightErr := getValueFromCoords(puzzleMap, rightPos, inBoundsFunc)
		upVal, upErr := getValueFromCoords(puzzleMap, upPos, inBoundsFunc)
		downVal, downErr := getValueFromCoords(puzzleMap, downPos, inBoundsFunc)

		if leftErr == nil && leftVal-trailValue == 1 {
			trailsForward = append(trailsForward, leftPos)
		}
		if rightErr == nil && rightVal-trailValue == 1 {
			trailsForward = append(trailsForward, rightPos)
		}
		if upErr == nil && upVal-trailValue == 1 {
			trailsForward = append(trailsForward, upPos)
		}
		if downErr == nil && downVal-trailValue == 1 {
			trailsForward = append(trailsForward, downPos)
		}

		// walk forward
		switch len(trailsForward) {
		case 0:
			countChannel <- nilResult
			return
		case 1:
			currentPos = trailsForward[0]
		default:
			currentPos = trailsForward[0]

			// queue sub tasks
			otherTrails := trailsForward[1:]
			for _, pos := range otherTrails {
				waitGroup.Add(1)
				go findTrailScore(pos, puzzleMap, countChannel, waitGroup, inBoundsFunc, resultTemplate, nilResult)
			}
		}

		// set variables for next iteration. If we are already here, we know that we are within the map
		trailValue, _ = getValueFromCoords(puzzleMap, currentPos, inBoundsFunc)
	}
	// if we make it here, we found the top
	result := (*resultTemplate).deepCopy()
	result.endI = currentPos[0]
	result.endJ = currentPos[1]
	countChannel <- result
}

func findTrailScore2(startingPos []int, puzzleMap *[][]int, countChannel chan int, waitGroup *sync.WaitGroup, inBoundsFunc func([]int) bool) {
	defer waitGroup.Done()
	currentPos := append(make([]int, 0, 2), startingPos...)
	trailValue, err := getValueFromCoords(puzzleMap, currentPos, inBoundsFunc)

	if err != nil {
		countChannel <- 0
		return
	}

	for trailValue < 9 {
		// Get hte possible forks in the road
		trailsForward := make([][]int, 0, 3)

		leftPos := getTranslation(currentPos, []int{0, -1})
		rightPos := getTranslation(currentPos, []int{0, 1})
		upPos := getTranslation(currentPos, []int{-1, 0})
		downPos := getTranslation(currentPos, []int{1, 0})

		leftVal, leftErr := getValueFromCoords(puzzleMap, leftPos, inBoundsFunc)
		rightVal, rightErr := getValueFromCoords(puzzleMap, rightPos, inBoundsFunc)
		upVal, upErr := getValueFromCoords(puzzleMap, upPos, inBoundsFunc)
		downVal, downErr := getValueFromCoords(puzzleMap, downPos, inBoundsFunc)

		if leftErr == nil && leftVal-trailValue == 1 {
			trailsForward = append(trailsForward, leftPos)
		}
		if rightErr == nil && rightVal-trailValue == 1 {
			trailsForward = append(trailsForward, rightPos)
		}
		if upErr == nil && upVal-trailValue == 1 {
			trailsForward = append(trailsForward, upPos)
		}
		if downErr == nil && downVal-trailValue == 1 {
			trailsForward = append(trailsForward, downPos)
		}

		// walk forward
		switch len(trailsForward) {
		case 0:
			countChannel <- 0
			return
		case 1:
			currentPos = trailsForward[0]
		default:
			currentPos = trailsForward[0]

			// queue sub tasks
			otherTrails := trailsForward[1:]
			waitGroup.Add(len(otherTrails))
			for _, pos := range otherTrails {
				go findTrailScore2(pos, puzzleMap, countChannel, waitGroup, inBoundsFunc)
			}
		}

		// set variables for next iteration. If we are already here, we know that we are within the map
		trailValue, _ = getValueFromCoords(puzzleMap, currentPos, inBoundsFunc)
	}
	// if we make it here, we found the top
	countChannel <- 1
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	problemArr, trailHeads, spaceBounds := parseInput(input)
	inBoundsFunc := func(pos []int) bool {
		return 0 <= pos[0] && pos[0] <= spaceBounds[0] && 0 <= pos[1] && pos[1] <= spaceBounds[1]
	}

	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		var wg sync.WaitGroup
		wg.Add(len(trailHeads))
		resChannel := make(chan int)
		for _, startPosition := range trailHeads {
			go findTrailScore2(startPosition, &problemArr, resChannel, &wg, inBoundsFunc)
		}

		// Close the channel once all the tasks are finished
		go func() {
			defer close(resChannel)
			wg.Wait()
		}()

		numResultsReceived := 0
		totalScore := 0
		for result := range resChannel {
			numResultsReceived++
			totalScore += result
		}
		return totalScore
	}
	// solve part 1 here
	var wg sync.WaitGroup
	resChannel := make(chan TrailResult)
	nilResult := TrailResult{-1, -1, -1, -1}
	wg.Add(len(trailHeads))
	for _, startPosition := range trailHeads {
		trailResultTemplate := TrailResult{startI: startPosition[0], startJ: startPosition[1]}
		go findTrailScore(startPosition, &problemArr, resChannel, &wg, inBoundsFunc, &trailResultTemplate, nilResult)
	}

	// Close the channel once all the tasks are finished
	go func() {
		defer close(resChannel)
		wg.Wait()
	}()

	numResultsReceived := 0
	results := make([]TrailResult, 0)
	totalScore := 0
	for result := range resChannel {
		numResultsReceived++

		if result != nilResult && !slices.Contains(results, result) {
			totalScore += 1
			results = append(results, result)
		}
	}
	return totalScore
}
