# registrar

Registrar handles provisioning a IoT Kubernetes Cluster.

## How?

We use WireGuard to bridge all of our devices together into a single vxlan, and then use flannel in host-gw mode.

## Instructions

Creating `registrard` secret:

````bash
$ kubectl create secret --namespace registrar generic --from-literal="REGISTRARD_TOKEN=$(base64 /dev/urandom | tr -d '/+' | dd bs=128 count=1 2>/dev/null)" --from-literal="CLUSTER_TOKEN=<from /var/lib/rancher/k3s/server/node-token on server>" registrard

### Creating a Server Node

There's currently a "chicken" and an egg problem when it comes to provisioning a new Kubernetes cluster, which is telling Kubernetes to run and publish the WireGuard IP. However, we create WireGuard using `registrard`. To get around we temporarily create a 10.10.0.1/24 address, which registrar will clean up when it runs.

```bash
sudo ip addr add 10.10.0.1/24 dev lo
````

Now we need to run `registrard`, you can do this:

```bash
# TODO(jaredallard): Add this when we have manifests... and the thing actually works

# Create the rancher secret
kubectl create secret --namespace registrard generic --from-literal="RANCHER_TOKEN=$RANCHER_TOKEN" rancher

# Create TLS secrets
kubectl create secret --namespace registrard generic --from-file="service.pem=../credentials/service.pem" --from-file="service.key=../credentials/service.key" tls
```

Needed IPTables rules:

```
iptables -A INPUT -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i wg0 -o wg0 -m conntrack --ctstate NEW -j ACCEPT
iptables -t nat -A POSTROUTING -s 10.10.0.0/24 -o wg0 -j MASQUERADE
```

## License

Apache-2.0
