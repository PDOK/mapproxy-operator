#!/bin/bash

TAG=$1

currentcontext=$(kubectx)

if [[ $? != 0 || $currentcontext != "default" ]]; then
  echo "Local cluster could not be detected"
  exit 1
fi

echo "Running: make generate"
make generate

echo ""
echo "Running: build -t local-registry:5000/mapproxy-operator:$TAG --build-context repos=./.. ."
docker build -t "local-registry:5000/mapproxy-operator:$TAG" --build-context repos=./.. .

echo ""
echo "Running: push local-registry:5000/mapproxy-operator:$TAG"
docker push "local-registry:5000/mapproxy-operator:$TAG"

if [[ $(kubectl get pod -l app=webhook -n cert-manager | grep "cert-manager") ]]; then
  echo "Cert-manager already installed"
else
  echo ""
  echo "Installing cert-manager"
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.17.0/cert-manager.yaml
fi

echo "Waiting for cert-manager"
while [[ $(kubectl get pod -l app=webhook -n cert-manager -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do
  sleep 1
done
echo "Cert-manager ready"

echo ""
echo "Running: make install"
make install

echo ""
echo "Running: deploy IMG=local-registry:5000/mapproxy-operator:$TAG"
make deploy "IMG=local-registry:5000/mapproxy-operator:$TAG"