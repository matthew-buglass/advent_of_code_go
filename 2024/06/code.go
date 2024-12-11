package main

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jpillora/puzzler/harness/aoc"
)

var expectedObs []string

func main() {
	aoc.Harness(run)
}

var filename string

func appendAuditToFile(content string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		panic(err)
	}
}

func arrsEqual(a []string, b []string) bool {
	for i, value := range a {
		if !(value == b[i]) {
			return false
		}
	}
	return true
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

func parseInput(input string) (position []int, direction []int, spaceBounds []int, obstructions [][]int) {
	rows := strings.Split(input, "\n")
	spaceBounds = []int{len(rows) - 1, len(rows[0]) - 1}
	strippedInput := strings.ReplaceAll(input, "\n", "")

	startRe := regexp.MustCompile(`\^`)
	obstructionsRe := regexp.MustCompile("#")

	startIdx := getLeadingIndices(startRe.FindAllStringSubmatchIndex(strippedInput, -1))
	obstructionIdx := getLeadingIndices(obstructionsRe.FindAllStringSubmatchIndex(strippedInput, -1))

	position = []int{startIdx[0] / (spaceBounds[0] + 1), startIdx[0] % (spaceBounds[0] + 1)}
	direction = []int{-1, 0} // up
	for _, obsIdx := range obstructionIdx {
		obstructions = append(obstructions, []int{obsIdx / (spaceBounds[0] + 1), obsIdx % (spaceBounds[0] + 1)})
	}

	return position, direction, spaceBounds, obstructions
}

func waitAndClose(channel chan int, wg *sync.WaitGroup) {
	defer close(channel)
	wg.Wait()
}

func getKey(position []int, direction []int) string {
	return fmt.Sprintf("%d,%d|%d,%d", position[0], position[1], direction[0], direction[1])
}

type PathLeg struct {
	src    []int
	dst    []int
	travel [][]int
}

func findLocationsAndDirectionsToObstruction(
	obstruction []int,
	obstructionsKeys []string,
	directionVectors [][]int,
	channel chan map[string]PathLeg,
	inBoundsFunc func([]int) bool) {
	// for each direction (2d vector) that we could come to an obstruction from, build a map of how
	// we could get to that obstruction. We build a map that form a location and direction to where
	// they would end up.
	srcToDstMap := make(map[string]PathLeg)
	obsDir := []int{0, 0}

	for _, arrivalVector := range directionVectors {
		departureVector := []int{arrivalVector[0] * -1, arrivalVector[1] * -1}
		dstLocation := []int{obstruction[0] + departureVector[0], obstruction[1] + departureVector[1]}
		currentLocation := append(make([]int, 0), dstLocation...)
		locationsTraversed := make([][]int, 0)

		// Work backwards in this direction until we hit another obstruction or we leave the bounds of the problem
		for !slices.Contains(obstructionsKeys, getKey(currentLocation, obsDir)) && inBoundsFunc(currentLocation) {
			// fmt.Println(arrivalVector, departureVector, dstLocation, currentLocation)
			locationsTraversed = append([][]int{append(make([]int, 0), currentLocation...)}, locationsTraversed...)
			if !reflect.DeepEqual(currentLocation, dstLocation) {
				pathLeg := PathLeg{
					src:    append(make([]int, 0), currentLocation...),
					dst:    append(make([]int, 0), dstLocation...),
					travel: append(make([][]int, 0), locationsTraversed...),
				}
				srcToDstMap[getKey(currentLocation, arrivalVector)] = pathLeg
			}

			currentLocation[0] += departureVector[0]
			currentLocation[1] += departureVector[1]
		}
	}
	// fmt.Println("Found map")
	channel <- srcToDstMap
}

func getSrcToDstMap(obstructions [][]int, directions [][]int, inBoundsFunc func([]int) bool) map[string]PathLeg {
	// Get the lookup keys for the obstructions
	obsDir := []int{0, 0}
	obsKeys := make([]string, len(obstructions))
	for _, obs := range obstructions {
		obsKeys = append(obsKeys, getKey(obs, obsDir))
	}

	// Build the location to destination lookup map
	subMapChannel := make(chan map[string]PathLeg)
	for _, obs := range obstructions {
		go findLocationsAndDirectionsToObstruction(
			obs,
			obsKeys,
			directions,
			subMapChannel,
			inBoundsFunc)
	}

	srcToDstMap := make(map[string]PathLeg)
	for i := 0; i < len(obstructions); i++ {
		obsMap := <-subMapChannel
		for k, v := range obsMap {
			srcToDstMap[k] = v
		}
	}

	return srcToDstMap
}

func locationToString(location []int) string {
	return fmt.Sprintf("%d,%d", location[0], location[1])
}

func findThePath(startPosition []int, direction []int, inBoundsFunc func([]int) bool, srcToDstMap map[string]PathLeg, turnMap map[string][]int) [][]int {
	currPosition := append(make([]int, 0), startPosition...)
	locations := make([][]int, 0)
	locationStrings := make([]string, 0)
	for inBoundsFunc(currPosition) {
		// Get next position and add to path
		pathLeg := srcToDstMap[getKey(currPosition, direction)]
		for _, location := range pathLeg.travel {
			locationString := locationToString(location)
			if !slices.Contains(locationStrings, locationString) {
				locationStrings = append(locationStrings, locationString)
				locations = append(locations, location)
			}
		}
		newPosition := pathLeg.dst

		if newPosition != nil {
			// Set new variables
			currPosition = newPosition
			direction = turnMap[fmt.Sprintf("%d,%d", direction[0], direction[1])]
		} else { // Now would be out of bounds
			newPosition := append(make([]int, 0), currPosition...)
			for inBoundsFunc(newPosition) { // find out puzzle exit
				locationString := locationToString(newPosition)
				if !slices.Contains(locationStrings, locationString) {
					locationStrings = append(locationStrings, locationString)
					locations = append(locations, append(make([]int, 0), newPosition...))
				}
				newPosition[0] += direction[0]
				newPosition[1] += direction[1]
			}
			break
		}
	}
	return locations
}

func advance(currPos []int, stepVector []int) (nextPos []int) {
	return []int{currPos[0] + stepVector[0], currPos[1] + stepVector[1]}
}

func isLoop(startPosition []int, direction []int, turnMap *map[string][]int, inBoundsFunc func([]int) bool, obstructionStrings []string, previousTurnLocations []string, channel chan int, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	// Setup
	currPos := []int{startPosition[0], startPosition[1]}
	currDirection := []int{direction[0], direction[1]}
	uniqueTurnLocationStrings := append(make([]string, 0, len(previousTurnLocations)), previousTurnLocations...)

	for inBoundsFunc(currPos) {

		// if we have already been here from this direction, we are in a loop
		if slices.Contains(uniqueTurnLocationStrings, getKey(currPos, currDirection)) {
			channel <- 1
			// appendAuditToFile(fmt.Sprintf("%s\n", obstructionStrings[len(obstructionStrings)-1]))
			return
		}
		// otherwise, we need to look at the next direction
		nextPos := advance(currPos, currDirection)

		if slices.Contains(obstructionStrings, locationToString(nextPos)) {
			// if we are hitting an obstruction, record it and turn
			uniqueTurnLocationStrings = append(uniqueTurnLocationStrings, getKey(currPos, currDirection))
			currDirection = (*turnMap)[locationToString(currDirection)]
		} else {
			// if we aren't, walk forward
			currPos = nextPos
		}

	}
	channel <- 0
}

func concurrentLoopPathFinding(startPosition []int, direction []int, turnMap *map[string][]int, inBoundsFunc func([]int) bool, obstructionStrings []string) int {
	// Setup
	currPos := []int{startPosition[0], startPosition[1]}
	currDirection := []int{direction[0], direction[1]}

	// At every step, if we find a location that we haven't been to before, we are going to launch a sub-problem to see
	// if introducing a new obstruction would cause a loop
	subProblemChannel := make(chan int)
	var wg sync.WaitGroup

	uniqueTurnLocationStrings := make([]string, 0)
	addedBlockers := append(make([]string, 0), locationToString(startPosition))

	for inBoundsFunc(currPos) {
		nextPos := advance(currPos, currDirection)

		if slices.Contains(obstructionStrings, locationToString(nextPos)) {
			// if we are hitting an obstruction, record it and turn
			uniqueTurnLocationStrings = append(uniqueTurnLocationStrings, getKey(currPos, currDirection))
			currDirection = (*turnMap)[locationToString(currDirection)]
		} else {
			if !slices.Contains(addedBlockers, locationToString(nextPos)) {
				// If we haven't been here before from this, put a blocker here and kick off a sub-task
				if nextPos[0] == 10 && nextPos[1] == 31 {
					fmt.Println("Queueing")
				}
				newObsStrings := append(make([]string, 0), obstructionStrings...)
				newObsStrings = append(newObsStrings, locationToString(nextPos))
				wg.Add(1)
				go isLoop(currPos, currDirection, turnMap, inBoundsFunc, newObsStrings, uniqueTurnLocationStrings, subProblemChannel, &wg)

				// Mark that we have been here
				addedBlockers = append(addedBlockers, locationToString(currPos))
				// appendAuditToFile(fmt.Sprintf("%v\n", addedBlockers))
			}

			// advance
			currPos = nextPos
		}
	}

	go waitAndClose(subProblemChannel, &wg)

	numLoops := 0
	numResults := 0

	for result := range subProblemChannel {
		numLoops += result
		numResults++
	}
	fmt.Println(numResults)
	return numLoops
}

func basicPathFinding(startPosition []int, direction []int, turnMap map[string][]int, inBoundsFunc func([]int) bool, obstructions [][]int) []string {
	// Setup
	currPos := append(make([]int, 0), startPosition...)
	currDirection := append(make([]int, 0), direction...)
	obstructionStrings := make([]string, 0)
	for _, obs := range obstructions {
		obstructionStrings = append(obstructionStrings, locationToString(obs))
	}

	uniqueLocations := make([]string, 0)
	for inBoundsFunc(currPos) {
		if !slices.Contains(uniqueLocations, locationToString(currPos)) {
			uniqueLocations = append(uniqueLocations, locationToString(currPos))
		}

		nextPos := advance(currPos, currDirection)

		// if we would hit a wall, just rotate and don't advance
		if slices.Contains(obstructionStrings, locationToString(nextPos)) {
			currDirection = turnMap[locationToString(currDirection)]
			nextPos = currPos
		}

		currPos = nextPos
	}

	return uniqueLocations

}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	partMap := map[bool]int{
		true:  2,
		false: 1,
	}
	filename = fmt.Sprintf("audit_part-%d_time-%s", partMap[part2], time.Now())

	startPosition, direction, spaceBounds, obstructions := parseInput(input)

	// solve part 1 here
	// Variable setup
	directions := [][]int{
		{0, 1},  // right
		{1, 0},  // down
		{0, -1}, // left
		{-1, 0}, // up
	}
	turnMap := make(map[string][]int)
	turnMap["0,1"] = []int{1, 0}
	turnMap["1,0"] = []int{0, -1}
	turnMap["0,-1"] = []int{-1, 0}
	turnMap["-1,0"] = []int{0, 1}

	inBoundsFunc := func(pos []int) bool {
		return 0 <= pos[0] && pos[0] <= spaceBounds[0] && 0 <= pos[1] && pos[1] <= spaceBounds[1]
	}

	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		obstructionStrings := make([]string, 0, len(obstructions))
		for _, obs := range obstructions {
			obstructionStrings = append(obstructionStrings, locationToString(obs))
		}
		return concurrentLoopPathFinding(startPosition, direction, &turnMap, inBoundsFunc, obstructionStrings)

		// expectedObs = append(make([]string, 0, len(obstructionStrings)), obstructionStrings...)
		// resChannel := make(chan int)
		// // The answer is 1663
		// var wg sync.WaitGroup
		// wg.Add((spaceBounds[0] + 1) * (spaceBounds[1] + 1))

		// for i := 0; i < spaceBounds[0]+1; i++ {
		// 	for j := 0; j < spaceBounds[1]+1; j++ {
		// 		newObs := append(obstructionStrings, locationToString([]int{i, j}))
		// 		go isLoop(startPosition, direction, &turnMap, inBoundsFunc, newObs, []string{}, resChannel, &wg)
		// 	}
		// }

		// go waitAndClose(resChannel, &wg)

		// numLoops := 0
		// numResults := 0
		// for result := range resChannel {
		// 	numLoops += result
		// 	numResults++
		// }
		// return numLoops
	} else {
		// Before you ask, yes this is 100% overkill and slower than just doing a basic search synchronously
		// I am trying to break down problems so they can be parallelized, even if I shouldn't.
		// As a comparison:
		// 	run(async, input-example) returned in 741µs => 41
		// 	run(async, input-user) returned in 69ms => 4982
		// 	run(sync, input-example) returned in 52µs => 41
		// 	run(sync, input-user) returned in 31ms => 4982
		srcToDstMap := getSrcToDstMap(obstructions, directions, inBoundsFunc)

		// Find the path
		locationsTraversed := findThePath(startPosition, append(make([]int, 0), direction...), inBoundsFunc, srcToDstMap, turnMap)
		fmt.Println(locationsTraversed)
		return len(locationsTraversed)
	}
}
