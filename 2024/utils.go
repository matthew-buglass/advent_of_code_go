package main

import (
	"math"
	"strconv"
	"strings"
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
