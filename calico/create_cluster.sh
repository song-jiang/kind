#!/bin/bash -e

# kubectl binary.
: ${kubectl:=kubectl}

# kind binary.
: ${KIND:=../bin/kind}

function checkModule(){
  MODULE="$1"
  echo "Checking kernel module $MODULE ..."
  if lsmod | grep "$MODULE" &> /dev/null ; then
    return 0
  else
    return 1
  fi
}

echo "kubernetes dualstack requires ipvs mode kube-proxy for the moment."
MODULES=("ip_vs" "ip_vs_rr" "ip_vs_wrr" "ip_vs_sh" "nf_conntrack_ipv4")
for m in "${MODULES[@]}"; do
  checkModule $m
  if [[ $? -eq 1 ]]; then
    echo "Could not find kernel module $m. Please install it..."
    exit 1
  fi
done
echo

${KIND} create cluster --image songtjiang/kindnode-dualstack:1.17.0 -v3 --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true
  ipFamily: DualStack
nodes:
# the control plane node
- role: control-plane
- role: worker
- role: worker
- role: worker
kubeadmConfigPatches:
- |
  apiVersion: kubeadm.k8s.io/v1beta2
  kind: ClusterConfiguration
  metadata:
    name: config
  featureGates:
    IPv6DualStack: true
- |
  apiVersion: kubeproxy.config.k8s.io/v1alpha1
  kind: KubeProxyConfiguration
  metadata:
    name: config
  mode: ipvs
EOF

${kubectl} get no -o wide
${kubectl} get po --all-namespaces -o wide

echo "Set ipv6 address on each node"
docker exec kind-control-plane ip -6 a a 2001:20::8/64 dev eth0
docker exec kind-worker ip -6 a a 2001:20::1/64 dev eth0
docker exec kind-worker2 ip -6 a a 2001:20::2/64 dev eth0
docker exec kind-worker3 ip -6 a a 2001:20::3/64 dev eth0
echo

echo "Install Calico for dualstack"
${kubectl} apply -f calico-3.10.0-dualstack.yaml
echo

echo "Wait Calico to be ready..."
while ! time ${kubectl} wait pod -l k8s-app=calico-node --for=condition=Ready -n kube-system --timeout=300s; do
    # This happens when no matching resources exist yet,
    # i.e. immediately after application of the Calico YAML.
    sleep 5
done
echo "Calico is running."
echo

echo "Create test-webserver deployment..."
kubectl apply -f test-webserver.yaml

echo "Wait webserver pods to be ready..."
while ! time ${kubectl} wait pod -l app=webserver --for=condition=Ready --timeout=300s; do
    sleep 5
done
echo "webserver pods are running."
echo

${kubectl} get po --all-namespaces -o wide
${kubectl} get svc
