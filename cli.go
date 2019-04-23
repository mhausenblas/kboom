package main

import (
	"log"
	"strconv"
	"strings"
)

func parseAllLoad(load string) (numpods, numsvc, numdeploy int) {
	// make sure we cover the case only
	// a single load item is specified:
	if !strings.Contains(load, ",") {
		load += ","
	}
	resources := strings.Split(load, ",")
	for _, res := range resources {
		switch {
		case strings.HasPrefix(res, "po"), strings.HasPrefix(res, "pods"):
			numpods = parseLoad(res)
		case strings.HasPrefix(res, "svc"), strings.HasPrefix(res, "services"):
			numsvc = parseLoad(res)
		case strings.HasPrefix(res, "deploy"), strings.HasPrefix(res, "deployments"):
			numdeploy = parseLoad(res)
		}
	}
	return numpods, numsvc, numdeploy
}

func parseLoad(res string) int {
	if strings.Contains(res, ":") {
		n, err := strconv.Atoi(strings.Split(res, ":")[1])
		if err != nil {
			log.Printf("Load format for %s not recognized, defaulting to 0: %v", strings.Split(res, ":")[0], err)
		}
		return n
	}
	return 0
}
