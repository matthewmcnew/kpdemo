# Minikube Setup

Minikube allows you to run the entire demo on your local machine.

### Setup:

Start a minikube cluster
```bash
minikube start --memory=4096 --cpus=4 --vm-driver=hyperkit --bootstrapper=kubeadm --insecure-registry "registry.default.svc.cluster.local:5000"
```

Update `/etc/hosts` by adding the name registry.default.svc.cluster.local on the same line as the entry for localhost. It should look something like this:
```bash
##
127.0.0.1       localhost registry.default.svc.cluster.local
255.255.255.255 broadcasthost
::1             localhost
```

Update the minikube `/etc/hosts` with the host ip for registry.default.svc.cluster.local
 ```bash
minikube ssh \
"echo \"192.168.64.1       registry.default.svc.cluster.local\" \
| sudo tee -a  /etc/hosts"
```

Install the kubernetes dependencies

```bash
kubectl apply -f https://raw.githubusercontent.com/matthewmcnew/build-service-visualization/master/setup/minikube/service.yaml
```

Make sure the registry(s) are running on your machine
```bash
docker run -d -p 5000:5000 registry:2
```

Install kpack
```bash
kubectl apply -f https://storage.googleapis.com/beam-releases/out.yaml
```

### Demo:

```bash
pbdemo populate --registry registry.default.svc.cluster.local:5000/please --count 15
```
