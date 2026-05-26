#!/bin/sh
set -e

unformatted=$(goimports -local github.com/dev-au/CodeStream -l .)

if [ -n "$unformatted" ]; then
    echo "These files need formatting:"
    echo "$unformatted"
    exit 1
fi

echo "All files properly formatted."
