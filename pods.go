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
	Image      string
}

type Result struct {
	Totalsuccess int
	Totaltime    time.Duration
	Cumulative   time.Duration
	Min          time.Duration
	Max          time.Duration
	P50          time.Duration
	P95          time.Duration
}

func (run *podrun) launch() {
	run.Start = time.Now()
	run.Success = false
	pod := genpod(run.Namespace, fmt.Sprintf("%s-sleeper-%d", run.Loadtype, run.Ordinalnum), run.Image)
	err := run.Client.Create(context.Background(), pod)
	if err != nil {
		log.Printf("Can't create pod %v: %v", *pod.Metadata.Name, err)
	}
	run.Pod = pod
}

func launchPods(client *k8s.Client, namespace, image string, timeoutinsec time.Duration, numpods int) Result {
	c := tachymeter.New(&tachymeter.Config{Size: numpods})
	start := time.Now()
	var podruns []*podrun

	// launch the pods in parallel, as fast as we can:
	for i := 0; i < numpods; i++ {
		pr := &podrun{
			Loadtype:   "scale",
			Client:     client,
			Namespace:  namespace,
			Ordinalnum: i,
			Image:      image,
		}
		podruns = append(podruns, pr)
		go pr.launch()
	}

	// check every second for successful running pods and capture
	// their overall time, that is, launch to phase 'Running':
	timeout := time.After(timeoutinsec)
	tick := time.Tick(1000 * time.Millisecond)
	l := new(k8s.LabelSelector)
	l.Eq("generator", "kboom")
Check:
	for {
		select {
		case <-timeout: // we're done checking
			break Check
		case <-tick: // check pods we generated
			var pods corev1.PodList
			if err := client.List(context.Background(), namespace, &pods, l.Selector()); err != nil {
				log.Printf("Can't check pods: %v", err)
				continue
			}
			for _, pod := range pods.Items {
				podname := *pod.Metadata.Name
				podphase := pod.GetStatus().GetPhase()
				if podphase == "Running" {
					if !podruns[name2ord(podname)].Success {
						podruns[name2ord(podname)].End = time.Now()
						podruns[name2ord(podname)].Success = true
					}
				}
			}
		}
	}

	// record stats and clean up pods:
	var numsuccess int
	for _, run := range podruns {
		if run.Success {
			c.AddTime(run.End.Sub(run.Start))
			numsuccess++
		}
		if err := client.Delete(context.Background(), run.Pod); err != nil {
			log.Printf("Can't delete pod %v: %v", *run.Pod.Metadata.Name, err)
		}
	}
	results := c.Calc()
	return Result{
		Totalsuccess: numsuccess,
		Totaltime:    time.Since(start),
		Cumulative:   results.Time.Cumulative,
		Min:          results.Time.Min,
		Max:          results.Time.Max,
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
func genpod(namespace, name, image string) *corev1.Pod {
	var userID int64 = 65534

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
					Image:   k8s.String(image),
					Command: []string{"/bin/sh", "-ec", "sleep 3600"},
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: &userID,
					},
				},
			},
		},
	}
}
