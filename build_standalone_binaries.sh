#!/bin/bash
set -e
cd frontend
[[ -s $HOME/.nvm/nvm.sh ]] && . $HOME/.nvm/nvm.sh  # This loads NVM
nvm install
nvm use
npm ci
npm run build
rm -rf ../backend/frontend_blobs
cp -r build ../backend/frontend_blobs
cd ../backend
rm frontend_blobs/index.html
rm -rf ../binaries
mkdir ../binaries
GOOS=linux GOARCH=amd64 go build -o ../binaries/backend-linux-amd64
GOOS=linux GOARCH=arm64 go build -o ../binaries/backend-linux-arm64
