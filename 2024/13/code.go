package main

import (
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/jpillora/puzzler/harness/aoc"
	"gonum.org/v1/gonum/mat"
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

func waitAndClose(channel chan int, wg *sync.WaitGroup) {
	defer close(channel)
	wg.Wait()
}

func main() {
	if os.Getenv("DEBUG") == "1" {
		b1, _ := readInput("input-example")
		b2, _ := readInput("input-user")
		input1 := string(b1)
		input2 := string(b2)
		run(false, input1)
		run(false, input2)
		run(true, input1)
	}
	aoc.Harness(run)
}

type Button struct {
	x    int
	y    int
	cost int
}

type Prize struct {
	x int
	y int
}

type ClawGame struct {
	buttonA Button
	buttonB Button
	prize   Prize
}

func parseInput(input string) []ClawGame {
	subProblemStrings := strings.Split(input, "\n\n")
	clawGames := make([]ClawGame, 0, len(subProblemStrings))

	for _, subProb := range subProblemStrings {
		sections := strings.Split(subProb, "\n")
		buttonAString := strings.ReplaceAll(sections[0], " ", "")
		buttonBString := strings.ReplaceAll(sections[1], " ", "")
		prizeString := strings.ReplaceAll(sections[2], " ", "")

		// get buttons
		buttonACoords := strings.Split(strings.Split(buttonAString, ":")[1], ",")
		buttonAX, _ := strconv.Atoi(strings.Split(buttonACoords[0], "+")[1])
		buttonAY, _ := strconv.Atoi(strings.Split(buttonACoords[1], "+")[1])
		buttonA := Button{
			x:    buttonAX,
			y:    buttonAY,
			cost: 3,
		}

		buttonBCoords := strings.Split(strings.Split(buttonBString, ":")[1], ",")
		buttonBX, _ := strconv.Atoi(strings.Split(buttonBCoords[0], "+")[1])
		buttonBY, _ := strconv.Atoi(strings.Split(buttonBCoords[1], "+")[1])
		buttonB := Button{
			x:    buttonBX,
			y:    buttonBY,
			cost: 1,
		}

		prizeCoords := strings.Split(strings.Split(prizeString, ":")[1], ",")
		prizeX, _ := strconv.Atoi(strings.Split(prizeCoords[0], "=")[1])
		prizeY, _ := strconv.Atoi(strings.Split(prizeCoords[1], "=")[1])
		prize := Prize{
			x: prizeX,
			y: prizeY,
		}

		// build the game
		game := ClawGame{
			buttonA: buttonA,
			buttonB: buttonB,
			prize:   prize,
		}
		clawGames = append(clawGames, game)
	}

	return clawGames
}

func solveBiVariateEquations(clawGame ClawGame) []float64 {
	a := clawGame.buttonA
	b := clawGame.buttonB
	prize := clawGame.prize

	coefficients := mat.NewDense(2, 2, []float64{float64(a.x), float64(b.x), float64(a.y), float64(b.y)})
	constants := mat.NewVecDense(2, []float64{float64(prize.x), float64(prize.y)})

	// Solve the equations using the Solve function.
	var x mat.VecDense
	if err := x.SolveVec(coefficients, constants); err != nil {
		return nil
	}

	// Print the solution vector x.
	return x.RawVector().Data
}

func isIntSolution(coordinates []float64) ([]int, error) {
	result := make([]int, 0, len(coordinates))
	intThreshold := 0.0000000000001

	for _, coordinate := range coordinates {
		diff := coordinate - float64(int(coordinate))
		if math.Abs(diff) < intThreshold {
			result = append(result, int(coordinate))
		} else {
			return nil, errors.New("Not close enough to an integer")
		}
	}
	return result, nil
}

func findMinValue(clawGame ClawGame, channel chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	intersection := solveBiVariateEquations(clawGame)
	if intersection == nil {
		return
	}

	intArr, err := isIntSolution(intersection)
	if err == nil {
		result := clawGame.buttonA.cost*intArr[0] + clawGame.buttonB.cost*intArr[1]
		channel <- result
	}
}

func findMinValueSync(clawGame ClawGame) int {
	result := 0
	intersection := solveBiVariateEquations(clawGame)
	if intersection == nil {
		return result
	}

	intArr, err := isIntSolution(intersection)
	if err == nil {
		result = clawGame.buttonA.cost*intArr[0] + clawGame.buttonB.cost*intArr[1]
	}
	return result
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	// when you're ready to do part 2, remove this "not implemented" block
	clawGames := parseInput(input)

	if part2 {
		return "not implemented"
	}
	// solve part 1 here
	// This is a system of 2 bi-variate linear equations, to they will have 1 intersection points
	// We have a solution if they intersect at a positive integer coordinate. Becasue there is only
	// one possible solution, that will be the cheapest.
	if os.Getenv("DEBUG") == "1" {
		total := 0
		for _, game := range clawGames {
			total += findMinValueSync(game)
		}
		return total
	} else {
		var wg sync.WaitGroup
		resultChannel := make(chan int)
		for _, game := range clawGames {
			wg.Add(1)
			go findMinValue(game, resultChannel, &wg)
		}

		go waitAndClose(resultChannel, &wg)

		total := 0
		for result := range resultChannel {
			total += result
		}
		return total
	}
}
