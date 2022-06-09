/*
 Navicat Premium Data Transfer

 Source Server         : local_docker
 Source Server Type    : PostgreSQL
 Source Server Version : 140003
 Source Host           : localhost:5432
 Source Catalog        : zecreyLegend
 Source Schema         : public

 Target Server Type    : PostgreSQL
 Target Server Version : 140003
 File Encoding         : 65001

 Date: 09/06/2022 15:35:59
*/


-- ----------------------------
-- Sequence structure for account_history_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."account_history_id_seq";
CREATE SEQUENCE "public"."account_history_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for account_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."account_id_seq";
CREATE SEQUENCE "public"."account_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for block_for_commit_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."block_for_commit_id_seq";
CREATE SEQUENCE "public"."block_for_commit_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for block_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."block_id_seq";
CREATE SEQUENCE "public"."block_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for fail_tx_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."fail_tx_id_seq";
CREATE SEQUENCE "public"."fail_tx_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l1_amount_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l1_amount_id_seq";
CREATE SEQUENCE "public"."l1_amount_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l1_block_monitor_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l1_block_monitor_id_seq";
CREATE SEQUENCE "public"."l1_block_monitor_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l1_tx_sender_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l1_tx_sender_id_seq";
CREATE SEQUENCE "public"."l1_tx_sender_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_asset_info_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_asset_info_id_seq";
CREATE SEQUENCE "public"."l2_asset_info_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_block_event_monitor_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_block_event_monitor_id_seq";
CREATE SEQUENCE "public"."l2_block_event_monitor_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_nft_collection_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_nft_collection_id_seq";
CREATE SEQUENCE "public"."l2_nft_collection_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_nft_exchange_history_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_nft_exchange_history_id_seq";
CREATE SEQUENCE "public"."l2_nft_exchange_history_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_nft_exchange_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_nft_exchange_id_seq";
CREATE SEQUENCE "public"."l2_nft_exchange_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_nft_history_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_nft_history_id_seq";
CREATE SEQUENCE "public"."l2_nft_history_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_nft_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_nft_id_seq";
CREATE SEQUENCE "public"."l2_nft_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_nft_withdraw_history_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_nft_withdraw_history_id_seq";
CREATE SEQUENCE "public"."l2_nft_withdraw_history_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for l2_tx_event_monitor_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."l2_tx_event_monitor_id_seq";
CREATE SEQUENCE "public"."l2_tx_event_monitor_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for liquidity_history_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."liquidity_history_id_seq";
CREATE SEQUENCE "public"."liquidity_history_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for liquidity_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."liquidity_id_seq";
CREATE SEQUENCE "public"."liquidity_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for mempool_tx_detail_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."mempool_tx_detail_id_seq";
CREATE SEQUENCE "public"."mempool_tx_detail_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for mempool_tx_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."mempool_tx_id_seq";
CREATE SEQUENCE "public"."mempool_tx_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for offer_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."offer_id_seq";
CREATE SEQUENCE "public"."offer_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for proof_sender_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."proof_sender_id_seq";
CREATE SEQUENCE "public"."proof_sender_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for sys_config_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."sys_config_id_seq";
CREATE SEQUENCE "public"."sys_config_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for tx_detail_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."tx_detail_id_seq";
CREATE SEQUENCE "public"."tx_detail_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for tx_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."tx_id_seq";
CREATE SEQUENCE "public"."tx_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Table structure for account
-- ----------------------------
DROP TABLE IF EXISTS "public"."account";
CREATE TABLE "public"."account" (
  "id" int8 NOT NULL DEFAULT nextval('account_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "account_index" int8,
  "account_name" text COLLATE "pg_catalog"."default",
  "public_key" text COLLATE "pg_catalog"."default",
  "account_name_hash" text COLLATE "pg_catalog"."default",
  "l1_address" text COLLATE "pg_catalog"."default",
  "nonce" int8,
  "collection_nonce" int8,
  "asset_info" text COLLATE "pg_catalog"."default",
  "asset_root" text COLLATE "pg_catalog"."default",
  "status" int8
)
;

-- ----------------------------
-- Records of account
-- ----------------------------
INSERT INTO "public"."account" VALUES (1, '2022-06-08 08:25:58.048507+00', '2022-06-08 08:25:58.048507+00', NULL, 0, 'treasury.legend', 'fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998', '167c5363088a40a4839912a872f43164270740c7e986ec55397b2d583317ab4a', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 0);
INSERT INTO "public"."account" VALUES (2, '2022-06-08 08:25:58.048507+00', '2022-06-08 08:25:58.048507+00', NULL, 1, 'gas.legend', '1ec94e497abe0fbb87f9ed2843e21163e17e3e97f6bbbae7a88399b826474f93', '0a48e9892a45a04d0c5b0f235a3aeb07b92137ba71a59b9c457774bafde95983', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 0);
INSERT INTO "public"."account" VALUES (3, '2022-06-08 08:25:58.048507+00', '2022-06-08 08:25:58.048507+00', NULL, 2, 'sher.legend', 'b0b6f7466154578ec66d51a335ead65ffd6a7210567fad9e68b6df8a5ce5dd85', '214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 0);
INSERT INTO "public"."account" VALUES (4, '2022-06-08 08:25:58.048507+00', '2022-06-08 08:25:58.048507+00', NULL, 3, 'gavin.legend', '0500ccea3ca064968f5292b850ac8d4d3ee48d499357351a5ebfa2f30bb6070e', '1c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 0);

-- ----------------------------
-- Table structure for account_history
-- ----------------------------
DROP TABLE IF EXISTS "public"."account_history";
CREATE TABLE "public"."account_history" (
  "id" int8 NOT NULL DEFAULT nextval('account_history_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "account_index" int8,
  "nonce" int8,
  "collection_nonce" int8,
  "asset_info" text COLLATE "pg_catalog"."default",
  "asset_root" text COLLATE "pg_catalog"."default",
  "l2_block_height" int8
)
;

-- ----------------------------
-- Records of account_history
-- ----------------------------

-- ----------------------------
-- Table structure for block
-- ----------------------------
DROP TABLE IF EXISTS "public"."block";
CREATE TABLE "public"."block" (
  "id" int8 NOT NULL DEFAULT nextval('block_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "block_commitment" text COLLATE "pg_catalog"."default",
  "block_height" int8,
  "state_root" text COLLATE "pg_catalog"."default",
  "priority_operations" int8,
  "pending_on_chain_operations_hash" text COLLATE "pg_catalog"."default",
  "pending_on_chain_operations_pub_data" text COLLATE "pg_catalog"."default",
  "committed_tx_hash" text COLLATE "pg_catalog"."default",
  "committed_at" int8,
  "verified_tx_hash" text COLLATE "pg_catalog"."default",
  "verified_at" int8,
  "block_status" int8
)
;

-- ----------------------------
-- Records of block
-- ----------------------------
INSERT INTO "public"."block" VALUES (1, '2022-06-08 08:24:27.340951+00', '2022-06-08 08:24:27.340951+00', NULL, '0000000000000000000000000000000000000000000000000000000000000000', 0, '14e4e8ad4848558d7200530337052e1ad30f5385b3c7187c80ad85f48547b74f', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 3);

-- ----------------------------
-- Table structure for block_for_commit
-- ----------------------------
DROP TABLE IF EXISTS "public"."block_for_commit";
CREATE TABLE "public"."block_for_commit" (
  "id" int8 NOT NULL DEFAULT nextval('block_for_commit_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "block_height" int8,
  "state_root" text COLLATE "pg_catalog"."default",
  "public_data" text COLLATE "pg_catalog"."default",
  "timestamp" int8,
  "public_data_offsets" text COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Records of block_for_commit
-- ----------------------------

-- ----------------------------
-- Table structure for fail_tx
-- ----------------------------
DROP TABLE IF EXISTS "public"."fail_tx";
CREATE TABLE "public"."fail_tx" (
  "id" int8 NOT NULL DEFAULT nextval('fail_tx_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "tx_hash" text COLLATE "pg_catalog"."default",
  "tx_type" int8,
  "gas_fee" text COLLATE "pg_catalog"."default",
  "gas_fee_asset_id" int8,
  "tx_status" int8,
  "asset_a_id" int8,
  "asset_b_id" int8,
  "tx_amount" text COLLATE "pg_catalog"."default",
  "native_address" text COLLATE "pg_catalog"."default",
  "tx_info" text COLLATE "pg_catalog"."default",
  "extra_info" text COLLATE "pg_catalog"."default",
  "memo" text COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Records of fail_tx
-- ----------------------------

-- ----------------------------
-- Table structure for l1_amount
-- ----------------------------
DROP TABLE IF EXISTS "public"."l1_amount";
CREATE TABLE "public"."l1_amount" (
  "id" int8 NOT NULL DEFAULT nextval('l1_amount_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "asset_id" int8,
  "block_height" int8,
  "total_amount" int8
)
;

-- ----------------------------
-- Records of l1_amount
-- ----------------------------

-- ----------------------------
-- Table structure for l1_block_monitor
-- ----------------------------
DROP TABLE IF EXISTS "public"."l1_block_monitor";
CREATE TABLE "public"."l1_block_monitor" (
  "id" int8 NOT NULL DEFAULT nextval('l1_block_monitor_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "l1_block_height" int8,
  "block_info" text COLLATE "pg_catalog"."default",
  "monitor_type" int8
)
;

-- ----------------------------
-- Records of l1_block_monitor
-- ----------------------------
INSERT INTO "public"."l1_block_monitor" VALUES (1, '2022-06-08 08:24:45.516427+00', '2022-06-08 08:24:45.516427+00', NULL, 320000, 'null', 1);
INSERT INTO "public"."l1_block_monitor" VALUES (2, '2022-06-08 08:24:56.731448+00', '2022-06-08 08:24:56.731448+00', NULL, 400645, '[{"EventType":6,"TxHash":"0xb9b9c017a978399c97013d23b40d1423c30f009c4d22c97deb824ea30eb4f90b"},{"EventType":7,"TxHash":"0xb9b9c017a978399c97013d23b40d1423c30f009c4d22c97deb824ea30eb4f90b"},{"EventType":5,"TxHash":"0xb9b9c017a978399c97013d23b40d1423c30f009c4d22c97deb824ea30eb4f90b"},{"EventType":4,"TxHash":"0xbc8ad0a59bec5bc874fee604ec6fc94f8067bac6e9e3f0a88b6910e090f91508"},{"EventType":4,"TxHash":"0x115fe0fd9d7c293528507f061915c3f667f796549bdfa8cf3d08d0d00c43dee8"}]', 1);
INSERT INTO "public"."l1_block_monitor" VALUES (3, '2022-06-08 08:25:22.693508+00', '2022-06-08 08:25:22.693508+00', NULL, 320000, 'null', 0);
INSERT INTO "public"."l1_block_monitor" VALUES (4, '2022-06-08 08:25:32.568327+00', '2022-06-08 08:25:32.568327+00', NULL, 400658, '[{"EventType":0,"TxHash":"0x168dace8281eee3d90f81db58754161166af5c6c25b9db744fce8551a0c87af6"},{"EventType":0,"TxHash":"0xc6e7f209ac872a25879b76f125b842f99556cff8451e7ec4af1300ef6502b4f3"},{"EventType":0,"TxHash":"0x2ababebda27a5441798f88b10b9b75ca29095b35c3965ee07aa8098067314e5d"},{"EventType":0,"TxHash":"0x1a1f35d1e50f4b9aa9abd110679e6aba0586f620a4c56a52308cd847b07490ed"},{"EventType":0,"TxHash":"0x4d5c3b3d15b48a17cd3920535581119be48a3ec8fb44eb621b1d587e14ca5a77"},{"EventType":0,"TxHash":"0x9ddacae8cde4948d91826f1badbce9136c50d2be70363b7f2d38b0bfc0d5ff5c"},{"EventType":0,"TxHash":"0x418cdeec66b95e6d9d0a9460d3c44479c15951101143a86035f2cd800f5fab98"},{"EventType":0,"TxHash":"0x8b6f1bf70751be929f0f77d8b1ea204b045881f01494ae3a7d7d73657812d38b"},{"EventType":0,"TxHash":"0xa5c17af78376a85c9f213ea488adefbee9966d263e9047f8c31dea5408d227e7"},{"EventType":0,"TxHash":"0xe5d077d6f88dafa65e9e0058dab2e852e8302dc7032d36e005dad37418cbe439"},{"EventType":0,"TxHash":"0xbad7534ff00fac01c27cf86f3c1e0039ed8d9bcecd918a0ac1e9c1a56652bc6d"},{"EventType":0,"TxHash":"0x463dcd43d99faeea3d3db3001ee9b464696ec9200d2955b84ffba08ff92df38a"},{"EventType":0,"TxHash":"0x390fb0013d788c35a9418f348e1b44c565d6f73b587e6c5d77f47ff858d8322a"},{"EventType":0,"TxHash":"0x9cc1335db2ad0cdd2e006c64da7e2b63adacab522a15050949b5cab827f9f2a9"},{"EventType":0,"TxHash":"0xb0e1317c18b03378e3699829339135b7afbf090a6aaf985cd825abf109c08066"}]', 0);

-- ----------------------------
-- Table structure for l1_tx_sender
-- ----------------------------
DROP TABLE IF EXISTS "public"."l1_tx_sender";
CREATE TABLE "public"."l1_tx_sender" (
  "id" int8 NOT NULL DEFAULT nextval('l1_tx_sender_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "l1_tx_hash" text COLLATE "pg_catalog"."default",
  "tx_status" int8,
  "tx_type" int2,
  "l2_block_height" int8
)
;

-- ----------------------------
-- Records of l1_tx_sender
-- ----------------------------

-- ----------------------------
-- Table structure for l2_asset_info
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_asset_info";
CREATE TABLE "public"."l2_asset_info" (
  "id" int8 NOT NULL DEFAULT nextval('l2_asset_info_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "asset_id" int8,
  "asset_address" text COLLATE "pg_catalog"."default",
  "asset_name" text COLLATE "pg_catalog"."default",
  "asset_symbol" text COLLATE "pg_catalog"."default",
  "decimals" int8,
  "status" int8
)
;

-- ----------------------------
-- Records of l2_asset_info
-- ----------------------------
INSERT INTO "public"."l2_asset_info" VALUES (1, '2022-06-08 08:24:27.334984+00', '2022-06-08 08:24:27.334984+00', NULL, 0, '0x00', 'BNB', 'BNB', 18, 0);
INSERT INTO "public"."l2_asset_info" VALUES (2, '2022-06-08 08:24:56.735319+00', '2022-06-08 08:24:56.735319+00', NULL, 1, '0x3E72bC3842c47d5B63B634F0c7f2E5a56Ad94124', 'LEG', 'LEG', 18, 0);
INSERT INTO "public"."l2_asset_info" VALUES (3, '2022-06-08 08:24:56.735319+00', '2022-06-08 08:24:56.735319+00', NULL, 2, '0x6403c9a361Df1276c1568EAB2141aceD24F53eF6', 'REY', 'REY', 18, 0);

-- ----------------------------
-- Table structure for l2_block_event_monitor
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_block_event_monitor";
CREATE TABLE "public"."l2_block_event_monitor" (
  "id" int8 NOT NULL DEFAULT nextval('l2_block_event_monitor_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "block_event_type" int2,
  "l1_block_height" int8,
  "l1_tx_hash" text COLLATE "pg_catalog"."default",
  "l2_block_height" int8,
  "status" int8
)
;

-- ----------------------------
-- Records of l2_block_event_monitor
-- ----------------------------

-- ----------------------------
-- Table structure for l2_nft
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_nft";
CREATE TABLE "public"."l2_nft" (
  "id" int8 NOT NULL DEFAULT nextval('l2_nft_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "nft_index" int8,
  "creator_account_index" int8,
  "owner_account_index" int8,
  "nft_content_hash" text COLLATE "pg_catalog"."default",
  "nft_l1_address" text COLLATE "pg_catalog"."default",
  "nft_l1_token_id" text COLLATE "pg_catalog"."default",
  "creator_treasury_rate" int8,
  "collection_id" int8
)
;

-- ----------------------------
-- Records of l2_nft
-- ----------------------------
INSERT INTO "public"."l2_nft" VALUES (1, '2022-06-08 08:25:58.056967+00', '2022-06-08 08:25:58.056967+00', NULL, 0, 0, 2, 'abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b', '0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1', '0', 0, 0);

-- ----------------------------
-- Table structure for l2_nft_collection
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_nft_collection";
CREATE TABLE "public"."l2_nft_collection" (
  "id" int8 NOT NULL DEFAULT nextval('l2_nft_collection_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "account_index" int8,
  "name" text COLLATE "pg_catalog"."default",
  "introduction" text COLLATE "pg_catalog"."default",
  "status" int8
)
;

-- ----------------------------
-- Records of l2_nft_collection
-- ----------------------------

-- ----------------------------
-- Table structure for l2_nft_exchange
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_nft_exchange";
CREATE TABLE "public"."l2_nft_exchange" (
  "id" int8 NOT NULL DEFAULT nextval('l2_nft_exchange_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "buyer_account_index" int8,
  "owner_account_index" int8,
  "nft_index" int8,
  "asset_id" int8,
  "asset_amount" text COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Records of l2_nft_exchange
-- ----------------------------

-- ----------------------------
-- Table structure for l2_nft_exchange_history
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_nft_exchange_history";
CREATE TABLE "public"."l2_nft_exchange_history" (
  "id" int8 NOT NULL DEFAULT nextval('l2_nft_exchange_history_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "buyer_account_index" int8,
  "owner_account_index" int8,
  "nft_index" int8,
  "asset_id" int8,
  "asset_amount" int8,
  "l2_block_height" int8
)
;

-- ----------------------------
-- Records of l2_nft_exchange_history
-- ----------------------------

-- ----------------------------
-- Table structure for l2_nft_history
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_nft_history";
CREATE TABLE "public"."l2_nft_history" (
  "id" int8 NOT NULL DEFAULT nextval('l2_nft_history_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "nft_index" int8,
  "creator_account_index" int8,
  "owner_account_index" int8,
  "nft_content_hash" text COLLATE "pg_catalog"."default",
  "nft_l1_address" text COLLATE "pg_catalog"."default",
  "nft_l1_token_id" text COLLATE "pg_catalog"."default",
  "creator_treasury_rate" int8,
  "collection_id" int8,
  "status" int8,
  "l2_block_height" int8
)
;

-- ----------------------------
-- Records of l2_nft_history
-- ----------------------------

-- ----------------------------
-- Table structure for l2_nft_withdraw_history
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_nft_withdraw_history";
CREATE TABLE "public"."l2_nft_withdraw_history" (
  "id" int8 NOT NULL DEFAULT nextval('l2_nft_withdraw_history_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "nft_index" int8,
  "creator_account_index" int8,
  "owner_account_index" int8,
  "nft_content_hash" text COLLATE "pg_catalog"."default",
  "nft_l1_address" text COLLATE "pg_catalog"."default",
  "nft_l1_token_id" text COLLATE "pg_catalog"."default",
  "creator_treasury_rate" int8,
  "collection_id" int8
)
;

-- ----------------------------
-- Records of l2_nft_withdraw_history
-- ----------------------------

-- ----------------------------
-- Table structure for l2_tx_event_monitor
-- ----------------------------
DROP TABLE IF EXISTS "public"."l2_tx_event_monitor";
CREATE TABLE "public"."l2_tx_event_monitor" (
  "id" int8 NOT NULL DEFAULT nextval('l2_tx_event_monitor_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "l1_tx_hash" text COLLATE "pg_catalog"."default",
  "l1_block_height" int8,
  "sender_address" text COLLATE "pg_catalog"."default",
  "request_id" int8,
  "tx_type" int8,
  "pubdata" text COLLATE "pg_catalog"."default",
  "expiration_block" int8,
  "status" int8
)
;

-- ----------------------------
-- Records of l2_tx_event_monitor
-- ----------------------------
INSERT INTO "public"."l2_tx_event_monitor" VALUES (1, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.058004+00', NULL, '0x168dace8281eee3d90f81db58754161166af5c6c25b9db744fce8551a0c87af6', 399780, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 0, 1, '01000000007472656173757279000000000000000000000000000000000000000000000000167c5363088a40a4839912a872f43164270740c7e986ec55397b2d583317ab4a2005db7af2bdcfae1fa8d28833ae2f1995e9a8e0825377cff121db64b0db21b718a96ca582a72b16f464330c89ab73277cb96e42df105ebf5c9ac5330d47b8fc', 440100, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (2, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.059215+00', NULL, '0xc6e7f209ac872a25879b76f125b842f99556cff8451e7ec4af1300ef6502b4f3', 399782, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 1, 1, '010000000067617300000000000000000000000000000000000000000000000000000000000a48e9892a45a04d0c5b0f235a3aeb07b92137ba71a59b9c457774bafde959832c24415b75651673b0d7bbf145ac8d7cb744ba6926963d1d014836336df1317a134f4726b89983a8e7babbf6973e7ee16311e24328edf987bb0fbe7a494ec91e', 440102, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (3, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.059769+00', NULL, '0x2ababebda27a5441798f88b10b9b75ca29095b35c3965ee07aa8098067314e5d', 399785, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 2, 1, '01000000007368657200000000000000000000000000000000000000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45235fdbbbf5ef1665f3422211702126433c909487c456e594ef3a56910810396a05dde55c8adfb6689ead7f5610726afd5fd6ea35a3516dc68e57546146f7b6b0', 440105, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (4, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.060276+00', NULL, '0x1a1f35d1e50f4b9aa9abd110679e6aba0586f620a4c56a52308cd847b07490ed', 399788, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 3, 1, '0100000000676176696e0000000000000000000000000000000000000000000000000000001c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb0649fef47f6cf3dfb767cf5599eea11677bb6495956ec4cf75707d3aca7c06ed0e07b60bf3a2bf5e1a355793498de43e4d8dac50b892528f9664a03ceacc0005', 440108, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (5, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.060504+00', NULL, '0x4d5c3b3d15b48a17cd3920535581119be48a3ec8fb44eb621b1d587e14ca5a77', 399796, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 4, 4, '0400000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad4500000000000000000000016345785d8a0000', 440116, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (6, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.061004+00', NULL, '0x9ddacae8cde4948d91826f1badbce9136c50d2be70363b7f2d38b0bfc0d5ff5c', 399798, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 5, 4, '04000000001c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb00000000000000000000016345785d8a0000', 440118, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (7, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.061504+00', NULL, '0x418cdeec66b95e6d9d0a9460d3c44479c15951101143a86035f2cd800f5fab98', 399805, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 6, 4, '0400000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45000100000000000000056bc75e2d63100000', 440125, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (8, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.062503+00', NULL, '0x8b6f1bf70751be929f0f77d8b1ea204b045881f01494ae3a7d7d73657812d38b', 399808, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 7, 4, '0400000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45000200000000000000056bc75e2d63100000', 440128, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (9, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.06257+00', NULL, '0xa5c17af78376a85c9f213ea488adefbee9966d263e9047f8c31dea5408d227e7', 399815, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 8, 2, '02000000000002001e000000000005', 440135, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (10, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.063316+00', NULL, '0xe5d077d6f88dafa65e9e0058dab2e852e8302dc7032d36e005dad37418cbe439', 399817, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 9, 2, '02000100000001001e000000000005', 440137, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (11, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.063777+00', NULL, '0xbad7534ff00fac01c27cf86f3c1e0039ed8d9bcecd918a0ac1e9c1a56652bc6d', 399820, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 10, 2, '02000200010002001e000000000005', 440140, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (12, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.064004+00', NULL, '0x463dcd43d99faeea3d3db3001ee9b464696ec9200d2955b84ffba08ff92df38a', 399827, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 11, 3, '030001003200000000000a', 440147, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (13, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.0647+00', NULL, '0x390fb0013d788c35a9418f348e1b44c565d6f73b587e6c5d77f47ff858d8322a', 399837, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 12, 5, '05000000000000000000b7ad4a7e9459d0c1541db2eececeacc7dba803e1000000000000abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b0000000000000000000000000000000000000000000000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad450000', 440157, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (14, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.065229+00', NULL, '0x9cc1335db2ad0cdd2e006c64da7e2b63adacab522a15050949b5cab827f9f2a9', 399844, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 13, 17, '1100000000000100000000000000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45', 440164, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (15, '2022-06-08 08:25:32.572194+00', '2022-06-08 08:25:58.06572+00', NULL, '0xb0e1317c18b03378e3699829339135b7afbf090a6aaf985cd825abf109c08066', 399851, '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 14, 18, '1200000000000000000000000000000000000000000000000000000000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 440171, 2);

-- ----------------------------
-- Table structure for liquidity
-- ----------------------------
DROP TABLE IF EXISTS "public"."liquidity";
CREATE TABLE "public"."liquidity" (
  "id" int8 NOT NULL DEFAULT nextval('liquidity_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "pair_index" int8,
  "asset_a_id" int8,
  "asset_a" text COLLATE "pg_catalog"."default",
  "asset_b_id" int8,
  "asset_b" text COLLATE "pg_catalog"."default",
  "lp_amount" text COLLATE "pg_catalog"."default",
  "k_last" text COLLATE "pg_catalog"."default",
  "fee_rate" int8,
  "treasury_account_index" int8,
  "treasury_rate" int8
)
;

-- ----------------------------
-- Records of liquidity
-- ----------------------------
INSERT INTO "public"."liquidity" VALUES (1, '2022-06-08 08:25:58.055259+00', '2022-06-08 08:25:58.055259+00', NULL, 0, 0, '0', 2, '0', '0', '0', 30, 0, 5);
INSERT INTO "public"."liquidity" VALUES (2, '2022-06-08 08:25:58.055259+00', '2022-06-08 08:25:58.055259+00', NULL, 1, 0, '0', 1, '0', '0', '0', 50, 0, 10);
INSERT INTO "public"."liquidity" VALUES (3, '2022-06-08 08:25:58.055259+00', '2022-06-08 08:25:58.055259+00', NULL, 2, 1, '0', 2, '0', '0', '0', 30, 0, 5);

-- ----------------------------
-- Table structure for liquidity_history
-- ----------------------------
DROP TABLE IF EXISTS "public"."liquidity_history";
CREATE TABLE "public"."liquidity_history" (
  "id" int8 NOT NULL DEFAULT nextval('liquidity_history_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "pair_index" int8,
  "asset_a_id" int8,
  "asset_a" text COLLATE "pg_catalog"."default",
  "asset_b_id" int8,
  "asset_b" text COLLATE "pg_catalog"."default",
  "lp_amount" text COLLATE "pg_catalog"."default",
  "k_last" text COLLATE "pg_catalog"."default",
  "fee_rate" int8,
  "treasury_account_index" int8,
  "treasury_rate" int8,
  "l2_block_height" int8
)
;

-- ----------------------------
-- Records of liquidity_history
-- ----------------------------

-- ----------------------------
-- Table structure for mempool_tx
-- ----------------------------
DROP TABLE IF EXISTS "public"."mempool_tx";
CREATE TABLE "public"."mempool_tx" (
  "id" int8 NOT NULL DEFAULT nextval('mempool_tx_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "tx_hash" text COLLATE "pg_catalog"."default",
  "tx_type" int8,
  "gas_fee_asset_id" int8,
  "gas_fee" text COLLATE "pg_catalog"."default",
  "nft_index" int8,
  "pair_index" int8,
  "asset_id" int8,
  "tx_amount" text COLLATE "pg_catalog"."default",
  "native_address" text COLLATE "pg_catalog"."default",
  "tx_info" text COLLATE "pg_catalog"."default",
  "extra_info" text COLLATE "pg_catalog"."default",
  "memo" text COLLATE "pg_catalog"."default",
  "account_index" int8,
  "nonce" int8,
  "expired_at" int8,
  "l2_block_height" int8,
  "status" int8
)
;

-- ----------------------------
-- Records of mempool_tx
-- ----------------------------
INSERT INTO "public"."mempool_tx" VALUES (1, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f5005a9-e704-11ec-b6f3-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":0,"AccountName":"treasury.legend","AccountNameHash":"FnxTYwiKQKSDmRKocvQxZCcHQMfphuxVOXstWDMXq0o=","PubKey":"fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998"}', '', '', 0, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (2, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f50d170-e704-11ec-b6f3-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":1,"AccountName":"gas.legend","AccountNameHash":"CkjpiSpFoE0MWw8jWjrrB7khN7pxpZucRXd0uv3pWYM=","PubKey":"1ec94e497abe0fbb87f9ed2843e21163e17e3e97f6bbbae7a88399b826474f93"}', '', '', 1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (3, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f50f093-e704-11ec-b6f3-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":2,"AccountName":"sher.legend","AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","PubKey":"b0b6f7466154578ec66d51a335ead65ffd6a7210567fad9e68b6df8a5ce5dd85"}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (4, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6f3-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":3,"AccountName":"gavin.legend","AccountNameHash":"HFTAnJj3renV7rpBJKx8kS5laZo/dvpl1x6vY1nZvOs=","PubKey":"0500ccea3ca064968f5292b850ac8d4d3ee48d499357351a5ebfa2f30bb6070e"}', '', '', 3, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (5, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6f4-988fe0603efa', 4, 0, '0', -1, -1, 0, '100000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (6, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6f5-988fe0603efa', 4, 0, '0', -1, -1, 0, '100000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":3,"AccountNameHash":"HFTAnJj3renV7rpBJKx8kS5laZo/dvpl1x6vY1nZvOs=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 3, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (7, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6f6-988fe0603efa', 4, 0, '0', -1, -1, 1, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (8, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6f7-988fe0603efa', 4, 0, '0', -1, -1, 2, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":2,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (9, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6f8-988fe0603efa', 2, 0, '0', -1, 0, 0, '0', '0', '{"TxType":2,"PairIndex":0,"AssetAId":0,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (10, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6f9-988fe0603efa', 2, 0, '0', -1, 1, 0, '0', '0', '{"TxType":2,"PairIndex":1,"AssetAId":0,"AssetBId":1,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (11, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6fa-988fe0603efa', 2, 0, '0', -1, 2, 0, '0', '0', '{"TxType":2,"PairIndex":2,"AssetAId":1,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (12, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f510cb1-e704-11ec-b6fb-988fe0603efa', 3, 0, '0', -1, 1, 0, '0', '0', '{"TxType":3,"PairIndex":1,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', '', '', -1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (13, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f51d005-e704-11ec-b6fb-988fe0603efa', 5, 0, '0', 0, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":5,"AccountIndex":2,"NftIndex":0,"NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftContentHash":"q9G2rnlQf3tKMqhKtklbyf7mdFDtMW27p2us6KPFGXs=","NftL1TokenId":0,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","CollectionId":0}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (14, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f51d005-e704-11ec-b6fc-988fe0603efa', 17, 0, '0', -1, -1, 1, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":17,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (15, '2022-06-08 08:25:58.050504+00', '2022-06-08 08:25:58.050504+00', NULL, '9f51d005-e704-11ec-b6fd-988fe0603efa', 18, 0, '0', 0, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":18,"AccountIndex":2,"CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftIndex":0,"CollectionId":0,"NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","CreatorAccountNameHash":"AA==","NftContentHash":"q9G2rnlQf3tKMqhKtklbyf7mdFDtMW27p2us6KPFGXs=","NftL1TokenId":0}', '', '', 2, 0, 0, -1, 0);

-- ----------------------------
-- Table structure for mempool_tx_detail
-- ----------------------------
DROP TABLE IF EXISTS "public"."mempool_tx_detail";
CREATE TABLE "public"."mempool_tx_detail" (
  "id" int8 NOT NULL DEFAULT nextval('mempool_tx_detail_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "tx_id" int8,
  "asset_id" int8,
  "asset_type" int8,
  "account_index" int8,
  "account_name" text COLLATE "pg_catalog"."default",
  "balance_delta" text COLLATE "pg_catalog"."default",
  "order" int8,
  "account_order" int8
)
;

-- ----------------------------
-- Records of mempool_tx_detail
-- ----------------------------
INSERT INTO "public"."mempool_tx_detail" VALUES (1, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 5, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (2, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 6, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (3, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 7, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (4, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 8, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (5, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 9, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (6, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 10, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (7, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 11, 2, 2, -1, '', '{"PairIndex":2,"AssetAId":1,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (8, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 12, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (9, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 13, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (10, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 13, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":2,"NftContentHash":"abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b","NftL1TokenId":"0","NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","CreatorTreasuryRate":0,"CollectionId":0}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (12, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 15, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (13, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 15, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (11, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 14, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);

-- ----------------------------
-- Table structure for offer
-- ----------------------------
DROP TABLE IF EXISTS "public"."offer";
CREATE TABLE "public"."offer" (
  "id" int8 NOT NULL DEFAULT nextval('offer_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "offer_type" int8,
  "offer_id" int8,
  "account_index" int8,
  "nft_index" int8,
  "asset_id" int8,
  "asset_amount" text COLLATE "pg_catalog"."default",
  "listed_at" int8,
  "expired_at" int8,
  "treasury_rate" int8,
  "sig" text COLLATE "pg_catalog"."default",
  "status" int8
)
;

-- ----------------------------
-- Records of offer
-- ----------------------------

-- ----------------------------
-- Table structure for proof_sender
-- ----------------------------
DROP TABLE IF EXISTS "public"."proof_sender";
CREATE TABLE "public"."proof_sender" (
  "id" int8 NOT NULL DEFAULT nextval('proof_sender_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "proof_info" text COLLATE "pg_catalog"."default",
  "block_number" int8,
  "status" int8
)
;

-- ----------------------------
-- Records of proof_sender
-- ----------------------------

-- ----------------------------
-- Table structure for sys_config
-- ----------------------------
DROP TABLE IF EXISTS "public"."sys_config";
CREATE TABLE "public"."sys_config" (
  "id" int8 NOT NULL DEFAULT nextval('sys_config_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "name" text COLLATE "pg_catalog"."default",
  "value" text COLLATE "pg_catalog"."default",
  "value_type" text COLLATE "pg_catalog"."default",
  "comment" text COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Records of sys_config
-- ----------------------------
INSERT INTO "public"."sys_config" VALUES (1, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'SysGasFee', '1', 'float', 'based on ETH');
INSERT INTO "public"."sys_config" VALUES (2, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'MaxAssetId', '9', 'int', 'max number of asset id');
INSERT INTO "public"."sys_config" VALUES (3, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'TreasuryAccountIndex', '0', 'int', 'treasury index');
INSERT INTO "public"."sys_config" VALUES (4, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'GasAccountIndex', '1', 'int', 'gas index');
INSERT INTO "public"."sys_config" VALUES (5, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'ZecreyLegendContract', '0x39c6354FdB9009E15B4006205E5Aa4C08c558c35', 'string', 'Zecrey contract on BSC');
INSERT INTO "public"."sys_config" VALUES (6, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'GovernanceContract', '0x5b7adDf0882aB683E5aC0BD880830eb0947B2BD1', 'string', 'Governance contract on BSC');
INSERT INTO "public"."sys_config" VALUES (7, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'BscTestNetworkRpc', 'http://tf-dex-preview-validator-nlb-6fd109ac8b9d390a.elb.ap-northeast-1.amazonaws.com:8545', 'string', 'BSC network rpc');
INSERT INTO "public"."sys_config" VALUES (8, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'Local_Test_Network_RPC', 'http://127.0.0.1:8545/', 'string', 'Local network rpc');
INSERT INTO "public"."sys_config" VALUES (9, '2022-06-08 08:24:56.738464+00', '2022-06-08 08:24:56.738464+00', NULL, 'AssetGovernanceContract', '0x4C7B3D1c2aafcE6Ca3a7c35c25fC717178565DE2', 'string', 'asset governance contract');
INSERT INTO "public"."sys_config" VALUES (10, '2022-06-08 08:24:56.738464+00', '2022-06-08 08:24:56.738464+00', NULL, 'Validators', '{"0x9A973e0b7dB1935Ffb59Ff35272332e6feE00182":{"Address":"0x9A973e0b7dB1935Ffb59Ff35272332e6feE00182","IsActive":true}}', 'map[string]*ValidatorInfo', 'validator info');
INSERT INTO "public"."sys_config" VALUES (11, '2022-06-08 08:24:56.738464+00', '2022-06-08 08:24:56.738464+00', NULL, 'Governor', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 'string', 'governor');

-- ----------------------------
-- Table structure for tx
-- ----------------------------
DROP TABLE IF EXISTS "public"."tx";
CREATE TABLE "public"."tx" (
  "id" int8 NOT NULL DEFAULT nextval('tx_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "tx_hash" text COLLATE "pg_catalog"."default",
  "tx_type" int8,
  "gas_fee" text COLLATE "pg_catalog"."default",
  "gas_fee_asset_id" int8,
  "tx_status" int8,
  "block_height" int8,
  "block_id" int8,
  "state_root" text COLLATE "pg_catalog"."default",
  "nft_index" int8,
  "pair_index" int8,
  "asset_id" int8,
  "tx_amount" text COLLATE "pg_catalog"."default",
  "native_address" text COLLATE "pg_catalog"."default",
  "tx_info" text COLLATE "pg_catalog"."default",
  "extra_info" text COLLATE "pg_catalog"."default",
  "memo" text COLLATE "pg_catalog"."default",
  "account_index" int8,
  "nonce" int8,
  "expired_at" int8
)
;

-- ----------------------------
-- Records of tx
-- ----------------------------

-- ----------------------------
-- Table structure for tx_detail
-- ----------------------------
DROP TABLE IF EXISTS "public"."tx_detail";
CREATE TABLE "public"."tx_detail" (
  "id" int8 NOT NULL DEFAULT nextval('tx_detail_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "tx_id" int8,
  "asset_id" int8,
  "asset_type" int8,
  "account_index" int8,
  "account_name" text COLLATE "pg_catalog"."default",
  "balance" text COLLATE "pg_catalog"."default",
  "balance_delta" text COLLATE "pg_catalog"."default",
  "order" int8,
  "account_order" int8,
  "nonce" int8,
  "collection_nonce" int8
)
;

-- ----------------------------
-- Records of tx_detail
-- ----------------------------

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."account_history_id_seq"
OWNED BY "public"."account_history"."id";
SELECT setval('"public"."account_history_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."account_id_seq"
OWNED BY "public"."account"."id";
SELECT setval('"public"."account_id_seq"', 4, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."block_for_commit_id_seq"
OWNED BY "public"."block_for_commit"."id";
SELECT setval('"public"."block_for_commit_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."block_id_seq"
OWNED BY "public"."block"."id";
SELECT setval('"public"."block_id_seq"', 1, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."fail_tx_id_seq"
OWNED BY "public"."fail_tx"."id";
SELECT setval('"public"."fail_tx_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l1_amount_id_seq"
OWNED BY "public"."l1_amount"."id";
SELECT setval('"public"."l1_amount_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l1_block_monitor_id_seq"
OWNED BY "public"."l1_block_monitor"."id";
SELECT setval('"public"."l1_block_monitor_id_seq"', 4, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l1_tx_sender_id_seq"
OWNED BY "public"."l1_tx_sender"."id";
SELECT setval('"public"."l1_tx_sender_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_asset_info_id_seq"
OWNED BY "public"."l2_asset_info"."id";
SELECT setval('"public"."l2_asset_info_id_seq"', 3, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_block_event_monitor_id_seq"
OWNED BY "public"."l2_block_event_monitor"."id";
SELECT setval('"public"."l2_block_event_monitor_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_collection_id_seq"
OWNED BY "public"."l2_nft_collection"."id";
SELECT setval('"public"."l2_nft_collection_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_exchange_history_id_seq"
OWNED BY "public"."l2_nft_exchange_history"."id";
SELECT setval('"public"."l2_nft_exchange_history_id_seq"', 2, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_exchange_id_seq"
OWNED BY "public"."l2_nft_exchange"."id";
SELECT setval('"public"."l2_nft_exchange_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_history_id_seq"
OWNED BY "public"."l2_nft_history"."id";
SELECT setval('"public"."l2_nft_history_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_id_seq"
OWNED BY "public"."l2_nft"."id";
SELECT setval('"public"."l2_nft_id_seq"', 1, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_withdraw_history_id_seq"
OWNED BY "public"."l2_nft_withdraw_history"."id";
SELECT setval('"public"."l2_nft_withdraw_history_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_tx_event_monitor_id_seq"
OWNED BY "public"."l2_tx_event_monitor"."id";
SELECT setval('"public"."l2_tx_event_monitor_id_seq"', 15, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."liquidity_history_id_seq"
OWNED BY "public"."liquidity_history"."id";
SELECT setval('"public"."liquidity_history_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."liquidity_id_seq"
OWNED BY "public"."liquidity"."id";
SELECT setval('"public"."liquidity_id_seq"', 3, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."mempool_tx_detail_id_seq"
OWNED BY "public"."mempool_tx_detail"."id";
SELECT setval('"public"."mempool_tx_detail_id_seq"', 13, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."mempool_tx_id_seq"
OWNED BY "public"."mempool_tx"."id";
SELECT setval('"public"."mempool_tx_id_seq"', 15, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."offer_id_seq"
OWNED BY "public"."offer"."id";
SELECT setval('"public"."offer_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."proof_sender_id_seq"
OWNED BY "public"."proof_sender"."id";
SELECT setval('"public"."proof_sender_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."sys_config_id_seq"
OWNED BY "public"."sys_config"."id";
SELECT setval('"public"."sys_config_id_seq"', 11, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."tx_detail_id_seq"
OWNED BY "public"."tx_detail"."id";
SELECT setval('"public"."tx_detail_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."tx_id_seq"
OWNED BY "public"."tx"."id";
SELECT setval('"public"."tx_id_seq"', 1, false);

-- ----------------------------
-- Indexes structure for table account
-- ----------------------------
CREATE UNIQUE INDEX "idx_account_account_index" ON "public"."account" USING btree (
  "account_index" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_account_account_name" ON "public"."account" USING btree (
  "account_name" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_account_account_name_hash" ON "public"."account" USING btree (
  "account_name_hash" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_account_deleted_at" ON "public"."account" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_account_public_key" ON "public"."account" USING btree (
  "public_key" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table account
-- ----------------------------
ALTER TABLE "public"."account" ADD CONSTRAINT "account_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table account_history
-- ----------------------------
CREATE INDEX "idx_account_history_account_index" ON "public"."account_history" USING btree (
  "account_index" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_account_history_deleted_at" ON "public"."account_history" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table account_history
-- ----------------------------
ALTER TABLE "public"."account_history" ADD CONSTRAINT "account_history_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table block
-- ----------------------------
CREATE INDEX "idx_block_deleted_at" ON "public"."block" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table block
-- ----------------------------
ALTER TABLE "public"."block" ADD CONSTRAINT "block_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table block_for_commit
-- ----------------------------
CREATE INDEX "idx_block_for_commit_deleted_at" ON "public"."block_for_commit" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table block_for_commit
-- ----------------------------
ALTER TABLE "public"."block_for_commit" ADD CONSTRAINT "block_for_commit_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table fail_tx
-- ----------------------------
CREATE INDEX "idx_fail_tx_deleted_at" ON "public"."fail_tx" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_fail_tx_tx_hash" ON "public"."fail_tx" USING btree (
  "tx_hash" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table fail_tx
-- ----------------------------
ALTER TABLE "public"."fail_tx" ADD CONSTRAINT "fail_tx_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l1_amount
-- ----------------------------
CREATE INDEX "idx_l1_amount_asset_id" ON "public"."l1_amount" USING btree (
  "asset_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_l1_amount_block_height" ON "public"."l1_amount" USING btree (
  "block_height" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_l1_amount_deleted_at" ON "public"."l1_amount" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l1_amount
-- ----------------------------
ALTER TABLE "public"."l1_amount" ADD CONSTRAINT "l1_amount_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l1_block_monitor
-- ----------------------------
CREATE INDEX "idx_l1_block_monitor_deleted_at" ON "public"."l1_block_monitor" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l1_block_monitor
-- ----------------------------
ALTER TABLE "public"."l1_block_monitor" ADD CONSTRAINT "l1_block_monitor_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l1_tx_sender
-- ----------------------------
CREATE INDEX "idx_l1_tx_sender_deleted_at" ON "public"."l1_tx_sender" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l1_tx_sender
-- ----------------------------
ALTER TABLE "public"."l1_tx_sender" ADD CONSTRAINT "l1_tx_sender_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_asset_info
-- ----------------------------
CREATE UNIQUE INDEX "idx_l2_asset_info_asset_id" ON "public"."l2_asset_info" USING btree (
  "asset_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_l2_asset_info_deleted_at" ON "public"."l2_asset_info" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_asset_info
-- ----------------------------
ALTER TABLE "public"."l2_asset_info" ADD CONSTRAINT "l2_asset_info_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_block_event_monitor
-- ----------------------------
CREATE INDEX "idx_l2_block_event_monitor_block_event_type" ON "public"."l2_block_event_monitor" USING btree (
  "block_event_type" "pg_catalog"."int2_ops" ASC NULLS LAST
);
CREATE INDEX "idx_l2_block_event_monitor_deleted_at" ON "public"."l2_block_event_monitor" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_l2_block_event_monitor_l2_block_height" ON "public"."l2_block_event_monitor" USING btree (
  "l2_block_height" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_block_event_monitor
-- ----------------------------
ALTER TABLE "public"."l2_block_event_monitor" ADD CONSTRAINT "l2_block_event_monitor_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_nft
-- ----------------------------
CREATE INDEX "idx_l2_nft_deleted_at" ON "public"."l2_nft" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_l2_nft_nft_index" ON "public"."l2_nft" USING btree (
  "nft_index" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_nft
-- ----------------------------
ALTER TABLE "public"."l2_nft" ADD CONSTRAINT "l2_nft_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_nft_collection
-- ----------------------------
CREATE INDEX "idx_l2_nft_collection_deleted_at" ON "public"."l2_nft_collection" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_nft_collection
-- ----------------------------
ALTER TABLE "public"."l2_nft_collection" ADD CONSTRAINT "l2_nft_collection_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_nft_exchange
-- ----------------------------
CREATE INDEX "idx_l2_nft_exchange_deleted_at" ON "public"."l2_nft_exchange" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_nft_exchange
-- ----------------------------
ALTER TABLE "public"."l2_nft_exchange" ADD CONSTRAINT "l2_nft_exchange_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_nft_exchange_history
-- ----------------------------
CREATE INDEX "idx_l2_nft_exchange_history_deleted_at" ON "public"."l2_nft_exchange_history" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_nft_exchange_history
-- ----------------------------
ALTER TABLE "public"."l2_nft_exchange_history" ADD CONSTRAINT "l2_nft_exchange_history_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_nft_history
-- ----------------------------
CREATE INDEX "idx_l2_nft_history_deleted_at" ON "public"."l2_nft_history" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_nft_history
-- ----------------------------
ALTER TABLE "public"."l2_nft_history" ADD CONSTRAINT "l2_nft_history_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_nft_withdraw_history
-- ----------------------------
CREATE INDEX "idx_l2_nft_withdraw_history_deleted_at" ON "public"."l2_nft_withdraw_history" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_l2_nft_withdraw_history_nft_index" ON "public"."l2_nft_withdraw_history" USING btree (
  "nft_index" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_nft_withdraw_history
-- ----------------------------
ALTER TABLE "public"."l2_nft_withdraw_history" ADD CONSTRAINT "l2_nft_withdraw_history_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table l2_tx_event_monitor
-- ----------------------------
CREATE INDEX "idx_l2_tx_event_monitor_deleted_at" ON "public"."l2_tx_event_monitor" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table l2_tx_event_monitor
-- ----------------------------
ALTER TABLE "public"."l2_tx_event_monitor" ADD CONSTRAINT "l2_tx_event_monitor_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table liquidity
-- ----------------------------
CREATE INDEX "idx_liquidity_deleted_at" ON "public"."liquidity" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table liquidity
-- ----------------------------
ALTER TABLE "public"."liquidity" ADD CONSTRAINT "liquidity_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table liquidity_history
-- ----------------------------
CREATE INDEX "idx_liquidity_history_deleted_at" ON "public"."liquidity_history" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table liquidity_history
-- ----------------------------
ALTER TABLE "public"."liquidity_history" ADD CONSTRAINT "liquidity_history_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table mempool_tx
-- ----------------------------
CREATE INDEX "idx_mempool_tx_deleted_at" ON "public"."mempool_tx" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_mempool_tx_status" ON "public"."mempool_tx" USING btree (
  "status" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_mempool_tx_tx_hash" ON "public"."mempool_tx" USING btree (
  "tx_hash" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table mempool_tx
-- ----------------------------
ALTER TABLE "public"."mempool_tx" ADD CONSTRAINT "mempool_tx_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table mempool_tx_detail
-- ----------------------------
CREATE INDEX "idx_mempool_tx_detail_account_index" ON "public"."mempool_tx_detail" USING btree (
  "account_index" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_mempool_tx_detail_deleted_at" ON "public"."mempool_tx_detail" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_mempool_tx_detail_tx_id" ON "public"."mempool_tx_detail" USING btree (
  "tx_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table mempool_tx_detail
-- ----------------------------
ALTER TABLE "public"."mempool_tx_detail" ADD CONSTRAINT "mempool_tx_detail_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table offer
-- ----------------------------
CREATE INDEX "idx_offer_deleted_at" ON "public"."offer" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table offer
-- ----------------------------
ALTER TABLE "public"."offer" ADD CONSTRAINT "offer_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table proof_sender
-- ----------------------------
CREATE INDEX "idx_proof_sender_block_number" ON "public"."proof_sender" USING btree (
  "block_number" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_proof_sender_deleted_at" ON "public"."proof_sender" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table proof_sender
-- ----------------------------
ALTER TABLE "public"."proof_sender" ADD CONSTRAINT "proof_sender_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table sys_config
-- ----------------------------
CREATE INDEX "idx_sys_config_deleted_at" ON "public"."sys_config" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table sys_config
-- ----------------------------
ALTER TABLE "public"."sys_config" ADD CONSTRAINT "sys_config_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table tx
-- ----------------------------
CREATE INDEX "idx_tx_block_height" ON "public"."tx" USING btree (
  "block_height" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_tx_block_id" ON "public"."tx" USING btree (
  "block_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_tx_deleted_at" ON "public"."tx" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_tx_tx_hash" ON "public"."tx" USING btree (
  "tx_hash" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table tx
-- ----------------------------
ALTER TABLE "public"."tx" ADD CONSTRAINT "tx_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table tx_detail
-- ----------------------------
CREATE INDEX "idx_tx_detail_account_index" ON "public"."tx_detail" USING btree (
  "account_index" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_tx_detail_deleted_at" ON "public"."tx_detail" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_tx_detail_tx_id" ON "public"."tx_detail" USING btree (
  "tx_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table tx_detail
-- ----------------------------
ALTER TABLE "public"."tx_detail" ADD CONSTRAINT "tx_detail_pkey" PRIMARY KEY ("id");
