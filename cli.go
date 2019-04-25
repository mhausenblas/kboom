package main

import (
	"log"
	"strconv"
	"strings"
)

func parseParams(mode, load string) (testmode string, timeout, numpods int) {
	testmode, timeout = parseParam(mode)
	switch {
	case strings.HasPrefix(load, "po"), strings.HasPrefix(load, "pods"):
		_, numpods = parseParam(load)
	case strings.HasPrefix(load, "crd"):
		numpods = 0
	}
	return testmode, timeout, numpods
}

// parseParam extracts the parameter and the numeric value of a parameter,
// for example pods:5 -> pods, 5
func parseParam(param string) (string, int) {
	if strings.Contains(param, ":") {
		n, err := strconv.Atoi(strings.Split(param, ":")[1])
		if err != nil {
			log.Printf("Format for %s not recognized, defaulting to 0: %v", strings.Split(param, ":")[0], err)
			return strings.Split(param, ":")[0], 0
		}
		return strings.Split(param, ":")[0], n
	}
	return "", 0
}
