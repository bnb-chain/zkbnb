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

 Date: 16/06/2022 11:03:16
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
-- Sequence structure for asset_info_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."asset_info_id_seq";
CREATE SEQUENCE "public"."asset_info_id_seq" 
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
INSERT INTO "public"."account" VALUES (1, '2022-06-16 03:02:50.036268+00', '2022-06-16 03:02:50.036268+00', NULL, 0, 'treasury.legend', 'fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998', 'c0d201aace9a2c17ce7066dc6ffefaf7930f1317c4c95d0661b164a1c584d676', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 0);
INSERT INTO "public"."account" VALUES (2, '2022-06-16 03:02:50.036268+00', '2022-06-16 03:02:50.036268+00', NULL, 1, 'gas.legend', '1ec94e497abe0fbb87f9ed2843e21163e17e3e97f6bbbae7a88399b826474f93', '68fbd17e77eec501c677ccc31c260f30ee8ed049c893900e084ba8b7f7569ce6', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 0);
INSERT INTO "public"."account" VALUES (3, '2022-06-16 03:02:50.036268+00', '2022-06-16 03:02:50.036268+00', NULL, 2, 'sher.legend', 'b0b6f7466154578ec66d51a335ead65ffd6a7210567fad9e68b6df8a5ce5dd85', '04b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 0);
INSERT INTO "public"."account" VALUES (4, '2022-06-16 03:02:50.036268+00', '2022-06-16 03:02:50.036268+00', NULL, 3, 'gavin.legend', '0500ccea3ca064968f5292b850ac8d4d3ee48d499357351a5ebfa2f30bb6070e', 'f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a3', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 0);

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
-- Table structure for asset_info
-- ----------------------------
DROP TABLE IF EXISTS "public"."asset_info";
CREATE TABLE "public"."asset_info" (
  "id" int8 NOT NULL DEFAULT nextval('asset_info_id_seq'::regclass),
  "created_at" timestamptz(6),
  "updated_at" timestamptz(6),
  "deleted_at" timestamptz(6),
  "asset_id" int8,
  "asset_name" text COLLATE "pg_catalog"."default",
  "asset_symbol" text COLLATE "pg_catalog"."default",
  "l1_address" text COLLATE "pg_catalog"."default",
  "decimals" int8,
  "status" int8
)
;

-- ----------------------------
-- Records of asset_info
-- ----------------------------
INSERT INTO "public"."asset_info" VALUES (1, '2022-06-16 03:01:29.40839+00', '2022-06-16 03:01:29.40839+00', NULL, 0, 'BNB', 'BNB', '0x00', 18, 0);
INSERT INTO "public"."asset_info" VALUES (2, '2022-06-16 03:02:04.096602+00', '2022-06-16 03:02:04.096602+00', NULL, 1, 'LEG', 'LEG', '0xDFF05aF25a5A56A3c7afFcB269235caE21eE53d8', 18, 0);
INSERT INTO "public"."asset_info" VALUES (3, '2022-06-16 03:02:04.096602+00', '2022-06-16 03:02:04.096602+00', NULL, 2, 'REY', 'REY', '0xE2Bd0916DFC2f5B9e05a4936982B67013Fbd338F', 18, 0);

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
INSERT INTO "public"."block" VALUES (1, '2022-06-16 03:01:29.413451+00', '2022-06-16 03:01:29.413451+00', NULL, '0000000000000000000000000000000000000000000000000000000000000000', 0, '14e4e8ad4848558d7200530337052e1ad30f5385b3c7187c80ad85f48547b74f', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 3);

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
INSERT INTO "public"."l1_block_monitor" VALUES (1, '2022-06-16 03:02:04.09326+00', '2022-06-16 03:02:04.09326+00', NULL, 628, '[{"EventType":6,"TxHash":"0xb38b074f0ee26bac5564982245882b917f58866cebfb9625e9d15359fb6c2054"},{"EventType":7,"TxHash":"0xb38b074f0ee26bac5564982245882b917f58866cebfb9625e9d15359fb6c2054"},{"EventType":5,"TxHash":"0xb38b074f0ee26bac5564982245882b917f58866cebfb9625e9d15359fb6c2054"},{"EventType":4,"TxHash":"0x4b64139d0696ef64b76f44ebd5b367bb6e887363a4bee19ca4d549ba7a2d7d4f"},{"EventType":4,"TxHash":"0xedc14dd231ebb62ec501ad9b74b42e4c23257c0c175cfd5501a2ef9455cdad2f"}]', 1);
INSERT INTO "public"."l1_block_monitor" VALUES (2, '2022-06-16 03:02:24.114168+00', '2022-06-16 03:02:24.114168+00', NULL, 628, '[{"EventType":0,"TxHash":"0xb47eac43aeaa52c33ae80b1fe7433e221b6c8003a7c4ba54984eff42ac0770ef"},{"EventType":0,"TxHash":"0xebba40069a895f43d4f1cae8a114c4f416d8354a63fec0b845ee3ae296f1783b"},{"EventType":0,"TxHash":"0x4640fb00a59749a95cc08823d4ca95c8936ddd89503633b7076678df2da5df1e"},{"EventType":0,"TxHash":"0xea08984a16136275a6849daa793c20dc272a37fd1266e3502923950509516e92"},{"EventType":0,"TxHash":"0x7184bd484e97ae61dcdddeb3282470a191dee273383369c275875144ed7022fb"},{"EventType":0,"TxHash":"0xdb8dd7b95f720396d1fb19a6789989983bfa3deab0c6b68af03ed0f34ed65e95"},{"EventType":0,"TxHash":"0x3f31a9fa97d4ce93e9858331e8ecce6699043b5c22b384f1b362a58e8d8a51f0"},{"EventType":0,"TxHash":"0xd50ede290f499c12156ff59dad8dfef00082f980ecf53a3ab8aa6596c926ad4e"},{"EventType":0,"TxHash":"0x33a5529ba1899f41b52d26ee4167808876f28bcdd4b1961919dd70432f64bb1f"},{"EventType":0,"TxHash":"0x4f62a0d4fb2a17e225f39cf832b101aa5daadfa2895e1b19f9fa53f8f21ffea7"},{"EventType":0,"TxHash":"0xed3c1beb710e7be01f88e5eee08a4e6d40e74af08e854a761a6e0a721f90d687"},{"EventType":0,"TxHash":"0xa97822618abccde3fa3fc38753bb720751337e0a2e86a82f9a600380cbf12e2b"},{"EventType":0,"TxHash":"0xf2cd0648ddeacb4234e725966e36d642ab47be5e011c83602ab9d5f84e5de62a"},{"EventType":0,"TxHash":"0xc0cdeaa451c6678d9368858ae55a76288c383c199871daabd0165e6c8a3e1237"},{"EventType":0,"TxHash":"0xcd9f5635ee8a285f545afa70d30cc9448a782ea6f72ac2baa4c6eef1ba2278e5"}]', 0);

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
INSERT INTO "public"."l2_asset_info" VALUES (1, '2022-06-14 06:43:24.621929+00', '2022-06-14 06:43:24.621929+00', NULL, 0, '0x00', 'BNB', 'BNB', 18, 0);
INSERT INTO "public"."l2_asset_info" VALUES (2, '2022-06-14 06:44:37.378403+00', '2022-06-14 06:44:37.378403+00', NULL, 1, '0x6b8bdbAACf09C562409Eb5f811A619D5c1A38c9D', 'LEG', 'LEG', 18, 0);
INSERT INTO "public"."l2_asset_info" VALUES (3, '2022-06-14 06:44:37.378403+00', '2022-06-14 06:44:37.378403+00', NULL, 2, '0xdDD0811dAD9d7Ef6518e0275c2e52BD9B837b6cD', 'REY', 'REY', 18, 0);

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
INSERT INTO "public"."l2_nft" VALUES (1, '2022-06-16 03:02:50.046873+00', '2022-06-16 03:02:50.046873+00', NULL, 0, 0, 2, '8fa3059a7c68daddcdf9c03b1cd1e6d0342b7c4a90ed610372c681bfea7ee478', '0x464ed8Ce7076Abaf743F760468230B9d71fB7D90', '0', 0, 0);

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
INSERT INTO "public"."l2_tx_event_monitor" VALUES (1, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.047909+00', NULL, '0xb47eac43aeaa52c33ae80b1fe7433e221b6c8003a7c4ba54984eff42ac0770ef', 605, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 0, 1, '01000000007472656173757279000000000000000000000000000000000000000000000000c0d201aace9a2c17ce7066dc6ffefaf7930f1317c4c95d0661b164a1c584d6762005db7af2bdcfae1fa8d28833ae2f1995e9a8e0825377cff121db64b0db21b718a96ca582a72b16f464330c89ab73277cb96e42df105ebf5c9ac5330d47b8fc', 40925, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (2, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.049562+00', NULL, '0xebba40069a895f43d4f1cae8a114c4f416d8354a63fec0b845ee3ae296f1783b', 606, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 1, 1, '0100000001676173000000000000000000000000000000000000000000000000000000000068fbd17e77eec501c677ccc31c260f30ee8ed049c893900e084ba8b7f7569ce62c24415b75651673b0d7bbf145ac8d7cb744ba6926963d1d014836336df1317a134f4726b89983a8e7babbf6973e7ee16311e24328edf987bb0fbe7a494ec91e', 40926, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (3, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.05065+00', NULL, '0x4640fb00a59749a95cc08823d4ca95c8936ddd89503633b7076678df2da5df1e', 607, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 2, 1, '0100000002736865720000000000000000000000000000000000000000000000000000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f235fdbbbf5ef1665f3422211702126433c909487c456e594ef3a56910810396a05dde55c8adfb6689ead7f5610726afd5fd6ea35a3516dc68e57546146f7b6b0', 40927, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (4, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.051304+00', NULL, '0xea08984a16136275a6849daa793c20dc272a37fd1266e3502923950509516e92', 608, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 3, 1, '0100000003676176696e000000000000000000000000000000000000000000000000000000f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a30649fef47f6cf3dfb767cf5599eea11677bb6495956ec4cf75707d3aca7c06ed0e07b60bf3a2bf5e1a355793498de43e4d8dac50b892528f9664a03ceacc0005', 40928, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (5, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.051809+00', NULL, '0x7184bd484e97ae61dcdddeb3282470a191dee273383369c275875144ed7022fb', 610, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 4, 4, '040000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f00000000000000000000016345785d8a0000', 40930, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (6, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.052336+00', NULL, '0xdb8dd7b95f720396d1fb19a6789989983bfa3deab0c6b68af03ed0f34ed65e95', 611, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 5, 4, '0400000000f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a300000000000000000000016345785d8a0000', 40931, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (7, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.053406+00', NULL, '0x3f31a9fa97d4ce93e9858331e8ecce6699043b5c22b384f1b362a58e8d8a51f0', 614, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 6, 4, '040000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f000100000000000000056bc75e2d63100000', 40934, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (8, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.053931+00', NULL, '0xd50ede290f499c12156ff59dad8dfef00082f980ecf53a3ab8aa6596c926ad4e', 615, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 7, 4, '040000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f000200000000000000056bc75e2d63100000', 40935, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (9, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.054595+00', NULL, '0x33a5529ba1899f41b52d26ee4167808876f28bcdd4b1961919dd70432f64bb1f', 617, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 8, 2, '02000000000002001e000000000005', 40937, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (10, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.055102+00', NULL, '0x4f62a0d4fb2a17e225f39cf832b101aa5daadfa2895e1b19f9fa53f8f21ffea7', 618, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 9, 2, '02000100000001001e000000000005', 40938, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (11, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.056161+00', NULL, '0xed3c1beb710e7be01f88e5eee08a4e6d40e74af08e854a761a6e0a721f90d687', 619, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 10, 2, '02000200010002001e000000000005', 40939, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (12, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.056817+00', NULL, '0xa97822618abccde3fa3fc38753bb720751337e0a2e86a82f9a600380cbf12e2b', 621, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 11, 3, '030001003200000000000a', 40941, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (13, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.057379+00', NULL, '0xf2cd0648ddeacb4234e725966e36d642ab47be5e011c83602ab9d5f84e5de62a', 624, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 12, 5, '05000000000000000000464ed8ce7076abaf743f760468230b9d71fb7d900000000000008fa3059a7c68daddcdf9c03b1cd1e6d0342b7c4a90ed610372c681bfea7ee478000000000000000000000000000000000000000000000000000000000000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f0000', 40944, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (14, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.057914+00', NULL, '0xc0cdeaa451c6678d9368858ae55a76288c383c199871daabd0165e6c8a3e1237', 626, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 13, 17, '110000000000010000000000000000000000000000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f', 40946, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (15, '2022-06-16 03:02:24.117031+00', '2022-06-16 03:02:50.058445+00', NULL, '0xcd9f5635ee8a285f545afa70d30cc9448a782ea6f72ac2baa4c6eef1ba2278e5', 628, '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 14, 18, '120000000000000000000000000000000000000000000000000000000000000000000000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 40948, 2);

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
INSERT INTO "public"."liquidity" VALUES (1, '2022-06-16 03:02:50.044467+00', '2022-06-16 03:02:50.044467+00', NULL, 0, 0, '0', 2, '0', '0', '0', 30, 0, 5);
INSERT INTO "public"."liquidity" VALUES (2, '2022-06-16 03:02:50.044467+00', '2022-06-16 03:02:50.044467+00', NULL, 1, 0, '0', 1, '0', '0', '0', 50, 0, 10);
INSERT INTO "public"."liquidity" VALUES (3, '2022-06-16 03:02:50.044467+00', '2022-06-16 03:02:50.044467+00', NULL, 2, 1, '0', 2, '0', '0', '0', 30, 0, 5);

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
INSERT INTO "public"."mempool_tx" VALUES (1, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce76403c-ed20-11ec-8b10-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":1,"AccountIndex":0,"AccountName":"treasury.legend","AccountNameHash":"wNIBqs6aLBfOcGbcb/7695MPExfEyV0GYbFkocWE1nY=","PubKey":"fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998"}', '', '', 0, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (2, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce771ece-ed20-11ec-8b10-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":1,"AccountIndex":1,"AccountName":"gas.legend","AccountNameHash":"aPvRfnfuxQHGd8zDHCYPMO6O0EnIk5AOCEuot/dWnOY=","PubKey":"1ec94e497abe0fbb87f9ed2843e21163e17e3e97f6bbbae7a88399b826474f93"}', '', '', 1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (3, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce7736d6-ed20-11ec-8b10-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":1,"AccountIndex":2,"AccountName":"sher.legend","AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","PubKey":"b0b6f7466154578ec66d51a335ead65ffd6a7210567fad9e68b6df8a5ce5dd85"}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (4, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b10-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":1,"AccountIndex":3,"AccountName":"gavin.legend","AccountNameHash":"9KZJFrMtD0ZzaZct0Vb30r2FnAoQijs5WiUPGU9GgKM=","PubKey":"0500ccea3ca064968f5292b850ac8d4d3ee48d499357351a5ebfa2f30bb6070e"}', '', '', 3, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (5, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b11-988fe0603efa', 4, 0, '0', -1, -1, 0, '100000000000000000', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (6, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b12-988fe0603efa', 4, 0, '0', -1, -1, 0, '100000000000000000', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":4,"AccountIndex":3,"AccountNameHash":"9KZJFrMtD0ZzaZct0Vb30r2FnAoQijs5WiUPGU9GgKM=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 3, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (7, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b13-988fe0603efa', 4, 0, '0', -1, -1, 1, '100000000000000000000', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (8, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b14-988fe0603efa', 4, 0, '0', -1, -1, 2, '100000000000000000000', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":2,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (9, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b15-988fe0603efa', 2, 0, '0', -1, 0, 0, '0', '0', '{"TxType":2,"PairIndex":0,"AssetAId":0,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (10, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b16-988fe0603efa', 2, 0, '0', -1, 1, 0, '0', '0', '{"TxType":2,"PairIndex":1,"AssetAId":0,"AssetBId":1,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (11, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b17-988fe0603efa', 2, 0, '0', -1, 2, 0, '0', '0', '{"TxType":2,"PairIndex":2,"AssetAId":1,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (12, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce774ff9-ed20-11ec-8b18-988fe0603efa', 3, 0, '0', -1, 1, 0, '0', '0', '{"TxType":3,"PairIndex":1,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', '', '', -1, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (13, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce77cdf5-ed20-11ec-8b18-988fe0603efa', 5, 0, '0', 0, -1, 0, '0', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":5,"AccountIndex":2,"NftIndex":0,"NftL1Address":"0x464ed8Ce7076Abaf743F760468230B9d71fB7D90","CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftContentHash":"j6MFmnxo2t3N+cA7HNHm0DQrfEqQ7WEDcsaBv+p+5Hg=","NftL1TokenId":0,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","CollectionId":0}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (14, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce77cdf5-ed20-11ec-8b19-988fe0603efa', 17, 0, '0', -1, -1, 1, '100000000000000000000', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":17,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, -1, 0);
INSERT INTO "public"."mempool_tx" VALUES (15, '2022-06-16 03:02:50.039037+00', '2022-06-16 03:02:50.039037+00', NULL, 'ce77cdf5-ed20-11ec-8b1a-988fe0603efa', 18, 0, '0', 0, -1, 0, '0', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', '{"TxType":18,"AccountIndex":2,"CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftIndex":0,"CollectionId":0,"NftL1Address":"0x464ed8Ce7076Abaf743F760468230B9d71fB7D90","AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","CreatorAccountNameHash":"AA==","NftContentHash":"j6MFmnxo2t3N+cA7HNHm0DQrfEqQ7WEDcsaBv+p+5Hg=","NftL1TokenId":0}', '', '', 2, 0, 0, -1, 0);

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
INSERT INTO "public"."mempool_tx_detail" VALUES (1, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 5, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (2, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 6, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (3, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 7, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (4, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 8, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (5, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 9, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (6, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 10, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (7, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 11, 2, 2, -1, '', '{"PairIndex":2,"AssetAId":1,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (8, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 12, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (9, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 13, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (10, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 13, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":2,"NftContentHash":"8fa3059a7c68daddcdf9c03b1cd1e6d0342b7c4a90ed610372c681bfea7ee478","NftL1TokenId":"0","NftL1Address":"0x464ed8Ce7076Abaf743F760468230B9d71fB7D90","CreatorTreasuryRate":0,"CollectionId":0}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (11, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 14, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (12, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 15, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (13, '2022-06-16 03:02:50.041786+00', '2022-06-16 03:02:50.041786+00', NULL, 15, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1);

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
INSERT INTO "public"."sys_config" VALUES (1, '2022-06-16 03:01:29.411105+00', '2022-06-16 03:01:29.411105+00', NULL, 'SysGasFee', '1', 'float', 'based on ETH');
INSERT INTO "public"."sys_config" VALUES (2, '2022-06-16 03:01:29.411105+00', '2022-06-16 03:01:29.411105+00', NULL, 'MaxAssetId', '9', 'int', 'max number of asset id');
INSERT INTO "public"."sys_config" VALUES (3, '2022-06-16 03:01:29.411105+00', '2022-06-16 03:01:29.411105+00', NULL, 'TreasuryAccountIndex', '0', 'int', 'treasury index');
INSERT INTO "public"."sys_config" VALUES (4, '2022-06-16 03:01:29.411105+00', '2022-06-16 03:01:29.411105+00', NULL, 'GasAccountIndex', '1', 'int', 'gas index');
INSERT INTO "public"."sys_config" VALUES (5, '2022-06-16 03:01:29.411105+00', '2022-06-16 03:01:29.411105+00', NULL, 'ZkbasContract', '0x045A98016DF9C1790caD1be1c4d69ba1fd2aB9d9', 'string', 'Zecrey contract on BSC');
INSERT INTO "public"."sys_config" VALUES (6, '2022-06-16 03:01:29.411105+00', '2022-06-16 03:01:29.411105+00', NULL, 'GovernanceContract', '0x45E486062b952225c97621567fCdD29eCE730B87', 'string', 'Governance contract on BSC');
INSERT INTO "public"."sys_config" VALUES (7, '2022-06-16 03:01:29.411105+00', '2022-06-16 03:01:29.411105+00', NULL, 'BscTestNetworkRpc', 'http://tf-dex-preview-validator-nlb-6fd109ac8b9d390a.elb.ap-northeast-1.amazonaws.com:8545', 'string', 'BSC network rpc');
INSERT INTO "public"."sys_config" VALUES (8, '2022-06-16 03:01:29.411105+00', '2022-06-16 03:01:29.411105+00', NULL, 'Local_Test_Network_RPC', 'http://127.0.0.1:8545/', 'string', 'Local network rpc');
INSERT INTO "public"."sys_config" VALUES (9, '2022-06-16 03:02:04.098828+00', '2022-06-16 03:02:04.098828+00', NULL, 'AssetGovernanceContract', '0x74ad9cd2e0656C49B3DB427a9aF8AC704C71DBbC', 'string', 'asset governance contract');
INSERT INTO "public"."sys_config" VALUES (10, '2022-06-16 03:02:04.098828+00', '2022-06-16 03:02:04.098828+00', NULL, 'Validators', '{"0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57":{"Address":"0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57","IsActive":true}}', 'map[string]*ValidatorInfo', 'validator info');
INSERT INTO "public"."sys_config" VALUES (11, '2022-06-16 03:02:04.098828+00', '2022-06-16 03:02:04.098828+00', NULL, 'Governor', '0x7dD2Ac589eFCC8888474d95Cb4b084CCa2d8aA57', 'string', 'governor');

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
ALTER SEQUENCE "public"."asset_info_id_seq"
OWNED BY "public"."asset_info"."id";
SELECT setval('"public"."asset_info_id_seq"', 3, true);

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
SELECT setval('"public"."l1_block_monitor_id_seq"', 2, true);

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
-- Indexes structure for table asset_info
-- ----------------------------
CREATE UNIQUE INDEX "idx_asset_info_asset_id" ON "public"."asset_info" USING btree (
  "asset_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "idx_asset_info_deleted_at" ON "public"."asset_info" USING btree (
  "deleted_at" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table asset_info
-- ----------------------------
ALTER TABLE "public"."asset_info" ADD CONSTRAINT "asset_info_pkey" PRIMARY KEY ("id");

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
