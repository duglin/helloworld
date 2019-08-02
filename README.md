# HelloWorld!

## An Introductory Knative Demo

Note: this version uses Knative v0.6.0. Look at the
[releases](https://github.com/duglin/helloworld/releases) to find instructions
for how to run the demo on older versions.

This repo contains the source code, tools and instructions for a demo that
shows using Knative on the
[IBM Cloud Kubernetes Service](https://cloud.ibm.com). You can pretty easily
convert it to work on other platforms.

The main reasons behind this are:
- showcase the basics of Knative so you can jump-start your usage of it
- showcase some of the newer features of Knative, meaning this demo will
  grow over time as new features are added to Knative
- to share my experiences during this exercise
- highlight some areas where I think Knative can be improved

If you want use the `demo` script, it will show the commands in bold
and the output of the commands in normal text. When it pauses, just press
the spacebar to move to the next step. If the slow typing is annoying, press
`f` when it pauses and it'll stop that.

Also, these scripts are tested on Ubuntu, I haven't tried them on MacOS
yet so don't be surprised if things don't work there yet. PR are welcome
though.

So, with that, let's get started...

## Canned demo!

If you just want to see what a successful demo looks like without actually
doing anything, like even installing Kubernetes or Knative, just run this:

```
$ USESAVED=1 ./demo
```

And press the spacebar to walk through each command. The output should look
basically the same as a live-run (even with delays at the right times) to make
it look-n-feel like it's live!  Great for situations where your network
isn't reliable, or the demo Gods are mad at you.

## Prep

To run the demo yourself, there are a couple of things you'll need to do.
First, fork and clone this repo. Cloning alone isn't good enough because as
part of the demo we will be updating some files and pushing them back to
Github to demonstrate how to build new container images as your source code
changes.

Once you've cloned the repo, you'll need to create a `.secrets` file in
the root of the clone directory that looks like this:

```
git_accesstoken=...
git_secrettoken=...
username=...
password=...
```

Where:
- `git_accesstoken` is a [Github Personal Access Token](https://github.com/settings/tokens)
- `git_secrettoken` is the secret token you want the events from Github to use to verify they're authenticated with your Knative subscription. This can basically be any random string you want.
- `username` is your container registry (e.g. Dockerhub) username
- `password` is your container registry (e.g. Dockerhub) password

I store all of the secret information I use during the demo in there. It's
safer to put them in there than to put them into the real files of the repo
and run the risk of checking them into github by mistake.

The rest of this assumes you:
- are already logged into the IBM Cloud (`ibmcloud login`)
- and have all of the appropriate CLI tools installed - see:
  https://github.com/IBM/knative101/workshop/exercise-0/README.md

## Creating a Kubernetes cluster

Note: in order to run this demo you'll need a cluster with at least 3
worker nodes in it. As of now this means that you can not use a free-tier
cluster.

To help with the process of creating a Kubernetes cluster, I've automated this
via the `mkcluster` script in the repo. The script assumes you're using the
Dallas data center, so if not you'll need to modify it.

To create the cluster just run:
```
$ ./mkcluster CLUSTER_NAME
```

And you can then skip to the next section.

If you decide to create it manually, then you'll first need to get some info
about our LANs since the `ibmcloud ks cluster-create` command requires that:

```
$ ibmcloud ks vlans --zone dal13
OK
ID        Name   Number   Type      Router         Supports Virtual Workers
PRIV_??          1240     private   bcr02a.dal13   true
PUB_??           996      public    fcr02a.dal13   true
```

Now we can create the cluster. In this case we'll create it with 3 worker
nodes. Replace the `PRIV_??` and `PUB_??` with  the values you see from the
`ibmcloud ks vlans` command - and replace `dal13` with whatever data center you
want to use.

```
$ ibmcloud ks cluster-create --name CLUSTER_NAME --zone dal13 \
    --machine-type ${MACHINE} --workers 3 \
	--private-vlan PRIV_?? --public-vlan PUB_??  \
    --kube-version 1.13.2
Creating cluster...
OK
```

Once the cluster is ready you'll need to set your `KUBECONFIG`
environment variable to point to your customized Kubernetes config file:

```
$(ibmcloud ks cluster-config --export CLUSTER_NAME)
```

Now we can install Knative...

Knative requires Istio, but luckily IBM Cloud's Kubernetes service's install
of Knative will install Istio too - just run:

```
$ ibmcloud ks cluster-addon-enable knative CLUSTER_NAME
```
If it prompts you to install Istio, just say `yes`, even if you have
Istio already installed - worst case, it'll upgrade it for you.

Since Knative is still very much a work-in-progress, and while the IKS
team will try to keep up with the latest versions, if you happen to catch
things at a time when the IKS install of Knative is behind the latest
version, and you want the very latest version just follow
[these instructions](https://github.com/knative/docs/blob/master/install/Knative-with-IKS.md)
for how to install the latest Knative manually.

This will take a moment or two, and you can see it's done when two
things happen, first you see the Istio and Knative namespaces:

```
$ kubectl get ns -w
NAME                 STATUS   AGE
default              Active   39h
ibm-cert-store       Active   39h
ibm-system           Active   39h
istio-system         Active   21h    <----
knative-build        Active   18h    <----
knative-eventing     Active   18h    <----
knative-monitoring   Active   18h    <----
knative-serving      Active   18h    <----
knative-sources      Active   18h    <----
kube-public          Active   39h
kube-system          Active   39h
```

When you see the `istio-system` and the `knatve-...` ones appear then
you're almost done. Press control-C to stop the watch.

Next, check to see if all of the pods are running. I find it easiest
to just see if there are any pods that are not running, via this:

```
$ kubectl get pods --all-namespaces | grep -v Running
```

And if that list is empty, or only shows non-Istio and non-Knative pods
(due to other things running in your cluster) then you should be good to go.
If the list isn't empty, then give it more time for things to initialize.

## Running the demo

You can now either run the `demo` script to have it automatically run through
the demo, providing commentary, or you can do it all manually.

Which ever way you run the demo, I'd run `./pods` in another window so
you can see the Knative Services and Pods as they come-n-go. Make sure
you run `$(ibmcloud ks cluster-config --export CLUSTER_NAME)` in that other
window too.

Before you begin, set these environment variables:

```
$ export APP_IMAGE=duglin/helloworld
$ export REBUILD_IMAGE=duglin/rebuild
$ export GITREPO=duglin/helloworld
```

Set `APP_IMAGE` and `REBUILD_IMAGE` values to use your
Dockerhub namespace name instead of `duglin`, and change `GITREPO`
to be the name of your Github clone of this repo - typically you should
just need to swap `duglin` for your Github name.

### Using the `demo` script

To have the computer type all of the demo commands for you, just run:

```
$ ./demo [ CLUSTER_NAME ]
```
`CLUSTER_NAME` is optional if your `KUBECONFIG` environment variable
already points to your IKS cluster.

When the demo is done you can jump the the [Cleaning Up](#cleaning-up) section.


### Manually running the demo / Demo details

This section will walk through the steps executed by the `demo` script
providing commentary as you do each step.

#### Secrets

Before we get to the real point of this, which is deploying an application,
we needed to create a Kuberneres Secret that holds all of the private
keys/tokens/usernames/etc... that will be used during the demo.
We also needed to create some new RBAC rules so that our "rebuild" service,
which is running under the "default" Service Account,
has the proper permissions to access and edit the Knative Service to force it
to rebuild - which I'll talk more about later.

For all of these things I have the `secrets.yaml` file:

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
  username: ${.secrets.username}
  password: ${.secrets.password}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-bot
secrets:
- name: mysecrets

---
# Give our "default" ServiceAccount permission to touch Knative Services
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: rebuild
rules:
- apiGroups:
  - serving.knative.dev
  resources:
  - services
  verbs:
  - get
  - list
  - update
  - patch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rebuild-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rebuild
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
```

You'll notice some environment variable looking values in there. Obviously,
those are not normal Kube yaml things. To make life easier, I created
a script called `kapply` which takes a yaml file and replaces references
like those with their real values before invoking `kubectl apply`. This allows
me to share my yaml with you w/o asking you to modify these files and run
the risk of you checking them into your github repo by mistake. All you need
to do is create the `.secrets` files I mentioned previously.

Then you can create the secret via:

```
$ ./kapply secrets.yaml
secret/mysecrets created
serviceaccount/build-bot created
clusterrole.rbac.authorization.k8s.io/rebuild created
clusterrolebinding.rbac.authorization.k8s.io/rebuild-binding created
```

#### Our application

For this demo I'm just using a very simple HTTP server that responds
to any request with `Hello World!` plus whatever text is in the `MSG`
environment variable.

Here's the source
([`helloworld.go`](https://github.com/duglin/helloworld/blob/master/helloworld.go)):

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

    if os.Getenv("MSG") != "" {
        text += " " + os.Getenv("MSG")
    }

    rev := os.Getenv("K_REVISION") // K_REVISION=helloworld-7vh75
    if i := strings.LastIndex(rev, "-"); i > 0 {
        rev = rev[i+1:] + ": "
    }

    msg := fmt.Sprintf("%s%s\n", rev, text)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Printf("Got request\n")
        time.Sleep(500 * time.Millisecond)
        fmt.Fprint(w, msg)
    })

    fmt.Printf("Listening on port 8080 (rev: %s)\n", rev)
    http.ListenAndServe(":8080", nil)
}
```

The `sleep` is in there just to slow things down a bit so that when we
increase the load on the app it'll cause one instance of the app to be
created for each client we have generating requests. It makes for a better
live demo experience.

The app will print part of the `K_REVISION` environment variable so
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
as a Knative Service. First, do not confuse a Knative Service with
a Kubernetes Service - they're not the same thing. This is an on-going
point of discussion within the Knative community, so for now we just
need to live with it.

We'll first deploy our Knative Service the really easy way, via the
Knative `kn` command line tool. The source code is available at
https://github.com/knative/client .
For convinience, I've included the `kn` exectuable in this directory - but
 it's only for Linux.

```
$ ./kn service create helloworld --image ${APP_IMAGE}
Service 'helloworld' successfully created in namespace 'default'.
Waiting for service 'helloworld' to become ready ... OK

Service URL:
http://helloworld-default.v06.us-south.containers.appdomain.cloud
```

As should be clear from the first two arguments, this `kn` command is creating
a service. The next argument is the service name (`helloworld`) and
then we give it the name/location of the container image to use.

Once the command is done, you'll see that it shows you the URL where the
service is available. IKS will automatically configure the networking
infrastructure for you and give you a nice DNS name so you don't need
to do any of that yourself.

With that URL you can now hit (curl) it:

```
$ curl -sf http://helloworld-default.v06.us-south.containers.appdomain.cloud
krc9z: Hello World!
```

With that, you've now successfully deploy a Knative Service and behind
the scenes it created all of the Kubernetes resources to host it, scale
it, route traffic to it and even give it a relatively nice URL.

One more thing.... you can also access it via SSL:
```
$ curl -sf https://helloworld-default.v06.us-south.containers.appdomain.cloud
krc9z: Hello World!
```

So, you also get security too! All with one simple command.

Let's take a quick look at all of the resources that were created
for us:

```
$ ./showresources all
deployment.apps/helloworld-krc9z-deployment
endpoint/helloworld-krc9z
endpoint/helloworld-krc9z-metrics
endpoint/helloworld-krc9z-priv
endpoint/kubernetes
pod/helloworld-krc9z-deployment-7875dcb55-cdksl
replicaset.apps/helloworld-krc9z-deployment-7875dcb55
service/helloworld
service/helloworld-krc9z
service/helloworld-krc9z-metrics
service/helloworld-krc9z-priv
service/kubernetes

clusterchannelprovisioner.eventing.knative.dev/in-memory
clusteringress.networking.internal.knative.dev/route-032c8cb8-6faa-4bdb-8fcf-4db63337b91b
configuration.serving.knative.dev/helloworld
image.caching.internal.knative.dev/helloworld-krc9z-cache
podautoscaler.autoscaling.internal.knative.dev/helloworld-krc9z
revision.serving.knative.dev/helloworld-krc9z
route.serving.knative.dev/helloworld
serverlessservice.networking.internal.knative.dev/helloworld-krc9z
service.serving.knative.dev/helloworld
```

The first list is the list of core Kube resources, and the second list
contains the Knative ones. That's a lot of stuff!  And I mean that in a
good way! One of my hopes for Knative is to offer up a more user
friendly user experience for Kube users. Sure Kube has a ton of features
but with that flexibility has come complexity. Think about how much learning
and work is required to setup all of these resources that Knative as done
for us. We no longer need to understand Ingress, load-balancing, auto-scaling,
etc. Very nice!

Back to our demo...

The `kn` command line is still under development so for the rest of this
write-up I'm going to switch back to the normal Kubernetes CLI, `kubectl`.
And to do that, let's delete this service so we can see how to do it
with `kubectl`.

```
$ ./kn service delete helloworld
Service 'helloworld' successfully deleted in namespace 'default'.
```

Let's first look at our yaml file that defines our Knative Service
defined in `service1.yaml`:

```yaml
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: helloworld
spec:
 template:
    spec:
      containers:
        - image: duglin/helloworld
```

Let's explain what some of these fields do:
- `metadata.name`: Like all Kube resources, this names our Service, and
  matches what we put on the `kn` command line.
- `spec.template`: One concept that might take some getting used to is the
  notion that the `spec.template` section of a Knative Service resource doesn't
  really define the Service itself directly. The better way to think of it is
  that you provide it a "template" for what the next version (revision) of your
  service should look like. While in practice, yes the `spec.template` will
  normally match what the service currently looks like, it may be important
  at times to understand that behind the scenes each version of your service
  has an associated Revison resource with its own unique configuration 
  information - which is populated from this `template`.  But for now,
  it's ok to think of `spec.template` as your desired state for the latest
  version of the Service.`
- `spec.template.spec`: This is taken directly from the Kubernetes Pod
  spec defintion. Meaning, you can put anything from the Pod spec in here
  to define your app. However, while from a syntax perspective you should
  be able to copy-n-paste your Pod spec into here, since Knative doesn't
  support all of the Pod's features you may get an error message in
  some cases if you use a feature that Knative doesn't support yet.
- `spec.template.spec.containers.image`: This just defines which container
  image this version of the Service should use.  Same as what we saw on
  the `kn` command line.

For the most part you'll see that there really are only 2 pieces of information
in here - the service name and the image name. Same as the `kn` command line.
There's just more yaml because, well, it's yaml and Kubernetes.

Now we can deploy it:

```
$ ./kapply service1.yaml
service.serving.knative.dev/helloworld created
```

If you're not running `pods` in another window, run it now to see what
happened:

```
$ ./pods
Cluster: v06
NS/NAME                        LATESTREADY                    READY
helloworld                     helloworld-9frrr               True    

POD_NAME                                                STATUS           AGE
helloworld-9frrr-deployment-65b66d976d-j8q7w            Running          5s
```

The output of `pods` shows the list of Knative services (at the top)
followed by the list of active pods.

When the pod is in the `Running` state, press control-C to stop it.

You should see your `helloworld` Knative service with one revision
called something like `helloworld-9frrr`, and a pod with a really funky name
but that starts with that revision name.

Notice the word "deployment" in there - that's because under the covers
Knative created a Kubernetes Deployment resource and this pod is related
to that Deployment.

So, it's running - let's invoke it, and we can use the same URL as before
when we used the `kn` command. But if you forgot it, you can do this:

```
$ kubectl get ksvc
NAME         DOMAIN                                                             LATESTCREATED      LATESTREADY        READY   REASON
helloworld   http://helloworld-default.v06.us-south.containers.appdomain.cloud  helloworld-9frrr   helloworld-9frrr   True
```

You'll notice that once the Service is ready the "DOMAIN" column will show
the full URL of the Service.

```
$ curl -sf http://helloworld-default.v06.us-south.containers.appdomain.cloud
9frrr: Hello World!
```

That's it. Notice that the deployment of the Service was pretty easy whether
you use the `kn` or the `kubectl` command line tool. And that's a huge
step forward for Kubernetes users.

#### Updating Our Service

Now that we have our app running, let's make some changes and roll out
a new version. Let's look at the next yaml file we'll use:

```
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: helloworld
spec:
  template:
    metadata:
      name: helloworld-v1
    spec:
      containers:
      - image: duglin/helloworld
      containerConcurrency: 1
  traffic:
  - tag: helloworld-v1
    revisionName: helloworld-v1
    percent: 100
```

You should notice 2 main differences here:

- `spec.template.metadata.name`: This is different from the Service
  name you'll see above on the 4th line. This provides a name for **this**
  revision of your service. We're going to provide a name so that we can
  reference it in the `traffic` section below.

- `spec.traffic`: This section allows for us to tell Knative how to
  route traffic between the various versions (revisions) of our
  Service. This is useful when you want to do a rolling (A/B) update
  of your service rather than switch all traffic to the new version
  immediately.

  In this case we're defining just one `traffic` section and telling
  Knative to send all (100%) of the traffic to the revision called
  `helloworld-v1` - which is what we're naming the new revision.
  You'll notice there's `tag` property too - this allows us to give
  a unique prefix to a specialize URL for just that one revision if
  you want to send request directly to it and avoid the percentage
  based routing logic. In this case, to keep it simple, we picked the
  same name as the revision.

Once final thing, there's also now `containerConcurrency` property in
there. We're setting that to `1` so that each instance of the service will
only process one request at time, instead of multiple. This makes for a
better demo when we generate a load since it'll force the Service to scale
up pretty quickly.

With that, let's go ahead and apply this yaml to update our service:

```
$ ./kapply service2.yaml
service.serving.knative.dev/helloworld configured
```

In the "pods" window you should see something like this:

```
$ ./pods
Cluster: v06
NS/NAME                        LATESTREADY                    READY
helloworld                     helloworld-v1                  True

POD_NAME                                                STATUS           AGE
helloworld-9frrr-deployment-65b66d976d-j8q7w            Running          23s
helloworld-v1-deployment-6bfbd447c-8x8wp                Running          11s
```

Notice that we now have a 2nd revision and it's name is `helloworld-v1`.

Hitting it you should see:

```
$ curl -sf http://helloworld-default.v06.us-south.containers.appdomain.cloud
v1: Hello World!
```

Notice we use the same URL but now the version number on the output shows
`v1`.

#### Traffic Splitting

With that brief introduction to traffic management, let's actually do some
real traffic splitting by deploying a 3rd version of the app:

```
$ ./kapply -t service3.yaml
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: helloworld
spec:
  template:
    metadata:
      name: helloworld-v2
    spec:
      containers:
      - image: duglin/helloworld
        env:
        - name: MSG
          value: Goodnight Moon!
      containerConcurrency: 1
  traffic:
  - tag: helloworld-v1
    revisionName: helloworld-v1
    percent: 50
  - tag: helloworld-v2
    revisionName: helloworld-v2
    percent: 50
```

Couple of things to notice here:
- We named this version of the app `helloworld-v2`.
- We added an environment variable called `MSG` to the app. This just
  allows us to see some different output when we hit this version.
- We've modified the `traffic` section so that we now have two
  versions of the app that will get traffic. Each gets 50% and the two
  versions are the v1 and v2 versions. The initial version we deployed
  will eventually be removed from the system since we never reference it
  in a traffic section.

Let's deploy this:

```
$ ./kapply service3.yaml
service.serving.knative.dev/helloworld configured
```

We can now generate some load on our service and we should see roughly
1/2 of the traffic go to v1 and 1/2 go to v2:

```
$ ./load 10 10 http://helloworld-default.v06.us-south.containers.appdomain.cloud

01: v1: Hello World!                                                            
02: v1: Hello World!                                                            
03: v1: Hello World!                                                            
04: v2: Hello World! Goodnight Moon!                                            
05: v1: Hello World!                                                            
06: v2: Hello World! Goodnight Moon!                                            
07: v2: Hello World! Goodnight Moon!                                            
08: v2: Hello World! Goodnight Moon!                                            
09: v2: Hello World! Goodnight Moon!                                            
10: v2: Hello World! Goodnight Moon!                                            
```

The `load` command (as specified) will simulate 10 clients hitting the URL
for 10 seconds. Notice we do in fact see both versions getting hit.

#### Hidden Traffic

The traffic split we did the previous section was nice, but often times
when you deploy a new version of an app you don't want it exposed to your
users right away. Perhaps you only want your tests to hit it. Knative
can help here too. Let's look at another version of the yaml:

```
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: helloworld
spec:
  template:
    metadata:
      name: helloworld-v2
    spec:
      containers:
      - image: duglin/helloworld
        env:
        - name: MSG
          value: Goodnight Moon!
      containerConcurrency: 1
  traffic:
  - tag: helloworld-v1
    revisionName: helloworld-v1
    percent: 50
  - tag: helloworld-v2
    revisionName: helloworld-v2
    percent: 50
  - tag: helloworld-test
    latestRevision: true
    percent: 0
```

Here we didn't modify anything about the `template` but we did add a new
`traffic` section. In there we told Knative a couple of things. First,
we gave it a name via the `tag` - `helloworld-test`. As before, this is
just some unique prefix for a dedicated URL we can use to talk to the revision
pointed to by this traffic section. Next, the `latestRevision` property.
If we have a name of a version we wanted to reference we could have put
that name on `revisionName` property, but in this case we're asking
Knative to make this traffic section point to the very latest version
of the app regardless of what its name is.

Finally, we told Knative to send no traffic (`0` percent)
to this version we're pointing to. This means that when people hit our main
URL they'll never talk to this revision. The only way we can talk to this
one is via the dedicated URL that Knative will setup for us. One thing
to note is that as of right now the "latest" version of the app just
happens to be "v2" so both the 2nd and 3rd traffic sections actually
point to the same version - and that's ok.

We're doing all of this because in the next section we'll be updating our
app and I want to show how only traffic to the `helloworld-test` URL
will be able to see it.

Let's deploy it:

```
$ ./kapply service4.yaml
service.serving.knative.dev/helloworld configured
```

#### Adding Build

In this part of the demo we're going to connect our Github account up to
our service so that when a new "push" to the "master" branch happens we'll
kick off a build of our app and deploy it.

To do this we'll first need to deply a "rebuild" service. This is really
nothing more than a Knative Service that waits for an incoming HTTP request
from Github's webhook system. When that request indicates that it was a
"push" event on our git repo, we'll kick off a build of the latest version
of the source, build a new container image and the deploy a new version of
our app using that image.

I'm not going to get into how the "rebuild" service works, but if you
really want to see it you can look at the source code:
[rebuild.go](./rebuild.go).

To deploy it we'll apply this yaml:

```
apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:
  name: rebuild
spec:
  template:
    spec:
      containers:
        - image: ${REBUILD_IMAGE}
          env:
          - name: KSVC
            value: helloworld
          - name: GITREPO
            value: ${GITREPO}
          - name: APP_IMAGE
            value: ${APP_IMAGE}
```

You'll notice that we pass in some environment variables telling it which
git repo to get the source from, which Knative service to update, and
where to store the built image.

Let's execute that:

```
$ ./kapply rebuild.yaml
service.serving.knative.dev/rebuild created
```

Now we need to connect this "rebuild" service up to Github. To do this
we'll use Knative Eventing. As part of Eventing there are things called
"importers". You can think of these as utilities that help you subscribe and
manage events from event producers. In this case we're going to deploy
a Github importer (also known as an "event source") that will talk to Github
for us to create the webhook for our git repo and tell it to send events
to the importer, who will then pass it on to our rebuild service.

One thing I need to mention, the event that is passed on to the Knative
service is a [CloudEvent](https://cloudevents.io). All that means is that
some common metadata about the event (such as who sent it or what its
`type` is) are in well-defined locations so that regardless of who the
event producer is, your code should be able to get that information by
just looking at one specific location. Just to make life a little easier.

To set this up we'll use this yaml:

```
apiVersion: sources.eventing.knative.dev/v1alpha1
kind: GitHubSource
metadata:
  name: githubsource
spec:
  eventTypes:
    - push
  ownerAndRepository: duglin/helloworld
  accessToken:
    secretKeyRef:
      name: mysecrets
      key: git_accesstoken
  secretToken:
    secretKeyRef:
      name: mysecrets
      key: git_secrettoken
  sink:
    apiVersion: serving.knative.dev/v1alpha1
    kind: Service
    name: rebuild
```

Walking through this:
- you'll see the types of events (`push`) that we're asking for.
- the `ownerAndRepository` property tells it which git repo to subscribe to.
- the `accessToken` contains the credentials needed for it to talk to Github
  for me - and it gets those credentials from the Kubernetes secret referenced
  in there.
- `secretToken` is just the auth token we expect Github to use on the events
  sent to us so we can verify they're actually related to this
  subscription/webhook.
- And finally, we provide a `sink` which is where the event should be sent to.
  Which in this case is our 'rebuild' service.

Technically, we could have done of this ourselves, but this just makes it
easier. And when we delete this resource it'll delete the webhook for us too.

Let's deploy it:

```
$ ./kapply github.yaml
githubsource.sources.eventing.knative.dev/githubsource created
```

If you go to your gitrepo repo's webhook page you should see an entry
listed in there for it - and it should look something like this:

![Github Webhooks](./webhooks.png "Github Webhooks")

Don't be surprised if the green check is actually a red "X", sometimes
the first (a verification message) has issues.

#### Editing Our App

With that infrastructure ready, let's edit our app and see it all in action.

In your favorite editor, edit the "Hello World" message to something else.
Or just run this command:

```
$ sed -i "s/text :=.*/text := \"Dogs rule!! Cats drool!!\"/" helloworld.go
```

Then let's commit it and push it to github:

```
$ git add helloworld.go

$ git commit -s -m "demo - Fri Aug 2 07:39:39 PDT 2019"
[master 2173aed] demo - Fri Aug  2 07:39:39 PDT 2019
 2 files changed, 4 insertions(+), 2 deletions(-)

$ git push origin master
To git@github.com:duglin/helloworld.git
   6b580c0..2173aed  master -> master
```

The entire build process takes about 30 seconds, but in your "pods" window
you should see the "rebuild" and "githubsource" pods appear, as well as some
"build" pod. Eventually a new revision pod for the service should appear too.

If you `curl` the "helloworld-test" revision (via the special URL setup
under the Traffic section), you should see something like this:

```

$ curl -sf http://helloworld-helloworld-test-default.v06.us-south.containers.appdomain.cloud
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
hwwx2: Dogs rule!! Cats drool!! Goodnight Moon!
```

Notice that it will show v2 of the app until the re-deploy is done and then
all traffic will switch over the latest version. Keep in mind, this is
only what the special "helloworld-test" URL will see because that's the only
one pointing to the latest version of the app.

Likewise, we could also hit just v1 or v2 directly:

```
$ curl -sf http://helloworld-helloworld-v1-default.v06.us-south.containers.appdomain.cloud
v1: Hello World!

$ curl -sf http://helloworld-helloworld-v2-default.v06.us-south.containers.appdomain.cloud
v2: Hello World! Goodnight Moon!
```

The main URL for the app is still routing traffic between v1 and v2.

To prove that run this:

```
$ for i in `seq 1 20` ; do curl -s http://helloworld-default.v06.us-south.containers.appdomain.cloud ; done
v1: Hello World!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v1: Hello World!
v1: Hello World!
v1: Hello World!
v2: Hello World! Goodnight Moon!
v1: Hello World!
v1: Hello World!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v1: Hello World!
v2: Hello World! Goodnight Moon!
v2: Hello World! Goodnight Moon!
v1: Hello World!
v2: Hello World! Goodnight Moon!
v1: Hello World!
v1: Hello World!
```

Notice that we never see the latet version that talks about dogs and cats.

#### Scary!

Now, just to show you how much Knative has saved us with respect to managing
Kubernetes resources, let's run `showresources` again:

```
$ ./showresources all
deployment.apps/githubsource-7bk5d-24srp-deployment
deployment.apps/helloworld-9frrr-deployment
deployment.apps/helloworld-hwwx2-deployment
deployment.apps/helloworld-v1-deployment
deployment.apps/helloworld-v2-deployment
deployment.apps/rebuild-xkdrt-deployment
endpoint/githubsource-7bk5d-24srp
endpoint/githubsource-7bk5d-24srp-metrics
endpoint/githubsource-7bk5d-24srp-priv
endpoint/helloworld-9frrr
endpoint/helloworld-9frrr-metrics
endpoint/helloworld-9frrr-priv
endpoint/helloworld-hwwx2
endpoint/helloworld-hwwx2-metrics
endpoint/helloworld-hwwx2-priv
endpoint/helloworld-v1
endpoint/helloworld-v1-metrics
endpoint/helloworld-v1-priv
endpoint/helloworld-v2
endpoint/helloworld-v2-metrics
endpoint/helloworld-v2-priv
endpoint/kubernetes
endpoint/rebuild-xkdrt
endpoint/rebuild-xkdrt-metrics
endpoint/rebuild-xkdrt-priv
pod/build-helloww94q-pod-50369b
pod/githubsource-7bk5d-24srp-deployment-ffb59c865-6wq5g
pod/helloworld-hwwx2-deployment-7d7df9558c-wl55z
pod/helloworld-v1-deployment-6bfbd447c-5wm8m
pod/helloworld-v2-deployment-77bbc9d98c-9jn8g
pod/rebuild-xkdrt-deployment-7b5c4648c6-v2pxk
replicaset.apps/githubsource-7bk5d-24srp-deployment-ffb59c865
replicaset.apps/helloworld-9frrr-deployment-65b66d976d
replicaset.apps/helloworld-hwwx2-deployment-7d7df9558c
replicaset.apps/helloworld-v1-deployment-6bfbd447c
replicaset.apps/helloworld-v2-deployment-77bbc9d98c
replicaset.apps/rebuild-xkdrt-deployment-7b5c4648c6
service/githubsource-7bk5d
service/githubsource-7bk5d-24srp
service/githubsource-7bk5d-24srp-metrics
service/githubsource-7bk5d-24srp-priv
service/helloworld
service/helloworld-9frrr
service/helloworld-9frrr-metrics
service/helloworld-9frrr-priv
service/helloworld-helloworld-test
service/helloworld-helloworld-v1
service/helloworld-helloworld-v2
service/helloworld-hwwx2
service/helloworld-hwwx2-metrics
service/helloworld-hwwx2-priv
service/helloworld-v1
service/helloworld-v1-metrics
service/helloworld-v1-priv
service/helloworld-v2
service/helloworld-v2-metrics
service/helloworld-v2-priv
service/kubernetes
service/rebuild
service/rebuild-xkdrt
service/rebuild-xkdrt-metrics
service/rebuild-xkdrt-priv
taskrun.tekton.dev/build-helloww94q

clusterchannelprovisioner.eventing.knative.dev/in-memory
clusteringress.networking.internal.knative.dev/route-42b0f459-578b-4fbe-845c-90017af34a74
clusteringress.networking.internal.knative.dev/route-7422c2f8-9359-4f58-a858-af21a24d8fa3
clusteringress.networking.internal.knative.dev/route-752e3da4-e6f2-42fb-aae3-cb885ab55a36
configuration.serving.knative.dev/githubsource-7bk5d
configuration.serving.knative.dev/helloworld
configuration.serving.knative.dev/rebuild
githubsource.sources.eventing.knative.dev/githubsource
image.caching.internal.knative.dev/githubsource-7bk5d-24srp-cache
image.caching.internal.knative.dev/helloworld-9frrr-cache
image.caching.internal.knative.dev/helloworld-hwwx2-cache
image.caching.internal.knative.dev/helloworld-v1-cache
image.caching.internal.knative.dev/helloworld-v2-cache
image.caching.internal.knative.dev/rebuild-xkdrt-cache
podautoscaler.autoscaling.internal.knative.dev/githubsource-7bk5d-24srp
podautoscaler.autoscaling.internal.knative.dev/helloworld-9frrr
podautoscaler.autoscaling.internal.knative.dev/helloworld-hwwx2
podautoscaler.autoscaling.internal.knative.dev/helloworld-v1
podautoscaler.autoscaling.internal.knative.dev/helloworld-v2
podautoscaler.autoscaling.internal.knative.dev/rebuild-xkdrt
revision.serving.knative.dev/githubsource-7bk5d-24srp
revision.serving.knative.dev/helloworld-9frrr
revision.serving.knative.dev/helloworld-hwwx2
revision.serving.knative.dev/helloworld-v1
revision.serving.knative.dev/helloworld-v2
revision.serving.knative.dev/rebuild-xkdrt
route.serving.knative.dev/githubsource-7bk5d
route.serving.knative.dev/helloworld
route.serving.knative.dev/rebuild
serverlessservice.networking.internal.knative.dev/githubsource-7bk5d-24srp
serverlessservice.networking.internal.knative.dev/helloworld-9frrr
serverlessservice.networking.internal.knative.dev/helloworld-hwwx2
serverlessservice.networking.internal.knative.dev/helloworld-v1
serverlessservice.networking.internal.knative.dev/helloworld-v2
serverlessservice.networking.internal.knative.dev/rebuild-xkdrt
service.serving.knative.dev/githubsource-7bk5d
service.serving.knative.dev/helloworld
service.serving.knative.dev/rebuild
```

That's a ton of stuff we'd have to manage ourselves!

#### Tekton

I didn't talk about how the build of the app itself is done, and I won't go
into too many details other than to say it uses [Tekton](https://tekton.dev)
under the covers.

You can think of Tekton in the same light as a Jenkins system. It will run
a set of steps that you tell it. And if you look at [task.yaml](./task.yaml)
you'll see the yaml file that defines those steps. In this case, it's just
two steps

- run a container called `kaniko-project/executor` that will do a Docker
  build and push the resulting image to a container registry

- issue a curl command (via the `duglin/curl` container) to poke the
  rebuild service so it knows when the new image is ready so it can
  deploy the next version of the service

Ideally, I would have preferred to make it so that the "rebuild" service
got a real event from DockerHub directly when the push happened so that I
didn't have to do the `curl` command, but there is no DockerHub importer
yet - but maybe one day.

There are also some niceties in there, like by specifying an input as a
github repo it'll do the clone for me automatically. Which is kind of nice.

The intersting thing about this to me is that I still can't quite decide
how Tekton will be used. By that I mean, will people create tasks with lots
of individual steps (similar to the steps they might see in a Makefile),
or will the steps be fairly large which would results in a relatively small
number of them in each job.

I also wonder if most jobs will end up being close to just one step, and
people will encode a ton of logic into there - more like just wrappering
your `make` command in a container. And if so, it'll be interesting to see
where people see the benefit of Tekton vs just running the container
directly.

## Summary

With that we're run through a pretty extensive demo. To recap, we:
- deployed a pre-built container image as a Knative service, but via
  `kn` and `kubectl`
- showed how to create multiple version of the service and route traffic
  to each based on a percentage.
- showed how to create an event "importer" to subscribe to an event producer
  and have those events sent on to a Knative service
- showed how to bring it all together to have a mini CI/CD pipeline
  that will build and redeploy our service as we make changes to its
  github repo

Overall, what's out there today is a HUGE useability improvement
for developers. The idea of being able to create and manipulate the Kubernetes
resource to do all of these (what I consider) advanced semantics is really
very exciting, and I can't wait to see how things progress.

## Cleaning up

To clean the system so you can run things over and over, just do:

```
$ ./demo --clean
```

It should delete everything except the cluster, Istio and Knative.
