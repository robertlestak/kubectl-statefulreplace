# kubectl statefulreplace

This is a simple tool to replace a stateful workload in Kubernetes in a manner which ensures that downstream clients never gets load balanced between the old and new versions during a rolling update. This is useful with client-facing services which are not tolerant of downtime and yet cannot serve traffic from multiple versions at the same time (such as PWAs without a caching CDN).

A more robust strategy would be to use a service mesh like Istio, and take advantage of the traffic shifting features it provides. However, this tool is useful in situations where a service mesh and CDN is not available.

## How it works

Given an existing Kubernetes scaled resource, this tool will:

- Scale down the replicas to 1
- Wait for all traffic to be drained from the old version
- Patch the resource with the new image
- Scale up the replicas to the original value
- Wait for all traffic to be routed to the new version

## Installation

The tool is designed to be used as a kubectl plugin. To install it, run:

```bash
make install
```

## Usage

Once installed in your `PATH`, you can run it as follows:

```bash
~ kubectl statefulreplace -h
Usage:
kubectl statefulreplace -n <namespace> [kind]/[name] [container]/[image] [container]/[image] ...
kubectl statefulreplace -f <config-file>
  -f string
        Config file
  -log-level string
        Log level (default "info")
  -n string
        Namespace
  -version
        Version
```

`kubectl statefulreplace` can be used in two ways:
- By specifying the namespace, kind, name, and container/image pairs as arguments
- By specifying a replacement file

### Example: Updating a Deployment using arguments

```bash
# create a deployment
kubectl create deployment nginx --image=nginx:1.7.8 --replicas=3

# replace the deployment
kubectl statefulreplace deployment/nginx nginx/nginx:1.7.9
```

### Example: Updating a Deployment using a replacement file

```bash
# create a deployment
kubectl create deployment nginx --image=nginx:1.7.8 --replicas=3

# create a replacement file
cat > replacement.yaml <<EOF
kind: Deployment
name: nginx
namespace: default
replacements:
- container: nginx
  image: nginx:1.7.9
EOF

# replace the deployment
kubectl statefulreplace -f replacement.yaml
```

See `examples` for more manifest examples.

