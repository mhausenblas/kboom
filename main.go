package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ericchiang/k8s"
	"github.com/mhausenblas/kubecuddler"
)

func main() {
	var namespace string
	var mode string
	var load string
	flag.StringVar(&namespace, "namespace", "kboom", "The namespace to run in, must exist. Create with 'kubectl create ns' if not done yet.")
	flag.StringVar(&mode, "mode", "scale:20", "The mode to operate in: 'scale' for perf testing, 'soak' for long-term testing with timeout in seconds, defaults to scale:20.")
	flag.StringVar(&load, "load", "pods:1", "The load, as in number of pods, defaults to pods:1.")
	flag.Parse()
	res, _ := kubecuddler.Kubectl(false, false, "/kubectl", "version", "--short")
	fmt.Println(res)
	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}
	testmode, timeout, numpods := parseParams(mode, load)
	timeoutinsec := time.Duration(timeout) * time.Second
	fmt.Printf("Running a scale test, launching %v pod(s) with a %v timeout ...\n\n", numpods, timeoutinsec)
	fmt.Println("-------- Results --------")
	switch testmode {
	case "scale":
		if numpods > 0 {
			r := launchPods(client, namespace, timeoutinsec, numpods)
			fmt.Printf("Overall pods successful: %v out of %v\n", r.Totalsuccess, numpods)
			fmt.Printf("Total runtime: %v\n", r.Totaltime)
			fmt.Printf("Fastest pod: %v\n", r.Min)
			fmt.Printf("Slowest pod: %v\n", r.Max)
			fmt.Printf("p50 pods: %v\n", r.P50)
			fmt.Printf("p95 pods: %v\n", r.P95)
		}
	case "soak":
		log.Println("Not yet implemented, aborting.")
	default:
		log.Println("Unknown mode, aborting.")
	}
}
