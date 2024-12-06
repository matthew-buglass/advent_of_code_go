package main

import (
	"fmt"
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

func intAbsDiff(a int, b int) int {
	if a < b {
		return b - a
	} else {
		return a - b
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

func calculateDistance(start []int, end []int) int {
	return intAbsDiff(start[0], end[0]) + intAbsDiff(start[1], end[1])
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
			pathLeg := PathLeg{
				src:    append(make([]int, 0), currentLocation...),
				dst:    append(make([]int, 0), dstLocation...),
				travel: append(make([][]int, 0), locationsTraversed...),
			}
			srcToDstMap[getKey(currentLocation, arrivalVector)] = pathLeg
			currentLocation[0] += departureVector[0]
			currentLocation[1] += departureVector[1]
		}
	}

	channel <- srcToDstMap
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	startPosition, direction, spaceBounds, obstructions := parseInput(input)
	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		return "not implemented"
	}
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

	// Get the lookup keys for the obstructions
	obsDir := []int{0, 0}
	obsKeys := make([]string, len(obstructions))
	for _, obs := range obstructions {
		obsKeys = append(obsKeys, getKey(obs, obsDir))
	}

	// Build the location to destination lookup map
	subMapChannel := make(chan map[string]PathLeg)
	for _, obs := range obstructions[:1] {
		go findLocationsAndDirectionsToObstruction(
			obs,
			obsKeys,
			directions,
			subMapChannel,
			inBoundsFunc)
	}

	srcToDstMap := make(map[string]PathLeg)
	for i := 0; i < len(obstructions[:1]); i++ {
		obsMap := <-subMapChannel
		for k, v := range obsMap {
			srcToDstMap[k] = v
		}
	}

	// Find the path
	currPosition := append(make([]int, 0), startPosition...)
	locations := make([][]int, 0)
	path := append(make([][]int, 0), currPosition)
	for inBoundsFunc(currPosition) {
		// Get next position and add to path
		pathLeg := srcToDstMap[getKey(currPosition, direction)]
		fmt.Println(pathLeg)
		for _, location := range pathLeg.travel {
			if !slices.Contains(locations, location) {

			}
			locations = append(locations, location)
		}
		newPosition := pathLeg.dst

		if newPosition != nil {
			path = append(path, newPosition)
			fmt.Println(currPosition, newPosition, path)

			// Set new variables
			currPosition = newPosition
			direction = turnMap[fmt.Sprintf("%d,%d", direction[0], direction[1])]
		} else { // Now would be out of bounds
			fmt.Println(direction)
			newPosition := append(make([]int, 0), currPosition...)
			for inBoundsFunc(newPosition) { // find out puzzle exit
				newPosition[0] += direction[0]
				newPosition[1] += direction[1]
			}
			path = append(path, newPosition)
			fmt.Println(direction, path)
			break
		}
	}

	// Find the the locations traveled between parts of the path
	// locationsChannel := make(chan [][]int)
	// uniqueLocations := make()
	return 42
}
