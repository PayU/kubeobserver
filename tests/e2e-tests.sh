export PORT=8080
export LOG_LEVEL=debug
export NUMBER_OF_WATCHERS="2"
export DOCKER_NAME=kubeobserver-test
export DOCKER_IMAGE=kubeobserver-local
export DEFAULT_RECEIVER=log

# clean all test resources from the kind cluster
function deleteOldK8sResource() {
    echo "waiting for old k8s resources to be removed form the cluster"
    deleteRes=$(kubectl delete deployment hello-world 2>&1 | grep "(NotFound)")
    if [ -z "$deleteRes" ]; then
        until kubectl get deployment hello-world 2>&1 | grep "(NotFound)" ; do sleep 3; done
    fi

    deleteRes=$(kubectl delete hpa hello-world-v2-beta1 2>&1 | grep "(NotFound)")
    if [ -z "$deleteRes" ]; then
        until kubectl delete hpa hello-world-v2-beta1 2>&1 | grep "(NotFound)" ; do sleep 3; done
    fi
}

function waitForApp() {
    watch_ready=0
    HEALTH_CHECK_TIMEOUT=40;
    HEALTH_CHECK_INTERVAL=1;

    while [[ $watch_ready -lt $NUMBER_OF_WATCHERS && $HEALTH_CHECK_TIMEOUT -gt 0 ]]; do
        echo "waiting for $DOCKER_NAME to be ready.."
        watch_ready=$(docker logs $DOCKER_NAME 2>&1 | grep "controller is ready and starting" | wc -l)
        let HEALTH_CHECK_TIMEOUT=$HEALTH_CHECK_TIMEOUT-1
        sleep $HEALTH_CHECK_INTERVAL
    done

    if [[ $watch_ready -lt $NUMBER_OF_WATCHERS ]];then
        echo "Couldn't start the application on time"
        exit 1
    fi
}

function deleteContainer() {
    NAME=$1
    isExists=$(docker ps -af name=$NAME | grep -v IMAGE)
    if [ ! -z isExists ];then
        docker rm -f $NAME
    fi
}

echo "************* VERIFY CLEAN ENVIRONMENT *************"
deleteContainer $DOCKER_NAME
deleteOldK8sResource
# delete, build and re-run the kubeobserver docker 
docker build -t $DOCKER_IMAGE .
COMMAND="docker run \
                -d \
                --name $DOCKER_NAME \
                --network=host \
                -e DEFAULT_RECEIVER=$DEFAULT_RECEIVER \
                -e PORT=$PORT \
                -e LOG_LEVEL=$LOG_LEVEL \
                -v ${HOME}/.kube/:/root/.kube \
                -e K8S_CLUSTER_NAME=local-cluster \
                $DOCKER_IMAGE"
echo -e "Starting $DOCKER_NAME\n"${COMMAND/\s+/ }
$COMMAND
COMMAND_EXIT_CODE=$?
if [ ${COMMAND_EXIT_CODE} != 0 ]; then
    printf "Error when executing: '${DOCKER_NAME}'\n"
    exit ${COMMAND_EXIT_CODE}
fi
waitForApp
echo "$DOCKER_NAME is ready"

echo "sleeping for 3 seconds for kind cluster to be valid"
sleep 3
docker restart $DOCKER_NAME

echo "******************* RUNNING TESTS *********************"

echo "1) pod watcher test"
kubectl apply -f $PWD/tests/manifests/deployment.yaml
sleep 5
res=$(docker logs $DOCKER_NAME 2>&1 | grep "A \`pod\` in namesapce \`default\` has been \`Created\`")

if [ -z "$res" ]; then
    # we can miss the create event. so we will verify it happend
    podDetails=$(kubectl get pods | grep hello-world)
    podDetailsArray=($podDetails)
    podName=${podDetailsArray[0]}
    res=$(docker logs $DOCKER_NAME 2>&1 | grep "found 1 event receivers for pod default/$podName in namespace default. receivers:log. event-type: Add.")
    if [ -z "$res" ]; then
        echo "error: kubeobserver did not process pod create event. exit tests with exit code 1.."
        exit 1
    fi
fi

echo "2) hpa-v2beta1 test"
kubectl apply -f $PWD/tests/manifests/hpa-v2beta1.yaml
sleep 5
res=$(docker logs $DOCKER_NAME 2>&1 | grep "og recevier event message[New HorizontalPodAutoscaler resource [\`default/hello-world-v2-beta1\`]")
if [ -z "$res" ]; then
    echo "error: kubeobserver did not process hpa create event. exit tests with exit code 1.."
    exit 1
fi

echo "************* TESTS FINISHED SUCCESSFULLY *************"