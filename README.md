# A simple Kubernetes load testing tool

![kboom logo](img/kboom-logo.png)


> NOTE: this is WIP and also not an official AWS tool. Use at your own risk.

Think of `kboom` as the Kubernetes equivalent of [boom](https://github.com/tarekziade/boom), allowing you to create short-term load for scale-testing and long-term load for soak-testing. Supported load out of the box are pods, services, and deployments as well as custom resource via CRDs.

## Why bother?

I didn't find a usable tool to do Kubernetes-native load testing, for scalability and/or soak purposes. Here's where I can imagine `kboom` might be useful for you:

- You are a cluster admin and want to test how much "fits" in the cluster. You use `kboom` for a scale test and see how many pods can be placed and how long it takes.
- You are a cluster or namespace admin and want to test how long it takes to launch a set number of pods in a new cluster, comparing it with what you already know from an existing cluster.
- You are developer and want to test your custom controller or operator. You use `kboom` for a long-term soak test of your controller.

## Install

```bash
$ curl https://raw.githubusercontent.com/mhausenblas/kboom/master/kboom -o kubectl-kboom
$ sudo mv ./kubectl-kboom /usr/local/bin
```

## Use

Here's how you'd use `kboom` to do some scale-testing using 10 pods. Note that the load test is run in-cluster as a [Kubernetes job](https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/) so you do multiple runs and compare outcomes in a straight-forward manner.

So, first we use the `generate` command to generate the scale load, launching 10 pods (that is, using `busybox` containers that just sleep) with a timeout of 14 seconds (that is, if a pod is not running within that time, it's considered a failure):

```bash
$ kubectl kboom generate --mode=scale:14 --load=pods:10
job.batch/kboom created
```

From now on you can execute the `results` command as often as you like, you can see the live progress there:


```bash
$ kubectl kboom results
Client Version: v1.14.0
Server Version: v1.12.6-eks-d69f1b
Running a scale test, launching 10 pod(s) with a 14s timeout ...

-------- Results --------
Overall pods successful: 7 out of 10
Total runtime: 14.106868399s sec
Fastest/slowest pod: 13.005129718s sec/13.006064873s sec
p50 pods: 13.005613259s
p95 pods: 13.006064873s
```

When you're done, and don't need the results anymore, use `kubectl kboom cleanup` to get rid of the run. Note: should you execute the `cleanup` command too soon for `kboom` to terminate all its test pods, you can use `kubectl delete po -l=generator=kboom` to get rid of all orphaned pods.