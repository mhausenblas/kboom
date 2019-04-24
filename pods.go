package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

func launchPods(client *k8s.Client, namespace string, numpods int) (totaltime time.Duration) {
	if numpods > 0 {
		var launchedpods []*corev1.Pod
		start := time.Now()
		// create the pods:
		for i := 0; i < numpods; i++ {
			pod := genpod(namespace, fmt.Sprintf("scale-sleeper-%d", i))
			err := client.Create(context.Background(), pod)
			if err != nil {
				log.Printf("Can't create pod %v: %v", *pod.Metadata.Name, err)
				continue
			}
			launchedpods = append(launchedpods, pod)
		}

		// wait until all are running:
		for {
			allrunning, err := checkpods(client, namespace)
			if err != nil {
				log.Printf("Can't check pods: %v", err)
			}
			if allrunning {
				break
			}
			time.Sleep(2 * time.Second)
		}
		totaltime = time.Now().Sub(start)
		// clean up pods:
		for _, pod := range launchedpods {
			if err := client.Delete(context.Background(), pod); err != nil {
				log.Printf("Can't delete pod %v: %v", *pod.Metadata.Name, err)
			}
		}
		return totaltime
	}
	return time.Duration(0)
}

func checkpods(client *k8s.Client, namespace string) (allrunning bool, err error) {
	allrunning = true
	l := new(k8s.LabelSelector)
	l.Eq("generator", "kboom")
	var pods corev1.PodList
	if err = client.List(context.Background(), namespace, &pods, l.Selector()); err != nil {
		return false, err
	}
	for _, pod := range pods.Items {
		podphase := pod.GetStatus().GetPhase()
		if podphase != "Running" {
			allrunning = false
		}
	}
	return allrunning, nil
}

func genpod(namespace, name string) *corev1.Pod {
	return &corev1.Pod{
		Metadata: &metav1.ObjectMeta{
			Name:      k8s.String(name),
			Namespace: k8s.String(namespace),
			Labels:    map[string]string{"generator": "kboom"},
		},
		Spec: &corev1.PodSpec{
			Containers: []*corev1.Container{
				&corev1.Container{
					Name:    k8s.String("main"),
					Image:   k8s.String("busybox"),
					Command: []string{"/bin/sh", "-ec", "sleep 3600"},
				},
			},
		},
	}
}
