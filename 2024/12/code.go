package main

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/jpillora/puzzler/harness/aoc"
)

func readInput(name string) ([]byte, bool) {
	b, err := os.ReadFile(name + ".txt")
	if err != nil {
		return nil, false
	}
	if len(b) == 0 {
		return nil, false
	}
	return b, true
}

func main() {
	if os.Getenv("DEBUG") == "1" {
		file := "input-example"
		b1, _ := readInput(file)
		input := string(b1)
		run(false, input)
		run(true, input)
	}
	aoc.Harness(run)
}

func waitAndClose(channel chan GardenRegion, wg *sync.WaitGroup) {
	defer close(channel)
	wg.Wait()
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

func parseInput(input string) (plotRuneToLocations map[string][]*GardenPlot) {
	rows := strings.Split(input, "\n")
	uniqueCharRe := regexp.MustCompile(`.`)
	spaceBounds := []int{len(rows) - 1, len(rows[0]) - 1}
	strippedInput := strings.ReplaceAll(input, "\n", "")

	plotRunes := uniqueCharRe.FindAllString(strippedInput, -1)
	plotRunesIdx := getLeadingIndices(uniqueCharRe.FindAllStringSubmatchIndex(strippedInput, -1))

	plotRuneToLocations = make(map[string][]*GardenPlot, 0)
	for i, plotRune := range plotRunes {
		offset := plotRunesIdx[i]
		plot := GardenPlot{
			gardenRune: plotRune,
			perimeter:  4,
			area:       1,
			i:          offset / (spaceBounds[1] + 1),
			j:          offset % (spaceBounds[1] + 1),
		}
		plotRuneToLocations[plotRune] = append(plotRuneToLocations[plotRune], &plot)
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
	fmt.Println(p.perimeter)
}

func (p *GardenPlot) calculateArea() {
	p.area = 1
}

func (p *GardenPlot) isEdge() bool {
	return p.perimeter != 0
}

func markAdjacent(src *GardenPlot, dst *GardenPlot) {
	(*src).adjacentPlots = append((*src).adjacentPlots, dst)
	(*dst).adjacentPlots = append((*dst).adjacentPlots, src)

	src.calculatePerimeter()
	dst.calculatePerimeter()
}

func (r *GardenRegion) calculatePerimeter() {
	r.perimeter = 0
	for _, plot := range r.edgePlots {
		r.perimeter += plot.perimeter
	}
}

func (r *GardenRegion) calculateArea() {
	r.area = len(r.gardenPlots)
	fmt.Println("num plots", r.area)
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
	fmt.Println(r.getAdjacentPlots(plot))
	for _, edgePlot := range r.getAdjacentPlots(plot) {
		fmt.Println("adjacent")
		markAdjacent(edgePlot, plot)
		// if !edgePlot.isEdge() {
		// 	r.removeEdge(edgePlot)
		// }
	}
	r.gardenPlots = append(r.gardenPlots, plot)
	if plot.isEdge() {
		r.edgePlots = append(r.edgePlots, plot)
	}
	r.calculateArea()
}

func (r *GardenRegion) removeEdge(plot *GardenPlot) {
	edgeIdx := slices.Index(r.edgePlots, plot)
	r.edgePlots = append(r.edgePlots[:edgeIdx], r.edgePlots[edgeIdx+1:]...)
}

func (r *GardenRegion) isAdjacent(plot *GardenPlot) bool {
	for _, edgePlot := range r.edgePlots {
		if areAdjacent(edgePlot, plot) {
			return true
		}
	}
	return false
}

func areAdjacent(plotA *GardenPlot, plotB *GardenPlot) bool {
	return intAbs(plotA.i-plotB.i)+intAbs(plotA.j-plotB.j) == 1
}

// Solver functions
func buildRegionsFromLikePlots(plotRune string, gardenPlots []*GardenPlot, wg *sync.WaitGroup, channel chan GardenRegion) {
	defer wg.Done()
	gardenRegions := make([]GardenRegion, 0)

	fmt.Println("finding regions for", plotRune)

	for _, plot := range gardenPlots {
		numAdded := 0
		for _, region := range gardenRegions {
			if region.isAdjacent(plot) {
				fmt.Println("pre added to region for", plotRune, len(region.gardenPlots), len(region.edgePlots))
				region.addPlot(plot)
				fmt.Println("post added to region for", plotRune, len(region.gardenPlots), len(region.edgePlots))
				numAdded++
			}
		}
		if numAdded > 1 {
			fmt.Println("Added plot more than once. This means there are plots we need to join")
		} else if numAdded == 0 { // need to make a new region
			fmt.Println("created new region for", plotRune)
			gardenRegions = append(gardenRegions, GardenRegion{
				gardenPlots: []*GardenPlot{plot},
				edgePlots:   []*GardenPlot{plot},
				gardenRune:  plotRune,
				perimeter:   4,
				area:        1,
			})
		}
	}

	fmt.Println("found number of regions for", len(gardenRegions), plotRune)
	for _, region := range gardenRegions {
		region.calculateArea()
		region.calculatePerimeter()
		channel <- region
	}
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	// when you're ready to do part 2, remove this "not implemented" block
	plotRuneToLocations := parseInput(input)

	if part2 {
		return "not implemented"
	}
	// solve part 1 here

	// Async vars
	var wg sync.WaitGroup
	regionChannel := make(chan GardenRegion)

	wg.Add(1)
	go buildRegionsFromLikePlots("O", plotRuneToLocations["O"], &wg, regionChannel)

	// for plotRune, gardenPlots := range plotRuneToLocations {
	// 	wg.Add(1)
	// 	buildRegionsFromLikePlots(plotRune, gardenPlots, &wg, regionChannel)
	// }

	// wait for the results
	go waitAndClose(regionChannel, &wg)

	for region := range regionChannel {
		fmt.Println(region)
	}

	return 42
}
