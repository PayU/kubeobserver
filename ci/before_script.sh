echo "***************** Running Travis.ci before script *****************"
export VERSION=$(cat $PWD/version)
export DOCKER_IMAGE=zooz/kubeobserver:$VERSION

# set release type
if [[ $TRAVIS_COMMIT_MESSAGE == *"feat"* ]]; then
  export RELEASE_TYPE = "minor"
elif [[ $TRAVIS_COMMIT_MESSAGE == *"future"* ]]; then
    export RELEASE_TYPE = "minor"
else 
     export RELEASE_TYPE = "patch"   
fi

function reportVars() {
    echo "DOCKER_IMAGE: $DOCKER_IMAGE"
    echo "VERSION: $VERSION"
    echo "RELEASE_TYPE: $RELEASE_TYPE"
}

reportVars

