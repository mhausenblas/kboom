package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ericchiang/k8s"
	"github.com/mhausenblas/kubecuddler"
)

func main() {
	var namespace string
	var mode string
	var load string
	flag.StringVar(&namespace, "namespace", "kboom", "The namespace to run in, must exist. Create with 'kubectl create ns' if not done yet.")
	flag.StringVar(&mode, "mode", "scale", "The mode to operate in: scale for short-term/perf testing, soak for long-term testing.")
	flag.StringVar(&load, "load", "pods:1", "The load in the format resource:number comma-separated, defaults to pods:1.")
	flag.Parse()
	res, _ := kubecuddler.Kubectl(false, false, "/kubectl", "version", "--short")
	fmt.Println(res)
	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}
	numpods, numsvc, numdeploy := parseAllLoad(load)
	fmt.Printf("Generating load: %v pods, %v services, %v deployments\n", numpods, numsvc, numdeploy)
	fmt.Println("-------- Results --------")
	switch mode {
	case "scale":
		if numpods > 0 {
			r := launchPods(client, namespace, numpods)
			fmt.Printf("Overall pods: %v out of %v successful\n", r.Totalsuccess, numpods)
			fmt.Printf("Total time pods: %v\n", r.Totaltime)
			fmt.Printf("p50 pods: %v\n", r.P50)
			fmt.Printf("p95 pods: %v\n", r.P95)
		}
	case "soak":
		log.Println("Not yet implemented, aborting.")
	default:
		log.Println("Unknown mode, aborting.")
	}
}
