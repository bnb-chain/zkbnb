
performexodus --m activateDesert  --privateKey   --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m performAsset  --proof ./tools/exodusexit/proofdata/performDesertAsset.json  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m performNft    --proof  ./tools/exodusexit/proofdata/performDesertNft.json  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m cancelOutstandingDeposit  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m withdrawNFT   --nftIndex 1  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m withdrawAsset    --owner --token --amount --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml



generateproof --m run  --address --token --amount --nftIndexList --proofFolder  --config ./tools/exodusexit/generateproof/etc/config.yaml

generateproof --m continue  --address --token --amount --nftIndexList --proofFolder  --config ./tools/exodusexit/generateproof/etc/config.yaml


performexodus --m activateDesert  --privateKey 444537256fd840bfcd7074781db2764da1e74bf95a7fbb3fdcab364307b5821e   --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m performAsset  --proof ./tools/exodusexit/proofdata/performDesertAsset.json  --privateKey 444537256fd840bfcd7074781db2764da1e74bf95a7fbb3fdcab364307b5821e  --config ./tools/exodusexit/performexodus/etc/config.yaml


