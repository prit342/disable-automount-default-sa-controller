#!/bin/bash

set -euo pipefail

# Download and setup binaries required by envtest https://book.kubebuilder.io/reference/envtest.html
curl -sSLo envtest-bins.tar.gz "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-1.25.0-linux-amd64.tar.gz"
rm -rf ~/envtest-binaries
mkdir -p ~/envtest-binaries
tar -zvxf envtest-bins.tar.gz
mv kubebuilder ~/envtest-binaries
ls -ltraR ~/envtest-binaries/kubebuilder/bin
rm -rf envtest-bins.tar.gz
export KUBEBUILDER_ASSETS=~/envtest-binaries/kubebuilder/bin
echo $KUBEBUILDER_ASSETS
