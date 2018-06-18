#!/bin/bash

if [ ! -d "$HOME/google-cloud-sdk/bin" ]; then rm -rf $HOME/google-cloud-sdk; export CLOUDSDK_CORE_DISABLE_PROMPTS=1; curl https://sdk.cloud.google.com | bash; fi
source /home/travis/google-cloud-sdk/path.bash.inc
gcloud --quiet version
gcloud --quiet components update
echo $KUBERNETES_SERVICE_ACCOUNT > /tmp/kubernetes-service-account-key.json
gcloud auth activate-service-account --key-file /tmp/kubernetes-service-account-key.json
gcloud config set project $GCP_PROJECT
gcloud container clusters get-credentials $KUBERNETES_CLUSTER_NAME --zone $KUBERNETES_CLUSTER_ZONE --project $GCP_PROJECT
kubectl set image deployment/teddycare teddycare=eu.gcr.io/teddy-care/teddycare:$TRAVIS_COMMIT
kubectl set image deployment/event-manager event-manager=eu.gcr.io/teddy-care/event-manager:$TRAVIS_COMMIT