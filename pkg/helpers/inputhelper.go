package helpers

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Processes the user's input if it's a range
func ValidateInputRange(userRange string, pipelineCount int) ([]string, error) {
	// Ensure range is within number of pipelines retrieved
	s := strings.Split(userRange, "-")
	min, _ := strconv.Atoi(s[0])
	max, _ := strconv.Atoi(s[1])
	if min >= max {
		return nil, errors.New("invalid range provided")
	} else if max-min > pipelineCount {
		return nil, errors.New("range provided is larger than number of pipelines retrieved")
	}

	// TODO: expand range out (1-5 becomes array of 1,2,3,4,5) or use range keyword?
	return s, nil

	return nil, err
}

// Processes the user's input if it's a selection
func ValidateInputSelection(userSelection string, pipelineCount int) ([]string, error) {
	// Ensure user provided selection is valid (1,2,3 or 1, 2, 3)
	match, err := regexp.MatchString(`(\d+)(,\s*\d+)*`, userSelection)
	if err != nil {
		return nil, err
	}

	if match {
		// Ensure lower and upper values are within number of pipelines retrieved
		userSelection = strings.ReplaceAll(userSelection, " ", "")
		s := strings.Split(userSelection, ",")
		sort.Strings(s)
		min, _ := strconv.Atoi(s[0])
		max, _ := strconv.Atoi(s[len(s)])
		if min > pipelineCount || max > pipelineCount {
			return nil, errors.New("specified range is out of bounds")
		}

		return s, nil
	}

	return nil, err
}
