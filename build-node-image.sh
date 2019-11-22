./bin/kind build node-image --base-image aojea/kindbase:latest --image songtjiang/kindnode-dualstack:1.17.0-beta.1 --type docker --kube-root $GOPATH/src/k8s.io/kubernetes --loglevel debug
