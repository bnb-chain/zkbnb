INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES( now(), now(), NULL, 'SysGasFee', '{"0":{"1":150000000000000,"10":20000000000000,"11":800000000000000,"4":20000000000000,"5":200000000000000,"6":40000000000000,"7":33000000000000,"8":20000000000000,"9":0},"1":{"1":10000000000000,"10":12000000000000,"11":20000000000000,"4":10000000000000,"5":20000000000000,"6":10000000000000,"7":10000000000000,"8":12000000000000,"9":18000000000000}}', 'string', 'based on BNB');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'ProtocolRate', '200', 'int', 'protocol rate');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'ProtocolAccountIndex', '0', 'int', 'protocol index');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'GasAccountIndex', '1', 'int', 'gas index');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'ZkBNBContract', '0x90fbdf7f6ff15E162D7E822FD35259aB70a95fA7', 'string', 'ZkBNB contract on BSC');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'CommitAddress', '0x83a1f1BaBF815056fa56586f752F116B2A14D26b', 'string', 'ZkBNB commit on BSC');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'VerifyAddress', '0xc785309fee44Fa66848135b58BfDdBb74d75b38D', 'string', 'ZkBNB verify on BSC');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'GovernanceContract', '0x924132EC10170A18656f98b8E09DB5F979e44564', 'string', 'Governance contract on BSC');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'BscTestNetworkRpc', 'https://bsc-testnet.nodereal.io/v1/a1cee760ac744f449416a711f20d99dd', 'string', 'BSC network rpc');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'LocalTestNetworkRpc', 'http://127.0.0.1:8545/', 'string', 'Local network rpc');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'DefaultNftFactory', '0x5b420bA8D0E94f5fC53c0f94a40B513DEC71051e', 'string', 'ZkBNB default nft factory contract on BSC');
INSERT INTO public.sys_config
(created_at, updated_at, deleted_at, "name", value, value_type, "comment")
VALUES(now(), now(), NULL, 'OptionalBlockSizes', '[8,16,32,64]', 'string', 'OptionalBlockSizes config for committer and prover');
