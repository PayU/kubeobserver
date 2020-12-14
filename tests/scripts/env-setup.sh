set -e
set -o pipefail

case $(uname -m) in
	aarch64)	ARCH="arm64";;
	x86_64)		ARCH="amd64";;
	*)		ARCH="$(uname -m)";;
esac

KUBE_STATE_METRICS_LOG_DIR=./log
E2E_SETUP_KIND=${E2E_SETUP_KIND:-}
E2E_SETUP_KUBECTL=${E2E_SETUP_KUBECTL:-}
KIND_VERSION=v0.9.0
CLUSTER_NAME=kind-kubeobserver-test
TEST_APP_DOCKER_NAME=test-web-app
SUDO=${SUDO:-}

OS=$(uname -s | awk '{print tolower($0)}')
OS=${OS:-linux}


function finish() {
    echo "calling cleanup function"
    # kill kubectl proxy in background
    kill %1 || true
    kubectl delete -f examples/standard/ || true
    kubectl delete -f tests/manifests/ || true
}

function setup_kind() {
    curl -sLo kind "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-${OS}-${ARCH}" \
        && chmod +x kind \
        && ${SUDO} mv kind /usr/local/bin/
}

function setup_kubectl() {
    curl -sLo kubectl https://storage.googleapis.com/kubernetes-release/release/"$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"/bin/"${OS}"/"${ARCH}"/kubectl \
        && chmod +x kubectl \
        && ${SUDO} mv kubectl /usr/local/bin/
}

[[ -n "${E2E_SETUP_KIND}" ]] && setup_kind

echo "************** KIND VERSION **************"
kind version

[[ -n "${E2E_SETUP_KUBECTL}" ]] && setup_kubectl

mkdir "${HOME}"/.kube || true
touch "${HOME}"/.kube/config

export KUBECONFIG=$HOME/.kube/config

if [ -z ${SKIP_KIND_CLUSTER_CREATE+x} ]; then 
    kind create cluster --name $CLUSTER_NAME --config $PWD/tests/scripts/cloud.yaml

    set +e

    is_kube_running="false"

    # this for loop waits until kubectl can access the api server that kind has created
    for _ in {1..90}; do # timeout for 3 minutes
    kubectl get po 1>/dev/null 2>&1
    if [[ $? -ne 1 ]]; then
        is_kube_running="true"
        break
    fi

    echo "waiting for Kubernetes cluster to come up"
    sleep 2
    done

    if [[ ${is_kube_running} == "false" ]]; then
    echo "Kubernetes does not start within 3 minutes"
    exit 1
    fi

    set -e
else 
    echo "SKIP_KIND_CLUSTER_CREATE is set. skipping kind cluster creation";
fi

echo "************** KUBECTL VERSION **************"
kubectl version

echo "************** BUILDING TEST WEB-SERVER **************"
cd $PWD/tests/test-web-server
docker build -t $TEST_APP_DOCKER_NAME .
kind load docker-image $TEST_APP_DOCKER_NAME --name $CLUSTER_NAME

echo "************** ADDING METRIC SERVER **************"
kubectl apply -f $PWD/tests/scripts/metrics-server.yaml
kubectl patch deployment metrics-server -n kube-system -p '{"spec":{"template":{"spec":{"containers":[{"name":"metrics-server","args":["--cert-dir=/tmp", "--secure-port=4443", "--kubelet-insecure-tls","--kubelet-preferred-address-types=InternalIP"]}]}}}}'
echo "sleeping for 20 seconds to make sure metric server is working"


