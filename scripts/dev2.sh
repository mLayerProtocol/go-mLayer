#!/bin/bash

#--private-key a12b76e5ab05842dd110b9cf709a3ca4e965ec149a1e6d40f09f55a4fc6279d2  \
# --testing \
go run . daemon \
--private-key a93fc8ff64318b2c73257cf98eb8347261e09aefb98a1b23ffb24daca74c2b77   \
--rpc-port 9090 \
--ws-address localhost:9091 \
--rest-address localhost:8080   \
--quic-host 0.0.0.0:9544  \
-l "/ip4/0.0.0.0/udp/5002/quic-v1" \
-l "/ip4/0.0.0.0/udp/5002/quic-v1/webtransport" \
-l "/ip4/127.0.0.1/tcp/7001/ws" \
-l "/ip4/127.0.0.1/tcp/6001" \
--verbose true \
--data-dir "./data/dev2/" \
--bootstrap-node true \
 --no-sync

 