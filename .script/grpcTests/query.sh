#!/bin/bash

# tests

echo -e "\nendpoint: cosmos.bank.v1beta1.Query/AllBalances\n"
grpcurl -plaintext -d '{"address":"mantle1pkkayn066msg6kn33wnl5srhdt3tnu2vuet86j"}' localhost:9090 cosmos.bank.v1beta1.Query/AllBalances

echo -e "\nendpoint: cosmos.bank.v1beta1.Query/Balance\n"
grpcurl -plaintext -d '{"address":"mantle1pkkayn066msg6kn33wnl5srhdt3tnu2vuet86j","denom":"stake"}' localhost:9090 cosmos.bank.v1beta1.Query/Balance
