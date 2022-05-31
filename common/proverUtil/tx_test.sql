/*
 Navicat Premium Data Transfer

 Source Server         : local-zecrey
 Source Server Type    : PostgreSQL
 Source Server Version : 140001
 Source Host           : localhost:5432
 Source Catalog        : zecreyLegend
 Source Schema         : public

 Target Server Type    : PostgreSQL
 Target Server Version : 140001
 File Encoding         : 65001

 Date: 30/05/2022 13:45:38
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
INSERT INTO "public"."account" VALUES (1, '0001-01-01 00:00:00+00', '2022-05-30 05:13:27.184024+00', NULL, 0, 'treasury.legend', '412805eb224e8c10de9ee037f55c92f32266f057fad3279cf4bab0a49d8f4080', 'c0d201aace9a2c17ce7066dc6ffefaf7930f1317c4c95d0661b164a1c584d676', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 0, 0, '{"0":{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 1);
INSERT INTO "public"."account" VALUES (2, '0001-01-01 00:00:00+00', '2022-05-30 05:13:27.301124+00', NULL, 1, 'gas.legend', '53aa127ef258d5311bb9d8736d087e1c81204d356f876e7c42c42befcd679827', '68fbd17e77eec501c677ccc31c260f30ee8ed049c893900e084ba8b7f7569ce6', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 0, 0, '{"0":{"AssetId":0,"Balance":20200,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":35000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '219d2d2c0bb8cba744ec53ea8388da6c961b555f62bd5aa290e97109d186c467', 1);
INSERT INTO "public"."account" VALUES (4, '0001-01-01 00:00:00+00', '2022-05-30 05:13:27.298799+00', NULL, 3, 'gavin.legend', 'c9e9ccb618f4825496506f70551d725dec7aeb2e3f31da262ea45ab88a174909', 'f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a3', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 2, 0, '{"0":{"AssetId":0,"Balance":100000000000080000,"LpAmount":0,"OfferCanceledOrFinalized":1},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '20e11089ec56b54159ea65fc328d75c7b15011b11f5c73653073ddd0bdf1423e', 1);
INSERT INTO "public"."account" VALUES (3, '0001-01-01 00:00:00+00', '2022-05-30 05:13:27.299881+00', NULL, 2, 'sher.legend', '7f70064f2c485996dc2acb397d0b4fe63eec854aad09b6fd3c41549e6d046586', '04b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 9, 1, '{"0":{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":3},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999865000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135f35977f2abf9cb4029cc418b45ba79cc45cf39685be661c67da75ead45d9a', 1);

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
INSERT INTO "public"."account_history" VALUES (1, '2022-05-30 05:11:11.046236+00', '2022-05-30 05:11:11.046236+00', NULL, 0, 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 1);
INSERT INTO "public"."account_history" VALUES (2, '2022-05-30 05:11:11.067729+00', '2022-05-30 05:11:11.067729+00', NULL, 1, 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 2);
INSERT INTO "public"."account_history" VALUES (3, '2022-05-30 05:11:11.082503+00', '2022-05-30 05:11:11.082503+00', NULL, 2, 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 3);
INSERT INTO "public"."account_history" VALUES (4, '2022-05-30 05:11:11.093825+00', '2022-05-30 05:11:11.093825+00', NULL, 3, 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 4);
INSERT INTO "public"."account_history" VALUES (5, '2022-05-30 05:11:11.108998+00', '2022-05-30 05:11:11.108998+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '06ce582922720755debe04d60415a9c28bc4e788d012d3ea1700549f0e190c9a', 5);
INSERT INTO "public"."account_history" VALUES (6, '2022-05-30 05:11:11.119292+00', '2022-05-30 05:11:11.119292+00', NULL, 3, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '06ce582922720755debe04d60415a9c28bc4e788d012d3ea1700549f0e190c9a', 6);
INSERT INTO "public"."account_history" VALUES (7, '2022-05-30 05:11:11.130033+00', '2022-05-30 05:11:11.130033+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '069e6e659595ff010898f90e61614c5a8d77de0d2984715be6f5b3f8505ae10c', 7);
INSERT INTO "public"."account_history" VALUES (8, '2022-05-30 05:11:11.140185+00', '2022-05-30 05:11:11.140185+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '17d8b1c33a32922ce0838eed568beb32728b6271d97e008e481edd92cca55f08', 8);
INSERT INTO "public"."account_history" VALUES (9, '2022-05-30 05:11:11.20751+00', '2022-05-30 05:11:11.20751+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '17d8b1c33a32922ce0838eed568beb32728b6271d97e008e481edd92cca55f08', 13);
INSERT INTO "public"."account_history" VALUES (10, '2022-05-30 05:11:11.220507+00', '2022-05-30 05:11:11.220507+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '09fe19fc526b3e67753a6d91cc709feb45f6b281f6a3a71773a0abebe50f517f', 14);
INSERT INTO "public"."account_history" VALUES (11, '2022-05-30 05:11:11.233488+00', '2022-05-30 05:11:11.233488+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '09fe19fc526b3e67753a6d91cc709feb45f6b281f6a3a71773a0abebe50f517f', 15);
INSERT INTO "public"."account_history" VALUES (12, '2022-05-30 05:13:27.071273+00', '2022-05-30 05:13:27.071273+00', NULL, 2, 1, 0, '{"0":{"AssetId":0,"Balance":99999999999900000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999995000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1cd1016e23d9e514928a567cc8a4cddcce67b01e817021eb916edaac7e166242', 16);
INSERT INTO "public"."account_history" VALUES (13, '2022-05-30 05:13:27.071273+00', '2022-05-30 05:13:27.071273+00', NULL, 3, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000100000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '14012ade6c7b76679cc709bbc6fe865ec94f0b55a7c976a9b93d8e214f2bf5e5', 16);
INSERT INTO "public"."account_history" VALUES (14, '2022-05-30 05:13:27.071273+00', '2022-05-30 05:13:27.071273+00', NULL, 1, 0, 0, '{"2":{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '08e7c9a1858f6ad9986887426fdddc7231a93a806c81b5841171ec5cb834eabe', 16);
INSERT INTO "public"."account_history" VALUES (15, '2022-05-30 05:13:27.125544+00', '2022-05-30 05:13:27.125544+00', NULL, 1, 0, 0, '{"2":{"AssetId":2,"Balance":10000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1e6cf281636a0d207da108b38aaada12c903f0f7531b3e60ff935675b9d64644', 17);
INSERT INTO "public"."account_history" VALUES (16, '2022-05-30 05:13:27.125544+00', '2022-05-30 05:13:27.125544+00', NULL, 2, 2, 0, '{"0":{"AssetId":0,"Balance":99999999989900000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999990000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '25cc5a90b005abb6c7c0d5d1fbd34907b12c70b7d2f11a6901cd5622186e584e', 17);
INSERT INTO "public"."account_history" VALUES (17, '2022-05-30 05:13:27.149317+00', '2022-05-30 05:13:27.149317+00', NULL, 2, 3, 0, '{"0":{"AssetId":0,"Balance":99999999989800000,"LpAmount":100000,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999885000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '3046d6422f86f1ab6e9cbe2a0e449604df61bbdd3f3199fadd7a5bc4f046d289', 18);
INSERT INTO "public"."account_history" VALUES (18, '2022-05-30 05:13:27.149317+00', '2022-05-30 05:13:27.149317+00', NULL, 0, 0, 0, '{"0":{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 18);
INSERT INTO "public"."account_history" VALUES (19, '2022-05-30 05:13:27.149317+00', '2022-05-30 05:13:27.149317+00', NULL, 1, 0, 0, '{"2":{"AssetId":2,"Balance":15000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '12aeb69e38371c4ef60475f6d1d5bd15fb602de9a3ac9d8ce98cb11b95685bee', 18);
INSERT INTO "public"."account_history" VALUES (20, '2022-05-30 05:13:27.16853+00', '2022-05-30 05:13:27.16853+00', NULL, 2, 4, 0, '{"0":{"AssetId":0,"Balance":99999999989795099,"LpAmount":100000,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999884900,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '136b21d7d137ada052e748f45719da132e2a344e5ae6fe334d0b67012c331d6d', 19);
INSERT INTO "public"."account_history" VALUES (21, '2022-05-30 05:13:27.16853+00', '2022-05-30 05:13:27.16853+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":15000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '0a8e87b9a27934661653c3d37ea4b6b9cb7257d23d0ec85e0a77b0c62f6ca453', 19);
INSERT INTO "public"."account_history" VALUES (22, '2022-05-30 05:13:27.187348+00', '2022-05-30 05:13:27.187348+00', NULL, 2, 5, 0, '{"0":{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999880000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '13f7575c9228694a34eaec2e080115ac5cc06ec248abb0e5f9cdb7151b9acedd', 20);
INSERT INTO "public"."account_history" VALUES (23, '2022-05-30 05:13:27.187348+00', '2022-05-30 05:13:27.187348+00', NULL, 0, 0, 0, '{"0":{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 20);
INSERT INTO "public"."account_history" VALUES (24, '2022-05-30 05:13:27.187348+00', '2022-05-30 05:13:27.187348+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":20000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1b0c02b49e7d799975e98665fc0f2062251e7e295001f43b0fc5013360d9f3cf', 20);
INSERT INTO "public"."account_history" VALUES (25, '2022-05-30 05:13:27.202898+00', '2022-05-30 05:13:27.202898+00', NULL, 2, 6, 1, '{"0":{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999875000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2e7c81f2815d8f11d39097bb1a5eb9f10d3ba9b7b56c0ac0f7a49c3eba397579', 21);
INSERT INTO "public"."account_history" VALUES (26, '2022-05-30 05:13:27.202898+00', '2022-05-30 05:13:27.202898+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":25000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '02f3efe09b203142ae196d4555d58e060da742fa15d9213fe26e75d8c5505539', 21);
INSERT INTO "public"."account_history" VALUES (27, '2022-05-30 05:13:27.22281+00', '2022-05-30 05:13:27.22281+00', NULL, 2, 7, 1, '{"0":{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999870000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135175067e21e4d0a1ec1f01d1eaacbb65a0ec3df762bf586df1c49a3a554e6d', 22);
INSERT INTO "public"."account_history" VALUES (28, '2022-05-30 05:13:27.22281+00', '2022-05-30 05:13:27.22281+00', NULL, 3, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000100000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '14012ade6c7b76679cc709bbc6fe865ec94f0b55a7c976a9b93d8e214f2bf5e5', 22);
INSERT INTO "public"."account_history" VALUES (29, '2022-05-30 05:13:27.22281+00', '2022-05-30 05:13:27.22281+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":30000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2b264f5337dc9d06629ff7099ad6e0653eb3cdf7056dd2cd46752d50c1050b93', 22);
INSERT INTO "public"."account_history" VALUES (30, '2022-05-30 05:13:27.24559+00', '2022-05-30 05:13:27.24559+00', NULL, 3, 1, 0, '{"0":{"AssetId":0,"Balance":100000000000095000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2e2137673dbe998c6dce6b1555760686b57af98a8c23820337ef881703f534d2', 23);
INSERT INTO "public"."account_history" VALUES (31, '2022-05-30 05:13:27.24559+00', '2022-05-30 05:13:27.24559+00', NULL, 2, 7, 1, '{"0":{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999870000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135175067e21e4d0a1ec1f01d1eaacbb65a0ec3df762bf586df1c49a3a554e6d', 23);
INSERT INTO "public"."account_history" VALUES (32, '2022-05-30 05:13:27.24559+00', '2022-05-30 05:13:27.24559+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":10000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":30000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '0ade8414224cd97b2841b34519db998c7873d2c386e87fcc93ac94f056424b9a', 23);
INSERT INTO "public"."account_history" VALUES (33, '2022-05-30 05:13:27.269673+00', '2022-05-30 05:13:27.269673+00', NULL, 2, 8, 1, '{"0":{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":1},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999870000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1a9e697d30b8b2f43aedc7251497f4a4308f221089f43c8794d8966d5bbf4769', 24);
INSERT INTO "public"."account_history" VALUES (34, '2022-05-30 05:13:27.269673+00', '2022-05-30 05:13:27.269673+00', NULL, 3, 1, 0, '{"0":{"AssetId":0,"Balance":100000000000085000,"LpAmount":0,"OfferCanceledOrFinalized":1},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '06d9b9fd9b2b3ab3ea7c44de16deef25a7dea1314ec20c5b92dcc1049d221f49', 24);
INSERT INTO "public"."account_history" VALUES (35, '2022-05-30 05:13:27.269673+00', '2022-05-30 05:13:27.269673+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":15200,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":30000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2118e6f94c540e6f3676a6cbae245679a59432115d48f7571a6ae5565edca611', 24);
INSERT INTO "public"."account_history" VALUES (36, '2022-05-30 05:13:27.287005+00', '2022-05-30 05:13:27.287005+00', NULL, 2, 9, 1, '{"0":{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":3},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999865000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135f35977f2abf9cb4029cc418b45ba79cc45cf39685be661c67da75ead45d9a', 25);
INSERT INTO "public"."account_history" VALUES (37, '2022-05-30 05:13:27.287005+00', '2022-05-30 05:13:27.287005+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":15200,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":35000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1a94962e42cbd751dc8fb4975ab18ee52493c4fd400d94fb1065719e24d019f3', 25);
INSERT INTO "public"."account_history" VALUES (38, '2022-05-30 05:13:27.302166+00', '2022-05-30 05:13:27.302166+00', NULL, 3, 2, 0, '{"0":{"AssetId":0,"Balance":100000000000080000,"LpAmount":0,"OfferCanceledOrFinalized":1},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '20e11089ec56b54159ea65fc328d75c7b15011b11f5c73653073ddd0bdf1423e', 26);
INSERT INTO "public"."account_history" VALUES (39, '2022-05-30 05:13:27.302166+00', '2022-05-30 05:13:27.302166+00', NULL, 2, 9, 1, '{"0":{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":3},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999865000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135f35977f2abf9cb4029cc418b45ba79cc45cf39685be661c67da75ead45d9a', 26);
INSERT INTO "public"."account_history" VALUES (40, '2022-05-30 05:13:27.302166+00', '2022-05-30 05:13:27.302166+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":20200,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":35000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '219d2d2c0bb8cba744ec53ea8388da6c961b555f62bd5aa290e97109d186c467', 26);

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
  "account_root" text COLLATE "pg_catalog"."default",
  "priority_operations" int8,
  "pending_onchain_operations_hash" text COLLATE "pg_catalog"."default",
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
INSERT INTO "public"."block" VALUES (1, '2022-05-30 05:07:44.336761+00', '2022-05-30 05:07:44.336761+00', NULL, '0000000000000000000000000000000000000000000000000000000000000000', 0, '2cddfbdf6acee742cb4d86d2bd482ea953bbe565c994ef61ea53e55511a1b4bf', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 3);
INSERT INTO "public"."block" VALUES (2, '2022-05-30 05:11:11.032+00', '2022-05-30 05:11:11.036549+00', NULL, '051f81ae0ab58ceaac24eb32499c26bfd1bf96bd3d6055450c89248011a32483', 1, '0896c0f1c9e8fbc0b71870e545386ef20b89e546d1abb8d10a0235bead6b7364', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (3, '2022-05-30 05:11:11.058+00', '2022-05-30 05:11:11.062381+00', NULL, '086c1e56a31ee7a99473e4c94c1ce8244173800dbf0f54e8724a7db865bdaad2', 2, '141d76c64d7ea9aae087ebcd990e116e0e918044f3f1d4e75d3e32ec01f8d209', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (4, '2022-05-30 05:11:11.069+00', '2022-05-30 05:11:11.076078+00', NULL, '17e4bc4241338c568f237c6d727ea817b0489b0d7503cc6512a246db2848db9c', 3, '2e378e466dab807e2d769076b5e16a6d3e043fbc01f5c526d239c12f3cf9fa2c', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (5, '2022-05-30 05:11:11.085+00', '2022-05-30 05:11:11.088474+00', NULL, '19db28dcfb4b906cba5502ac18877c39ffef7bc9d9633ba60970eb63493b3554', 4, '2efa2a41c1168b8aecd5f5d1d67c0a442df6bd8fa37e589757aada6841d6ebf2', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (6, '2022-05-30 05:11:11.096+00', '2022-05-30 05:11:11.098169+00', NULL, '28410e9868d0907d2e392d4bf46babf6654ba5c786f9613222ef575f997cc2e2', 5, '06f601cc791636973388cea255953926a1ac70107a6230b77945cf74fe46f9f6', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (7, '2022-05-30 05:11:11.111+00', '2022-05-30 05:11:11.112725+00', NULL, '145f59d864a6c01807e20673413657d6afdeedb592f72d468ef0f4015ec74662', 6, '0e38c1b0f794dcd7ddd947fe82e5f75efaaae7747459de52c94d92f9df0b64ef', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (8, '2022-05-30 05:11:11.121+00', '2022-05-30 05:11:11.123558+00', NULL, '202154c2763e404f26afdbd1dcaab2e2149b174b6c56fdfcd1a7d241c5fff505', 7, '1f7b6a15037e6b1fb1688221ad16f5cbe0f7f2cdc030aacb3c37d1bec2813fc3', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (9, '2022-05-30 05:11:11.132+00', '2022-05-30 05:11:11.134349+00', NULL, '0627ebeda171c9bf1df7406e3f1a7bfdb1474ef344007e3300d5637d369fdbff', 8, '2e0bc9bad4f2cf0dbd5b2a3570879832ef2579c975b1dcb43840cdc523914d80', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (10, '2022-05-30 05:11:11.142+00', '2022-05-30 05:11:11.14548+00', NULL, '20cfcf9d5aadbf10986016c78f3b7c9d30fbde3e8bb4e7e23f316b1a2a9161ab', 9, '228199bd4964493fd18425a3ba2bfed71160c8302232c3c8c693512a76626779', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (11, '2022-05-30 05:11:11.159+00', '2022-05-30 05:11:11.161633+00', NULL, '253a6c185573728b582a4e7364f8ad7408fee035e520fe15676158bfa44804ff', 10, '00fdfc52d138f9ba5ce9611c805426b95616f518732bcfc8e40a2a2019e15287', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (12, '2022-05-30 05:11:11.17+00', '2022-05-30 05:11:11.17229+00', NULL, '01bc7baa58a4b7d7047e775a4eeb5422b74f396ef9eb08c3f1cb2f254348b799', 11, '0cec93680cf90d14634ce669e90b157fe2029d18ccecc856a07c3b1b8ecb7b23', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (13, '2022-05-30 05:11:11.182+00', '2022-05-30 05:11:11.183735+00', NULL, '17df73a817530bbc07731a2795ac3f71a8f8d9f4cc3d303821b41bb63f91fe7f', 12, '096665c5754506b34a69d625bbaec71f137e0d397c480dc488aa067187dd0db7', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (14, '2022-05-30 05:11:11.192+00', '2022-05-30 05:11:11.199476+00', NULL, '2e5bd8dff4d283b8af03bc1dd9a07a52ca7977c91ce2a732654fcf5bf5051ca3', 13, '00f39cb5e4b0361d022655c730f402936bd3459583134c17aa9d9580e97d66b2', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (15, '2022-05-30 05:11:11.213+00', '2022-05-30 05:11:11.215158+00', NULL, '1352afd3407282891d9d0a020e2049add0d26647f349e7ac3a6c35367c170eee', 14, '09cf39a8a8d7787da03f683c147cef5939704f3ead516e8a2a157ec8fe27cd94', 1, '0ae215a10fae016a8f33b3394ce8e8ead69c527f8262a8c628684e6d438ac624', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (16, '2022-05-30 05:11:11.222+00', '2022-05-30 05:11:11.225421+00', NULL, '14cc16124b68b69c70ed7064f8648308a56f9a9e975cce928ed49f56fc67a484', 15, '0be5445077dd00cddefcb6539fb91f344fcdf62c2ab4be11e652b6cc88c0e9f9', 1, '508d199df647c54ee03c53bbc201a3b3588b82982a68de1eea83a6cbc7521b6a', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (17, '2022-05-30 05:13:27.05+00', '2022-05-30 05:13:27.057345+00', NULL, '24da5f0ebd2cd3325962eeb8a5cc010038ee13c282a8dc0920ea0d3f933c8609', 16, '18ab34ac1cb16907fe00e43f9adc4342d0d68e39b79767cba0e7947f6a373c35', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (18, '2022-05-30 05:13:27.112+00', '2022-05-30 05:13:27.115288+00', NULL, '02ca98084cdf6b809a90fb26af7e455d1464fe9745371a53d34b80b17cc8053d', 17, '12ff91cd0cab6f32781bef0fb9e12a16f14380860e4bda177d9c4c0de34239a6', 0, '79df78b2014d4c7930df848eb80ea9e16db3851cc11a06cde22bd4446160d127', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (19, '2022-05-30 05:13:27.13+00', '2022-05-30 05:13:27.137552+00', NULL, '1f6d1a842035a5c6696753f5920e70969dce34a29092082d15ffb9055f0edba2', 18, '0502e76193590e0737cdd109bd000898ea41d23c4d2b7bb6fff773df7214e8e6', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (20, '2022-05-30 05:13:27.156+00', '2022-05-30 05:13:27.159828+00', NULL, '1e89f07d2c82da7b4b1d1eb94f7c7bf7675445a767cac08e3d14e37f10e627bc', 19, '08088d6698c046bec4fa2371bc023831c1ead344ca5f11b0c85084a0ea208fd2', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (21, '2022-05-30 05:13:27.174+00', '2022-05-30 05:13:27.178877+00', NULL, '1f1f0532a96e9bd1690c92a9b0cc49fc6f06677b75a43a36978fc7f607227d04', 20, '23ac82db1260a22ae3fb448ac3c06b6dacd76ceb24a587d2e9e6256bdcafaf5c', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (22, '2022-05-30 05:13:27.192+00', '2022-05-30 05:13:27.195352+00', NULL, '0900445273b91508ddb52b47e0460ff6533fc7fe8e0102f3803acb59e0d499a5', 21, '18016dc2d985df4327b74915a4e595d7770abd84a4d2ba02a101662d683209a0', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (23, '2022-05-30 05:13:27.205+00', '2022-05-30 05:13:27.211952+00', NULL, '1c7b433e1defb4fd0725ce39d7cb254587f11729289ecf03865d9d99d59a6cf9', 22, '1c2f58272ada607379fd722177a3a455c8cfeb6965b9706d00f7bfd003ef71f1', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (24, '2022-05-30 05:13:27.229+00', '2022-05-30 05:13:27.233854+00', NULL, '07479b62d7e7c613a75b5764a7544ef93b960ed4992cf86407a025b58db35f71', 23, '2e0f826aba027333a11b3a355166e9080663595eacc68727bc0d8fd9d46e8d4b', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (25, '2022-05-30 05:13:27.252+00', '2022-05-30 05:13:27.257613+00', NULL, '107bafffd78fb760d77b0445c8f0341a2d6c164a6181187fceb5096a6c0d8ad6', 24, '050556f4ee047254af35995b4d2fbbf6d553d5119931e646b53512d368e12d1d', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (26, '2022-05-30 05:13:27.275+00', '2022-05-30 05:13:27.278893+00', NULL, '11f113261057acdedb4181ad19602b2e15f7f34da4c55e41335c7d5b7f5dcf54', 25, '290c51f8cb09e7a32efd05f623b0181427a0357630f31c86e6208ddcfdb4251c', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (27, '2022-05-30 05:13:27.289+00', '2022-05-30 05:13:27.293952+00', NULL, '1e9b0ccd97ca02a2ef55f50b4bf0e581a141673fe9d51c25085431afc214fd7c', 26, '1a8a0465a857f6b1278ef72b0e42637b7febb0d32072294348ff27392335ac1c', 0, 'a232f9dd8e3aac1fed2e18d14441e881faabded73817d60ea452607ed85c16cd', '', 0, '', 0, 1);

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
  "block_info" text COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Records of l1_block_monitor
-- ----------------------------
INSERT INTO "public"."l1_block_monitor" VALUES (1, '2022-05-30 05:10:23.005927+00', '2022-05-30 05:10:23.005927+00', NULL, 582, '[{"EventType":0,"TxHash":"0xe589a46bf698d79a5d4aa961efec3099fce2a8c2c9e4e929fbaea865f246187d"},{"EventType":0,"TxHash":"0xa4a9c268c4bb38b5e2a84ecabdfa00caf70207472f4f56c117ce0a8e86a24d08"},{"EventType":0,"TxHash":"0xb40a64aba50c256aca39058dfff6a4f5a4933d9c3d2cf2e983fdfacd76080d61"},{"EventType":0,"TxHash":"0xd6c4c8d06bc03f7c33cac356e1ea4c2bf0383f059e020b7dc9d6a8c3281c8c1e"},{"EventType":0,"TxHash":"0xabaa1263b912f69324269e087036f961e09d8dcc42107766186b1ce8f4838a8c"},{"EventType":0,"TxHash":"0xf0da83de6b12251c248276fd2aca7874df535564364631730b195bf3a39cccaa"},{"EventType":0,"TxHash":"0xf6ef1adb8e3076438af6fbbb8971719157b90e610bcaaef29a22517fe67d8418"},{"EventType":0,"TxHash":"0x3b9930b1cac58c31252ebb4331e2881f3d9d47c94df2c17a4cd7a30792f1d970"},{"EventType":0,"TxHash":"0xe726c8ac58a27de860d0fc415b8f504e39a70994a6cb7f09e7eec4a1253be3e2"},{"EventType":0,"TxHash":"0x9e884269087bdab88ede8f4ae0dce1fef838a8da0c55c3bdd3debb9674688754"},{"EventType":0,"TxHash":"0xa8970751d98d865ac01f2484c117ca8df20ca4f99646016dd3a02bb1075a4f3a"},{"EventType":0,"TxHash":"0xa0b6475c914331371534674d19aace5f1d5c8e662a9b75b53aa1870721309a50"},{"EventType":0,"TxHash":"0x78ccc749096a1702fc9bf4b67c8165a34f16da93ba22cb4cbfe91e69d32aba04"},{"EventType":0,"TxHash":"0xaaaec6117af4fce5324dd6dee37a9c357872a9375e32a801a55cac52f74a8edf"},{"EventType":0,"TxHash":"0x9f09994bda2e87e315c8723870d684782f0d4070a6d494464170d5adb314d2ff"}]');

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
  "asset_name" text COLLATE "pg_catalog"."default",
  "asset_symbol" text COLLATE "pg_catalog"."default",
  "decimals" int8,
  "status" int8
)
;

-- ----------------------------
-- Records of l2_asset_info
-- ----------------------------
INSERT INTO "public"."l2_asset_info" VALUES (1, '2022-05-30 05:07:44.32558+00', '2022-05-30 05:07:44.32558+00', NULL, 0, 'BNB', 'BNB', 18, 0);
INSERT INTO "public"."l2_asset_info" VALUES (2, '2022-05-30 05:07:44.32558+00', '2022-05-30 05:07:44.32558+00', NULL, 1, 'LEG', 'LEG', 18, 0);
INSERT INTO "public"."l2_asset_info" VALUES (3, '2022-05-30 05:07:44.32558+00', '2022-05-30 05:07:44.32558+00', NULL, 2, 'REY', 'REY', 18, 0);

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
INSERT INTO "public"."l2_nft" VALUES (1, '2022-05-30 05:10:50.080236+00', '2022-05-30 05:11:11.238808+00', NULL, 0, 0, 0, '0', '0', '0', 0, 0);
INSERT INTO "public"."l2_nft" VALUES (2, '2022-05-30 05:12:26.764311+00', '2022-05-30 05:13:27.305909+00', NULL, 1, 0, 0, '0', '0', '0', 0, 0);

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
INSERT INTO "public"."l2_nft_exchange" VALUES (1, '2022-05-30 05:12:46.779979+00', '2022-05-30 05:12:46.779979+00', NULL, 3, 2, 1, 0, '10000');

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
INSERT INTO "public"."l2_nft_history" VALUES (1, '2022-05-30 05:11:11.209203+00', '2022-05-30 05:11:11.209203+00', NULL, 0, 0, 2, '0c1293246402e53a2701d0d764bd3ce7ac9a2b82c8a1cd3c6c1d1ef3ec8076f2', '0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE', '0', 0, 0, 0, 13);
INSERT INTO "public"."l2_nft_history" VALUES (2, '2022-05-30 05:11:11.2427+00', '2022-05-30 05:11:11.2427+00', NULL, 0, 0, 0, '0', '0', '0', 0, 0, 0, 15);
INSERT INTO "public"."l2_nft_history" VALUES (3, '2022-05-30 05:13:27.22617+00', '2022-05-30 05:13:27.22617+00', NULL, 1, 2, 3, '09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1', '0', '0', 0, 1, 0, 22);
INSERT INTO "public"."l2_nft_history" VALUES (4, '2022-05-30 05:13:27.249485+00', '2022-05-30 05:13:27.249485+00', NULL, 1, 2, 2, '09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1', '0', '0', 0, 1, 0, 23);
INSERT INTO "public"."l2_nft_history" VALUES (5, '2022-05-30 05:13:27.272912+00', '2022-05-30 05:13:27.272912+00', NULL, 1, 2, 3, '09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1', '0', '0', 0, 1, 0, 24);
INSERT INTO "public"."l2_nft_history" VALUES (6, '2022-05-30 05:13:27.308187+00', '2022-05-30 05:13:27.308187+00', NULL, 1, 0, 0, '0', '0', '0', 0, 0, 0, 26);

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
INSERT INTO "public"."l2_nft_withdraw_history" VALUES (1, '2022-05-30 05:11:11.235094+00', '2022-05-30 05:11:11.235094+00', NULL, 0, 0, 2, '0c1293246402e53a2701d0d764bd3ce7ac9a2b82c8a1cd3c6c1d1ef3ec8076f2', '0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE', '0', 0, 0);
INSERT INTO "public"."l2_nft_withdraw_history" VALUES (2, '2022-05-30 05:13:27.304307+00', '2022-05-30 05:13:27.304307+00', NULL, 1, 2, 3, '09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1', '0', '0', 0, 1);

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
INSERT INTO "public"."l2_tx_event_monitor" VALUES (1, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.082967+00', NULL, '0xe589a46bf698d79a5d4aa961efec3099fce2a8c2c9e4e929fbaea865f246187d', 559, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 0, 1, '01000000007472656173757279000000000000000000000000000000000000000000000000c0d201aace9a2c17ce7066dc6ffefaf7930f1317c4c95d0661b164a1c584d676412805eb224e8c10de9ee037f55c92f32266f057fad3279cf4bab0a49d8f4080', 40879, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (2, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.084589+00', NULL, '0xa4a9c268c4bb38b5e2a84ecabdfa00caf70207472f4f56c117ce0a8e86a24d08', 560, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 1, 1, '0100000000676173000000000000000000000000000000000000000000000000000000000068fbd17e77eec501c677ccc31c260f30ee8ed049c893900e084ba8b7f7569ce653aa127ef258d5311bb9d8736d087e1c81204d356f876e7c42c42befcd679827', 40880, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (3, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.085115+00', NULL, '0xb40a64aba50c256aca39058dfff6a4f5a4933d9c3d2cf2e983fdfacd76080d61', 561, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 2, 1, '0100000000736865720000000000000000000000000000000000000000000000000000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f7f70064f2c485996dc2acb397d0b4fe63eec854aad09b6fd3c41549e6d046586', 40881, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (4, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.086181+00', NULL, '0xd6c4c8d06bc03f7c33cac356e1ea4c2bf0383f059e020b7dc9d6a8c3281c8c1e', 562, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 3, 1, '0100000000676176696e000000000000000000000000000000000000000000000000000000f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a3c9e9ccb618f4825496506f70551d725dec7aeb2e3f31da262ea45ab88a174909', 40882, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (5, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.086718+00', NULL, '0xabaa1263b912f69324269e087036f961e09d8dcc42107766186b1ce8f4838a8c', 564, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 4, 4, '040000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f00000000000000000000016345785d8a0000', 40884, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (6, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.087793+00', NULL, '0xf0da83de6b12251c248276fd2aca7874df535564364631730b195bf3a39cccaa', 565, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 5, 4, '0400000000f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a300000000000000000000016345785d8a0000', 40885, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (7, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.088332+00', NULL, '0xf6ef1adb8e3076438af6fbbb8971719157b90e610bcaaef29a22517fe67d8418', 568, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 6, 4, '040000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f000100000000000000056bc75e2d63100000', 40888, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (8, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.089417+00', NULL, '0x3b9930b1cac58c31252ebb4331e2881f3d9d47c94df2c17a4cd7a30792f1d970', 569, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 7, 4, '040000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f000200000000000000056bc75e2d63100000', 40889, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (9, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.090495+00', NULL, '0xe726c8ac58a27de860d0fc415b8f504e39a70994a6cb7f09e7eec4a1253be3e2', 571, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 8, 2, '02000000000002001e000000000005', 40891, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (10, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.091595+00', NULL, '0x9e884269087bdab88ede8f4ae0dce1fef838a8da0c55c3bdd3debb9674688754', 572, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 9, 2, '02000100000001001e000000000005', 40892, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (11, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.092135+00', NULL, '0xa8970751d98d865ac01f2484c117ca8df20ca4f99646016dd3a02bb1075a4f3a', 573, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 10, 2, '02000200010002001e000000000005', 40893, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (12, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.093224+00', NULL, '0xa0b6475c914331371534674d19aace5f1d5c8e662a9b75b53aa1870721309a50', 575, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 11, 3, '030001003200000000000a', 40895, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (13, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.093751+00', NULL, '0x78ccc749096a1702fc9bf4b67c8165a34f16da93ba22cb4cbfe91e69d32aba04', 578, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 12, 5, '0500000000000000000078c34ad5641ae34edec94dd463c61298070ff7be0000000000000c1293246402e53a2701d0d764bd3ce7ac9a2b82c8a1cd3c6c1d1ef3ec8076f2000000000000000000000000000000000000000000000000000000000000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f0000', 40898, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (14, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.094296+00', NULL, '0xaaaec6117af4fce5324dd6dee37a9c357872a9375e32a801a55cac52f74a8edf', 580, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 13, 17, '110000000000010000000000000000000000000000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f', 40900, 2);
INSERT INTO "public"."l2_tx_event_monitor" VALUES (15, '2022-05-30 05:10:23.011335+00', '2022-05-30 05:10:50.095384+00', NULL, '0x9f09994bda2e87e315c8723870d684782f0d4070a6d494464170d5adb314d2ff', 582, '0x677d65A350c9FB84b14bDDF591043eb8243960D1', 14, 18, '120000000000000000000000000000000000000000000000000000000000000000000000000004b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 40902, 2);

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
INSERT INTO "public"."liquidity" VALUES (3, '2022-05-30 05:10:50.077029+00', '2022-05-30 05:11:11.176566+00', NULL, 2, 1, '0', 2, '0', '0', '0', 30, 0, 5);
INSERT INTO "public"."liquidity" VALUES (2, '2022-05-30 05:10:50.077029+00', '2022-05-30 05:11:11.187988+00', NULL, 1, 0, '0', 1, '0', '0', '0', 50, 0, 10);
INSERT INTO "public"."liquidity" VALUES (1, '2022-05-30 05:10:50.077029+00', '2022-05-30 05:13:27.188425+00', NULL, 0, 0, '99802', 2, '100000', '99900', '9980200000', 30, 0, 5);

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
INSERT INTO "public"."liquidity_history" VALUES (1, '2022-05-30 05:11:11.153647+00', '2022-05-30 05:11:11.153647+00', NULL, 0, 0, '0', 2, '0', '0', '0', 30, 0, 5, 9);
INSERT INTO "public"."liquidity_history" VALUES (2, '2022-05-30 05:11:11.168087+00', '2022-05-30 05:11:11.168087+00', NULL, 1, 0, '0', 1, '0', '0', '0', 30, 0, 5, 10);
INSERT INTO "public"."liquidity_history" VALUES (3, '2022-05-30 05:11:11.178317+00', '2022-05-30 05:11:11.178317+00', NULL, 2, 1, '0', 2, '0', '0', '0', 30, 0, 5, 11);
INSERT INTO "public"."liquidity_history" VALUES (4, '2022-05-30 05:11:11.190118+00', '2022-05-30 05:11:11.190118+00', NULL, 1, 0, '0', 1, '0', '0', '0', 50, 0, 10, 12);
INSERT INTO "public"."liquidity_history" VALUES (5, '2022-05-30 05:13:27.152554+00', '2022-05-30 05:13:27.152554+00', NULL, 0, 0, '100000', 2, '100000', '100000', '10000000000', 30, 0, 5, 18);
INSERT INTO "public"."liquidity_history" VALUES (6, '2022-05-30 05:13:27.171225+00', '2022-05-30 05:13:27.171225+00', NULL, 0, 0, '99901', 2, '100100', '100000', '10000000000', 30, 0, 5, 19);
INSERT INTO "public"."liquidity_history" VALUES (7, '2022-05-30 05:13:27.190547+00', '2022-05-30 05:13:27.190547+00', NULL, 0, 0, '99802', 2, '100000', '99900', '9980200000', 30, 0, 5, 20);

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
INSERT INTO "public"."mempool_tx" VALUES (1, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.041979+00', NULL, 'df14aac2-dfd6-11ec-a855-7cb27d9ca483', 1, 0, '0', -1, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":1,"AccountIndex":0,"AccountName":"treasury.legend","AccountNameHash":"wNIBqs6aLBfOcGbcb/7695MPExfEyV0GYbFkocWE1nY=","PubKey":"412805eb224e8c10de9ee037f55c92f32266f057fad3279cf4bab0a49d8f4080"}', '', '', 0, 0, 0, 1, 1);
INSERT INTO "public"."mempool_tx" VALUES (2, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.065049+00', NULL, 'df15a86d-dfd6-11ec-a855-7cb27d9ca483', 1, 0, '0', -1, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":1,"AccountIndex":0,"AccountName":"gas.legend","AccountNameHash":"aPvRfnfuxQHGd8zDHCYPMO6O0EnIk5AOCEuot/dWnOY=","PubKey":"53aa127ef258d5311bb9d8736d087e1c81204d356f876e7c42c42befcd679827"}', '', '', 1, 0, 0, 2, 1);
INSERT INTO "public"."mempool_tx" VALUES (3, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.0793+00', NULL, 'df15d330-dfd6-11ec-a855-7cb27d9ca483', 1, 0, '0', -1, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":1,"AccountIndex":0,"AccountName":"sher.legend","AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","PubKey":"7f70064f2c485996dc2acb397d0b4fe63eec854aad09b6fd3c41549e6d046586"}', '', '', 2, 0, 0, 3, 1);
INSERT INTO "public"."mempool_tx" VALUES (4, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.09062+00', NULL, 'df15e800-dfd6-11ec-a855-7cb27d9ca483', 1, 0, '0', -1, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":1,"AccountIndex":0,"AccountName":"gavin.legend","AccountNameHash":"9KZJFrMtD0ZzaZct0Vb30r2FnAoQijs5WiUPGU9GgKM=","PubKey":"c9e9ccb618f4825496506f70551d725dec7aeb2e3f31da262ea45ab88a174909"}', '', '', 3, 0, 0, 4, 1);
INSERT INTO "public"."mempool_tx" VALUES (5, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.104197+00', NULL, 'df15e800-dfd6-11ec-a856-7cb27d9ca483', 4, 0, '0', -1, -1, 0, '100000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 2, 0, 0, 5, 1);
INSERT INTO "public"."mempool_tx" VALUES (6, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.114862+00', NULL, 'df15e800-dfd6-11ec-a857-7cb27d9ca483', 4, 0, '0', -1, -1, 0, '100000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":4,"AccountIndex":3,"AccountNameHash":"9KZJFrMtD0ZzaZct0Vb30r2FnAoQijs5WiUPGU9GgKM=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 3, 0, 0, 6, 1);
INSERT INTO "public"."mempool_tx" VALUES (7, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.126837+00', NULL, 'df15e800-dfd6-11ec-a858-7cb27d9ca483', 4, 0, '0', -1, -1, 1, '100000000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, 7, 1);
INSERT INTO "public"."mempool_tx" VALUES (8, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.137011+00', NULL, 'df15e800-dfd6-11ec-a859-7cb27d9ca483', 4, 0, '0', -1, -1, 2, '100000000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":2,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, 8, 1);
INSERT INTO "public"."mempool_tx" VALUES (9, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.148862+00', NULL, 'df15e800-dfd6-11ec-a85a-7cb27d9ca483', 2, 0, '0', -1, 0, 0, '0', '0', '{"TxType":2,"PairIndex":0,"AssetAId":0,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, 9, 1);
INSERT INTO "public"."mempool_tx" VALUES (10, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.164332+00', NULL, 'df15e800-dfd6-11ec-a85b-7cb27d9ca483', 2, 0, '0', -1, 1, 0, '0', '0', '{"TxType":2,"PairIndex":1,"AssetAId":0,"AssetBId":1,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, 10, 1);
INSERT INTO "public"."mempool_tx" VALUES (11, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.174965+00', NULL, 'df15e800-dfd6-11ec-a85c-7cb27d9ca483', 2, 0, '0', -1, 2, 0, '0', '0', '{"TxType":2,"PairIndex":2,"AssetAId":1,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, 11, 1);
INSERT INTO "public"."mempool_tx" VALUES (12, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.186394+00', NULL, 'df15fcf9-dfd6-11ec-a85c-7cb27d9ca483', 3, 0, '0', -1, 1, 0, '0', '0', '{"TxType":3,"PairIndex":1,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', '', '', -1, 0, 0, 12, 1);
INSERT INTO "public"."mempool_tx" VALUES (13, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.203212+00', NULL, 'df1774e6-dfd6-11ec-a85c-7cb27d9ca483', 5, 0, '0', 0, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":5,"AccountIndex":2,"NftIndex":0,"NftL1Address":"0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE","CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftContentHash":"DBKTJGQC5TonAdDXZL0856yaK4LIoc08bB0e8+yAdvI=","NftL1TokenId":0,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","CollectionId":0}', '', '', 2, 0, 0, 13, 1);
INSERT INTO "public"."mempool_tx" VALUES (14, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.217282+00', NULL, 'df178a04-dfd6-11ec-a85c-7cb27d9ca483', 17, 0, '0', -1, -1, 1, '100000000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":17,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, 14, 1);
INSERT INTO "public"."mempool_tx" VALUES (15, '2022-05-30 05:10:50.05641+00', '2022-05-30 05:11:11.229722+00', NULL, 'df178a04-dfd6-11ec-a85d-7cb27d9ca483', 18, 0, '0', 0, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":18,"AccountIndex":2,"CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftIndex":0,"CollectionId":0,"NftL1Address":"0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE","AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","CreatorAccountNameHash":"AA==","NftContentHash":"DBKTJGQC5TonAdDXZL0856yaK4LIoc08bB0e8+yAdvI=","NftL1TokenId":0}', '', '', 2, 0, 0, 15, 1);
INSERT INTO "public"."mempool_tx" VALUES (17, '2022-05-30 05:11:38.233334+00', '2022-05-30 05:13:27.119588+00', NULL, 'e0725e08-7457-4add-b5f7-0bad35e7b131', 10, 2, '5000', -1, -1, 0, '10000000', '0x99AC8881834797ebC32f185ee27c2e96842e1a47', '{"FromAccountIndex":2,"AssetId":0,"AssetAmount":10000000,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ToAddress":"0x99AC8881834797ebC32f185ee27c2e96842e1a47","ExpiredAt":1653894698213,"Nonce":2,"Sig":"WSxUpuP9Aepr1BMPwcVJTlESeefnLCX/NyxWh2J01qcCJ8gMNzYAHv/vVqGB3Fmm3TwWuib+raAJdDDGv+Gq8g=="}', '', '', 2, 2, 1653894698213, 17, 1);
INSERT INTO "public"."mempool_tx" VALUES (18, '2022-05-30 05:11:48.100953+00', '2022-05-30 05:13:27.143489+00', NULL, 'e3e31b80-b5a1-4b76-b5c4-65a8f87c51f9', 8, 2, '5000', -1, 0, 0, '100000', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":0,"AssetAAmount":100000,"AssetBId":2,"AssetBAmount":100000,"LpAmount":100000,"KLast":10000000000,"TreasuryAmount":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894708071,"Nonce":3,"Sig":"RND1pOP4/14hseE7SiK5hZNUj/tMdnYkFk3bUDm6IR0D7IUW/h+ndzoTA0p79q9Hk5zMCHJT1tQtXYQ2LTX2Iw=="}', '', '', 2, 3, 1653894708071, 18, 1);
INSERT INTO "public"."mempool_tx" VALUES (16, '2022-05-30 05:11:27.581194+00', '2022-05-30 05:13:27.063798+00', NULL, '57d48626-6c44-46db-9408-0570909a1572', 6, 2, '5000', -1, -1, 0, '100000', '', '{"FromAccountIndex":2,"ToAccountIndex":3,"ToAccountNameHash":"f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a3","AssetId":0,"AssetAmount":100000,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"Memo":"transfer","CallData":"","CallDataHash":"Dd56AihX/sG4/6dmSpN6JQ065o81YGF1TTUx4mdBA9g=","ExpiredAt":1653894687554,"Nonce":1,"Sig":"kBZV9VlzsmcbBXSGl3/FnLZOULyzahQIPKsX9WoTgR4Est/ghAeg0Z8+RIZU6mwNh33UAw+HmgYCGIWnK5ZN8A=="}', '', 'transfer', 2, 1, 1653894687554, 16, 1);
INSERT INTO "public"."mempool_tx" VALUES (19, '2022-05-30 05:11:57.720139+00', '2022-05-30 05:13:27.163161+00', NULL, '2099988c-b889-4924-8929-5eb46ac07329', 7, 0, '5000', -1, 0, 0, '100', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":2,"AssetAAmount":100,"AssetBId":0,"AssetBMinAmount":98,"AssetBAmountDelta":99,"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"ExpiredAt":1653894717694,"Nonce":4,"Sig":"NzbsuPX6rEFysJitNKc5CyeEe1H3TirdEYTTvUYNqygAQHNToygl1L1hEepZxqQ0UaITSH6/2foUiNUdrsKz1g=="}', '', '', 2, 4, 1653894717694, 19, 1);
INSERT INTO "public"."mempool_tx" VALUES (20, '2022-05-30 05:12:07.322865+00', '2022-05-30 05:13:27.18135+00', NULL, 'bbab860d-9b58-40ab-98fb-d5aeead2befc', 9, 2, '5000', -1, 0, 0, '100', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":0,"AssetAMinAmount":98,"AssetBId":2,"AssetBMinAmount":99,"LpAmount":100,"AssetAAmountDelta":99,"AssetBAmountDelta":100,"KLast":9980200000,"TreasuryAmount":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894727296,"Nonce":5,"Sig":"B1VNlxQ+v9uE5N3/PwUZP6M6DWnmLQnbPPVmVtmUTg8CVaCnY8hoEHEOaVouWpBOHGAOvCIOC3D7xt6gXt7Gfg=="}', '', '', 2, 5, 1653894727296, 20, 1);
INSERT INTO "public"."mempool_tx" VALUES (21, '2022-05-30 05:12:17.063514+00', '2022-05-30 05:13:27.19803+00', NULL, '45b43cfe-a44b-4390-be26-e1691d1b7891', 11, 2, '5000', -1, -1, 0, 'sher.legend', '0', '{"AccountIndex":2,"CollectionId":1,"Name":"Zecrey Collection","Introduction":"Wonderful zecrey!","GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894737032,"Nonce":6,"Sig":"DHtTO/fLN1EtSnsuKiMtRP4nK+MaDXMyrXPlYcXu2pQDiYgHm8gy28CMzEcg2nKIonM2jVqLoifSSjVgYXe1fQ=="}', '', '', 2, 6, 1653894737032, 21, 1);
INSERT INTO "public"."mempool_tx" VALUES (22, '2022-05-30 05:12:26.760284+00', '2022-05-30 05:13:27.215214+00', NULL, '4f74db23-d540-42dc-b46e-fb94a0893e60', 12, 2, '5000', 1, -1, 0, '0', '', '{"CreatorAccountIndex":2,"ToAccountIndex":3,"ToAccountNameHash":"f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a3","NftIndex":1,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftCollectionId":1,"CreatorTreasuryRate":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894746723,"Nonce":7,"Sig":"DeLXCzbM5T21fJDR5mykshr7ry0hDsp+rwN9BPnT+5kBytoElZyKraEv9bicIdivuFvZtIa71vEx0UnlmdyCPQ=="}', '', '', 2, 7, 1653894746723, 22, 1);
INSERT INTO "public"."mempool_tx" VALUES (23, '2022-05-30 05:12:36.97637+00', '2022-05-30 05:13:27.238348+00', NULL, '94de5a20-0be2-4772-8aa6-369c8b89f6bb', 13, 0, '5000', 1, -1, 0, '0', '', '{"FromAccountIndex":3,"ToAccountIndex":2,"ToAccountNameHash":"04b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f","NftIndex":1,"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"CallData":"","CallDataHash":"Dd56AihX/sG4/6dmSpN6JQ065o81YGF1TTUx4mdBA9g=","ExpiredAt":1653894756947,"Nonce":1,"Sig":"OdlWJsxjuvGIlJcYzZmavMvAtcvvvO/ZjPk/buw7B60FBiD0VjOZu2lwY7ep871yDwm5DLhGf8azfQb2Sb02DQ=="}', '', '', 3, 1, 1653894756947, 23, 1);
INSERT INTO "public"."mempool_tx" VALUES (24, '2022-05-30 05:12:46.775741+00', '2022-05-30 05:13:27.262156+00', NULL, '88843228-98e8-4f01-b8a3-e5c21c99cd44', 14, 0, '5000', 1, -1, 0, '10000', '', '{"AccountIndex":2,"BuyOffer":{"Type":0,"OfferId":0,"AccountIndex":3,"NftIndex":1,"AssetId":0,"AssetAmount":10000,"ListedAt":1653887566744,"ExpiredAt":1653894766744,"TreasuryRate":200,"Sig":"AbILxtNPKmAIEY8fBmZyaHF/9jHWMmHGjbqYou5/HS4BwVwkQU9ZXuNc1XedTTk/aEeiTjFH8zXK3DdzSE38/Q=="},"SellOffer":{"Type":1,"OfferId":0,"AccountIndex":2,"NftIndex":1,"AssetId":0,"AssetAmount":10000,"ListedAt":1653887566744,"ExpiredAt":1653894766744,"TreasuryRate":200,"Sig":"/lX3kvKUAXtaRaryr8wXbmWDF33rN8kMwkZWZf/jb4sEcqSDU1VQFPVEL0Rf0YuZRL7Ud0VSLNSzEmjqA1PSvQ=="},"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"CreatorAmount":0,"TreasuryAmount":200,"Nonce":8,"ExpiredAt":1653894766744,"Sig":"9KLmDNsetYzx7V1I4gUo7YHbCIhNlDmHI4Migtf08w8BZJ9qw1+9b7A7EDNjALAVs1SV4s/C95DjsThFd45FEA=="}', '', '', 2, 8, 1653894766744, 24, 1);
INSERT INTO "public"."mempool_tx" VALUES (25, '2022-05-30 05:12:57.071876+00', '2022-05-30 05:13:27.282721+00', NULL, '07c849fd-0d54-4132-ba0a-2e4d0f79cfd0', 15, 2, '5000', -1, -1, 0, 'sher.legend', '0', '{"AccountIndex":2,"OfferId":1,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894777041,"Nonce":9,"Sig":"S+8/suwX1dKDEiPtZKxWRQhgGnF5eyQ7QtH6hkewJiED/Jby88sGpmbv5cQ2X4H4eTN5H+ZA+zsvQE6uMzzgYw=="}', '', '', 2, 9, 1653894777041, 25, 1);
INSERT INTO "public"."mempool_tx" VALUES (26, '2022-05-30 05:13:07.846481+00', '2022-05-30 05:13:27.297185+00', NULL, '755100ba-85fa-461b-8ef2-b25ed269a019', 16, 0, '5000', 1, -1, 0, '0', '', '{"AccountIndex":3,"CreatorAccountIndex":2,"CreatorAccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","CreatorTreasuryRate":0,"NftIndex":1,"NftContentHash":"CbvOME8CPnvrZB/lsVUIPtzNyjQjTnRjMgdO6w/fB9E=","NftL1Address":"0","NftL1TokenId":0,"CollectionId":1,"ToAddress":"0xd5Aa3B56a2E2139DB315CdFE3b34149c8ed09171","GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"ExpiredAt":1653894787817,"Nonce":2,"Sig":"ZF/HLauJDTGD1+qUMZwpNfi1oT38dF1snx2Tq4ukGZwDiRi2FWKws6xRIAsWezitCtGHuJOILjgfI+rnNlX5Rw=="}', '', '', 3, 2, 1653894787817, 26, 1);

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
INSERT INTO "public"."mempool_tx_detail" VALUES (1, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 5, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (2, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 6, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (3, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 7, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (4, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 8, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (5, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 9, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (6, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 10, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (7, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 11, 2, 2, -1, '', '{"PairIndex":2,"AssetAId":1,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (8, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 12, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (9, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 13, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (10, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 13, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":2,"NftContentHash":"0c1293246402e53a2701d0d764bd3ce7ac9a2b82c8a1cd3c6c1d1ef3ec8076f2","NftL1TokenId":"0","NftL1Address":"0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE","CreatorTreasuryRate":0,"CollectionId":0}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (11, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 14, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":null}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (12, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 15, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (13, '2022-05-30 05:10:50.071576+00', '2022-05-30 05:10:50.071576+00', NULL, 15, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (18, '2022-05-30 05:11:38.235472+00', '2022-05-30 05:11:38.235472+00', NULL, 17, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-10000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (19, '2022-05-30 05:11:38.235472+00', '2022-05-30 05:11:38.235472+00', NULL, 17, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (20, '2022-05-30 05:11:38.235472+00', '2022-05-30 05:11:38.235472+00', NULL, 17, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (21, '2022-05-30 05:11:48.102547+00', '2022-05-30 05:11:48.102547+00', NULL, 18, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (22, '2022-05-30 05:11:48.102547+00', '2022-05-30 05:11:48.102547+00', NULL, 18, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (23, '2022-05-30 05:11:48.102547+00', '2022-05-30 05:11:48.102547+00', NULL, 18, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (24, '2022-05-30 05:11:48.102547+00', '2022-05-30 05:11:48.102547+00', NULL, 18, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":100000,"OfferCanceledOrFinalized":0}', 3, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (28, '2022-05-30 05:11:57.722247+00', '2022-05-30 05:11:57.722247+00', NULL, 19, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-100,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (29, '2022-05-30 05:11:57.722247+00', '2022-05-30 05:11:57.722247+00', NULL, 19, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (30, '2022-05-30 05:11:57.722247+00', '2022-05-30 05:11:57.722247+00', NULL, 19, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (31, '2022-05-30 05:11:57.722247+00', '2022-05-30 05:11:57.722247+00', NULL, 19, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":-99,"AssetBId":2,"AssetB":100,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 3, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (32, '2022-05-30 05:11:57.722247+00', '2022-05-30 05:11:57.722247+00', NULL, 19, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 4, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (33, '2022-05-30 05:12:07.324459+00', '2022-05-30 05:12:07.324459+00', NULL, 20, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (34, '2022-05-30 05:12:07.324459+00', '2022-05-30 05:12:07.324459+00', NULL, 20, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":100,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (35, '2022-05-30 05:12:07.324459+00', '2022-05-30 05:12:07.324459+00', NULL, 20, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (36, '2022-05-30 05:12:07.324459+00', '2022-05-30 05:12:07.324459+00', NULL, 20, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":-100,"OfferCanceledOrFinalized":0}', 3, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (37, '2022-05-30 05:12:07.324459+00', '2022-05-30 05:12:07.324459+00', NULL, 20, 0, 1, 0, 'treasury.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 4, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (14, '2022-05-30 05:11:27.583386+00', '2022-05-30 05:11:27.583386+00', NULL, 16, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (15, '2022-05-30 05:11:27.583386+00', '2022-05-30 05:11:27.583386+00', NULL, 16, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (16, '2022-05-30 05:11:27.583386+00', '2022-05-30 05:11:27.583386+00', NULL, 16, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (17, '2022-05-30 05:11:27.583386+00', '2022-05-30 05:11:27.583386+00', NULL, 16, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (25, '2022-05-30 05:11:48.102547+00', '2022-05-30 05:11:48.102547+00', NULL, 18, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":100000,"AssetBId":2,"AssetB":100000,"LpAmount":100000,"KLast":10000000000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 4, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (26, '2022-05-30 05:11:48.102547+00', '2022-05-30 05:11:48.102547+00', NULL, 18, 0, 1, 0, 'treasury.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 5, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (27, '2022-05-30 05:11:48.102547+00', '2022-05-30 05:11:48.102547+00', NULL, 18, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 6, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (38, '2022-05-30 05:12:07.324459+00', '2022-05-30 05:12:07.324459+00', NULL, 20, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":-99,"AssetBId":2,"AssetB":-100,"LpAmount":-100,"KLast":9980200000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 5, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (39, '2022-05-30 05:12:07.324459+00', '2022-05-30 05:12:07.324459+00', NULL, 20, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 6, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (40, '2022-05-30 05:12:17.066154+00', '2022-05-30 05:12:17.066154+00', NULL, 21, 0, 4, 2, 'sher.legend', '1', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (41, '2022-05-30 05:12:17.066154+00', '2022-05-30 05:12:17.066154+00', NULL, 21, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (43, '2022-05-30 05:12:26.762042+00', '2022-05-30 05:12:26.762042+00', NULL, 22, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (44, '2022-05-30 05:12:26.762042+00', '2022-05-30 05:12:26.762042+00', NULL, 22, 2, 1, 3, 'gavin.legend', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (46, '2022-05-30 05:12:26.762042+00', '2022-05-30 05:12:26.762042+00', NULL, 22, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (47, '2022-05-30 05:12:36.979072+00', '2022-05-30 05:12:36.979072+00', NULL, 23, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (48, '2022-05-30 05:12:36.979072+00', '2022-05-30 05:12:36.979072+00', NULL, 23, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (49, '2022-05-30 05:12:36.979072+00', '2022-05-30 05:12:36.979072+00', NULL, 23, 1, 3, 2, 'sher.legend', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":2,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 2, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (50, '2022-05-30 05:12:36.979072+00', '2022-05-30 05:12:36.979072+00', NULL, 23, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (51, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (52, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":-10000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (53, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":1}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (54, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":9800,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (55, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":1}', 4, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (60, '2022-05-30 05:12:57.074126+00', '2022-05-30 05:12:57.074126+00', NULL, 25, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (61, '2022-05-30 05:12:57.074126+00', '2022-05-30 05:12:57.074126+00', NULL, 25, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":3}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (62, '2022-05-30 05:12:57.074126+00', '2022-05-30 05:12:57.074126+00', NULL, 25, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (63, '2022-05-30 05:13:07.848621+00', '2022-05-30 05:13:07.848621+00', NULL, 26, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (64, '2022-05-30 05:13:07.848621+00', '2022-05-30 05:13:07.848621+00', NULL, 26, 1, 3, -1, '', '{"NftIndex":1,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (65, '2022-05-30 05:13:07.848621+00', '2022-05-30 05:13:07.848621+00', NULL, 26, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (66, '2022-05-30 05:13:07.848621+00', '2022-05-30 05:13:07.848621+00', NULL, 26, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (42, '2022-05-30 05:12:17.066154+00', '2022-05-30 05:12:17.066154+00', NULL, 21, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (45, '2022-05-30 05:12:26.762042+00', '2022-05-30 05:12:26.762042+00', NULL, 22, 1, 3, 3, 'gavin.legend', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 2, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (56, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 5, 3);
INSERT INTO "public"."mempool_tx_detail" VALUES (57, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 1, 3, -1, '', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 6, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (58, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":200,"LpAmount":0,"OfferCanceledOrFinalized":0}', 7, 4);
INSERT INTO "public"."mempool_tx_detail" VALUES (59, '2022-05-30 05:12:46.77788+00', '2022-05-30 05:12:46.77788+00', NULL, 24, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 8, 4);

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
  "account_root" bytea,
  "commitment" bytea,
  "timestamp" int8,
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
INSERT INTO "public"."sys_config" VALUES (1, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'SysGasFee', '1', 'float', 'based on ETH');
INSERT INTO "public"."sys_config" VALUES (2, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'MaxAssetId', '9', 'int', 'max number of asset id');
INSERT INTO "public"."sys_config" VALUES (3, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'TreasuryAccountIndex', '0', 'int', 'treasury index');
INSERT INTO "public"."sys_config" VALUES (4, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'GasAccountIndex', '1', 'int', 'gas index');
INSERT INTO "public"."sys_config" VALUES (5, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'ZecreyLegendContract', '0x0C17367D14760D2a54A3D140c9F2f1c2EdB81E7D', 'string', 'Zecrey contract on BSC');
INSERT INTO "public"."sys_config" VALUES (6, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'GovernanceContract', '0x4B7635b2A882F94cB4E50CDc073bA8630f1759A6', 'string', 'Governance contract on BSC');
INSERT INTO "public"."sys_config" VALUES (7, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'AssetGovernanceContract', '0x3C4237AbEf419C7C76efAd854b7166F49C77F516', 'string', 'Asset_Governance contract on BSC');
INSERT INTO "public"."sys_config" VALUES (8, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'VerifierContract', '0x4EFCfA18c2cdf4661C028Df55F4911c7F82F253d', 'string', 'Verifier contract on BSC');
INSERT INTO "public"."sys_config" VALUES (9, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'BscTestNetworkRpc', 'https://data-seed-prebsc-1-s1.binance.org:8545/', 'string', 'BSC network rpc');
INSERT INTO "public"."sys_config" VALUES (10, '2022-05-30 05:07:44.330829+00', '2022-05-30 05:07:44.330829+00', NULL, 'Local_Test_Network_RPC', 'http://127.0.0.1:8545/', 'string', 'Local network rpc');

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
  "account_root" text COLLATE "pg_catalog"."default",
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
INSERT INTO "public"."tx" VALUES (1, '2022-05-30 05:11:11.038683+00', '2022-05-30 05:11:11.038683+00', NULL, 'df14aac2-dfd6-11ec-a855-7cb27d9ca483', 1, '0', 0, 1, 1, 2, '0896c0f1c9e8fbc0b71870e545386ef20b89e546d1abb8d10a0235bead6b7364', -1, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":1,"AccountIndex":0,"AccountName":"treasury.legend","AccountNameHash":"wNIBqs6aLBfOcGbcb/7695MPExfEyV0GYbFkocWE1nY=","PubKey":"412805eb224e8c10de9ee037f55c92f32266f057fad3279cf4bab0a49d8f4080"}', '', '', 0, 0, 0);
INSERT INTO "public"."tx" VALUES (2, '2022-05-30 05:11:11.063984+00', '2022-05-30 05:11:11.063984+00', NULL, 'df15a86d-dfd6-11ec-a855-7cb27d9ca483', 1, '0', 0, 1, 2, 3, '141d76c64d7ea9aae087ebcd990e116e0e918044f3f1d4e75d3e32ec01f8d209', -1, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":1,"AccountIndex":0,"AccountName":"gas.legend","AccountNameHash":"aPvRfnfuxQHGd8zDHCYPMO6O0EnIk5AOCEuot/dWnOY=","PubKey":"53aa127ef258d5311bb9d8736d087e1c81204d356f876e7c42c42befcd679827"}', '', '', 1, 0, 0);
INSERT INTO "public"."tx" VALUES (3, '2022-05-30 05:11:11.077682+00', '2022-05-30 05:11:11.077682+00', NULL, 'df15d330-dfd6-11ec-a855-7cb27d9ca483', 1, '0', 0, 1, 3, 4, '2e378e466dab807e2d769076b5e16a6d3e043fbc01f5c526d239c12f3cf9fa2c', -1, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":1,"AccountIndex":0,"AccountName":"sher.legend","AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","PubKey":"7f70064f2c485996dc2acb397d0b4fe63eec854aad09b6fd3c41549e6d046586"}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (4, '2022-05-30 05:11:11.089542+00', '2022-05-30 05:11:11.089542+00', NULL, 'df15e800-dfd6-11ec-a855-7cb27d9ca483', 1, '0', 0, 1, 4, 5, '2efa2a41c1168b8aecd5f5d1d67c0a442df6bd8fa37e589757aada6841d6ebf2', -1, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":1,"AccountIndex":0,"AccountName":"gavin.legend","AccountNameHash":"9KZJFrMtD0ZzaZct0Vb30r2FnAoQijs5WiUPGU9GgKM=","PubKey":"c9e9ccb618f4825496506f70551d725dec7aeb2e3f31da262ea45ab88a174909"}', '', '', 3, 0, 0);
INSERT INTO "public"."tx" VALUES (5, '2022-05-30 05:11:11.099757+00', '2022-05-30 05:11:11.099757+00', NULL, 'df15e800-dfd6-11ec-a856-7cb27d9ca483', 4, '0', 0, 1, 5, 6, '06f601cc791636973388cea255953926a1ac70107a6230b77945cf74fe46f9f6', -1, -1, 0, '100000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (6, '2022-05-30 05:11:11.11326+00', '2022-05-30 05:11:11.11326+00', NULL, 'df15e800-dfd6-11ec-a857-7cb27d9ca483', 4, '0', 0, 1, 6, 7, '0e38c1b0f794dcd7ddd947fe82e5f75efaaae7747459de52c94d92f9df0b64ef', -1, -1, 0, '100000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":4,"AccountIndex":3,"AccountNameHash":"9KZJFrMtD0ZzaZct0Vb30r2FnAoQijs5WiUPGU9GgKM=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 3, 0, 0);
INSERT INTO "public"."tx" VALUES (7, '2022-05-30 05:11:11.124686+00', '2022-05-30 05:11:11.124686+00', NULL, 'df15e800-dfd6-11ec-a858-7cb27d9ca483', 4, '0', 0, 1, 7, 8, '1f7b6a15037e6b1fb1688221ad16f5cbe0f7f2cdc030aacb3c37d1bec2813fc3', -1, -1, 1, '100000000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (8, '2022-05-30 05:11:11.135413+00', '2022-05-30 05:11:11.135413+00', NULL, 'df15e800-dfd6-11ec-a859-7cb27d9ca483', 4, '0', 0, 1, 8, 9, '2e0bc9bad4f2cf0dbd5b2a3570879832ef2579c975b1dcb43840cdc523914d80', -1, -1, 2, '100000000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":2,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (9, '2022-05-30 05:11:11.146552+00', '2022-05-30 05:11:11.146552+00', NULL, 'df15e800-dfd6-11ec-a85a-7cb27d9ca483', 2, '0', 0, 1, 9, 10, '228199bd4964493fd18425a3ba2bfed71160c8302232c3c8c693512a76626779', -1, 0, 0, '0', '0', '{"TxType":2,"PairIndex":0,"AssetAId":0,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0);
INSERT INTO "public"."tx" VALUES (10, '2022-05-30 05:11:11.16269+00', '2022-05-30 05:11:11.16269+00', NULL, 'df15e800-dfd6-11ec-a85b-7cb27d9ca483', 2, '0', 0, 1, 10, 11, '00fdfc52d138f9ba5ce9611c805426b95616f518732bcfc8e40a2a2019e15287', -1, 1, 0, '0', '0', '{"TxType":2,"PairIndex":1,"AssetAId":0,"AssetBId":1,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0);
INSERT INTO "public"."tx" VALUES (11, '2022-05-30 05:11:11.173359+00', '2022-05-30 05:11:11.173359+00', NULL, 'df15e800-dfd6-11ec-a85c-7cb27d9ca483', 2, '0', 0, 1, 11, 12, '0cec93680cf90d14634ce669e90b157fe2029d18ccecc856a07c3b1b8ecb7b23', -1, 2, 0, '0', '0', '{"TxType":2,"PairIndex":2,"AssetAId":1,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0);
INSERT INTO "public"."tx" VALUES (12, '2022-05-30 05:11:11.184798+00', '2022-05-30 05:11:11.184798+00', NULL, 'df15fcf9-dfd6-11ec-a85c-7cb27d9ca483', 3, '0', 0, 1, 12, 13, '096665c5754506b34a69d625bbaec71f137e0d397c480dc488aa067187dd0db7', -1, 1, 0, '0', '0', '{"TxType":3,"PairIndex":1,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', '', '', -1, 0, 0);
INSERT INTO "public"."tx" VALUES (13, '2022-05-30 05:11:11.200548+00', '2022-05-30 05:11:11.200548+00', NULL, 'df1774e6-dfd6-11ec-a85c-7cb27d9ca483', 5, '0', 0, 1, 13, 14, '00f39cb5e4b0361d022655c730f402936bd3459583134c17aa9d9580e97d66b2', 0, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":5,"AccountIndex":2,"NftIndex":0,"NftL1Address":"0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE","CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftContentHash":"DBKTJGQC5TonAdDXZL0856yaK4LIoc08bB0e8+yAdvI=","NftL1TokenId":0,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","CollectionId":0}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (14, '2022-05-30 05:11:11.215688+00', '2022-05-30 05:11:11.215688+00', NULL, 'df178a04-dfd6-11ec-a85c-7cb27d9ca483', 17, '0', 0, 1, 14, 15, '09cf39a8a8d7787da03f683c147cef5939704f3ead516e8a2a157ec8fe27cd94', -1, -1, 1, '100000000000000000000', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":17,"AccountIndex":2,"AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (15, '2022-05-30 05:11:11.226501+00', '2022-05-30 05:11:11.226501+00', NULL, 'df178a04-dfd6-11ec-a85d-7cb27d9ca483', 18, '0', 0, 1, 15, 16, '0be5445077dd00cddefcb6539fb91f344fcdf62c2ab4be11e652b6cc88c0e9f9', 0, -1, 0, '0', '0x677d65A350c9FB84b14bDDF591043eb8243960D1', '{"TxType":18,"AccountIndex":2,"CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftIndex":0,"CollectionId":0,"NftL1Address":"0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE","AccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","CreatorAccountNameHash":"AA==","NftContentHash":"DBKTJGQC5TonAdDXZL0856yaK4LIoc08bB0e8+yAdvI=","NftL1TokenId":0}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (16, '2022-05-30 05:13:27.05948+00', '2022-05-30 05:13:27.05948+00', NULL, '57d48626-6c44-46db-9408-0570909a1572', 6, '5000', 2, 1, 16, 17, '18ab34ac1cb16907fe00e43f9adc4342d0d68e39b79767cba0e7947f6a373c35', -1, -1, 0, '100000', '', '{"FromAccountIndex":2,"ToAccountIndex":3,"ToAccountNameHash":"f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a3","AssetId":0,"AssetAmount":100000,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"Memo":"transfer","CallData":"","CallDataHash":"Dd56AihX/sG4/6dmSpN6JQ065o81YGF1TTUx4mdBA9g=","ExpiredAt":1653894687554,"Nonce":1,"Sig":"kBZV9VlzsmcbBXSGl3/FnLZOULyzahQIPKsX9WoTgR4Est/ghAeg0Z8+RIZU6mwNh33UAw+HmgYCGIWnK5ZN8A=="}', '', 'transfer', 2, 1, 1653894687554);
INSERT INTO "public"."tx" VALUES (17, '2022-05-30 05:13:27.116359+00', '2022-05-30 05:13:27.116359+00', NULL, 'e0725e08-7457-4add-b5f7-0bad35e7b131', 10, '5000', 2, 1, 17, 18, '12ff91cd0cab6f32781bef0fb9e12a16f14380860e4bda177d9c4c0de34239a6', -1, -1, 0, '10000000', '0x99AC8881834797ebC32f185ee27c2e96842e1a47', '{"FromAccountIndex":2,"AssetId":0,"AssetAmount":10000000,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ToAddress":"0x99AC8881834797ebC32f185ee27c2e96842e1a47","ExpiredAt":1653894698213,"Nonce":2,"Sig":"WSxUpuP9Aepr1BMPwcVJTlESeefnLCX/NyxWh2J01qcCJ8gMNzYAHv/vVqGB3Fmm3TwWuib+raAJdDDGv+Gq8g=="}', '', '', 2, 2, 1653894698213);
INSERT INTO "public"."tx" VALUES (18, '2022-05-30 05:13:27.138724+00', '2022-05-30 05:13:27.138724+00', NULL, 'e3e31b80-b5a1-4b76-b5c4-65a8f87c51f9', 8, '5000', 2, 1, 18, 19, '0502e76193590e0737cdd109bd000898ea41d23c4d2b7bb6fff773df7214e8e6', -1, 0, 0, '100000', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":0,"AssetAAmount":100000,"AssetBId":2,"AssetBAmount":100000,"LpAmount":100000,"KLast":10000000000,"TreasuryAmount":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894708071,"Nonce":3,"Sig":"RND1pOP4/14hseE7SiK5hZNUj/tMdnYkFk3bUDm6IR0D7IUW/h+ndzoTA0p79q9Hk5zMCHJT1tQtXYQ2LTX2Iw=="}', '', '', 2, 3, 1653894708071);
INSERT INTO "public"."tx" VALUES (19, '2022-05-30 05:13:27.160432+00', '2022-05-30 05:13:27.160432+00', NULL, '2099988c-b889-4924-8929-5eb46ac07329', 7, '5000', 0, 1, 19, 20, '08088d6698c046bec4fa2371bc023831c1ead344ca5f11b0c85084a0ea208fd2', -1, 0, 0, '100', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":2,"AssetAAmount":100,"AssetBId":0,"AssetBMinAmount":98,"AssetBAmountDelta":99,"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"ExpiredAt":1653894717694,"Nonce":4,"Sig":"NzbsuPX6rEFysJitNKc5CyeEe1H3TirdEYTTvUYNqygAQHNToygl1L1hEepZxqQ0UaITSH6/2foUiNUdrsKz1g=="}', '', '', 2, 4, 1653894717694);
INSERT INTO "public"."tx" VALUES (20, '2022-05-30 05:13:27.179771+00', '2022-05-30 05:13:27.179771+00', NULL, 'bbab860d-9b58-40ab-98fb-d5aeead2befc', 9, '5000', 2, 1, 20, 21, '23ac82db1260a22ae3fb448ac3c06b6dacd76ceb24a587d2e9e6256bdcafaf5c', -1, 0, 0, '100', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":0,"AssetAMinAmount":98,"AssetBId":2,"AssetBMinAmount":99,"LpAmount":100,"AssetAAmountDelta":99,"AssetBAmountDelta":100,"KLast":9980200000,"TreasuryAmount":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894727296,"Nonce":5,"Sig":"B1VNlxQ+v9uE5N3/PwUZP6M6DWnmLQnbPPVmVtmUTg8CVaCnY8hoEHEOaVouWpBOHGAOvCIOC3D7xt6gXt7Gfg=="}', '', '', 2, 5, 1653894727296);
INSERT INTO "public"."tx" VALUES (21, '2022-05-30 05:13:27.195882+00', '2022-05-30 05:13:27.195882+00', NULL, '45b43cfe-a44b-4390-be26-e1691d1b7891', 11, '5000', 2, 1, 21, 22, '18016dc2d985df4327b74915a4e595d7770abd84a4d2ba02a101662d683209a0', -1, -1, 0, 'sher.legend', '0', '{"AccountIndex":2,"CollectionId":1,"Name":"Zecrey Collection","Introduction":"Wonderful zecrey!","GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894737032,"Nonce":6,"Sig":"DHtTO/fLN1EtSnsuKiMtRP4nK+MaDXMyrXPlYcXu2pQDiYgHm8gy28CMzEcg2nKIonM2jVqLoifSSjVgYXe1fQ=="}', '', '', 2, 6, 1653894737032);
INSERT INTO "public"."tx" VALUES (22, '2022-05-30 05:13:27.212999+00', '2022-05-30 05:13:27.212999+00', NULL, '4f74db23-d540-42dc-b46e-fb94a0893e60', 12, '5000', 2, 1, 22, 23, '1c2f58272ada607379fd722177a3a455c8cfeb6965b9706d00f7bfd003ef71f1', 1, -1, 0, '0', '', '{"CreatorAccountIndex":2,"ToAccountIndex":3,"ToAccountNameHash":"f4a64916b32d0f467369972dd156f7d2bd859c0a108a3b395a250f194f4680a3","NftIndex":1,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftCollectionId":1,"CreatorTreasuryRate":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894746723,"Nonce":7,"Sig":"DeLXCzbM5T21fJDR5mykshr7ry0hDsp+rwN9BPnT+5kBytoElZyKraEv9bicIdivuFvZtIa71vEx0UnlmdyCPQ=="}', '', '', 2, 7, 1653894746723);
INSERT INTO "public"."tx" VALUES (23, '2022-05-30 05:13:27.234938+00', '2022-05-30 05:13:27.234938+00', NULL, '94de5a20-0be2-4772-8aa6-369c8b89f6bb', 13, '5000', 0, 1, 23, 24, '2e0f826aba027333a11b3a355166e9080663595eacc68727bc0d8fd9d46e8d4b', 1, -1, 0, '0', '', '{"FromAccountIndex":3,"ToAccountIndex":2,"ToAccountNameHash":"04b2dd1162802d057ed00dcb516ea627b207970520d1ad583f712cd6e954691f","NftIndex":1,"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"CallData":"","CallDataHash":"Dd56AihX/sG4/6dmSpN6JQ065o81YGF1TTUx4mdBA9g=","ExpiredAt":1653894756947,"Nonce":1,"Sig":"OdlWJsxjuvGIlJcYzZmavMvAtcvvvO/ZjPk/buw7B60FBiD0VjOZu2lwY7ep871yDwm5DLhGf8azfQb2Sb02DQ=="}', '', '', 3, 1, 1653894756947);
INSERT INTO "public"."tx" VALUES (24, '2022-05-30 05:13:27.25815+00', '2022-05-30 05:13:27.25815+00', NULL, '88843228-98e8-4f01-b8a3-e5c21c99cd44', 14, '5000', 0, 1, 24, 25, '050556f4ee047254af35995b4d2fbbf6d553d5119931e646b53512d368e12d1d', 1, -1, 0, '10000', '', '{"AccountIndex":2,"BuyOffer":{"Type":0,"OfferId":0,"AccountIndex":3,"NftIndex":1,"AssetId":0,"AssetAmount":10000,"ListedAt":1653887566744,"ExpiredAt":1653894766744,"TreasuryRate":200,"Sig":"AbILxtNPKmAIEY8fBmZyaHF/9jHWMmHGjbqYou5/HS4BwVwkQU9ZXuNc1XedTTk/aEeiTjFH8zXK3DdzSE38/Q=="},"SellOffer":{"Type":1,"OfferId":0,"AccountIndex":2,"NftIndex":1,"AssetId":0,"AssetAmount":10000,"ListedAt":1653887566744,"ExpiredAt":1653894766744,"TreasuryRate":200,"Sig":"/lX3kvKUAXtaRaryr8wXbmWDF33rN8kMwkZWZf/jb4sEcqSDU1VQFPVEL0Rf0YuZRL7Ud0VSLNSzEmjqA1PSvQ=="},"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"CreatorAmount":0,"TreasuryAmount":200,"Nonce":8,"ExpiredAt":1653894766744,"Sig":"9KLmDNsetYzx7V1I4gUo7YHbCIhNlDmHI4Migtf08w8BZJ9qw1+9b7A7EDNjALAVs1SV4s/C95DjsThFd45FEA=="}', '', '', 2, 8, 1653894766744);
INSERT INTO "public"."tx" VALUES (25, '2022-05-30 05:13:27.280507+00', '2022-05-30 05:13:27.280507+00', NULL, '07c849fd-0d54-4132-ba0a-2e4d0f79cfd0', 15, '5000', 2, 1, 25, 26, '290c51f8cb09e7a32efd05f623b0181427a0357630f31c86e6208ddcfdb4251c', -1, -1, 0, 'sher.legend', '0', '{"AccountIndex":2,"OfferId":1,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1653894777041,"Nonce":9,"Sig":"S+8/suwX1dKDEiPtZKxWRQhgGnF5eyQ7QtH6hkewJiED/Jby88sGpmbv5cQ2X4H4eTN5H+ZA+zsvQE6uMzzgYw=="}', '', '', 2, 9, 1653894777041);
INSERT INTO "public"."tx" VALUES (26, '2022-05-30 05:13:27.295033+00', '2022-05-30 05:13:27.295033+00', NULL, '755100ba-85fa-461b-8ef2-b25ed269a019', 16, '5000', 0, 1, 26, 27, '1a8a0465a857f6b1278ef72b0e42637b7febb0d32072294348ff27392335ac1c', 1, -1, 0, '0', '', '{"AccountIndex":3,"CreatorAccountIndex":2,"CreatorAccountNameHash":"BLLdEWKALQV+0A3LUW6mJ7IHlwUg0a1YP3Es1ulUaR8=","CreatorTreasuryRate":0,"NftIndex":1,"NftContentHash":"CbvOME8CPnvrZB/lsVUIPtzNyjQjTnRjMgdO6w/fB9E=","NftL1Address":"0","NftL1TokenId":0,"CollectionId":1,"ToAddress":"0xd5Aa3B56a2E2139DB315CdFE3b34149c8ed09171","GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"ExpiredAt":1653894787817,"Nonce":2,"Sig":"ZF/HLauJDTGD1+qUMZwpNfi1oT38dF1snx2Tq4ukGZwDiRi2FWKws6xRIAsWezitCtGHuJOILjgfI+rnNlX5Rw=="}', '', '', 3, 2, 1653894787817);

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
INSERT INTO "public"."tx_detail" VALUES (1, '2022-05-30 05:11:11.100987+00', '2022-05-30 05:11:11.100987+00', NULL, 5, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (2, '2022-05-30 05:11:11.114327+00', '2022-05-30 05:11:11.114327+00', NULL, 6, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (3, '2022-05-30 05:11:11.125763+00', '2022-05-30 05:11:11.125763+00', NULL, 7, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (4, '2022-05-30 05:11:11.135945+00', '2022-05-30 05:11:11.135945+00', NULL, 8, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (5, '2022-05-30 05:11:11.147207+00', '2022-05-30 05:11:11.147207+00', NULL, 9, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":0,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":0,"TreasuryAccountIndex":0,"TreasuryRate":0}', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (6, '2022-05-30 05:11:11.163261+00', '2022-05-30 05:11:11.163261+00', NULL, 10, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":0,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":0,"TreasuryAccountIndex":0,"TreasuryRate":0}', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (7, '2022-05-30 05:11:11.173895+00', '2022-05-30 05:11:11.173895+00', NULL, 11, 2, 2, -1, '', '{"PairIndex":2,"AssetAId":0,"AssetA":0,"AssetBId":0,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":0,"TreasuryAccountIndex":0,"TreasuryRate":0}', '{"PairIndex":2,"AssetAId":1,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (8, '2022-05-30 05:11:11.185861+00', '2022-05-30 05:11:11.185861+00', NULL, 12, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (9, '2022-05-30 05:11:11.201621+00', '2022-05-30 05:11:11.201621+00', NULL, 13, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (10, '2022-05-30 05:11:11.201621+00', '2022-05-30 05:11:11.201621+00', NULL, 13, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":2,"NftContentHash":"0c1293246402e53a2701d0d764bd3ce7ac9a2b82c8a1cd3c6c1d1ef3ec8076f2","NftL1TokenId":"0","NftL1Address":"0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE","CreatorTreasuryRate":0,"CollectionId":0}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (11, '2022-05-30 05:11:11.216757+00', '2022-05-30 05:11:11.216757+00', NULL, 14, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":1,"Balance":-100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (12, '2022-05-30 05:11:11.228114+00', '2022-05-30 05:11:11.228114+00', NULL, 15, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (13, '2022-05-30 05:11:11.228114+00', '2022-05-30 05:11:11.228114+00', NULL, 15, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":2,"NftContentHash":"0c1293246402e53a2701d0d764bd3ce7ac9a2b82c8a1cd3c6c1d1ef3ec8076f2","NftL1TokenId":"0","NftL1Address":"0x78C34ad5641aE34eDEc94dd463C61298070Ff7BE","CreatorTreasuryRate":0,"CollectionId":0}', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (14, '2022-05-30 05:13:27.061612+00', '2022-05-30 05:13:27.061612+00', NULL, 16, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (15, '2022-05-30 05:13:27.061612+00', '2022-05-30 05:13:27.061612+00', NULL, 16, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (16, '2022-05-30 05:13:27.061612+00', '2022-05-30 05:13:27.061612+00', NULL, 16, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (17, '2022-05-30 05:13:27.061612+00', '2022-05-30 05:13:27.061612+00', NULL, 16, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (18, '2022-05-30 05:13:27.117446+00', '2022-05-30 05:13:27.117446+00', NULL, 17, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999999900000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-10000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (19, '2022-05-30 05:13:27.117446+00', '2022-05-30 05:13:27.117446+00', NULL, 17, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999995000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (20, '2022-05-30 05:13:27.117446+00', '2022-05-30 05:13:27.117446+00', NULL, 17, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (21, '2022-05-30 05:13:27.140308+00', '2022-05-30 05:13:27.140308+00', NULL, 18, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989900000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 2, 0);
INSERT INTO "public"."tx_detail" VALUES (22, '2022-05-30 05:13:27.140308+00', '2022-05-30 05:13:27.140308+00', NULL, 18, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999990000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 2, 0);
INSERT INTO "public"."tx_detail" VALUES (23, '2022-05-30 05:13:27.140308+00', '2022-05-30 05:13:27.140308+00', NULL, 18, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999890000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0, 2, 0);
INSERT INTO "public"."tx_detail" VALUES (24, '2022-05-30 05:13:27.140308+00', '2022-05-30 05:13:27.140308+00', NULL, 18, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989800000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":100000,"OfferCanceledOrFinalized":0}', 3, 0, 2, 0);
INSERT INTO "public"."tx_detail" VALUES (25, '2022-05-30 05:13:27.140308+00', '2022-05-30 05:13:27.140308+00', NULL, 18, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '{"PairIndex":0,"AssetAId":0,"AssetA":100000,"AssetBId":2,"AssetB":100000,"LpAmount":100000,"KLast":10000000000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 4, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (26, '2022-05-30 05:13:27.140308+00', '2022-05-30 05:13:27.140308+00', NULL, 18, 0, 1, 0, 'treasury.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 5, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (27, '2022-05-30 05:13:27.140308+00', '2022-05-30 05:13:27.140308+00', NULL, 18, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":10000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 6, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (28, '2022-05-30 05:13:27.161495+00', '2022-05-30 05:13:27.161495+00', NULL, 19, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999885000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-100,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 3, 0);
INSERT INTO "public"."tx_detail" VALUES (29, '2022-05-30 05:13:27.161495+00', '2022-05-30 05:13:27.161495+00', NULL, 19, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989800000,"LpAmount":100000,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":99,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 3, 0);
INSERT INTO "public"."tx_detail" VALUES (30, '2022-05-30 05:13:27.161495+00', '2022-05-30 05:13:27.161495+00', NULL, 19, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989800099,"LpAmount":100000,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0, 3, 0);
INSERT INTO "public"."tx_detail" VALUES (31, '2022-05-30 05:13:27.161495+00', '2022-05-30 05:13:27.161495+00', NULL, 19, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":100000,"AssetBId":2,"AssetB":100000,"LpAmount":100000,"KLast":10000000000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '{"PairIndex":0,"AssetAId":0,"AssetA":-99,"AssetBId":2,"AssetB":100,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 3, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (32, '2022-05-30 05:13:27.161495+00', '2022-05-30 05:13:27.161495+00', NULL, 19, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 4, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (33, '2022-05-30 05:13:27.18028+00', '2022-05-30 05:13:27.18028+00', NULL, 20, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989795099,"LpAmount":100000,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":99,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 4, 0);
INSERT INTO "public"."tx_detail" VALUES (34, '2022-05-30 05:13:27.18028+00', '2022-05-30 05:13:27.18028+00', NULL, 20, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999884900,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":100,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 4, 0);
INSERT INTO "public"."tx_detail" VALUES (35, '2022-05-30 05:13:27.18028+00', '2022-05-30 05:13:27.18028+00', NULL, 20, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999885000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0, 4, 0);
INSERT INTO "public"."tx_detail" VALUES (36, '2022-05-30 05:13:27.18028+00', '2022-05-30 05:13:27.18028+00', NULL, 20, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989795198,"LpAmount":100000,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":-100,"OfferCanceledOrFinalized":0}', 3, 0, 4, 0);
INSERT INTO "public"."tx_detail" VALUES (37, '2022-05-30 05:13:27.18028+00', '2022-05-30 05:13:27.18028+00', NULL, 20, 0, 1, 0, 'treasury.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 4, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (38, '2022-05-30 05:13:27.18028+00', '2022-05-30 05:13:27.18028+00', NULL, 20, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":99901,"AssetBId":2,"AssetB":100100,"LpAmount":100000,"KLast":10000000000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '{"PairIndex":0,"AssetAId":0,"AssetA":-99,"AssetBId":2,"AssetB":-100,"LpAmount":-100,"KLast":9980200000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 5, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (39, '2022-05-30 05:13:27.18028+00', '2022-05-30 05:13:27.18028+00', NULL, 20, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":15000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 6, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (40, '2022-05-30 05:13:27.196946+00', '2022-05-30 05:13:27.196946+00', NULL, 21, 0, 4, 2, 'sher.legend', '0', '1', 0, -1, 5, 0);
INSERT INTO "public"."tx_detail" VALUES (41, '2022-05-30 05:13:27.196946+00', '2022-05-30 05:13:27.196946+00', NULL, 21, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999880000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 5, 0);
INSERT INTO "public"."tx_detail" VALUES (42, '2022-05-30 05:13:27.196946+00', '2022-05-30 05:13:27.196946+00', NULL, 21, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":20000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (43, '2022-05-30 05:13:27.214058+00', '2022-05-30 05:13:27.214058+00', NULL, 22, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999875000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 6, 1);
INSERT INTO "public"."tx_detail" VALUES (44, '2022-05-30 05:13:27.214058+00', '2022-05-30 05:13:27.214058+00', NULL, 22, 2, 1, 3, 'gavin.legend', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (45, '2022-05-30 05:13:27.214058+00', '2022-05-30 05:13:27.214058+00', NULL, 22, 1, 3, 3, 'gavin.legend', '{"NftIndex":1,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 2, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (46, '2022-05-30 05:13:27.214058+00', '2022-05-30 05:13:27.214058+00', NULL, 22, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":25000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (47, '2022-05-30 05:13:27.236696+00', '2022-05-30 05:13:27.236696+00', NULL, 23, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (48, '2022-05-30 05:13:27.236696+00', '2022-05-30 05:13:27.236696+00', NULL, 23, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (49, '2022-05-30 05:13:27.236696+00', '2022-05-30 05:13:27.236696+00', NULL, 23, 1, 3, 2, 'sher.legend', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":2,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 2, -1, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (50, '2022-05-30 05:13:27.236696+00', '2022-05-30 05:13:27.236696+00', NULL, 23, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (51, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (52, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000095000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-10000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (53, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000085000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":1}', 2, 1, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (54, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989790198,"LpAmount":99900,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":9800,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (55, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":1}', 4, 2, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (56, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":1}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 5, 3, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (57, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 1, 3, -1, '', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":2,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 6, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (58, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":10000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":200,"LpAmount":0,"OfferCanceledOrFinalized":0}', 7, 4, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (59, '2022-05-30 05:13:27.259222+00', '2022-05-30 05:13:27.259222+00', NULL, 24, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":10200,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 8, 4, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (60, '2022-05-30 05:13:27.281609+00', '2022-05-30 05:13:27.281609+00', NULL, 25, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999870000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 8, 1);
INSERT INTO "public"."tx_detail" VALUES (61, '2022-05-30 05:13:27.281609+00', '2022-05-30 05:13:27.281609+00', NULL, 25, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":1}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":3}', 1, 0, 8, 1);
INSERT INTO "public"."tx_detail" VALUES (62, '2022-05-30 05:13:27.281609+00', '2022-05-30 05:13:27.281609+00', NULL, 25, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":30000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (63, '2022-05-30 05:13:27.296112+00', '2022-05-30 05:13:27.296112+00', NULL, 26, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000085000,"LpAmount":0,"OfferCanceledOrFinalized":1}', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (64, '2022-05-30 05:13:27.296112+00', '2022-05-30 05:13:27.296112+00', NULL, 26, 1, 3, -1, '', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"09bbce304f023e7beb641fe5b155083edccdca34234e746332074eeb0fdf07d1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', '{"NftIndex":1,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (65, '2022-05-30 05:13:27.296112+00', '2022-05-30 05:13:27.296112+00', NULL, 26, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":3}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 9, 1);
INSERT INTO "public"."tx_detail" VALUES (66, '2022-05-30 05:13:27.296112+00', '2022-05-30 05:13:27.296112+00', NULL, 26, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":15200,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 0, 0);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."account_history_id_seq"
OWNED BY "public"."account_history"."id";
SELECT setval('"public"."account_history_id_seq"', 41, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."account_id_seq"
OWNED BY "public"."account"."id";
SELECT setval('"public"."account_id_seq"', 5, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."block_id_seq"
OWNED BY "public"."block"."id";
SELECT setval('"public"."block_id_seq"', 28, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."fail_tx_id_seq"
OWNED BY "public"."fail_tx"."id";
SELECT setval('"public"."fail_tx_id_seq"', 2, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l1_amount_id_seq"
OWNED BY "public"."l1_amount"."id";
SELECT setval('"public"."l1_amount_id_seq"', 2, false);

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
SELECT setval('"public"."l1_tx_sender_id_seq"', 2, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_asset_info_id_seq"
OWNED BY "public"."l2_asset_info"."id";
SELECT setval('"public"."l2_asset_info_id_seq"', 4, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_block_event_monitor_id_seq"
OWNED BY "public"."l2_block_event_monitor"."id";
SELECT setval('"public"."l2_block_event_monitor_id_seq"', 2, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_collection_id_seq"
OWNED BY "public"."l2_nft_collection"."id";
SELECT setval('"public"."l2_nft_collection_id_seq"', 2, false);

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
SELECT setval('"public"."l2_nft_exchange_id_seq"', 2, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_history_id_seq"
OWNED BY "public"."l2_nft_history"."id";
SELECT setval('"public"."l2_nft_history_id_seq"', 7, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_id_seq"
OWNED BY "public"."l2_nft"."id";
SELECT setval('"public"."l2_nft_id_seq"', 3, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_withdraw_history_id_seq"
OWNED BY "public"."l2_nft_withdraw_history"."id";
SELECT setval('"public"."l2_nft_withdraw_history_id_seq"', 3, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_tx_event_monitor_id_seq"
OWNED BY "public"."l2_tx_event_monitor"."id";
SELECT setval('"public"."l2_tx_event_monitor_id_seq"', 16, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."liquidity_history_id_seq"
OWNED BY "public"."liquidity_history"."id";
SELECT setval('"public"."liquidity_history_id_seq"', 8, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."liquidity_id_seq"
OWNED BY "public"."liquidity"."id";
SELECT setval('"public"."liquidity_id_seq"', 4, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."mempool_tx_detail_id_seq"
OWNED BY "public"."mempool_tx_detail"."id";
SELECT setval('"public"."mempool_tx_detail_id_seq"', 67, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."mempool_tx_id_seq"
OWNED BY "public"."mempool_tx"."id";
SELECT setval('"public"."mempool_tx_id_seq"', 27, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."offer_id_seq"
OWNED BY "public"."offer"."id";
SELECT setval('"public"."offer_id_seq"', 2, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."proof_sender_id_seq"
OWNED BY "public"."proof_sender"."id";
SELECT setval('"public"."proof_sender_id_seq"', 2, false);

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
SELECT setval('"public"."tx_detail_id_seq"', 67, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."tx_id_seq"
OWNED BY "public"."tx"."id";
SELECT setval('"public"."tx_id_seq"', 27, true);

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
