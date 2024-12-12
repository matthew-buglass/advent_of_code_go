package main

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

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

func parseInputToInt(input string) [][]int {
	rows := strings.Split(input, "\n")

	elems := make([][]int, 0)
	for _, row := range rows {
		elems = append(elems, convertStrArrToIntArr(strings.Split(row, ",")))
	}

	return elems
}

func powInt(x, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

func waitAndClose(channel chan any, wg *sync.WaitGroup) {
	defer close(channel)
	wg.Wait()
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

func parseInputSymbolsAndLocations(input string) (symbolToLocation map[string][][]int, spaceBounds []int) {
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
