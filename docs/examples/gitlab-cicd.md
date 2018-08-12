# Gitlab + mgo integration

There are two deploy stages: `deployToStaging` and `deployToProduction`.
These consist only in calling the `mygitops` service.

```yaml
variables:
  MYGITOPS_URL: "http://mygitops-service-url:8080/deploy"

stages:
    - build
    - test
    - staging
    - production

build:
  stage: build
  script:
    - ...

test:
  stage: test
  script:
    - ...

deployToStaging:
  stage: staging
  image:
    name: byrnedo/alpine-curl
    entrypoint: [""]

  variables:
    # this is the cluster as specified in kubeconfig
    CLUSTER: "my-kubernetes-staging-cluster"

  script:
    - curl --fail ${MYGITOPS_URL}
           --data "triggerRepo=${CI_PROJECT_URL##https://}"
           --data "imageRepo=${IMAGE_REPO}"
           --data "imageTag=${IMAGE_TAG_PROD}"
           --data "cluster=${CLUSTER}"
           --data "author=${GITLAB_USER_EMAIL}"

  only:
    - master

deployToProduction:
  stage: production
  image:
    name: byrnedo/alpine-curl
    entrypoint: [""]

  variables:
    CLUSTER: "my-kubernetes-production-cluster"

  script:
    - curl --fail ${MYGITOPS_URL}
           --data "triggerRepo=${CI_PROJECT_URL##https://}"
           --data "imageRepo=${IMAGE_REPO}"
           --data "imageTag=${IMAGE_TAG_PROD}"
           --data "cluster=${CLUSTER}"
           --data "author=${GITLAB_USER_EMAIL}"

  when: manual

  only:
    - master
```

