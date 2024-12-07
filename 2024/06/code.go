package main

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
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

func intMin(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
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

// func calculateDistance(start []int, end []int) int {
// 	return intAbsDiff(start[0], end[0]) + intAbsDiff(start[1], end[1])
// }

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

func isLoop(startPosition []int, direction []int, turnMap map[string][]int, inBoundsFunc func([]int) bool, obstructionStrings []string, previousTurnLocations []string, channel chan int) {
	// Setup
	currPos := append(make([]int, 0), startPosition...)
	currDirection := append(make([]int, 0), direction...)

	uniqueTurnLocationStrings := append(make([]string, 0), previousTurnLocations...)
	for inBoundsFunc(currPos) {
		nextPos := advance(currPos, currDirection)
		if slices.Contains(uniqueTurnLocationStrings, getKey(currPos, currDirection)) {
			// If we have already turned here, we are about to enter a loop
			fmt.Println("    loop", currPos, obstructionStrings, uniqueTurnLocationStrings)
			channel <- 1
			return
		} else if slices.Contains(obstructionStrings, locationToString(nextPos)) {
			// rotate until we can advance
			for slices.Contains(obstructionStrings, locationToString(nextPos)) {
				if !slices.Contains(uniqueTurnLocationStrings, getKey(currPos, currDirection)) {
					uniqueTurnLocationStrings = append(uniqueTurnLocationStrings, getKey(currPos, currDirection))
				}
				currDirection = turnMap[locationToString(currDirection)]
				nextPos = advance(currPos, currDirection)
			}
		}
		currPos = nextPos
	}
	fmt.Println("not loop", currPos, obstructionStrings, uniqueTurnLocationStrings)
	channel <- 0
}

func basicLoopPathFinding(startPosition []int, direction []int, turnMap map[string][]int, inBoundsFunc func([]int) bool, obstructions [][]int) int {
	// Setup
	currPos := append(make([]int, 0), startPosition...)
	currDirection := append(make([]int, 0), direction...)
	obstructionStrings := make([]string, 0)
	for _, obs := range obstructions {
		obstructionStrings = append(obstructionStrings, locationToString(obs))
	}

	// // At every step, if we find a location that we haven't been to before, we are going to launch a sub-problem to see
	// // if introducing a new obstruction would cause a loop
	subProblemChannel := make(chan int)
	numSubTasks := 0

	uniqueTurnLocationStrings := make([]string, 0)
	uniqueLocations := make([]string, 0)
	numLoops := 0
	for inBoundsFunc(currPos) {
		nextPos := advance(currPos, currDirection)
		if slices.Contains(obstructionStrings, locationToString(nextPos)) {
			// rotate until we can advance
			for slices.Contains(obstructionStrings, locationToString(nextPos)) {
				if !slices.Contains(uniqueTurnLocationStrings, getKey(currPos, currDirection)) {
					uniqueTurnLocationStrings = append(uniqueTurnLocationStrings, getKey(currPos, currDirection))
				}
				currDirection = turnMap[locationToString(currDirection)]
				nextPos = advance(currPos, currDirection)
			}
		} else {
			if !slices.Contains(uniqueLocations, locationToString(currPos)) {
				// If we haven't been here before, put a blocker in front and kick off a sub-task
				go isLoop(currPos, currDirection, turnMap, inBoundsFunc, append(obstructionStrings, locationToString(nextPos)), uniqueTurnLocationStrings, subProblemChannel)
				numSubTasks += 1

				// Mark that we have been here
				uniqueLocations = append(uniqueLocations, locationToString(currPos))
			}
		}

		currPos = nextPos
	}

	for i := 0; i < numSubTasks; i++ {
		numLoops += <-subProblemChannel
	}

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
		numLoops := basicLoopPathFinding(startPosition, direction, turnMap, inBoundsFunc, obstructions)
		return numLoops
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
		return len(locationsTraversed)
	}
}
