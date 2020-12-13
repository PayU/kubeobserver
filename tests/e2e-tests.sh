export PORT=8080
export LOG_LEVEL=INFO
export NUMBER_OF_WATCHERS="2"
export DOCKER_NAME=kubeobserver-test
export DOCKER_IMAGE=kubeobserver-local

export LOG_FILE_PATH=$PWD/tests/kubeobserver-test.log

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
    # docker build -t $DOCKER_IMAGE .
    COMMAND="docker run \
                    -d \
                    --name $DOCKER_NAME \
                    --network=host \
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

echo "****************************************************"


echo "******************* RUNNING TESTS *********************"

echo "1) POD WATCHER TEST"
kubectl apply -f $PWD/tests/manifests/hpa-v1.yaml
echo "************* TESTS FINISHED SUCCESSFULLY *************"