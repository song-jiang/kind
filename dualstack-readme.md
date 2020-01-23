## Steps to create a dualstack kind cluster with Calico.

- Make sure kernel moduels for ipvs has been installed.
```
ip_vs
ip_vs_rr
ip_vs_wrr
ip_vs_sh
nf_conntrack_ipv4
```

- Git clone git@github.com:song-jiang/kind.git repo to $GOPATH/src/sigs.k8s.io/kind

- Checkout branch song-dualstack-01

- Run make build to build bin/kind

- cd into calico and run ./create_cluster.sh
