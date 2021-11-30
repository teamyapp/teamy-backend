#!/bin/bash

GIT_LONG_COMMIT_HASH=$(git rev-parse HEAD)
REPO_OWNER=teamyapp
REPO_NAME=teamy-backend

cat > .repo.env <<EOF
GIT_LONG_COMMIT_HASH=$GIT_LONG_COMMIT_HASH
REPO_OWNER=$REPO_OWNER
REPO_NAME=$REPO_NAME
EOF

cat .repo.env
