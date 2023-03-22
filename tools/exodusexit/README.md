
performexodus --m activateDesert  --privateKey   --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m performAsset  --proof ./tools/exodusexit/proofdata/performDesertNft.json  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m performNft    --proof  ./tools/exodusexit/proofdata/performDesertAsset.json  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m cancelOutstandingDeposit  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m withdrawNFT   --nftIndex 1  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m withdrawAsset    --owner --token --amount --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml




