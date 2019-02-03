#!/bin/bash
set -ex
echo
env | sort  # just for debugging
echo

#  Log into IBM Cloud and set our KUBECONFIG
bx login --apikey ${IC_KEY} -r us-south
bx config --check-version false
bx ks cluster-config -s --export ${CLUSTER} 2>&1 | tee out
$(cat out | grep KUBECONFIG=)
echo

# Get the service's yaml, tweak it then 'apply' it
kubectl get ksvc/helloworld -o yaml > yaml
grep trigger yaml
sed -i "s/\(trigger\).*/\1: \"$(date +%s)\"/" yaml
grep trigger yaml
kubectl apply -f yaml

# Show the results
kubectl get ksvc/helloworld -o yaml
