#!/bin/bash

testcases="/testdata/frr/test/frr.conf.*"
for tc in $testcases; do
    echo -n  "Testing ${FRR_VERSION} on ${OS_NAME}:${OS_VERSION} with input ${tc}: "
    if vtysh --dryrun --inputfile "${tc}";
    then
        echo "✅"
    else
        echo "❌"
        echo "FRR ${FRR_VERSION} on ${OS_NAME}:${OS_VERSION} produces an invalid configuration"
        exit 1
    fi
done

testcases="/testdata/nftables/test/nftrules*"
for tc in $testcases; do
    echo -n  "Testing nft rules on ${OS_NAME}:${OS_VERSION} with input ${tc}: "
    if nft -c -f "${tc}";
    then
        echo "✅"
    else
        echo "❌"
        echo "nft input ${tc} on ${OS_NAME}:${OS_VERSION} produces an invalid configuration"
        exit 1
    fi
done
