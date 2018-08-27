#!/bin/sh -ex

WORKTREE=/gitops-repo

mkdir -p $WORKTREE
cd $WORKTREE

git clone $GITOPS_REPO .
git checkout $GITOPS_BRANCH

/mgo serve --listen=$LISTEN --gitops-repo=$WORKTREE $@

