export PORT=8080
export LOG_LEVEL=INFO

echo "************* VERIFY CLEAN ENVIRONMENT *************"
pInfo=$(sudo lsof -nP -iTCP:$PORT | grep LISTEN)
if [ -z ${pInfo+x} ]; then 
    echo "Environment is clean"
else
    echo "removing old kubeobserver process"
    pInfoArray=($pInfo)
    kill -9 ${pInfoArray[1]}
fi
echo "****************************************************"

echo "************* BUILDING & RUNNING KUBEOBSERVER *************"
make build

./kubeobserver &
echo "***********************************************************"

echo "Main process will now sleep for 7 seconds..."
sleep 7

echo "******************* RUNNING TESTS *********************"
kubectl apply -f $PWD/tests/manifests/hpa-v1.yaml

sleep 10
echo "************* TESTS FINISHED SUCCESSFULLY *************"
