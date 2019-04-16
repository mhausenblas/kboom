package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
)

func main() {
	var namespace string
	var mode string
	flag.StringVar(&mode, "namespace", "kboom", "The namespace to run in, must exist.")
	flag.StringVar(&mode, "mode", "scale", "The mode to operate in: scale for short-term/perf testing, soak for long-term testing.")
	flag.Parse()

	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listing pods")
	var pods corev1.PodList
	if err := client.List(context.Background(), namespace, &pods); err != nil {
		log.Fatal(err)
	}
	for _, pod := range pods.Items {
		fmt.Printf("%v+", *pod)
	}
}
