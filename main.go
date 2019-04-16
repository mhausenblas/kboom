package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

func main() {
	var namespace string
	var mode string
	flag.StringVar(&namespace, "namespace", "kboom", "The namespace to run in, must exist.")
	flag.StringVar(&mode, "mode", "scale", "The mode to operate in: scale for short-term/perf testing, soak for long-term testing.")
	flag.Parse()

	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Creating load ...")
	launchPod(client, namespace, "test")
	log.Println("Listing pods I generated")
	l := new(k8s.LabelSelector)
	l.Eq("generator", "kboom")
	var pods corev1.PodList
	if err := client.List(context.Background(), namespace, &pods, l.Selector()); err != nil {
		log.Fatal(err)
	}
	for _, pod := range pods.Items {
		fmt.Printf("%v+", *pod)
	}
}

func launchPod(client *k8s.Client, namespace, name string) {
	pod := &corev1.Pod{
		Metadata: &metav1.ObjectMeta{
			Name:      k8s.String(name),
			Namespace: k8s.String(namespace),
			Labels:    map[string]string{"generator": "kboom"},
		},
		Spec: &corev1.PodSpec{
			Containers: []*corev1.Container{
				&corev1.Container{
					Name:  k8s.String("main"),
					Image: k8s.String("busybox"),
				},
			},
		},
	}

	if err := client.Create(context.Background(), pod); err != nil {
		log.Fatal(err)
	}
}
