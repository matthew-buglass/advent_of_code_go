package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
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
		return "not implemented"
	}
	// solve part 1 here
	re := regexp.MustCompile("[0-9]")
	list := strings.Split(input, "\n")
	out_sum := 0
	for _, line := range list {
		res := re.FindAllString(line, -1)
		calibrationStr := fmt.Sprintf("%s%s", res[0], res[len(res)-1])
		calibrationNum, err := strconv.Atoi(calibrationStr)
		if err == nil {
			// fmt.Println(calibrationNum)
			out_sum += calibrationNum
		}
	}
	return out_sum
}
