if [[ $TRAVIS_COMMIT_MESSAGE == *"[no release]"* ]]; then
    echo "skipping release stage since [no release] was found on commit message"
else
    echo "releasing new version: $DOCKER_IMAGE"
    (cd .. && make docker-build-and-push)
    git config user.email "travis@travis.org"
    git config user.name "travis" # this email and name can be set anything you like
    ./version-control.sh ${RELEASE_TYPE}
    git add .
    git commit -m "Release new version [skip ci]"
    git push https://shyimo:${GITHUB_API_KEY}@github.com/PayU/kubeobserver.git HEAD:master    
fi