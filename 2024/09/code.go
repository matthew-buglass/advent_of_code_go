package main

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jpillora/puzzler/harness/aoc"
)

func main() {
	aoc.Harness(run)
}

func intMin(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

type BlockPair struct {
	id         int
	usedSpace  int
	emptySpace int
}

func sortById(toSort []BlockPair) {
	slices.SortFunc(toSort, func(a BlockPair, b BlockPair) int { return cmp.Compare(a.id, b.id) })
}

func stringRepr(blockPairs []BlockPair) string {
	var builder strings.Builder
	for _, block := range blockPairs {
		for j := 0; j < block.usedSpace; j++ {
			builder.WriteString(fmt.Sprintf("%d", block.id))
			// builder.WriteString(fmt.Sprintf("(%d)", block.id))
		}
		for j := 0; j < block.emptySpace; j++ {
			builder.WriteString(".")
			// builder.WriteString("(.)")
		}
	}
	return builder.String()
}

func stringReprFromArray(arr []int) string {
	var builder strings.Builder
	for _, val := range arr {
		if val < 0 {
			builder.WriteString(".")
		} else {
			builder.WriteString(fmt.Sprintf("%v", val))
		}
	}
	return builder.String()
}

func decompressedArray(blockPairs []BlockPair) []int {
	decompressedMemory := make([]int, 0)
	for _, block := range blockPairs {
		for j := 0; j < block.usedSpace; j++ {
			decompressedMemory = append(decompressedMemory, block.id)
		}
		for j := 0; j < block.emptySpace; j++ {
			decompressedMemory = append(decompressedMemory, -1)
		}
	}
	return decompressedMemory
}

func parseInput(input string, startID int, channel chan BlockPair) {
	for i := 0; i < len(input); i += 2 {
		space, _ := strconv.Atoi(string(input[i]))
		empty := 0
		if i+1 < len(input) {
			empty, _ = strconv.Atoi(string(input[i+1]))
		}
		pair := BlockPair{
			id:         startID + (i / 2),
			usedSpace:  space,
			emptySpace: empty,
		}
		channel <- pair
	}
	// Send a poisson pill
	channel <- BlockPair{id: -1}
}

func optimizeStorage(decompressedArray []int) {
	frontPointer := 0
	backPointer := len(decompressedArray) - 1

	for frontPointer < backPointer {
		if decompressedArray[frontPointer] >= 0 {
			frontPointer++
		} else if decompressedArray[backPointer] < 0 {
			backPointer--
		} else {
			decompressedArray[frontPointer] = decompressedArray[backPointer]
			decompressedArray[backPointer] = -1
		}
	}
}

func printMemory(storagePairs []BlockPair) {
	fmt.Println(stringReprFromArray(decompressedArray(storagePairs)))
}

func moveForwardIndexIToJ(storagePairs []BlockPair, i int, j int) {
	itemToMove := storagePairs[i]
	for k := i; k > j; k-- {
		storagePairs[k] = storagePairs[k-1]
	}
	storagePairs[j] = itemToMove
}

func moveStorageBlock(storagePairs []BlockPair, frontPointer int, backPointer int) {
	frontVal := storagePairs[frontPointer]
	backVal := storagePairs[backPointer]
	backTotalSpace := backVal.usedSpace + backVal.emptySpace
	storagePairs[backPointer].emptySpace = frontVal.emptySpace - backVal.usedSpace
	storagePairs[frontPointer].emptySpace = 0
	moveForwardIndexIToJ(storagePairs, backPointer, frontPointer+1)
	storagePairs[backPointer].emptySpace += backTotalSpace
	// printMemory(storagePairs)
}

func optimizeStoragePairs(storagePairs []BlockPair) {
	// printMemory(storagePairs)
	// go from back to front
	backPointer := len(storagePairs) - 1
	frontPointer := 0
	for frontPointer < backPointer {
		for frontPointer < backPointer {
			// fmt.Println(frontPointer, backPointer, storagePairs[frontPointer].emptySpace, storagePairs[backPointer].usedSpace)
			if storagePairs[frontPointer].emptySpace >= storagePairs[backPointer].usedSpace {
				moveStorageBlock(storagePairs, frontPointer, backPointer)
				break
			}
			frontPointer++
		}
		if frontPointer == backPointer {
			backPointer--
		}
		frontPointer = 0
		// fmt.Println(frontPointer, backPointer)
	}

}

func calcCheckSum1(optimizedArray []int) int {
	checkSum := 0
	for i, val := range optimizedArray {
		if val < 0 {
			break
		} else {
			checkSum += val * i
		}
	}
	return checkSum
}

func calcCheckSum2(optimizedArray []int) int {
	checkSum := 0
	for i, val := range optimizedArray {
		if val >= 0 {
			checkSum += val * i
		}
	}
	return checkSum
}

// on code change, run will be executed 4 times:
// 1. with: false (part1), and example input
// 2. with: true (part2), and example input
// 3. with: false (part1), and user input
// 4. with: true (part2), and user input
// the return value of each run is printed to stdout
func run(part2 bool, input string) any {
	// Parse input
	parseChannel := make(chan BlockPair)
	numTasks := 0
	batchSize := 1
	for i := 0; i < len(input); i += batchSize * 2 {
		go parseInput(input[i:intMin(i+batchSize*2, len(input))], i/2, parseChannel)
		numTasks++
	}

	blockPairs := make([]BlockPair, 0)
	numPoison := 0
	for numPoison < numTasks {
		pair := <-parseChannel
		switch pair.id {
		case -1:
			numPoison++
		default:
			blockPairs = append(blockPairs, pair)
		}
	}
	sortById(blockPairs)

	// when you're ready to do part 2, remove this "not implemented" block
	if part2 {
		optimizeStoragePairs(blockPairs)
		decompressed := decompressedArray(blockPairs)
		return calcCheckSum2(decompressed)
	}

	// solve part 1 here
	decompressed := decompressedArray(blockPairs)
	optimizeStorage(decompressed)

	return calcCheckSum1(decompressed)
}
