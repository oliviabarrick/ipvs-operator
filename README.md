Kubernetes operator for setting IPVS weights on Kubernetes ClusterIP services.

The operator provides a new Kubernetes resource called a WeightedService.

![build status](https://ci.codesink.net/api/badges/justinbarrick/ipvs-operator/status.svg)
[![image version](https://images.microbadger.com/badges/version/justinbarrick/ipvs-operator.svg)](https://microbadger.com/images/justinbarrick/ipvs-operator)
[![image size](https://images.microbadger.com/badges/image/justinbarrick/ipvs-operator.svg)](https://microbadger.com/images/justinbarrick/ipvs-operator "Get your own image badge on microbadger.com")

This operator is in very early alpha and should be used with care.

Note that the service weights will not apply for anything that routes directly to
endpoint IPs (e.g., ingress-nginx), but will work for the service IP.

# Installation

Follow the [IPVS guide](https://github.com/kubernetes/kubernetes/tree/master/pkg/proxy/ipvs) to configure
your Kubernetes cluster to use IPVS instead of iptables.

To install the operator, simply apply the manifest:

```
kubectl apply -f https://raw.githubusercontent.com/justinbarrick/ipvs-operator/master/deploy/operator.yaml
```

This will install the IPVS operator as a DaemonSet on all of your nodes.

# WeightedServices

You can use a WeightedService to apply IPVS load balancing weights to pods matching certain labels.

A WeightService can enable canary deployments by setting different weights for your canary and production
deployments. See [the IPVS weighted round-robin documentation](http://kb.linuxvirtualserver.org/wiki/Weighted_Round-Robin_Scheduling)
for more information. If a scheduler is not set, it defaults to "wrr".

For example, to send 10% of traffic to your canary deployment:

```
apiVersion: codesink.net/v1alpha1
kind: WeightedService
metadata:
  name: example
spec:
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: www
  scheduler: wlc
  weights:
  - weight: 10
    selector:
      env: prod
  - weight: 1
    selector:
      env: canary
```

The WeightedService will create a Service matching the provided ServiceSpec and assign weights to
pods that match the specified labels.

See `test/example.yaml` for a full example.
