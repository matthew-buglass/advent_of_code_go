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

func parseInput(input string) (symbolToLocation map[string][][]int, spaceBounds []int) {
	rows := strings.Split(input, "\n")
	spaceBounds = []int{len(rows) - 1, len(rows[0]) - 1}
	strippedInput := strings.ReplaceAll(input, "\n", "")

	antennaRe := regexp.MustCompile(`([A-z]|[0-9])`)

	antennaSymbols := antennaRe.FindAllString(strippedInput, -1)
	antennaIdx := getLeadingIndices(antennaRe.FindAllStringSubmatchIndex(strippedInput, -1))

	symbolToLocation = make(map[string][][]int, 0)
	for i, symbol := range antennaSymbols {
		offset := antennaIdx[i]
		antennaLocation := []int{offset / (spaceBounds[1] + 1), offset % (spaceBounds[1] + 1)}
		symbolToLocation[symbol] = append(symbolToLocation[symbol], antennaLocation)
	}

	return symbolToLocation, spaceBounds
}

func getVectorOfPoints(src []int, dst []int) []int {
	return []int{dst[0] - src[0], dst[1] - src[1]}
}

func getTranslation(src []int, transVect []int) []int {
	return []int{src[0] + transVect[0], src[1] + transVect[1]}
}

func getAntinodeLocations(antennaLocations [][]int, channel chan []int) {
	for i, locA := range antennaLocations[:len(antennaLocations)-1] {
		for _, locB := range antennaLocations[i+1:] {
			channel <- getTranslation(locA, getVectorOfPoints(locB, locA))
			channel <- getTranslation(locB, getVectorOfPoints(locA, locB))
		}
	}
	// send a poison pill to indicate we are done.
	channel <- nil
}

func getAllAntinodeLocations(antennaLocations [][]int, inBoundsFunction func([]int) bool, channel chan []int) {
	for i, locA := range antennaLocations[:len(antennaLocations)-1] {
		for _, locB := range antennaLocations[i+1:] {
			lastA := getTranslation(locA, getVectorOfPoints(locB, locA))
			lastB := getTranslation(locB, getVectorOfPoints(locA, locB))
			for inBoundsFunction(lastA) {
				channel <- lastA
				lastA = getTranslation(lastA, getVectorOfPoints(locB, locA))
			}
			for inBoundsFunction(lastB) {
				channel <- lastB
				lastB = getTranslation(lastB, getVectorOfPoints(locA, locB))
			}
		}
	}
	// send a poison pill to indicate we are done.
	channel <- nil
}

func locationToString(location []int) string {
	return fmt.Sprintf("%d,%d", location[0], location[1])
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	symbolToLocationsMap, spaceBounds := parseInput(input)

	inBoundsFunc := func(pos []int) bool {
		return 0 <= pos[0] && pos[0] <= spaceBounds[0] && 0 <= pos[1] && pos[1] <= spaceBounds[1]
	}

	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		trackedLocations := []string{}
		resultChannel := make(chan []int)
		numTasks := 0
		for _, v := range symbolToLocationsMap {
			go getAllAntinodeLocations(v, inBoundsFunc, resultChannel)
			for _, vi := range v {
				trackedLocations = append(trackedLocations, locationToString(vi))
			}
			numTasks++
		}

		numPoisonPills := 0
		for numPoisonPills < numTasks {
			result := <-resultChannel
			switch result {
			case nil:
				numPoisonPills += 1
			default:
				inBoundsFunc(result)
				resString := locationToString(result)
				if inBoundsFunc(result) &&
					!slices.Contains(trackedLocations, resString) {
					trackedLocations = append(trackedLocations, resString)
				}
			}
		}
		return len(trackedLocations)
	}

	// solve part 1 here
	resultChannel := make(chan []int)
	numTasks := 0
	for _, v := range symbolToLocationsMap {
		go getAntinodeLocations(v, resultChannel)
		numTasks++
	}

	numPoisonPills := 0
	trackedLocations := []string{}
	for numPoisonPills < numTasks {
		result := <-resultChannel
		switch result {
		case nil:
			numPoisonPills += 1
		default:
			inBoundsFunc(result)
			resString := locationToString(result)
			if inBoundsFunc(result) &&
				!slices.Contains(trackedLocations, resString) {
				trackedLocations = append(trackedLocations, resString)
			}
		}
	}
	return len(trackedLocations)
}
