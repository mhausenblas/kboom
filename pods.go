package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/jamiealquiza/tachymeter"
	k8Yaml "k8s.io/apimachinery/pkg/util/yaml"
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

type podEvent struct {
	eventType string
	pod       *corev1.Pod
}

func (run *podrun) launch(podTemplate corev1.Pod) {
	run.Start = time.Now()
	run.Success = false
	pod := genpod(run.Namespace, fmt.Sprintf("%s-sleeper-%d", run.Loadtype, run.Ordinalnum), run.Image, podTemplate)
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
	successfulPods := 0

	content, err := ioutil.ReadFile("/tmp/template/template.yaml")
	if err != nil {
		log.Fatal(err)
	}

	podTemplate := &corev1.Pod{}
	dec := k8Yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(content)), 1000)

	if err := dec.Decode(&podTemplate); err != nil {
		log.Fatal(err)
	}

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
		go pr.launch(*podTemplate)
	}

	// check every second for successful running pods and capture
	// their overall time, that is, launch to phase 'Running':
	timeout := time.After(timeoutinsec)
	l := new(k8s.LabelSelector)
	l.Eq("generator", "kboom")

	ctx := context.TODO()
	var pod corev1.Pod
	watcher, err := client.Watch(ctx, namespace, &pod, l.Selector())
	if err != nil {
		// TODO handle error
	}
	defer watcher.Close()

	eventCh := make(chan podEvent, 1)

Check:
	for {
		go func() {
			pod := new(corev1.Pod)
			eventType, err := watcher.Next(pod)
			if err != nil {
				// TODO handle error
				return
			}
			eventCh <- podEvent{eventType, pod}
		}()

		select {
		case e := <-eventCh:
			if e.eventType == "Added" {
				continue
			}

			podname := *e.pod.Metadata.Name
			podphase := e.pod.GetStatus().GetPhase()
			if podphase == "Running" {
				if !podruns[name2ord(podname)].Success {
					podruns[name2ord(podname)].End = time.Now()
					podruns[name2ord(podname)].Success = true
					successfulPods++

					if successfulPods == numpods {
						watcher.Close()
						break Check
					}
				}
			}
		case <-timeout:
			fmt.Println("timeout")
			watcher.Close()
			break Check
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
func genpod(namespace, name, image string, podTemplate corev1.Pod) *corev1.Pod {

	podTemplate.Metadata = &metav1.ObjectMeta{
		Name:      k8s.String(name),
		Namespace: k8s.String(namespace),
		Labels:    map[string]string{"generator": "kboom"},
	}
	// TODO: currently only one container is supported
	podTemplate.Spec.Containers[0].Image = k8s.String(image)

	return &podTemplate
}
