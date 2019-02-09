#!/bin/bash
set -ex
echo
env | sort  # just for debugging
echo

patch="[{\"op\":\"replace\",\"path\":\"/spec/runLatest/configuration/build/metadata/annotations/trigger\",\"value\":\"$(date +%s)\"}]"

kubectl patch ksvc/helloworld --type json -p="$patch"

# Show the results
kubectl get ksvc/helloworld -o yaml
