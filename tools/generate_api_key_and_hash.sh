#!/bin/bash

API_KEY=$(openssl rand -hex 32)
HASH_HEX=$(echo -n "$API_KEY" | openssl dgst -sha256 | awk '{print $2}')

echo "Key:      $API_KEY"
echo "Hash:     $HASH_HEX"
