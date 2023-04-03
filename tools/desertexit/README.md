
performdesert --m activateDesert  --privateKey   --config ./tools/desertexit/etc/config.yaml

performdesert --m performAsset  --proof ./tools/desertexit/proofdata/performDesertAsset.json  --privateKey  --config ./tools/desertexit/performdesert/etc/config.yaml
performdesert --m performNft    --proof  ./tools/desertexit/proofdata/performDesertNft.json  --privateKey  --config ./tools/desertexit/performdesert/etc/config.yaml

performdesert --m cancelOutstandingDeposit  --privateKey  --config ./tools/desertexit/etc/config.yaml

performdesert --m withdrawNFT   --nftIndexList []  --privateKey  --config ./tools/desertexit/etc/config.yaml
performdesert --m withdrawAsset    --address --token --amount --privateKey  --config ./tools/desertexit/etc/config.yaml



generateproof --m run  --address --token --nftIndexList  --proofFolder  --config ./tools/desertexit/etc/config.yaml

generateproof --m continue  --address --token  --nftIndexList --proofFolder  --config ./tools/desertexit/etc/config.yaml


performdesert --m activateDesert  --privateKey c6182407eedcee00478ac16f6f25046633c91dcc8e664d9964adb26322839049   --config ./tools/desertexit/etc/config.yaml
performdesert --m performAsset  --proof ./tools/desertexit/proofdata/performDesertAsset.json  --privateKey 3f242374e0d7580e8c52a75790493539e514268ebc2e441e61cb5b86d5077698  --config ./tools/desertexit/etc/config.yaml


performdesert --m activateDesert  --privateKey c6182407eedcee00478ac16f6f25046633c91dcc8e664d9964adb26322839049   --config ./tools/desertexit/etc/config.yaml
performdesert --m performAsset  --proof ./tools/desertexit/proofdata/performDesertAsset.json  --privateKey 7f9294e5f4e6e7015434b89de498b096c761101d448695b28e632b581d6eb887  --config ./tools/desertexit/etc/config.yaml
performdesert --m performNft    --proof  ./tools/desertexit/proofdata/performDesertNft.json  --privateKey 7f9294e5f4e6e7015434b89de498b096c761101d448695b28e632b581d6eb887  --config ./tools/desertexit/etc/config.yaml

performdesert --m cancelOutstandingDeposit  --privateKey 7f9294e5f4e6e7015434b89de498b096c761101d448695b28e632b581d6eb887  --config ./tools/desertexit/etc/config.yaml


performdesert --m withdrawAsset  --address 0x9069Bfe50613D85E2125Fc529cba75a781E19622  --token 0x00 --amount 10000000000000000 --privateKey 7f9294e5f4e6e7015434b89de498b096c761101d448695b28e632b581d6eb887 --config ./tools/desertexit/etc/config.yaml
performdesert --m withdrawNFT  --config ./tools/desertexit/etc/config.yaml


performdesert --m getBalance  --address 0x9069Bfe50613D85E2125Fc529cba75a781E19622  --token 0x00 --config ./tools/desertexit/etc/config.yaml
performdesert --m getPendingBalance  --address 0x9069Bfe50613D85E2125Fc529cba75a781E19622  --token 0x00 --config ./tools/desertexit/etc/config.yaml

