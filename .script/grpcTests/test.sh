#!/bin/bash

NODE=""
GRPCURL=""

# params
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 describe identities.queries.identity.Query
grpcurl -plaintext localhost:9090 describe classifications.queries.classification.Query
grpcurl -plaintext localhost:9090 describe assets.queries.asset.Query
grpcurl -plaintext localhost:9090 describe maintainers.queries.maintainer.Query
grpcurl -plaintext localhost:9090 describe metas.queries.meta.Query
grpcurl -plaintext localhost:9090 describe splits.queries.split.Query
grpcurl -plaintext localhost:9090 describe splits.queries.ownable.Query
grpcurl -plaintext localhost:9090 describe orders.queries.order.Query

# tests
echo -e "\nendpoint: cosmos.bank.v1beta1.Query/AllBalances\n"
grpcurl -plaintext -d '{"address":"mantle1pkkayn066msg6kn33wnl5srhdt3tnu2vuet86j"}' localhost:9090 cosmos.bank.v1beta1.Query/AllBalances

echo -e "\nendpoint: cosmos.bank.v1beta1.Query/Balance\n"
grpcurl -plaintext -d '{"address":"mantle1pkkayn066msg6kn33wnl5srhdt3tnu2vuet86j","denom":"stake"}' localhost:9090 cosmos.bank.v1beta1.Query/Balance
