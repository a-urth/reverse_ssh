package static

import (
	_ "embed"
	"errors"
	"strings"
)

//go:embed destinations
var rawDestinations []byte

var (
	// not thread safe
	destinations []string
	currDest     int

	ErrNoExtraDestinations = errors.New("no additional destinations given")
)

func init() {
	dests := strings.Split(string(rawDestinations), "\n")

	destinations = make([]string, 0, len(dests))

	for _, d := range dests {
		if d != "" {
			destinations = append(destinations, d)
		}
	}
}

func GetDestination() (string, error) {
	if len(destinations) == 0 {
		return "", ErrNoExtraDestinations
	}

	// rotate given destinations in cycle
	if currDest == len(destinations) {
		currDest = 0
	}

	defer func() { currDest++ }()

	return destinations[currDest], nil
}
