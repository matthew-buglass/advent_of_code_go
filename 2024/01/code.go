package main

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
}

func getLeftAndRightLists(input string) ([]int, []int) {
	re := regexp.MustCompile("[0-9]+")
	location_pairs := strings.Split(input, "\n")
	locations_left := []int{}
	locations_right := []int{}
	for _, line := range location_pairs {
		res := re.FindAllString(line, -1)
		val_left, _ := strconv.Atoi(res[0])
		val_right, _ := strconv.Atoi(res[1])
		locations_left = append(locations_left, val_left)
		locations_right = append(locations_right, val_right)
	}
	slices.SortFunc(locations_left, func(i, j int) int {
		return j - i
	})
	slices.SortFunc(locations_right, func(i, j int) int {
		return j - i
	})
	return locations_left, locations_right
}

func compare(a int, b int, channel chan int) {
	if a < b {
		channel <- b - a
	} else {
		channel <- a - b
	}
}

func buildFreqMap(input_list []int, channel chan map[int]int) {
	freqMap := map[int]int{}
	for _, v := range input_list {
		count, ok := freqMap[v]
		if ok {
			freqMap[v] = count + 1
		} else {
			freqMap[v] = 1
		}
	}
	channel <- freqMap
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	locations_left, locations_right := getLeftAndRightLists(input)
	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		leftMapChan := make(chan map[int]int)
		rightMapChan := make(chan map[int]int)
		go buildFreqMap(locations_left, leftMapChan)
		go buildFreqMap(locations_right, rightMapChan)

		// Get the frequency map
		freqMapLeft := <-leftMapChan
		freqMapRight := <-rightMapChan

		// get the total dist
		total_dist := 0
		for leftLocation, leftCount := range freqMapLeft {
			rightCount, ok := freqMapRight[leftLocation]
			if !ok {
				rightCount = 0
			}
			total_dist += leftLocation * leftCount * rightCount
		}
		return total_dist
	}

	// solve part 1 here
	result_channel := make(chan int)
	// Send to sub-tasks
	for i := 0; i < len(locations_left); i++ {
		go compare(locations_left[i], locations_right[i], result_channel)
	}

	// get from sub-tasks
	total_dist := 0
	for i := 0; i < len(locations_left); i++ {
		res := <-result_channel
		total_dist += res
	}

	// return the result
	return total_dist
}
