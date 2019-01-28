# HelloWorld!

## Prep

After you fork and clone this repo you'll need to create a `.secrets` file in
the root of the clone directory that looks like this:

```
git_accesstoken=...
git_secrettoken=...
ic_apitoken=...
username=...
password=...
```

Where:
- `git_accesstoken` is a [Github Personal Access Token](https://github.com/settings/tokens)
- `git_secrettoken` is the secret token you want the events from Github to use to verify they're authenticated with your Knative subscription. This can basically be any random string you want.
- `ic_apitoken` is an IBM Cloud IAM api key - see `ic iam api-key-create`
- `username` is your Dockerhub username
- `password` is your Dockerhub password

If you're going to run the `demo` script then you'll also need to modify these
lines in there:

```
export APP_IMAGE=${APP_IMAGE:-duglin/helloworld}
export GITREPO=${GITREPO:-duglin/helloworld}
export REBUILD_IMAGE=${REBUILD_IMAGE:-duglin/rebuild}
```

Just change the `APP_IMAGE` and `REBUILD_IMAGE` values to use your
Dockerhub namespace name instead of `duglin`, and change `GITREPO`
to be the name of your Github clone of this repo - typically you should
just need to swap `duglin` for your Github name.

If you want to run through the demo manually, move on to the next
section, otherwise just run `demo` and it should do it all for you.
Just press the spacebar to move on to the next command when it pauses.
If the slow typing it annoying, press `f` when it pauses and it'll stop that.

## Manually running the demo | Demo details

This repo contains the source code, tools and instructions for a demo that
shows using Knative on the
(IBM Cloud Kubernetes Service)[https://cloud.ibm.com]. You can pretty easily
convert it to work on other platforms. The main reasons behind this
are:
- for me to learn more about Knative
- to share my experiences during this exercise
- highlight some areas where I think Knative can be improved and I'll
  be opening up issues for these things

So, with that, let's get started...

## Create a new cluster

Clearly we first need a Kubernetes cluster. And, I've automated this
(since I did it a lot durig my testing) via the `mkcluster` script
in the repo. It assumes that you're already logged into the IBM Cloud,
have all of the appropriate CLI tools installed and that we'll be using the
Dallas data center. If you need instructions on how to install the tooling
see our
(knative 101)[https://github.com/IBM/knative101/workshop/exercise-0/README.md]
docs.

```console
$ ic ks vlans --zone dal13
OK
ID        Name   Number   Type      Router         Supports Virtual Workers
PRIV_??          1240     private   bcr02a.dal13   true
PUB_??           996      public    fcr02a.dal13   true

$ ic ks cluster-create --name kndemo --zone dal13 --machine-type ${MACHINE} \
    --workers 3 --private-vlan PRIV_?? --public-vlan PUB_??  \
    --kube-version 1.13.2
Creating cluster...
OK
```

Just replace the `PRIV_??` and `PUB_??` with  the values you see from the
`ic ks vlans` command - and replace `dal13` with whatever data center you
want to use.

Once the cluster is ready you'll need to set your `KUBECONFIG`
environment variable to point to your customized Kubernetes config file:
```
$(ic ks cluster-config --export kndemo)

## Installing Knative

Knative requires Istio, so first we'll need to install that:

```console
$ kubectl apply -f  https://github.com/knative/serving/releases/download/v0.2.2/istio-crds.yaml
$ kubectl apply --filename https://github.com/knative/serving/releases/download/v0.2.2/istio.yaml
```

These should show you the list of all of the various resources that get
created. Once done, you'll need to tell Istio to automatically turn on
Istio's network management for your `default` namespace by using this:

```console
$ kubectl label namespace default istio-injection=enabled
namespace/default labeled
```

Now we can install Knative:
```console
kubectl apply \
  -f https://github.com/knative/serving/releases/download/v0.3.0/serving.yaml \
  -f https://github.com/knative/build/releases/download/v0.3.0/release.yaml \
  -f https://github.com/knative/eventing/releases/download/v0.3.0/release.yaml \
  -f https://github.com/knative/eventing-sources/releases/download/v0.3.0/release.yaml \
  -f https://github.com/knative/serving/releases/download/v0.3.0/monitoring.yaml
```

If during the install of Istio or Knative you get an error about some resource
type not being available, just run the command again - it's usually because
some of the resources being created use CRDs (custom resource definitons)
and it sometimes take a moment or two for those to be created and availble
for use.

## Setup our network

Before we can actually use Knative we need to do some additional setup around
how networking. To do this I have the `ingress.yaml` file:

```
# Route all *.containers.appdomain.cloud URLs to our istio gateway
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: iks-knative-ingress
  namespace: istio-system
  annotations:
    # give 30s to services to start from "cold start"
    ingress.bluemix.net/upstream-fail-timeout: "serviceName=knative-ingressgateway fail-timeout=30"
    ingress.bluemix.net/upstream-max-fails: "serviceName=knative-ingressgateway max-fails=0"
    ingress.bluemix.net/client-max-body-size: "size=200m"
spec:
  rules:
    - host: "*.containers.appdomain.cloud"
      http:
        paths:
          - path: /
            backend:
              serviceName: istio-ingressgateway
              servicePort: 80
---
# Allow for pods to talk to the internet
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","data":{"istio.sidecar.includeOutboundIPRanges":"*"},"kind":"ConfigMap","metadata":{"annotations":{},"name":"config-network","namespace":"knative-serving"}}
  name: config-network
  namespace: knative-serving
data:
  istio.sidecar.includeOutboundIPRanges: 172.30.0.0/16,172.20.0.0/16,10.10.10.0/24
```

The first resource in the file defines a new Ingress rule that maps all
HTTP traffic coming into the environment that has an HTTP `HOST` header
value that matches `*.containers.appdomain.cloud` to our `istio-gateway`
Kubernetes Service. Two interesting things about this:
- Unless we do something special, all of our apps we deploy will automatically
  get a URL (route) of the form 
  `<app-name>.<namespace>.containers.appdomain.cloud`. This Istio rule will
  (due to the wildcard) will allow us to deploy any application and not have
  to setup a special rule for each application to get it to route its
  traffic to Istio.
- The Istio gateway we're using here is what will manage all of the advanced
  networking that we'll leverage, such as load-balancing and traffic routing
  between multiple versions of our app.

The second resource in the yaml file will modify the Knative configuration
of Istio such that the only outbound traffic it blocks are to those IP
ranges listed on the `istio.sidecar.includeOutboundIPRanges` field. By default
Istio will block all outbound traffic - which, could take a while to figure
out if you're used to Kubernetes which lets all traffic through by default.


## Secrets

Before we get to the real point of this, which is deploying an application,
I needed to deploy a Kuberneres Secret that holds all of the private
keys/tokens/usernames/etc... that will be used during the exercises.




The second resource in the yaml file actually modifies an existing `ConfigMap`
in the Knative configuration. By default Knative will give all applications
a URL that uses the `example.com` domain. Since we want to piggy back on
IBM's DNS servers that already have `container.appdomain.cloud` routed
to IBM's container service, we need to tell Knative to use



## Our application

For this demo I'm just using a very simple HTTP server that responds
to any request with `Hello World!`, here's the source (`helloworld.go`):

```
package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Got request\n")
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintf(w, "Hello World!\n")
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
```

The `sleep` is in there just to slow things down a bit so that when we
increase the load on the app it'll cause one instance of the app to be
created for each client we have generating requests.

If you look in the `Makefile` you'll see how I built and pushed it
to my namespace in Dockerhub:

```console
docker build -t duglin/helloworld .
docker push duglin/helloworld
```

Now, our first adventure with Knative... let's deploy the application
as a Knative Service.  First, do not confuse a Knative Service with
a Kubernetes Service - they're not the same thing. This is an on-going
point of discussion within the Knative community, so for now we just
need to live with it.

So, let's first look at our yaml file that defines our Knative Service
defined in `service1.yaml`:

```yaml
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: helloworld
spec:
  runLatest:
    configuration:
      revisionTemplate:
        metadata:
          annotations:
            autoscaling.knative.dev/target: "1"
        spec:
          container:
            image: duglin/helloworld
          containerConcurrency: 1
```

Let's explain what some of these fields do:
- `metadata.name`: Like all Kube resources, this names our Service
- `runLatest`: here we have some choices as to how we define our Service.
  In particular we can specify one or multiple containers and that will
  not only control what is currently being run, but how we want to do a
  rolling upgrade to a new version - for example, how much traffic do we want
  to go to v1 versus v2 of our app. I'm not going to go into this now,
  so we'll just use `runLatest` which takes a single container definition.
- `revisionTemplate`: Each version of our application is called a
  revision. So, what we're doing here is defining a version of our app
  and for its image we're specifying `duglin/helloworld`, which we
  built before.
- `containerConcurrency` and `autoscaling.knative.dev/target`: these
  both appear to do the same thing - they tell the system how many concurrent
  requests to allow to each instance of our app at a single time. I'm
  not sure why I need to specify it twice, I'm hoping it's a bug.

Now we can deploy it:

```console
$ kubectl apply -f service1.yaml
service.serving.knative.dev/helloworld created
```




