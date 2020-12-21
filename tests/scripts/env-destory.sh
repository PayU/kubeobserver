CLUSTER_NAME=kind-kubeobserver-test

echo "************** DELETING KING CLUSTER **************"
kind delete cluster --name $CLUSTER_NAME
echo "***************************************************"