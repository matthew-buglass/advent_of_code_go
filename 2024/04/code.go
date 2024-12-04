package main

import (
	"regexp"
	"strings"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
}

func parseInputIntoMatrix(input string) [][]string {
	re := regexp.MustCompile(".")
	rows := strings.Split(input, "\n")
	letters := make([][]string, 0, len(rows))
	for _, row := range rows {
		cols := re.FindAllString(row, -1)
		letters = append(letters, cols)
	}
	return letters
}

// This was to create a sliding window, but it couldn't create independent sub-problems and hence double counted columns and rows
// It did run in less than half the time though...if only it was correct
func countXMAS(strMatrix [][]string, channel chan int) {
	target := "XMAS"
	reverseTarget := "SAMX"

	// Horizontal
	var row1Builder strings.Builder
	var row2Builder strings.Builder
	var row3Builder strings.Builder
	var row4Builder strings.Builder

	// Vertical
	var col1Builder strings.Builder
	var col2Builder strings.Builder
	var col3Builder strings.Builder
	var col4Builder strings.Builder

	// Diagonal
	var diagLeftBuilder strings.Builder
	var diagRightBuilder strings.Builder

	// Add to the builders
	for i, row := range strMatrix {
		for j, letter := range row {
			// Horizontal
			switch i {
			case 0:
				row1Builder.WriteString(letter)
			case 1:
				row2Builder.WriteString(letter)
			case 2:
				row3Builder.WriteString(letter)
			case 3:
				row4Builder.WriteString(letter)
			}

			// Vertical
			switch j {
			case 0:
				col1Builder.WriteString(letter)
			case 1:
				col2Builder.WriteString(letter)
			case 2:
				col3Builder.WriteString(letter)
			case 3:
				col4Builder.WriteString(letter)
			}

			// Diagonal
			if i == j {
				diagLeftBuilder.WriteString(letter)
			}
			if i+j == 3 {
				diagRightBuilder.WriteString(letter)
			}
		}
	}

	// Count the occurrences
	stringsToCompare := []string{
		// Horizontal
		row1Builder.String(),
		row2Builder.String(),
		row3Builder.String(),
		row4Builder.String(),

		// Vertical
		col1Builder.String(),
		col2Builder.String(),
		col3Builder.String(),
		col4Builder.String(),

		// Diagonal
		diagLeftBuilder.String(),
		diagRightBuilder.String(),
	}
	totalCount := 0

	for _, str := range stringsToCompare {
		if str == target || str == reverseTarget {
			totalCount += 1
		}
	}
	// if totalCount > 0 {
	// 	fmt.Println(strMatrix)
	// 	fmt.Println(row1Builder.String())
	// 	fmt.Println(row2Builder.String())
	// 	fmt.Println(row3Builder.String())
	// 	fmt.Println(row4Builder.String())
	// 	fmt.Println(stringsToCompare)
	// 	fmt.Println(totalCount)
	// 	fmt.Println()
	// }
	channel <- totalCount
}

func isTargetString(targetString string, problemSpace [][]string, startRowIdx int, startColIdx int, direction []int, channel chan int) {
	colTick := direction[0]
	rowTick := direction[1]

	maxRowIdx := len(problemSpace) - 1
	maxColIdx := len(problemSpace[0]) - 1

	var stringBuilder strings.Builder
	for i := 0; i < len(targetString); i++ {
		colIdx := startColIdx + colTick*i
		rowIdx := startRowIdx + rowTick*i

		if colIdx >= 0 && colIdx <= maxColIdx && rowIdx >= 0 && rowIdx <= maxRowIdx {
			stringBuilder.WriteString(problemSpace[rowIdx][colIdx])
		} else {
			break
		}
	}

	resultString := stringBuilder.String()
	if resultString == targetString {
		channel <- 1
	} else {
		channel <- 0
	}
}

func countX_MAS(strMatrix [][]string, channel chan int) {
	target := "MAS"
	reverseTarget := "SAM"

	// Diagonal
	var diagLeftBuilder strings.Builder
	var diagRightBuilder strings.Builder

	// Add to the builders
	if strMatrix[1][1] == "A" {
		for i, row := range strMatrix {
			for j, letter := range row {
				// Diagonal
				if i == j {
					diagLeftBuilder.WriteString(letter)
				}
				if i+j == len(target)-1 {
					diagRightBuilder.WriteString(letter)
				}
			}
		}

		// Count the occurrences
		diagLeft := diagLeftBuilder.String()
		diagRight := diagRightBuilder.String()

		if (diagLeft == target || diagLeft == reverseTarget) && (diagRight == target || diagRight == reverseTarget) {
			channel <- 1
		} else {
			channel <- 0
		}
	} else {
		channel <- 0
	}
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	letterMatrix := parseInputIntoMatrix(input)
	height := len(letterMatrix)
	width := len(letterMatrix[0])
	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		countChannel := make(chan int)

		numRoutines := 0
		for i := 0; i < height-2; i++ {
			for j := 0; j < width-2; j++ {
				strMatrix := [][]string{
					letterMatrix[i][j : j+3],
					letterMatrix[i+1][j : j+3],
					letterMatrix[i+2][j : j+3],
				}
				go countX_MAS(strMatrix, countChannel)
				numRoutines += 1
			}
		}

		// Count
		totalCount := 0
		for i := 0; i < numRoutines; i++ {
			totalCount += <-countChannel
		}
		return totalCount
	}
	// solve part 1 here
	searchDirections := [][]int{
		{-1, -1},
		{-1, 0},
		{-1, 1},
		{0, -1},
		{0, 1},
		{1, -1},
		{1, 0},
		{1, 1},
	}

	// Get sliding windows of length
	countChannel := make(chan int)

	// numRoutines := 0
	// for i := 0; i < height-3; i++ {
	// 	for j := 0; j < width-3; j++ {
	// 		strMatrix := [][]string{
	// 			letterMatrix[i][j : j+4],
	// 			letterMatrix[i+1][j : j+4],
	// 			letterMatrix[i+2][j : j+4],
	// 			letterMatrix[i+3][j : j+4],
	// 		}
	// 		go countXMAS(strMatrix, countChannel)
	// 		numRoutines += 1
	// 	}
	// }

	numRoutines := 0
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			for _, dir := range searchDirections {
				go isTargetString("XMAS", letterMatrix, i, j, dir, countChannel)
				numRoutines += 1
			}

		}
	}

	// Count
	totalCount := 0
	for i := 0; i < numRoutines; i++ {
		totalCount += <-countChannel
	}
	return totalCount
}
