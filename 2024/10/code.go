package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

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

// func subTask(channel chan int) {
// 	// send diff poison pill
// 	channel <- 1
// }

// func testIntPointer(numPoisonPills *int, channel chan int) {
// 	// increment the number of poison pills we expect
// 	*numPoisonPills++
// 	go subTask(channel)

// 	// send poison pill
// 	channel <- 0
// }

func getValueFromCoords(puzzleMap *[][]int, position []int, inBoundsFunc func([]int) bool) (int, error) {
	if inBoundsFunc(position) {
		return (*puzzleMap)[position[0]][position[1]], nil
	} else {
		return -1, errors.New("Position not in bounds")
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

func (r TrailResult) equals(t TrailResult) bool {
	return r.startI == t.startI && r.startJ == t.startJ && r.endI == t.endI && r.endJ == t.endJ
}

func appendAuditToFile(filename string, content string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		panic(err)
	}
}

var auditFileName string = "audit.csv"

func findTrailScore(startingPos []int, puzzleMap *[][]int, countChannel chan TrailResult, numTasks *int, inBoundsFunc func([]int) bool, resultTemplate *TrailResult, nilResult TrailResult) {
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
			*numTasks += len(otherTrails)
			appendAuditToFile(auditFileName, fmt.Sprintf(
				"%d,%d,%v,%v\n",
				currentPos[0],
				currentPos[1],
				len(otherTrails),
				otherTrails,
			))
			for _, pos := range otherTrails {
				go findTrailScore(pos, puzzleMap, countChannel, numTasks, inBoundsFunc, resultTemplate, nilResult)
			}
		}

		// set variables for next iteration. If we are already here, we know that we are within the map
		trailValue, _ = getValueFromCoords(puzzleMap, currentPos, inBoundsFunc)
		// fmt.Println(currentPos, trailValue)
	}
	// if we make it here, we found the top
	result := (*resultTemplate).deepCopy()
	result.endI = currentPos[0]
	result.endJ = currentPos[1]
	countChannel <- result
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	auditFileName = "audit_3.csv"

	problemArr, trailHeads, spaceBounds := parseInput(input)
	inBoundsFunc := func(pos []int) bool {
		return 0 <= pos[0] && pos[0] <= spaceBounds[0] && 0 <= pos[1] && pos[1] <= spaceBounds[1]
	}

	// fmt.Println(problemArr, trailHeads, spaceBounds)
	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		return "not implemented"
	}
	// solve part 1 here
	appendAuditToFile(auditFileName, fmt.Sprintf(
		"%s,%s,%s,%s\n",
		"Direction I",
		"Direction J",
		"Num other trails",
		"Other trails",
	))

	tasksToExpect := 0
	resChannel := make(chan TrailResult)
	nilResult := TrailResult{-1, -1, -1, -1}
	for _, startPosition := range trailHeads {
		trailResultTemplate := TrailResult{startI: startPosition[0], startJ: startPosition[1]}
		go findTrailScore(startPosition, &problemArr, resChannel, &tasksToExpect, inBoundsFunc, &trailResultTemplate, nilResult)
		tasksToExpect++
	}

	numResultsReceived := 0
	results := make([]TrailResult, 0)
	totalScore := 0
	for numResultsReceived < tasksToExpect {
		result := <-resChannel
		numResultsReceived++

		if result != nilResult && !slices.Contains(results, result) {
			totalScore += 1
			results = append(results, result)
		}

		// fmt.Println(results)
		// fmt.Println(numResultsReceived, tasksToExpect)
	}
	fmt.Println(numResultsReceived, tasksToExpect)
	return totalScore
}
