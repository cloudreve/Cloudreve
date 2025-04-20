#!/bin/bash
set -e
export NODE_OPTIONS="--max-old-space-size=8192"

# This script is used to build the assets for the application.
cd assets
rm -rf build
yarn install --network-timeout 1000000
yarn version --new-version $1 --no-git-tag-version
yarn run build

# Copy the build files to the application directory
cd ../
zip -r - assets/build >assets.zip
mv assets.zip application/statics