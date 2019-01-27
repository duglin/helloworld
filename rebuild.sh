#!/bin/bash
set -ex
echo
env   # just for debuggin
echo

#  Log into IBM Cloud and set our KUBECONFIG
bx login --apikey ${IC_KEY} -r us-south
$(bx ks cluster-config --export ${CLUSTER})

# Get the service's yaml, tweak it then 'apply' it
kubectl get ksvc/helloworld -o yaml > yaml
grep gitsha yaml
sed -i "s/gitsha.*/gitsha=$RANDOM/" yaml
grep gitsha yaml
kubectl apply -f yaml

# Show the results
kubectl get ksvc/helloworld -o yaml
