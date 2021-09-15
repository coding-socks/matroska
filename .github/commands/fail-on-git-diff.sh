#!/usr/bin/env bash

status=$(git status --porcelain | tee /dev/stderr)
if [[ $status ]]; then
    echo "Diff detected" 1>&2
    exit 1
else
    echo "OK"
fi
