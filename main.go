package main

import (
	"flag"
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

	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(load)
	numpods, numsvc, numdeploy := parseAllLoad(load)
	log.Printf("Creating load: %v pods, %v services, %v deployments\n", numpods, numsvc, numdeploy)

	res, _ := kubecuddler.Kubectl(false, false, "/kubectl", "version", "--short")
	log.Println(res)

	switch mode {
	case "scale":
		if numpods > 0 {
			r := launchPods(client, namespace, numpods)
			log.Printf("Stats:\n%v\n", r)
		}
	case "soak":
		log.Println("Not yet implemented, aborting.")
	default:
		log.Println("Unknown mode, aborting.")
	}
}
