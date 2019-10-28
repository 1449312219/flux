#!/usr/bin/env bash

set -o errexit

# This script runs the bats tests, first ensuring there is a kubernetes cluster available,
# with a flux namespace and a git secret ready to use

# Global variables to be used in the libraries/tests
export FLUX_NAMESPACE=flux-e2e
FLUX_ROOT_DIR=$(git rev-parse --show-toplevel)
export FLUX_ROOT_DIR
export E2E_DIR="${FLUX_ROOT_DIR}/test/e2e"
export FIXTURES_DIR="${E2E_DIR}/fixtures"
KNOWN_HOSTS=$(cat "${FIXTURES_DIR}/known_hosts")
export KNOWN_HOSTS
GITCONFIG=$(cat "${FIXTURES_DIR}/gitconfig")
export GITCONFIG
export DEMO_NAMESPACE=demo

KIND_VERSION="v0.5.1"
CACHE_DIR="${FLUX_ROOT_DIR}/cache/$CURRENT_OS_ARCH"
KIND_CACHE_PATH="${CACHE_DIR}/kind-$KIND_VERSION"
KIND_CLUSTER=flux-e2e
USING_KIND=false

# shellcheck disable=SC1090
source "${E2E_DIR}/lib/defer.bash"

# Check if there is a kubernetes cluster running, otherwise use Kind
if ! kubectl version > /dev/null 2>&1; then
  if [ ! -f "${KIND_CACHE_PATH}" ]; then
    echo '>>> Downloading Kind'
    mkdir -p "${CACHE_DIR}"
    curl -sL "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-${CURRENT_OS_ARCH}" -o "${KIND_CACHE_PATH}"
  fi
  echo '>>> Creating Kind Kubernetes cluster'
  cp "${KIND_CACHE_PATH}" "${FLUX_ROOT_DIR}/test/bin/kind"
  chmod +x "${FLUX_ROOT_DIR}/test/bin/kind"
  kind create cluster --name "${KIND_CLUSTER}" --wait 5m
  defer kind --name "${KIND_CLUSTER}" delete cluster > /dev/null 2>&1 || true
  KUBECONFIG="$(kind --name="${KIND_CLUSTER}" get kubeconfig-path)"
  export KUBECONFIG
  USING_KIND=true
  kubectl get pods --all-namespaces
fi

kubectl create namespace "$FLUX_NAMESPACE"
defer kubectl delete namespace "$FLUX_NAMESPACE"

if [ "${USING_KIND}" = 'true' ]; then
  echo '>>> Loading images into the Kind cluster'
  kind --name "${KIND_CLUSTER}" load docker-image 'docker.io/fluxcd/flux:latest'
fi

# Run the tests
echo '>>> Running the tests'
(
  cd "${E2E_DIR}"
  "${E2E_DIR}/bats/bin/bats" -t .
)
