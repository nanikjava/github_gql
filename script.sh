#!/bin/bash

# Fetch all key-value pairs from etcd
OUTPUT=$(/home/nanik/GolandProjects/etcd/etcdctl/etcdctl --endpoints=http://localhost:7777 get --prefix "")

# Start HTML structure
echo "<html>"
echo "<head><title>ETCD Keys and Values</title></head>"
echo "<body>"
echo "<h1>ETCD Key-Value List</h1>"

# Read keys and values in pairs
KEY=""
while read -r LINE; do
    if [[ -z "$KEY" ]]; then
        # First line is a key
        KEY="$LINE"
    else
        # Second line is the value
        VALUE="$LINE"
        echo "<p><a href=\"$KEY\" target=\"_blank\">$KEY</a>: $VALUE</p>"
        KEY=""  # Reset key for the next pair
    fi
done <<< "$OUTPUT"

# End HTML
echo "</body>"
echo "</html>"
