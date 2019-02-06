# HelloWorld!

## My First Knative Demo / App

Written with IKS in mind.

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

You can now either run the `demo` script to have it automatically run through
the demo, providing commentary, or you can do it all manually.

Which ever way you run the demo, I'd run `./pods` in another window so
you can see the Knative Services and Pods as they come-n-go. Again, make
sure you run:

```
$(ic ks cluster config --export <cluster-name>)
```

in that other window too.

Also, these scripts are tested on Ubuntu, I haven't tried them on MacOS
yet so don't be surprised if things don't work there yet. PR are welcome
though.

## Using the `demo` script

If you're going to run the `demo` script then you'll also need to modify these
lines in there:

```
export APP_IMAGE=${APP_IMAGE:-duglin/helloworld}
export REBUILD_IMAGE=${REBUILD_IMAGE:-duglin/rebuild}
export GITREPO=${GITREPO:-duglin/helloworld}
```

Change the `APP_IMAGE` and `REBUILD_IMAGE` values to use your
Dockerhub namespace name instead of `duglin`, and change `GITREPO`
to be the name of your Github clone of this repo - typically you should
just need to swap `duglin` for your Github name.

You'll need to create a cluster in advance. You can use `mkcluster` to do that
or see the (Create a new cluster)[#create_a_new_cluster] section below
if you want to do it manually.

Then you'll need to install Istio and Knative, see the sections before
for that as well.

When done, make sure you also run:

```
$(ic ks cluster config --export <cluster-name>)
```

before you run the demo so your `kubectl` points to the correct cluster.

### Running the demo

```
./demo [ CLUSTER_NAME ]
```
`CLUSTER_NAME` is optional if your `KUBECONFIG` environment variable
already points to your IKS cluster.

As the demo runs, press the spacebar to move on to the next command when it
pauses.  If the slow typing is annoying, press `f` when it pauses and it'll
stop that.


## Manually running the demo | Demo details

### Create a new cluster

Clearly we first need a Kubernetes cluster. And, I've automated this
(since I did it a lot durig my testing) via the `mkcluster` script
in the repo. It assumes that you're already logged into the IBM Cloud,
have all of the appropriate CLI tools installed and that we'll be using the
Dallas data center. If you need instructions on how to install the tooling
see our
(knative 101)[https://github.com/IBM/knative101/workshop/exercise-0/README.md]
docs.

First, we need to get some info about our LANs since the `ic ks cluster-create`
command requires that:

```
$ ic ks vlans --zone dal13
OK
ID        Name   Number   Type      Router         Supports Virtual Workers
PRIV_??          1240     private   bcr02a.dal13   true
PUB_??           996      public    fcr02a.dal13   true
```

Now we can create the cluster. In this case we'll create it with 3 worker
nodes.  Replace the `PRIV_??` and `PUB_??` with  the values you see from the
`ic ks vlans` command - and replace `dal13` with whatever data center you
want to use.

```
$ ic ks cluster-create --name kndemo --zone dal13 --machine-type ${MACHINE} \
    --workers 3 --private-vlan PRIV_?? --public-vlan PUB_??  \
    --kube-version 1.13.2
Creating cluster...
OK
```

Once the cluster is ready you'll need to set your `KUBECONFIG`
environment variable to point to your customized Kubernetes config file:

```
$(ic ks cluster-config --export kndemo)
```

### Installing Knative (and Istio)

Knative requires Istio, but luckily IKS's install of Knative will install
Istio too - just run:

```
$ ic ks cluster-addon-enable knative CLUSTER_NAME
```

This will take a moment or two, and you can see it's done when two
things happen, first you see the Istio and Knative namespaces:

```
$ kubectl get ns -w
NAME                 STATUS   AGE
default              Active   39h
ibm-cert-store       Active   39h
ibm-system           Active   39h
istio-system         Active   21h
knative-build        Active   18h
knative-eventing     Active   18h
knative-monitoring   Active   18h
knative-serving      Active   18h
knative-sources      Active   18h
kube-public          Active   39h
kube-system          Active   39h
```

When you see the `istio-system` and the `knatve-...` ones appear then
you're almost done.

Next, check to see if all of the pods are running. I find it easiest
to just see if there are any pods that are not running ;-)  via this:

```
$ kubectl get pods --all-namespaces | ep -v Running
```

And if that list is empty, or only shows non-Istio and non-Knative pods
(due to other things running in your cluster) then you should be go to go.
If the list isn't empty, then give it more time for things to initialize.

### Setup our network

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

Install these resouces:

```
$ ./kapply ingress.yaml
ingress.extensions/iks-knative-ingress created
configmap/config-network created
```

(We'll talk more about the `kapply` command later.)

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
out if you're used to using vanilla Kubernetes which lets all traffic through
by default.

Almost done with network, yes this is way too much work! Last, we need to
modify a Knative ConfigMap such that the default URL assigned to our apps
isn't `example.com`, which is the default that Knative uses.

Before we can do that, we need to know your cluster's domain name. You
can get this info by running:

```
$ ic ks cluster-get -s CLUSTER_NAME
Name:                   kntest03
State:                  normal
Created:                2019-02-04T21:36:18+0000
Location:               dal13
Master URL:             https://c2.us-south.containers.cloud.ibm.com:24730
Master Location:        Dallas
Master Status:          Ready (1 day ago)
Ingress Subdomain:      kntest03.us-south.containers.appdomain.cloud
Ingress Secret:         kntest03
Workers:                4
Worker Zones:           dal13
Version:                1.12.4_1534* (1.12.5_1537 latest)
Owner:                  me@us.ibm.com
Resource Group Name:    default
```

Notice the `Ingress Subdomain:` line - that's your domain name.

Now, edit the ConfigMap:

```
$ kubectl edit cm/config-domain -n knative-serving
```

In there, change any occurrence of `example.com` with your cluster's
domain name.

Now, we're finally done with the administrivial networking stuff.

### Secrets

Before we get to the real point of this, which is deploying an application,
I needed to create a Kuberneres Secret that holds all of the private
keys/tokens/usernames/etc... that will be used during the exercises.

For that I have the `secrets.yaml` file:

```
apiVersion: v1
kind: Secret
metadata:
  name: mysecrets
  annotations:
    build.knative.dev/docker-0: https://index.docker.io/v1/
type: kubernetes.io/basic-auth
stringData:
  git_accesstoken: ${.secrets.git_accesstoken}
  git_secrettoken: ${.secrets.git_secrettoken}
  ic_apitoken: ${.secrets.ic_apitoken}
  username: ${.secrets.username}
  password: ${.secrets.password}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-bot
secrets:
- name: mysecrets
```

You'll notice some environment variable looking fields in there. Obviously,
those are not normal Kube yaml things. To make life easier, I created
a script called `kapply` which takes a yaml file and replaces references
like those with their real values before invoking `kubectl`. This allows
me to share my yaml with you w/o asking you to modify these files and run
the risk of you checking them into your github repo by mistake. All you need
to do is create the `.secrets` files I mentioned at the top of this README.

The you can create the secret via:

```
$ ./kapply secrets.yal
secret/mysecrets configured
serviceaccount/build-bot configured
```

### Install the Kaniko build template

Almost there! Let's install the Kaniko build template:

```
$ kubectl apply -f https://raw.githubusercontent.com/knative/build-templates/master/kaniko/kaniko.yaml
buildtemplate.build.knative.dev/kaniko created
```

Build templates are like CloudFoundry buildpacks, they'll take your source
and create a container image from it, and then push it to some container
registry. If you look at the yaml for this resource you'll see something
like this:

```
apiVersion: build.knative.dev/v1alpha1
kind: BuildTemplate
metadata:
  name: kaniko
spec:
  parameters:
  - name: IMAGE
    description: The name of the image to push
  - name: DOCKERFILE
    description: Path to the Dockerfile to build.
    default: /workspace/Dockerfile

  steps:
  - name: build-and-push
    image: gcr.io/kaniko-project/executor
    args:
    - --dockerfile=${DOCKERFILE}
    - --destination=${IMAGE}
    env:
    - name: DOCKER_CONFIG
      value: /builder/home/.docker
```

Most of what's in there should be obvious:
- `image: gcr.io/kaniko-project/executor` defines the container image that will
  be used to actually do all of the work of building things.
- `args` are the command line flags to pass to running container
- `env` defines some enviornment variables for the container
` `parameters` define some parameters that users of the template can specify
- `steps` allows for you to define a list of things to do in order to build
  the image

What's innteresting about this to me is that I'm wondering if this is
overly complex and overly simplified at the same time. What I mean by that
is this... the template provides an image to do all sorts of magic to build
our image - that part makes sense to me. However, they then suggest that
people will want to mix-n-match the calling of multiple images/steps by
allowing template owners to define `steps`. Why not just put all of that
logic into the one image?

If the argument is that you may need to string these steps together, then
why do we think we can get by with a simple ordered list? It won't be long
before people need real control flow (like `if` statements) between the
steps. It seems to me it would be better to tell people to just put all of
the logic they need into an image and do whatever orchestration of steps
within that. Let's not head down the path of inventing some kind of scripting
language here. That's why I think it's overly complex (I don't see the need
for `steps`) and overly simplified (if you do see the need then a simple list
isn't sufficient in the long run).

And finally, originally I didn't use Build Templates, I just put a reference
to the Kaniko image directly into my Service definition's build section and
that, of course, worked with basically the same amount of yaml - but didn't
introduce a level of indirection that could only lead to added perception
of complexity. But I gave in and added Build Templates so I could
ramble about them here :-)

Anyway, moving on...

### Our application

For this demo I'm just using a very simple HTTP server that responds
to any request with `Hello World!`, here's the source (`helloworld.go`):

```
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	text := "Hello World!"

	rev := os.Getenv("K_REVISION")
	if i := strings.LastIndex(rev, "-"); i > 0 {
		rev = rev[i+1:]
	}

	msg := fmt.Sprintf("%s: %s\n", rev, text)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("Got request\n")
		fmt.Fprint(w, msg)
	})

	fmt.Printf("Listening on port 8080 (rev: %s)\n", rev)
	http.ListenAndServe(":8080", nil)
}
```

The `sleep` is in there just to slow things down a bit so that when we
increase the load on the app it'll cause one instance of the app to be
created for each client we have generating requests.

It will also print part of the `K_REVISION` environment variable so
we can see which revision number of our app we're hitting.

If you look in the `Makefile` you'll see how I built and pushed it
to my namespace in Dockerhub:

```
docker build -t $(APP_IMAGE) .
docker push $(APP_IMAGE)
```

You'll need to modify the `Makefile` to point to your DockerHub
namespace if you want to use `make`.

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
        spec:
          container:
            image: ${APP_IMAGE}
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
  and its `image` value.
- `containerConcurrency`: this tells the system how many concurrent
  requests to allow to each instance of our app at a single time.

Now we can deploy it:

```
export APP_IMAGE=duglin/helloworld
$ ./kapply service1.yaml
service.serving.knative.dev/helloworld created
```

Notice you'll need to set the `APP_IMAGE` environment variable so that
`kapply` can fill in correctly.


If you're not running `pods` in another window, run it now to see what
happened:

```
$ ./pods --once
Cluster: knative102
K_SVC_NAME                     LATESTREADY                    READY
helloworld                     helloworld-00001               True 

