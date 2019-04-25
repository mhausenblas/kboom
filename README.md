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

So, first we use the `generate` command to generate the scale load, launching 10 pods (that is, using `busybox` containers that just sleep):

```bash
$ kubectl kboom generate --load=pods:10
job.batch/kboom created
```

From now on you can execute the `results` command as often as you like, you can see the live progress there:


```bash
$ kubectl kboom results
Client Version: v1.14.0
Server Version: v1.12.6-eks-d69f1b
Generating load: 10 pods, 0 services, 0 deployments
-------- Results --------
Overall pods: 10 out of 10 successful
Total time pods: 2m51.26601059s
p50 pods: 17.126641281s
p95 pods: 17.12704589s
```

When you're done, and don't need the results anymore, use `kubectl kboom cleanup` to get rid of the run.