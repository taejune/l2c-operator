#!/bin/bash

cd "$(dirname "$0")"/.. || exit 1
source scripts/common.sh

operator-sdk run --local --watch-namespace=""