POD_NAME                                                STATUS           AGE
helloworld-00001-deployment-78796cb584-jswh6            Running          90s
```

This shows the list of Knative services and running pods in the cluster.
You should see your `helloworld` Knative service with one revisioned
called `helloworld-00001`, and a pod with a really funky name but that
starts with `helloworld-00001` - meaning it's related to revision 1.

Notice the word "deployment" in there - that's because under the covers
Knative create a Kubernetes deployment resource.

So, it's running - let's hit it:

```
$ curl -sf helloworld.default.knative102.us-south.containers.appdomain.cloud
00001: Hello World!
```

You'll need to replace the `knative02...` portion with the domain name
of your cluster that you determine above.

I won't bore you with the commands I used, but if you run the `demo`
script I show you all of the various resources created as a result of
deploying this ONE yaml files (yes, `kubectl get all` is a lie but it
gets the point across and it's close):

```
$ kubectl get all | grep -v knative.dev

deployment.apps/helloworld-00001-deployment
endpoint/helloworld-00001-service
endpoint/kubernetes
pod/helloworld-00001-deployment-78796cb584-jswh6
pod/helloworld-00001-deployment-78796cb584-sph6w
replicaset.apps/helloworld-00001-deployment-78796cb584
service/helloworld
service/helloworld-00001-service
service/kubernetes

$ kubectl get all | grep knative.dev

buildtemplate.build.knative.dev/kaniko
clusteringress.networking.internal.knative.dev/helloworld-rk28q
configuration.serving.knative.dev/helloworld
image.caching.internal.knative.dev/helloworld-00001-cache
image.caching.internal.knative.dev/kaniko-1622814-00000
podautoscaler.autoscaling.internal.knative.dev/helloworld-00001
revision.serving.knative.dev/helloworld-00001
route.serving.knative.dev/helloworld
service.serving.knative.dev/helloworld
```

The first list is the list of native Kube resources, and the second list
contains the Knative ones. That's a lot of stuff!  And I mean that in a
good way! One of my personal goals for Knative is to offer up a more user
friendly user experience for Kube users. Sure Kube has a ton of features
but with that flexibility has come complexity. Think about how much learning
and work is required to setup all of these resources that Knative as done
for us. I no longer need to understans Ingress, load-balancing, auto-scaling,
etc.  Very nice!

That's it. Once we got past the setup (which is a bit much, but I'm hoping
is just a one-time thing for most people), the deployment of the app itself
was a single `kuebctl` command with a single resource definition. That's
a huge step forward for Kube users

### Adding build

