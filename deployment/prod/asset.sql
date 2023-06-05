INSERT INTO public.asset
(created_at, updated_at, deleted_at, asset_id, asset_name, asset_symbol, l1_address, decimals, status, is_gas_asset)
VALUES(now(), now(), NULL, 0, 'BNB', 'BNB', '0x00', 18, 0, 1);
INSERT INTO public.asset
(created_at, updated_at, deleted_at, asset_id, asset_name, asset_symbol, l1_address, decimals, status, is_gas_asset)
VALUES(now(), now(),  NULL, 1, 'BUSD', 'BUSD', '', 18, 0, 1);