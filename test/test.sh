#!/bin/bash

set -e

# wait for up to $1 seconds for some command to return true
function wait_for {
    set +x
    set +e

    max_tries=$1
    count=0
    ret=1


    while [ $count -lt $max_tries ] && [ $ret -ne 0 ]; do
        ${@:2}
        ret=$?
        sleep 1
        count=$(($count + 1))
    done

    set -e
    set -x

    return $ret
}

function logs {
  kubectl logs -f $(kubectl get pods -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l name=ipvs-operator)
}

IMAGE=justinbarrick/ipvs-operator
TAG=test-$(date +%s)

make build
docker build -t $IMAGE:$TAG .
docker push $IMAGE:$TAG
kubectl set image deployment/ipvs-operator ipvs-operator=$IMAGE:$TAG

wait_for 30 logs
