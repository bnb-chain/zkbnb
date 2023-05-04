INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(1, now(), now(), NULL, 'SysGasFee', '{"0":{"1":150000000000000,"10":12000000000000,"11":800000000000000,"4":20000000000000,"5":200000000000000,"6":40000000000000,"7":33000000000000,"8":20000000000000,"9":18000000000000}}', 'string', 'based on BNB');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(2, now(), now(), NULL, 'ProtocolRate', '200', 'int', 'protocol rate');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(3, now(), now(), NULL, 'ProtocolAccountIndex', '0', 'int', 'protocol index');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(4, now(), now(), NULL, 'GasAccountIndex', '1', 'int', 'gas index');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(5, now(), now(), NULL, 'ZkBNBContract', '0xBd012395D9D85499Fc4BF60d7F024d34fD3a88FF', 'string', 'ZkBNB contract on BSC');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(6, now(), now(), NULL, 'CommitAddress', '0x83a1f1BaBF815056fa56586f752F116B2A14D26b', 'string', 'ZkBNB commit on BSC');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(7, now(), now(), NULL, 'VerifyAddress', '0xc785309fee44Fa66848135b58BfDdBb74d75b38D', 'string', 'ZkBNB verify on BSC');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(8, now(), now(), NULL, 'GovernanceContract', '0xB933CD36D937EB2430D4508DbC4470308Bb28813', 'string', 'Governance contract on BSC');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(9, now(), now(), NULL, 'BscTestNetworkRpc', 'https://bsc-testnet.nodereal.io/v1/a1cee760ac744f449416a711f20d99dd', 'string', 'BSC network rpc');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(10, now(), now(), NULL, 'LocalTestNetworkRpc', 'http://127.0.0.1:8545/', 'string', 'Local network rpc');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(11, now(), now(), NULL, 'ZnsPriceOracle', '0x67611D3E0fbB56C016C2B44d428Bb588B1943e9d', 'string', 'Zns Price Oracle');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(12, now(), now(), NULL, 'DefaultNftFactory', '0xDA8c0929ec116C81a85280cAaf73218553848e9D', 'string', 'ZkBNB default nft factory contract on BSC');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(13, now(), now(), NULL, 'Governor', '0x35888CD920AFbE82c7D4cDac9896e720E0aa8cb5', 'string', 'governor');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(14, now(), now(), NULL, 'AssetGovernanceContract', '0x80A3D3eDA8cCD58DC40b76c6004dc8388fa2475A', 'string', 'asset governance contract');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(15, now(), now(), NULL, 'Validators', '{"0x35888CD920AFbE82c7D4cDac9896e720E0aa8cb5":{"Address":"0x35888CD920AFbE82c7D4cDac9896e720E0aa8cb5","IsActive":true},"0x83a1f1BaBF815056fa56586f752F116B2A14D26b":{"Address":"0x83a1f1BaBF815056fa56586f752F116B2A14D26b","IsActive":true},"0xc785309fee44Fa66848135b58BfDdBb74d75b38D":{"Address":"0xc785309fee44Fa66848135b58BfDdBb74d75b38D","IsActive":true}}', 'map[string]*ValidatorInfo', 'validator info');
INSERT INTO public.sys_config
(id, created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(16, now(), now(), NULL, 'OptionalBlockSizes', '[8,16]', 'string', 'OptionalBlockSizes config for committer and prover');
