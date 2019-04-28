# A simple Kubernetes load testing tool

![kboom logo](img/kboom-logo.png)


> NOTE: this is WIP and also this is not an official AWS tool. Provided as is and use at your own risk.

Think of `kboom` as the Kubernetes equivalent of [boom](https://github.com/tarekziade/boom), allowing you to create short-term load for scale testing and long-term load for soak testing. Supported load out of the box for scale testing are pods and custom resources via CRDs for soak testing is planned.

Check out the interactive [demo](https://www.katacoda.com/petermbenjamin/scenarios/kboom).

## Why bother?

I didn't find a usable tool to do Kubernetes-native load testing, for scalability and/or soak purposes. Here's where I can imagine `kboom` might be useful for you:

- You are a cluster admin and want to test how much "fits" in the cluster. You use `kboom` for a scale test and see how many pods can be placed and how long it takes.
- You are a cluster or namespace admin and want to test how long it takes to launch a set number of pods in a new cluster, comparing it with what you already know from an existing cluster.
- You are developer and want to test your custom controller or operator. You use `kboom` for a long-term soak test of your controller.

## Install

Before you begin, you will need `kubectl` client version v1.12.0 or higher for [kubectl plugin support](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/).

To install `kboom`, do the following:

```bash
$ curl https://raw.githubusercontent.com/mhausenblas/kboom/master/kboom -o kubectl-kboom
$ chmod +x kubectl-kboom
$ sudo mv ./kubectl-kboom /usr/local/bin
```

From this point on you can use it as a `kubectl` plugin as in `kubectl kboom`. However, in order for you to generate the load, you'll have to also give it the necessary [permissions](permissions.yaml) (note: you only need to do this once, per cluster):

```bash
$ kubectl create ns kboom
$ kubectl apply -f https://raw.githubusercontent.com/mhausenblas/kboom/master/permissions.yaml
```

Now you're set up and good to go, next up, learn how to use `kboom`.

## Use

Here's how you'd use `kboom` to do some scale testing. The load test is run in-cluster as a [Kubernetes job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) so you do multiple runs and compare outcomes in a straight-forward manner. Note that by default `kboom` assumes there's a namespace `kboom` available and it will run in this namespace. If this namespace doesn't exist, create it with `kubectl create ns kboom` or otherwise use the `--namespace` parameter to overwrite it.

So, first we use the `generate` command to generate the load, launching 10 pods (that is, using `busybox` containers that just sleep) with a timeout of 14 seconds (that is, if a pod is not running within that time, it's considered a failure):

```bash
$ kubectl kboom generate --mode=scale:14 --load=pods:10
job.batch/kboom created
```

From now on you can execute the `results` command as often as you like, you can see the live progress there:


```bash
$ kubectl kboom results
Server Version: v1.12.6-eks-d69f1b
Running a scale test, launching 10 pod(s) with a 14s timeout ...

-------- Results --------
Overall pods successful: 6 out of 10
Total runtime: 14.061988653s
Fastest pod: 9.003997546s
Slowest pod: 13.003831951s
p50 pods: 12.003529448s
p95 pods: 13.003831951s
```

When you're done, and don't need the results anymore, use `kubectl kboom cleanup` to get rid of the run. Note: should you execute the `cleanup` command too soon for `kboom` to terminate all its test pods, you can use `kubectl delete po -l=generator=kboom` to get rid of all orphaned pods.

## Known issues and plans

- Need to come up with stricter permissions, currently too wide and not following the least privileges principle.
- Add support for custom resources and soak testing (running for many hours).
- Add support for other core resources, such as services or deployments.
