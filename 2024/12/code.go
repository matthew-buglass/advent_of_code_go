package main

import (
	"regexp"
	"slices"
	"strings"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
}

// General Functions
func intAbs(num int) int {
	if num < 0 {
		return -1 * num
	}
	return num
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

func parseInput(input string) (plotRuneToLocations map[string][][]int) {
	rows := strings.Split(input, "\n")
	uniqueCharRe := regexp.MustCompile(`.`)
	spaceBounds := []int{len(rows) - 1, len(rows[0]) - 1}
	strippedInput := strings.ReplaceAll(input, "\n", "")

	plotRunes := uniqueCharRe.FindAllString(strippedInput, -1)
	plotRunesIdx := getLeadingIndices(uniqueCharRe.FindAllStringSubmatchIndex(strippedInput, -1))

	plotRuneToLocations = make(map[string][][]int, 0)
	for i, rune := range plotRunes {
		offset := plotRunesIdx[i]
		antennaLocation := []int{offset / (spaceBounds[1] + 1), offset % (spaceBounds[1] + 1)}
		plotRuneToLocations[rune] = append(plotRuneToLocations[rune], antennaLocation)
	}

	return plotRuneToLocations
}

// Data structures and methods
type GardenPlot struct {
	gardenRune    string
	adjacentPlots []*GardenPlot
	perimeter     int
	area          int
	i             int
	j             int
}

type GardenRegion struct {
	gardenRune  string
	gardenPlots []*GardenPlot
	edgePlots   []*GardenPlot
	perimeter   int
	area        int
}

func (p *GardenPlot) calculatePerimeter() {
	p.perimeter = 4 - len(p.adjacentPlots)
}

func (p *GardenPlot) calculateArea() {
	p.area = 1
}

func (p *GardenPlot) isEdge() bool {
	return p.perimeter != 0
}

func markAdjacent(src *GardenPlot, dst *GardenPlot) {
	src.adjacentPlots = append(src.adjacentPlots, dst)
	dst.adjacentPlots = append(dst.adjacentPlots, src)

	src.calculatePerimeter()
	dst.calculatePerimeter()
}

func (r *GardenRegion) calculatePerimeter() {
	r.perimeter = 0
	for _, plot := range r.gardenPlots {
		r.perimeter += plot.perimeter
	}
}

func (r *GardenRegion) calculateArea() {
	r.area = len(r.gardenPlots)
}

func (r *GardenRegion) getAdjacentPlots(plot *GardenPlot) []*GardenPlot {
	adjPlots := make([]*GardenPlot, 0, 4)
	for _, edgePlot := range r.edgePlots {
		if areAdjacent(edgePlot, plot) {
			adjPlots = append(adjPlots, plot)
		}
	}
	return adjPlots
}

func (r *GardenRegion) addPlot(plot *GardenPlot) {
	for _, edgePlot := range r.getAdjacentPlots(plot) {
		if areAdjacent(edgePlot, plot) {
			markAdjacent(edgePlot, plot)
			if !edgePlot.isEdge() {
				r.removeEdge(edgePlot)
			}
		}
	}
}

func (r *GardenRegion) removeEdge(plot *GardenPlot) {
	edgeIdx := slices.Index(r.edgePlots, plot)
	r.edgePlots = append(r.edgePlots[:edgeIdx], r.edgePlots[edgeIdx+1:]...)
}

func (r *GardenRegion) isAdjacent(plot *GardenPlot) bool {
	isAdj := false
	for _, edgePlot := range r.edgePlots {
		isAdj = isAdj || areAdjacent(edgePlot, plot)
		if isAdj {
			return isAdj
		}
	}
	return isAdj
}

func areAdjacent(plotA *GardenPlot, plotB *GardenPlot) bool {
	return intAbs(plotA.i-plotB.i)+intAbs(plotA.j-plotB.j) == 1
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
	return 42
}
