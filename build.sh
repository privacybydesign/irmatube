#!/usr/bin/env bash

set -euxo pipefail

mkdir -p build
rm -rf build/*
mkdir build/www build/data

cp --parents -r data/* build
rm build/data/README.md
cp --parents -r www/content build
cp --parents -r www/css build
cp --parents -r www/img build
cp --parents -r www/js build
cp --parents -r www/php build
cp --parents -r www/vendor build
cp --parents www/config.* build
cp --parents www/explanation-* build
cp --parents www/index.php build
cp --parents www/node_modules/@privacybydesign/irmajs/dist/irma.js build
cp -r --parents www/node_modules/bootstrap/dist build
cp --parents www/node_modules/jquery/dist/jquery.min.js build
cp --parents www/node_modules/mustache/mustache.min.js build
