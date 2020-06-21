echo "***************** Running Travis.ci before script *****************"
export VERSION=$(cat $PWD/version)
export DOCKER_IMAGE=zooz/kubeobserver:$VERSION

reportVars

function reportVars() {
    echo "DOCKER_IMAGE: $DOCKER_IMAGE"
    echo "VERSION: $VERSION"
}

