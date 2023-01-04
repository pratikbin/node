#!/bin/bash

# smol wrapper for describe of Messages & Services

# params
echo -e
grpcurl -plaintext localhost:9090 list && echo -e
grpcurl -plaintext localhost:9090 describe identities.transactions.nub.Message && echo -e
grpcurl -plaintext localhost:9090 describe identities.queries.identity.Query && echo -e
grpcurl -plaintext localhost:9090 describe identities.transactions.define.Message && echo -e
grpcurl -plaintext localhost:9090 describe identities.transactions.issue.Message && echo -e
grpcurl -plaintext localhost:9090 describe identities.transactions.provision.Message && echo -e
grpcurl -plaintext localhost:9090 describe identities.transactions.unprovision.Message && echo -e

grpcurl -plaintext localhost:9090 describe classifications.queries.classification.Query && echo -e

grpcurl -plaintext localhost:9090 describe assets.transactions.define.Message && echo -e
grpcurl -plaintext localhost:9090 describe assets.queries.asset.Query && echo -e
grpcurl -plaintext localhost:9090 describe assets.transactions.mint.Message && echo -e
grpcurl -plaintext localhost:9090 describe assets.transactions.deputize.Message && echo -e
grpcurl -plaintext localhost:9090 describe assets.transactions.mutate.Message && echo -e

grpcurl -plaintext localhost:9090 describe maintainers.queries.maintainer.Query && echo -e

grpcurl -plaintext localhost:9090 describe metas.transactions.reveal.Message && echo -e
grpcurl -plaintext localhost:9090 describe metas.queries.meta.Query && echo -e

grpcurl -plaintext localhost:9090 describe splits.transactions.wrap.Message && echo -e
grpcurl -plaintext localhost:9090 describe splits.queries.split.Query && echo -e
grpcurl -plaintext localhost:9090 describe splits.transactions.unwrap.Message && echo -e
grpcurl -plaintext localhost:9090 describe splits.transactions.send.Message && echo -e

grpcurl -plaintext localhost:9090 describe orders.transactions.define.Message && echo -e
grpcurl -plaintext localhost:9090 describe orders.queries.order.Query && echo -e
grpcurl -plaintext localhost:9090 describe orders.transactions.make.Message && echo -e
grpcurl -plaintext localhost:9090 describe orders.transactions.take.Message && echo -e
grpcurl -plaintext localhost:9090 describe orders.transactions.cancel.Message && echo -e
