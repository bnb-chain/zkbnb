
performdesert --m activateDesert  --privateKey   --config ./tools/desertexit/etc/config.yaml

performdesert --m perform  --proof ./tools/desertexit/proofdata/performDesert.json  --privateKey  --config ./tools/desertexit/performdesert/etc/config.yaml

performdesert --m cancelOutstandingDeposit  --privateKey  --config ./tools/desertexit/etc/config.yaml

performdesert --m withdrawNFT   --nftIndex  --privateKey  --config ./tools/desertexit/etc/config.yaml
performdesert --m withdrawAsset    --address --token --amount --privateKey  --config ./tools/desertexit/etc/config.yaml



generateproof --m run  --address --token --nftIndex  --proofFolder  --config ./tools/desertexit/etc/config.yaml

generateproof --m continue  --address --token  --nftIndex --proofFolder  --config ./tools/desertexit/etc/config.yaml




performdesert --m activateDesert  --privateKey 7e110218a418b62e4bdd7bda4aff1e5e870b38fa3b4fc5e3462dc3c24e5ec1f4   --config ./tools/desertexit/etc/config.yaml
performdesert --m perform  --proof ./tools/desertexit/proofdata/performDesert.json  --privateKey 7f9294e5f4e6e7015434b89de498b096c761101d448695b28e632b581d6eb887  --config ./tools/desertexit/etc/config.yaml

performdesert --m cancelOutstandingDeposit  --privateKey 7f9294e5f4e6e7015434b89de498b096c761101d448695b28e632b581d6eb887  --config ./tools/desertexit/etc/config.yaml


performdesert --m withdrawAsset  --address 0x9069Bfe50613D85E2125Fc529cba75a781E19622  --token 0x00 --amount 10000000000000000 --privateKey 7f9294e5f4e6e7015434b89de498b096c761101d448695b28e632b581d6eb887 --config ./tools/desertexit/etc/config.yaml
performdesert --m withdrawNFT  --config ./tools/desertexit/etc/config.yaml


performdesert --m getBalance  --address 0x9069Bfe50613D85E2125Fc529cba75a781E19622  --token 0x00 --config ./tools/desertexit/etc/config.yaml
performdesert --m getPendingBalance  --address 0x9069Bfe50613D85E2125Fc529cba75a781E19622  --token 0x00 --config ./tools/desertexit/etc/config.yaml

