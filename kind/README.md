# Kubernetes in Docker (or podman)

This doc describes how to run the operator using
[kind](https://kind.sigs.k8s.io/), a tool for running a cluster on your local
machine using just a container runtime tool such as podman.

## Podman works great

As tested on Fedora, you can install the `podman-docker` RPM, which installs a shell
script as `/usr/bin/docker` that simply runs:

```
exec /usr/bin/podman "$@"
```

kind works fine in this setup.

## Create a cluster

Create a cluster using the config file in this directory. The config file
contains ingress-related port forwarding.

```
kind create cluster --config=kind.config
```

## Setup Ingress

Apply the nginx ingress manifest as [shown in the kind docs](https://kind.sigs.k8s.io/docs/user/ingress/#ingress-nginx):

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

Wait for rollout to complete:

```
kubectl wait --namespace ingress-nginx   --for=condition=ready pod   --selector=app.kubernetes.io/component=controller   --timeout=90s
```

## Access the application

If you've run the operator and created a BoutiqueShop resource to deploy the
application, then access your application at
`[http://localhost:18080](http://localhost:18080)`.
