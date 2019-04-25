package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/jamiealquiza/tachymeter"
)

type podrun struct {
	Loadtype   string
	Client     *k8s.Client
	Namespace  string
	Ordinalnum int
	Pod        *corev1.Pod
	Success    bool
	Start      time.Time
	End        time.Time
}

type Result struct {
	Totalsuccess int
	Totaltime    time.Duration
	P50          time.Duration
	P95          time.Duration
}

func (run *podrun) launch() {
	run.Start = time.Now()
	run.Success = false
	pod := genpod(run.Namespace, fmt.Sprintf("%s-sleeper-%d", run.Loadtype, run.Ordinalnum))
	err := run.Client.Create(context.Background(), pod)
	if err != nil {
		log.Printf("Can't create pod %v: %v", *pod.Metadata.Name, err)
	}
	run.Pod = pod
}

func launchPods(client *k8s.Client, namespace string, numpods int) Result {
	c := tachymeter.New(&tachymeter.Config{Size: numpods})
	var podruns []*podrun

	// launch the pods in parallel, as fast as we can:
	for i := 0; i < numpods; i++ {
		pr := &podrun{
			Loadtype:   "scale",
			Client:     client,
			Namespace:  namespace,
			Ordinalnum: i,
		}
		podruns = append(podruns, pr)
		go pr.launch()
	}

	// check for successful running pods and capture their
	// overall time, that is, launch time to 'Running':
	l := new(k8s.LabelSelector)
	l.Eq("generator", "kboom")
	for {
		var pods corev1.PodList
		if err := client.List(context.Background(), namespace, &pods, l.Selector()); err != nil {
			log.Printf("Can't check pods: %v", err)
		}
		allrunning := true
		for _, pod := range pods.Items {
			podname := *pod.Metadata.Name
			podphase := pod.GetStatus().GetPhase()
			if podphase == "Running" {
				podruns[name2ord(podname)].End = time.Now()
				podruns[name2ord(podname)].Success = true
				continue
			}
			allrunning = false
		}
		if allrunning {
			break
		}
		time.Sleep(1 * time.Second)
	}
	// record stats and clean up pods:
	var numsuccess int
	for _, run := range podruns {
		c.AddTime(run.End.Sub(run.Start))
		if run.Success {
			numsuccess++
		}
		if err := client.Delete(context.Background(), run.Pod); err != nil {
			log.Printf("Can't delete pod %v: %v", *run.Pod.Metadata.Name, err)
		}
	}
	results := c.Calc()
	return Result{
		Totalsuccess: numsuccess,
		Totaltime:    results.Time.Cumulative,
		P50:          results.Time.P50,
		P95:          results.Time.P95,
	}
}

// name2ord converts the name of a load tester pod to its
// ordinal number, for example: scale-sleeper-42 -> 42
func name2ord(name string) int {
	ordinalstr := strings.SplitAfterN(name, "-", 3)[2]
	ordinalnum, _ := strconv.Atoi(ordinalstr)
	return ordinalnum
}

// genpod generates the pod specification
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
