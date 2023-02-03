package helpers

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Processes the user's input if it's a range
func ValidateInputRange(userRange string, pipelineCount int) ([]int, error) {
	// Ensure range is within number of pipelines retrieved
	s := strings.Split(userRange, "-")
	min, _ := strconv.Atoi(s[0])
	max, _ := strconv.Atoi(s[1])
	if min >= max {
		return nil, errors.New("invalid range provided")
	} else if max-min > pipelineCount {
		return nil, errors.New("range provided is larger than number of pipelines retrieved")
	}

	return createNumbers(min, max), nil
}

// Processes the user's input if it's a selection
func ValidateInputSelection(userSelection string, pipelineCount int) ([]int, error) {
	// Ensure user provided selection is valid (1,2,3 or 1, 2, 3)
	match, err := regexp.MatchString(`(\d+)(,\s*\d+)*`, userSelection)
	if err != nil {
		return nil, errors.New("selection format is invalid, must be '1,2,3' or '1, 2, 3'")
	}

	if match {
		// Ensure lower and upper values are within number of pipelines retrieved
		userSelection = strings.ReplaceAll(userSelection, " ", "")
		s := strings.Split(userSelection, ",")
		sort.Strings(s)
		min, _ := strconv.Atoi(s[0])
		max, _ := strconv.Atoi(s[len(s)-1])
		if min > pipelineCount || max > pipelineCount {
			return nil, errors.New("specified range is out of bounds")
		}

		// Create int array
		intArray := make([]int, len(s))
		for i := range s {
			val, _ := strconv.Atoi(s[i])
			intArray[i] = val
		}

		return intArray, nil
	}

	return nil, err
}

// Takes the min and max and returns a number array
func createNumbers(min int, max int) []int {
	numbers := make([]int, max-min+1)

	for i := range numbers {
		numbers[i] = i + min
	}

	return numbers
}
