# MyGitops (mgo)

> This is WIP, freshly pushed. Needs some clarifications, updates on documentation, etc..

`mygitops` is tool to help you implement a [gitops](https://www.weave.works/blog/gitops-operations-by-pull-request) based
continuous delivery flow in a kubernetes environment.

## What is this?

```
gitops: git-oriented operations. the state of your system is recorded in a git repository.
        updates to your cluster are reflected in this repo.

gitops repository: the repository containing the full state of your kubernetes cluster/services
```

It's using [helm](https://helm.sh/) as the underlaying management tool for the services in the cluster.

To integrate it in your CI/CD pipeline, it's recommended you run it as a
service listening for incoming _new deployment_ requests.

## Features

- All deployment requests are commited and pushed to your gitops repository
- Synchronizes cluster with gitops repo state
- Http api for easy CI/CD pipeline integration

## How it works

The full cluster state is held in your gitops repo [structured as below](#repo-structure)

When releasing a new version, there are 2 proceses going on under the hood:
- a commit is created in the repo reflecting the pod image repo & tag updates
- the cluster is synced to the new state

## Gitops repo structure

[Here are some details](https://github.com/valer-cara/mgo/blob/master/docs/structure.md) on how to structure your repo.

## CI/CD pipeline integration

`mgo` is meant to be ran as a service. It can handle requests to deploy updates
to your cluster via it's HTTP API.

[Here's an example](https://github.com/valer-cara/mgo/blob/master/docs/examples/gitlab-cicd.md) integration with gitlab, a full `.gitlab-ci.yml` file.


## TODO

- [ ] statefulset upgrades: currently fails when STS are updated. the crude way is `k delete sts --cascade=false xxxx`. maybe something else works better?
- [ ] performance: too many helm upgrades/diffs can end up choking the master node/apiserver. need to limit. maybe rudder is lighter?
- [ ] handle those non-helm manifests (those `*-raw.yaml` files that are raw kubernetes manifests, prob via `kubectl apply -f xxxxx`)
- [ ] Define/Design authentication of clients
- [ ] gopkg.in vanity package urls
- [ ] handle empty commits in kube, mainly when running just re-deploy
- [ ] Check out [Rudder](https://github.com/AcalephStorage/rudder). Might be better than running `exec(helm)`
- [ ] Documentation
  - [ ] update readme, explain updated yaml structure
  - [ ] how to setup local development env
  - [ ] how to import a cluster into a mygitops-style repo
  - [ ] maybe do a demo of an end-to-end integration (CI pipeline -> deployment to cluster)

- [ ] maybe switch to `helm template` & `kubectl apply` instead of `helm upgrade`
  - also check out `helm template` and `kubediff` instead of `helm diff` that only
    diffs against helm’s configmaps rather than the cluster state itself

- [ ] rename mygitops -> mgo -> (maybe smth better?)

## Acknowledgements

The work on mgo is being supported by [Aleth.io](https://aleth.io).


