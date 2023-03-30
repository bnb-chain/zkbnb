
performexodus --m activateDesert  --privateKey   --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m performAsset  --proof ./tools/exodusexit/proofdata/performDesertAsset.json  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m performNft    --proof  ./tools/exodusexit/proofdata/performDesertNft.json  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m cancelOutstandingDeposit  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m withdrawNFT   --nftIndex 1  --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m withdrawAsset    --owner --token --amount --privateKey  --config ./tools/exodusexit/performexodus/etc/config.yaml



generateproof --m run  --address --token --nftIndexList  --proofFolder  --config ./tools/exodusexit/generateproof/etc/config.yaml

generateproof --m continue  --address --token  --nftIndexList --proofFolder  --config ./tools/exodusexit/generateproof/etc/config.yaml


performexodus --m activateDesert  --privateKey c6182407eedcee00478ac16f6f25046633c91dcc8e664d9964adb26322839049   --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m performAsset  --proof ./tools/exodusexit/proofdata/performDesertAsset.json  --privateKey 3f242374e0d7580e8c52a75790493539e514268ebc2e441e61cb5b86d5077698  --config ./tools/exodusexit/performexodus/etc/config.yaml


performexodus --m activateDesert  --privateKey c6182407eedcee00478ac16f6f25046633c91dcc8e664d9964adb26322839049   --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m performAsset  --proof ./tools/exodusexit/proofdata/performDesertAsset.json  --privateKey c6182407eedcee00478ac16f6f25046633c91dcc8e664d9964adb26322839049  --config ./tools/exodusexit/performexodus/etc/config.yaml
performexodus --m performNft    --proof  ./tools/exodusexit/proofdata/performDesertNft.json  --privateKey c6182407eedcee00478ac16f6f25046633c91dcc8e664d9964adb26322839049  --config ./tools/exodusexit/performexodus/etc/config.yaml

performexodus --m cancelOutstandingDeposit  --privateKey c6182407eedcee00478ac16f6f25046633c91dcc8e664d9964adb26322839049  --config ./tools/exodusexit/performexodus/etc/config.yaml


performexodus --m withdrawAsset  --owner 0x18c6Cdbd596c51cA253Da0248a95377338942a62  --token 0x00 --amount 997000000000000000 --privateKey c6182407eedcee00478ac16f6f25046633c91dcc8e664d9964adb26322839049 --config ./tools/exodusexit/performexodus/etc/config.yaml
