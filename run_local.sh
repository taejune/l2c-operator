#!/bin/bash

export OPERATOR_NAME=l2c=operator

operator-sdk run --local --watch-namespace=default
