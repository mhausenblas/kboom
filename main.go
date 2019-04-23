package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

func main() {
	var namespace string
	var mode string
	var load string
	flag.StringVar(&namespace, "namespace", "kboom", "The namespace to run in, must exist. Create with kubectl create ns if not done yet.")
	flag.StringVar(&mode, "mode", "scale", "The mode to operate in: scale for short-term/perf testing, soak for long-term testing.")
	flag.StringVar(&load, "load", "pods:5", "The load in the format resource:number comma-separated, defaults to pods:5.")
	flag.Parse()

	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(load)
	numpods, numsvc, numdeploy := parseAllLoad(load)
	log.Printf("Creating load: %v pods, %v services, %v deployments\n", numpods, numsvc, numdeploy)
	launchPods(client, namespace, "test")

}

func launchPods(client *k8s.Client, namespace, name string) {
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
	start := time.Now()
	if err := client.Create(context.Background(), pod); err != nil {
		log.Fatal(err)
	}
	// wait until all are running:
	// TBD
	diff := time.Now().Sub(start)
	log.Printf("Total time pods %v", diff)
}

func listAllPods(client *k8s.Client, namespace string) {
	l := new(k8s.LabelSelector)
	l.Eq("generator", "kboom")
	var pods corev1.PodList
	if err := client.List(context.Background(), namespace, &pods, l.Selector()); err != nil {
		log.Fatal(err)
	}
	for _, pod := range pods.Items {
		log.Printf("%v", *pod)
	}
}
