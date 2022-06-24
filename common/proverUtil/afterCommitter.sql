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

 Date: 09/06/2022 15:34:20
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
INSERT INTO "public"."account" VALUES (1, '0001-01-01 00:00:00+00', '2022-06-09 07:09:06.228049+00', NULL, 0, 'treasury.legend', 'fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998', '167c5363088a40a4839912a872f43164270740c7e986ec55397b2d583317ab4a', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 0, 0, '{"0":{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 1);
INSERT INTO "public"."account" VALUES (3, '0001-01-01 00:00:00+00', '2022-06-09 07:09:06.296772+00', NULL, 2, 'sher.legend', 'b0b6f7466154578ec66d51a335ead65ffd6a7210567fad9e68b6df8a5ce5dd85', '214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 9, 1, '{"0":{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":3},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999865000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135f35977f2abf9cb4029cc418b45ba79cc45cf39685be661c67da75ead45d9a', 1);
INSERT INTO "public"."account" VALUES (2, '0001-01-01 00:00:00+00', '2022-06-09 07:09:06.295772+00', NULL, 1, 'gas.legend', '1ec94e497abe0fbb87f9ed2843e21163e17e3e97f6bbbae7a88399b826474f93', '0a48e9892a45a04d0c5b0f235a3aeb07b92137ba71a59b9c457774bafde95983', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 0, 0, '{"0":{"AssetId":0,"Balance":20200,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":35000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '219d2d2c0bb8cba744ec53ea8388da6c961b555f62bd5aa290e97109d186c467', 1);
INSERT INTO "public"."account" VALUES (4, '0001-01-01 00:00:00+00', '2022-06-09 07:09:06.296273+00', NULL, 3, 'gavin.legend', '0500ccea3ca064968f5292b850ac8d4d3ee48d499357351a5ebfa2f30bb6070e', '1c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', 2, 0, '{"0":{"AssetId":0,"Balance":100000000000080000,"LpAmount":0,"OfferCanceledOrFinalized":1},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '20e11089ec56b54159ea65fc328d75c7b15011b11f5c73653073ddd0bdf1423e', 1);

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
INSERT INTO "public"."account_history" VALUES (1, '2022-06-09 07:09:06.051241+00', '2022-06-09 07:09:06.051241+00', NULL, 0, 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 1);
INSERT INTO "public"."account_history" VALUES (2, '2022-06-09 07:09:06.060525+00', '2022-06-09 07:09:06.060525+00', NULL, 1, 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 2);
INSERT INTO "public"."account_history" VALUES (3, '2022-06-09 07:09:06.069319+00', '2022-06-09 07:09:06.069319+00', NULL, 2, 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 3);
INSERT INTO "public"."account_history" VALUES (4, '2022-06-09 07:09:06.076707+00', '2022-06-09 07:09:06.076707+00', NULL, 3, 0, 0, '{}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 4);
INSERT INTO "public"."account_history" VALUES (5, '2022-06-09 07:09:06.085616+00', '2022-06-09 07:09:06.085616+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '06ce582922720755debe04d60415a9c28bc4e788d012d3ea1700549f0e190c9a', 5);
INSERT INTO "public"."account_history" VALUES (6, '2022-06-09 07:09:06.093199+00', '2022-06-09 07:09:06.093199+00', NULL, 3, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '06ce582922720755debe04d60415a9c28bc4e788d012d3ea1700549f0e190c9a', 6);
INSERT INTO "public"."account_history" VALUES (7, '2022-06-09 07:09:06.099703+00', '2022-06-09 07:09:06.099703+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '069e6e659595ff010898f90e61614c5a8d77de0d2984715be6f5b3f8505ae10c', 7);
INSERT INTO "public"."account_history" VALUES (8, '2022-06-09 07:09:06.10681+00', '2022-06-09 07:09:06.10681+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '17d8b1c33a32922ce0838eed568beb32728b6271d97e008e481edd92cca55f08', 8);
INSERT INTO "public"."account_history" VALUES (9, '2022-06-09 07:09:06.14888+00', '2022-06-09 07:09:06.14888+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '17d8b1c33a32922ce0838eed568beb32728b6271d97e008e481edd92cca55f08', 13);
INSERT INTO "public"."account_history" VALUES (10, '2022-06-09 07:09:06.157966+00', '2022-06-09 07:09:06.157966+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '09fe19fc526b3e67753a6d91cc709feb45f6b281f6a3a71773a0abebe50f517f', 14);
INSERT INTO "public"."account_history" VALUES (11, '2022-06-09 07:09:06.165729+00', '2022-06-09 07:09:06.165729+00', NULL, 2, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '09fe19fc526b3e67753a6d91cc709feb45f6b281f6a3a71773a0abebe50f517f', 15);
INSERT INTO "public"."account_history" VALUES (12, '2022-06-09 07:09:06.18161+00', '2022-06-09 07:09:06.18161+00', NULL, 2, 1, 0, '{"0":{"AssetId":0,"Balance":99999999999900000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999995000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1cd1016e23d9e514928a567cc8a4cddcce67b01e817021eb916edaac7e166242', 16);
INSERT INTO "public"."account_history" VALUES (13, '2022-06-09 07:09:06.18161+00', '2022-06-09 07:09:06.18161+00', NULL, 3, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000100000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '14012ade6c7b76679cc709bbc6fe865ec94f0b55a7c976a9b93d8e214f2bf5e5', 16);
INSERT INTO "public"."account_history" VALUES (14, '2022-06-09 07:09:06.18161+00', '2022-06-09 07:09:06.18161+00', NULL, 1, 0, 0, '{"2":{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '08e7c9a1858f6ad9986887426fdddc7231a93a806c81b5841171ec5cb834eabe', 16);
INSERT INTO "public"."account_history" VALUES (15, '2022-06-09 07:09:06.192937+00', '2022-06-09 07:09:06.192937+00', NULL, 2, 2, 0, '{"0":{"AssetId":0,"Balance":99999999989900000,"LpAmount":0,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999990000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '25cc5a90b005abb6c7c0d5d1fbd34907b12c70b7d2f11a6901cd5622186e584e', 17);
INSERT INTO "public"."account_history" VALUES (16, '2022-06-09 07:09:06.192937+00', '2022-06-09 07:09:06.192937+00', NULL, 1, 0, 0, '{"2":{"AssetId":2,"Balance":10000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1e6cf281636a0d207da108b38aaada12c903f0f7531b3e60ff935675b9d64644', 17);
INSERT INTO "public"."account_history" VALUES (17, '2022-06-09 07:09:06.206169+00', '2022-06-09 07:09:06.206169+00', NULL, 2, 3, 0, '{"0":{"AssetId":0,"Balance":99999999989800000,"LpAmount":100000,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999885000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '3046d6422f86f1ab6e9cbe2a0e449604df61bbdd3f3199fadd7a5bc4f046d289', 18);
INSERT INTO "public"."account_history" VALUES (18, '2022-06-09 07:09:06.206169+00', '2022-06-09 07:09:06.206169+00', NULL, 0, 0, 0, '{"0":{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 18);
INSERT INTO "public"."account_history" VALUES (19, '2022-06-09 07:09:06.206169+00', '2022-06-09 07:09:06.206169+00', NULL, 1, 0, 0, '{"2":{"AssetId":2,"Balance":15000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '12aeb69e38371c4ef60475f6d1d5bd15fb602de9a3ac9d8ce98cb11b95685bee', 18);
INSERT INTO "public"."account_history" VALUES (20, '2022-06-09 07:09:06.217449+00', '2022-06-09 07:09:06.217449+00', NULL, 2, 4, 0, '{"0":{"AssetId":0,"Balance":99999999989795099,"LpAmount":100000,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999884900,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '136b21d7d137ada052e748f45719da132e2a344e5ae6fe334d0b67012c331d6d', 19);
INSERT INTO "public"."account_history" VALUES (21, '2022-06-09 07:09:06.217449+00', '2022-06-09 07:09:06.217449+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":15000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '0a8e87b9a27934661653c3d37ea4b6b9cb7257d23d0ec85e0a77b0c62f6ca453', 19);
INSERT INTO "public"."account_history" VALUES (22, '2022-06-09 07:09:06.230155+00', '2022-06-09 07:09:06.230155+00', NULL, 2, 5, 0, '{"0":{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999880000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '13f7575c9228694a34eaec2e080115ac5cc06ec248abb0e5f9cdb7151b9acedd', 20);
INSERT INTO "public"."account_history" VALUES (23, '2022-06-09 07:09:06.230155+00', '2022-06-09 07:09:06.230155+00', NULL, 0, 0, 0, '{"0":{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2c642dc4ac8b021154b4248c4ab4a0b0fbcfebc1557ecc218fc3a3c19ece7f47', 20);
INSERT INTO "public"."account_history" VALUES (24, '2022-06-09 07:09:06.230155+00', '2022-06-09 07:09:06.230155+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":20000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1b0c02b49e7d799975e98665fc0f2062251e7e295001f43b0fc5013360d9f3cf', 20);
INSERT INTO "public"."account_history" VALUES (25, '2022-06-09 07:09:06.239599+00', '2022-06-09 07:09:06.239599+00', NULL, 2, 6, 1, '{"0":{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999875000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2e7c81f2815d8f11d39097bb1a5eb9f10d3ba9b7b56c0ac0f7a49c3eba397579', 21);
INSERT INTO "public"."account_history" VALUES (26, '2022-06-09 07:09:06.239599+00', '2022-06-09 07:09:06.239599+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":25000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '02f3efe09b203142ae196d4555d58e060da742fa15d9213fe26e75d8c5505539', 21);
INSERT INTO "public"."account_history" VALUES (27, '2022-06-09 07:09:06.250069+00', '2022-06-09 07:09:06.250069+00', NULL, 2, 7, 1, '{"0":{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999870000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135175067e21e4d0a1ec1f01d1eaacbb65a0ec3df762bf586df1c49a3a554e6d', 22);
INSERT INTO "public"."account_history" VALUES (28, '2022-06-09 07:09:06.250069+00', '2022-06-09 07:09:06.250069+00', NULL, 3, 0, 0, '{"0":{"AssetId":0,"Balance":100000000000100000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '14012ade6c7b76679cc709bbc6fe865ec94f0b55a7c976a9b93d8e214f2bf5e5', 22);
INSERT INTO "public"."account_history" VALUES (29, '2022-06-09 07:09:06.250069+00', '2022-06-09 07:09:06.250069+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":30000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2b264f5337dc9d06629ff7099ad6e0653eb3cdf7056dd2cd46752d50c1050b93', 22);
INSERT INTO "public"."account_history" VALUES (30, '2022-06-09 07:09:06.261097+00', '2022-06-09 07:09:06.261097+00', NULL, 3, 1, 0, '{"0":{"AssetId":0,"Balance":100000000000095000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2e2137673dbe998c6dce6b1555760686b57af98a8c23820337ef881703f534d2', 23);
INSERT INTO "public"."account_history" VALUES (31, '2022-06-09 07:09:06.261097+00', '2022-06-09 07:09:06.261097+00', NULL, 2, 7, 1, '{"0":{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999870000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135175067e21e4d0a1ec1f01d1eaacbb65a0ec3df762bf586df1c49a3a554e6d', 23);
INSERT INTO "public"."account_history" VALUES (32, '2022-06-09 07:09:06.261097+00', '2022-06-09 07:09:06.261097+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":10000,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":30000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '0ade8414224cd97b2841b34519db998c7873d2c386e87fcc93ac94f056424b9a', 23);
INSERT INTO "public"."account_history" VALUES (33, '2022-06-09 07:09:06.274718+00', '2022-06-09 07:09:06.274718+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":15200,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":30000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '2118e6f94c540e6f3676a6cbae245679a59432115d48f7571a6ae5565edca611', 24);
INSERT INTO "public"."account_history" VALUES (34, '2022-06-09 07:09:06.274718+00', '2022-06-09 07:09:06.274718+00', NULL, 2, 8, 1, '{"0":{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":1},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999870000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1a9e697d30b8b2f43aedc7251497f4a4308f221089f43c8794d8966d5bbf4769', 24);
INSERT INTO "public"."account_history" VALUES (35, '2022-06-09 07:09:06.274718+00', '2022-06-09 07:09:06.274718+00', NULL, 3, 1, 0, '{"0":{"AssetId":0,"Balance":100000000000085000,"LpAmount":0,"OfferCanceledOrFinalized":1},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '06d9b9fd9b2b3ab3ea7c44de16deef25a7dea1314ec20c5b92dcc1049d221f49', 24);
INSERT INTO "public"."account_history" VALUES (36, '2022-06-09 07:09:06.286505+00', '2022-06-09 07:09:06.286505+00', NULL, 2, 9, 1, '{"0":{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":3},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999865000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135f35977f2abf9cb4029cc418b45ba79cc45cf39685be661c67da75ead45d9a', 25);
INSERT INTO "public"."account_history" VALUES (37, '2022-06-09 07:09:06.286505+00', '2022-06-09 07:09:06.286505+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":15200,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":35000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '1a94962e42cbd751dc8fb4975ab18ee52493c4fd400d94fb1065719e24d019f3', 25);
INSERT INTO "public"."account_history" VALUES (38, '2022-06-09 07:09:06.297928+00', '2022-06-09 07:09:06.297928+00', NULL, 1, 0, 0, '{"0":{"AssetId":0,"Balance":20200,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":35000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '219d2d2c0bb8cba744ec53ea8388da6c961b555f62bd5aa290e97109d186c467', 26);
INSERT INTO "public"."account_history" VALUES (39, '2022-06-09 07:09:06.297928+00', '2022-06-09 07:09:06.297928+00', NULL, 3, 2, 0, '{"0":{"AssetId":0,"Balance":100000000000080000,"LpAmount":0,"OfferCanceledOrFinalized":1},"2":{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '20e11089ec56b54159ea65fc328d75c7b15011b11f5c73653073ddd0bdf1423e', 26);
INSERT INTO "public"."account_history" VALUES (40, '2022-06-09 07:09:06.297928+00', '2022-06-09 07:09:06.297928+00', NULL, 2, 9, 1, '{"0":{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":3},"1":{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0},"2":{"AssetId":2,"Balance":99999999999999865000,"LpAmount":0,"OfferCanceledOrFinalized":0}}', '135f35977f2abf9cb4029cc418b45ba79cc45cf39685be661c67da75ead45d9a', 26);

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
INSERT INTO "public"."block" VALUES (2, '2022-06-09 07:09:06.039+00', '2022-06-09 07:09:06.042917+00', NULL, '1a72f54f1286faefd0f05a774d75a9fc14a981226b52f93af7a8301bfaa1a9dd', 1, '21422f9bebac15af8ddc504da0dbb88020c1a4de7e7b6722fe00acb0ed968942', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (3, '2022-06-09 07:09:06.053+00', '2022-06-09 07:09:06.056523+00', NULL, '044d5308e567d8490b58f1416c701f308a45b886c99da304ac6f7cd7c02de1ae', 2, '1b2ff4ae0d507a971fb267849af6a28000b1d483865c5a610cc47db6f196c672', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (4, '2022-06-09 07:09:06.062+00', '2022-06-09 07:09:06.065022+00', NULL, '0927fa92c98b9d94d8b6af2855f9a09ee9562274540826f81d3820ca72538dfe', 3, '189517f4cfb59471e3539dae36b8f53cb1264d407daf6afbf86132917f1cbafc', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (5, '2022-06-09 07:09:06.07+00', '2022-06-09 07:09:06.073155+00', NULL, '242a431dd79dc30695f5f53f7f995444dd0fe97044dafe0d00086b54db844576', 4, '08b2dc20da16235e692de317d6134578159532d4f081827bd29a5fc783fcc2b7', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (6, '2022-06-09 07:09:06.078+00', '2022-06-09 07:09:06.079383+00', NULL, '1ca7efbd17cc00f4793cc499eadc87380c8b30b3a2a50184f48f850dc969ed2b', 5, '236e2c312a52cfbe96fc14a0693ea0f26d59fae774b35d44ddcf7737d965902f', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (7, '2022-06-09 07:09:06.087+00', '2022-06-09 07:09:06.088388+00', NULL, '181982d4f80f4a56b6df25550961a5d471eefa8db9854d79234f087c8912aad7', 6, '029cfe1c99565d3722f32b6bdb4ee5740d4f4c78bf318968c366c9c7e82d9ba7', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (8, '2022-06-09 07:09:06.094+00', '2022-06-09 07:09:06.095702+00', NULL, '21337368ad6b2c7fde547cdb84bbeaa286d26ed43987bb21f217f10ac346f196', 7, '25cade17a4affef4114a06b8ae6e8e18651a8c4aa0aa01e1c20abce23ad614ec', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (9, '2022-06-09 07:09:06.101+00', '2022-06-09 07:09:06.102555+00', NULL, '254dd9634abd61d13c24f435ad04d6d089b91aef10a8e68e6bc7fe08d1ad4768', 8, '17a21620fe89a6ef610ceea7b2c6230dba84731020a11bd081b46ba23c1cae94', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (10, '2022-06-09 07:09:06.108+00', '2022-06-09 07:09:06.110587+00', NULL, '05439be16da134ebfb53da4aea2088c97a23b7f88569db1ef8ef44266198925c', 9, '0f5cf7c3fa8452ccb12d87b99952cfde059999f3767ddbc032994d94f3fe24ba', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (11, '2022-06-09 07:09:06.118+00', '2022-06-09 07:09:06.119587+00', NULL, '2cd16cf22d1c33999e7ef4706413d8f93a60123dac8823995dc8991f06357eda', 10, '0945597849e7df9b43bfade724068c4d5a9d6039da208e6b829feb530ce784cd', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (12, '2022-06-09 07:09:06.125+00', '2022-06-09 07:09:06.127124+00', NULL, '12fc02c4e112c2062a416d8443e2a79b9425fc1984e23d66bd4384de288f3f4e', 11, '1671dd749a5a522f18908e28512d1c6c10034740923bbe9bab5664585b87411d', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (13, '2022-06-09 07:09:06.132+00', '2022-06-09 07:09:06.133802+00', NULL, '114681e463929e795b6143ef83e51c19d74d54b1f4dbc224ddb67c7cecee5e9c', 12, '08ef9af5048b3df61fe3bd025a8db3f47b591a0136281cb3325e7e89930f3925', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (14, '2022-06-09 07:09:06.139+00', '2022-06-09 07:09:06.143066+00', NULL, '2ebcf44ebde73f0a5f18b0691ce757580ad3b5a6a45ee89b4224d90f88c54936', 13, '23f9301b57dbde40b067fc04f2bb2e5241b58739845efa223de352a8a14dd2ae', 1, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (15, '2022-06-09 07:09:06.152+00', '2022-06-09 07:09:06.153338+00', NULL, '1f7e314a0141f18fe8ff32edae3bfe3204f435a35258be0c1642859858813d90', 14, '1785a0c0ef9c282c5dddde78ad80b9689d34cda4a59ed35fcf4f00966ff034e2', 1, '25519b0462cdac4689dce03a73cb3ac5d1200dc9ab3a43f3d0da8e2064c92205', '["EQAAAAIAAQAAAAAAAAAFa8deLWMQAAAAAAAAAAAAAAAhSi168gIt+u5J2tuJktPXwiXYrjYQm1McKEBt1pqtRQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"]', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (16, '2022-06-09 07:09:06.159+00', '2022-06-09 07:09:06.161147+00', NULL, '115a7ed40790d5c0f8faa54d867d07ea41652a613b1e8bb4743229326f1832a7', 15, '28ff96ba5f7e023a7ed9d446cb412fc6965a6ed68d1439b357bb4014ec57a8a4', 1, 'e557d59f7ebc4acb5a9d4de3fe645c08b591d83170bd01c80e7e5335f92e8450', '["EgAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAC3rUp+lFnQwVQdsu7OzqzH26gD4SFKLXryAi367kna24mS09fCJdiuNhCbUxwoQG3Wmq1FAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACr0baueVB/e0oyqEq2SVvJ/uZ0UO0xbbuna6zoo8UZewAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"]', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (17, '2022-06-09 07:09:06.171+00', '2022-06-09 07:09:06.174649+00', NULL, '08043522804cec1b46b7a3eae1afd4f577d00e8ea12fc9db6ec37a65f77eb290', 16, '0c599d212ed3641e0b6df735e8b04dd627accfdafbbfa38c173af5f38efb433e', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (18, '2022-06-09 07:09:06.184+00', '2022-06-09 07:09:06.186387+00', NULL, '2f8a924c6a3e9412341c3151fbf3e0e4ccb928926d80108796b7760c4e58ffea', 17, '2d425cddc3d5aaec5ed0dd1465e7038e7fbbcc679e4a4c6c620742134ef93714', 0, 'a213a63e48ec6e59019927c20dbeda00195c7a95e68a25d8dd25db55cfc6fd03', '["CgAAAAKZrIiBg0eX68MvGF7ifC6WhC4aRwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACYloAAAAABAAI+gQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"]', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (19, '2022-06-09 07:09:06.195+00', '2022-06-09 07:09:06.199381+00', NULL, '20d1b0f785f43e91a989bc29b6fb417bc0f9c3a1de3029615adee054e00271b3', 18, '1211d91f4e22bd2f1aa38daaec68431b35fd37f8a272d147ebe7ba3e73a58555', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (20, '2022-06-09 07:09:06.209+00', '2022-06-09 07:09:06.211937+00', NULL, '0e00c0df1a6d31d4c9da8845b419c1a63ca17e96374cfbf25cfd6649f4a708be', 19, '1db7fb69796667194858edf7aea403110c42cddc0907b3953181e1184907fb35', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (21, '2022-06-09 07:09:06.22+00', '2022-06-09 07:09:06.224079+00', NULL, '01b9b434b824c6ef586c29c2d210550cd22f163c35ef28c6287aa8fe2f724fb1', 20, '2e888850863cf0c2dffa40c8a0c162749f1f93da6ddf225030a01648cdfc26d6', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (22, '2022-06-09 07:09:06.233+00', '2022-06-09 07:09:06.234755+00', NULL, '0ece5e1e234d032f797f8c7002aa705d0cf2279a12ac08c7c907abe1384faa34', 21, '140622efbca882cddada16ee07f9cc8718b69998a5d8d7922fa7591f2f533edd', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (23, '2022-06-09 07:09:06.241+00', '2022-06-09 07:09:06.244865+00', NULL, '1e7e2fd098f75e8b5e8ac4714b5b467161cc9a8eb85d5c8c9d23dbd1efa96b28', 22, '12884f8bb4852d02ad1f654daa7a2fc230c539e5a1d3dcd83a474eed139e1f7f', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (24, '2022-06-09 07:09:06.252+00', '2022-06-09 07:09:06.255441+00', NULL, '09bf02a6510b15e9e22892aa8c1f6a509468d718fe7305a2a1f5013be4357800', 23, '19ca2bf9cca9b55f61c3f2d352ab486ac7e529670b6af93171054cb8d82f4fee', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (25, '2022-06-09 07:09:06.264+00', '2022-06-09 07:09:06.268598+00', NULL, '080e4feab8e0f5992a1ae7f65956ca02032377c3ecc5275601e8207dd5c5268b', 24, '137f5a5193ca65babef27c7f5be3ebc0eee3fd5c4de748d4cc0e736b80a99649', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (26, '2022-06-09 07:09:06.277+00', '2022-06-09 07:09:06.280578+00', NULL, '12380990e61da98f07e1997c6377f6ab6ae2e240882f97d2b7670636a001c332', 25, '16033680a98409353095c6679b48d1fe06a03ec709b55e448a4c4a56e229393e', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 1);
INSERT INTO "public"."block" VALUES (27, '2022-06-09 07:09:06.288+00', '2022-06-09 07:09:06.292098+00', NULL, '004030d6fc1ab25f9aad69781c2af86d779af098cedd06729e53f4786f44ecd4', 26, '278d08c3c1a50ed6e932abdfde1555b7843c43de10c1fded32f7cfc2987c9105', 0, '2e0da5cc3106784395dd76ba3f496389e9c033b84d523b96c1c30f011cd9d219', '["EAAAAAMAAAACAAAAAAAAAQABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADVqjtWouITnbMVzf47NBScjtCRcQAAAAEAAD6BBmpl0+Q5etBfsuf1DqwWBkenSGws7bvqxkYkyL7qIvEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACFKLXryAi367kna24mS09fCJdiuNhCbUxwoQG3Wmq1F"]', '', 0, '', 0, 1);

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
INSERT INTO "public"."block_for_commit" VALUES (1, '2022-06-09 07:09:06.046519+00', '2022-06-09 07:09:06.046519+00', NULL, 1, '21422f9bebac15af8ddc504da0dbb88020c1a4de7e7b6722fe00acb0ed968942', '01000000000000000000000000000000000000000000000000000000000000007472656173757279000000000000000000000000000000000000000000000000167c5363088a40a4839912a872f43164270740c7e986ec55397b2d583317ab4a2005db7af2bdcfae1fa8d28833ae2f1995e9a8e0825377cff121db64b0db21b718a96ca582a72b16f464330c89ab73277cb96e42df105ebf5c9ac5330d47b8fc0000000000000000000000000000000000000000000000000000000000000000', 1654758546039, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (2, '2022-06-09 07:09:06.057578+00', '2022-06-09 07:09:06.057578+00', NULL, 2, '1b2ff4ae0d507a971fb267849af6a28000b1d483865c5a610cc47db6f196c672', '010000000100000000000000000000000000000000000000000000000000000067617300000000000000000000000000000000000000000000000000000000000a48e9892a45a04d0c5b0f235a3aeb07b92137ba71a59b9c457774bafde959832c24415b75651673b0d7bbf145ac8d7cb744ba6926963d1d014836336df1317a134f4726b89983a8e7babbf6973e7ee16311e24328edf987bb0fbe7a494ec91e0000000000000000000000000000000000000000000000000000000000000000', 1654758546053, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (3, '2022-06-09 07:09:06.066523+00', '2022-06-09 07:09:06.066523+00', NULL, 3, '189517f4cfb59471e3539dae36b8f53cb1264d407daf6afbf86132917f1cbafc', '01000000020000000000000000000000000000000000000000000000000000007368657200000000000000000000000000000000000000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45235fdbbbf5ef1665f3422211702126433c909487c456e594ef3a56910810396a05dde55c8adfb6689ead7f5610726afd5fd6ea35a3516dc68e57546146f7b6b00000000000000000000000000000000000000000000000000000000000000000', 1654758546062, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (4, '2022-06-09 07:09:06.074432+00', '2022-06-09 07:09:06.074432+00', NULL, 4, '08b2dc20da16235e692de317d6134578159532d4f081827bd29a5fc783fcc2b7', '0100000003000000000000000000000000000000000000000000000000000000676176696e0000000000000000000000000000000000000000000000000000001c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb0649fef47f6cf3dfb767cf5599eea11677bb6495956ec4cf75707d3aca7c06ed0e07b60bf3a2bf5e1a355793498de43e4d8dac50b892528f9664a03ceacc00050000000000000000000000000000000000000000000000000000000000000000', 1654758546070, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (5, '2022-06-09 07:09:06.081888+00', '2022-06-09 07:09:06.081888+00', NULL, 5, '236e2c312a52cfbe96fc14a0693ea0f26d59fae774b35d44ddcf7737d965902f', '040000000200000000000000000000016345785d8a0000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad450000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546078, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (6, '2022-06-09 07:09:06.090131+00', '2022-06-09 07:09:06.090131+00', NULL, 6, '029cfe1c99565d3722f32b6bdb4ee5740d4f4c78bf318968c366c9c7e82d9ba7', '040000000300000000000000000000016345785d8a00000000000000000000001c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546087, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (7, '2022-06-09 07:09:06.097203+00', '2022-06-09 07:09:06.097203+00', NULL, 7, '25cade17a4affef4114a06b8ae6e8e18651a8c4aa0aa01e1c20abce23ad614ec', '0400000002000100000000000000056bc75e2d63100000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad450000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546094, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (8, '2022-06-09 07:09:06.10415+00', '2022-06-09 07:09:06.10415+00', NULL, 8, '17a21620fe89a6ef610ceea7b2c6230dba84731020a11bd081b46ba23c1cae94', '0400000002000200000000000000056bc75e2d63100000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad450000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546101, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (9, '2022-06-09 07:09:06.112169+00', '2022-06-09 07:09:06.112169+00', NULL, 9, '0f5cf7c3fa8452ccb12d87b99952cfde059999f3767ddbc032994d94f3fe24ba', '02000000000002001e000000000005000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546108, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (10, '2022-06-09 07:09:06.121441+00', '2022-06-09 07:09:06.121441+00', NULL, 10, '0945597849e7df9b43bfade724068c4d5a9d6039da208e6b829feb530ce784cd', '02000100000001001e000000000005000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546118, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (11, '2022-06-09 07:09:06.128738+00', '2022-06-09 07:09:06.128738+00', NULL, 11, '1671dd749a5a522f18908e28512d1c6c10034740923bbe9bab5664585b87411d', '02000200010002001e000000000005000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546125, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (12, '2022-06-09 07:09:06.135375+00', '2022-06-09 07:09:06.135375+00', NULL, 12, '08ef9af5048b3df61fe3bd025a8db3f47b591a0136281cb3325e7e89930f3925', '030001003200000000000a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546132, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (13, '2022-06-09 07:09:06.145391+00', '2022-06-09 07:09:06.145391+00', NULL, 13, '23f9301b57dbde40b067fc04f2bb2e5241b58739845efa223de352a8a14dd2ae', '05000000020000000000b7ad4a7e9459d0c1541db2eececeacc7dba803e100000000000000000000000000000000000000000000000000000000000000000000abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b0000000000000000000000000000000000000000000000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad450000000000000000000000000000000000000000000000000000000000000000', 1654758546139, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (14, '2022-06-09 07:09:06.155182+00', '2022-06-09 07:09:06.155182+00', NULL, 14, '1785a0c0ef9c282c5dddde78ad80b9689d34cda4a59ed35fcf4f00966ff034e2', '1100000002000100000000000000056bc75e2d63100000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad450000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546152, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (15, '2022-06-09 07:09:06.16275+00', '2022-06-09 07:09:06.16275+00', NULL, 15, '28ff96ba5f7e023a7ed9d446cb412fc6965a6ed68d1439b357bb4014ec57a8a4', '1200000002000000000000000000000000000000000000000000000000000000000000000000000000000000b7ad4a7e9459d0c1541db2eececeacc7dba803e1214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad450000000000000000000000000000000000000000000000000000000000000000abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b0000000000000000000000000000000000000000000000000000000000000000', 1654758546159, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (16, '2022-06-09 07:09:06.177047+00', '2022-06-09 07:09:06.177047+00', NULL, 16, '0c599d212ed3641e0b6df735e8b04dd627accfdafbbfa38c173af5f38efb433e', '0600000002000000030000000030d4000000000100023e8100000000000000000dde7a022857fec1b8ffa7664a937a250d3ae68f356061754d3531e2674103d80000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546171, 'null');
INSERT INTO "public"."block_for_commit" VALUES (17, '2022-06-09 07:09:06.189386+00', '2022-06-09 07:09:06.189386+00', NULL, 17, '2d425cddc3d5aaec5ed0dd1465e7038e7fbbcc679e4a4c6c620742134ef93714', '0a0000000299ac8881834797ebc32f185ee27c2e96842e1a47000000000000000000000000000000000000000000000000000000009896800000000100023e810000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546184, '[0]');
INSERT INTO "public"."block_for_commit" VALUES (18, '2022-06-09 07:09:06.202172+00', '2022-06-09 07:09:06.202172+00', NULL, 18, '1211d91f4e22bd2f1aa38daaec68431b35fd37f8a272d147ebe7ba3e73a58555', '08000000020000000030d400000030d400000030d4004a817c800000000000000000000000000000000000000000000000000000000000000000000100023e810000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546195, 'null');
INSERT INTO "public"."block_for_commit" VALUES (19, '2022-06-09 07:09:06.214306+00', '2022-06-09 07:09:06.214306+00', NULL, 19, '1db7fb69796667194858edf7aea403110c42cddc0907b3953181e1184907fb35', '070000000200000000000c800000000c600000000100003e810000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546209, 'null');
INSERT INTO "public"."block_for_commit" VALUES (20, '2022-06-09 07:09:06.22603+00', '2022-06-09 07:09:06.22603+00', NULL, 20, '2e888850863cf0c2dffa40c8a0c162749f1f93da6ddf225030a01648cdfc26d6', '090000000200000000000c600000000c800000000c804a5bb8880000000000000000000000000000000000000000000000000000000000000000000100023e810000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546220, 'null');
INSERT INTO "public"."block_for_commit" VALUES (21, '2022-06-09 07:09:06.236752+00', '2022-06-09 07:09:06.236752+00', NULL, 21, '140622efbca882cddada16ee07f9cc8718b69998a5d8d7922fa7591f2f533edd', '0b0000000200010000000100023e81000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546233, 'null');
INSERT INTO "public"."block_for_commit" VALUES (22, '2022-06-09 07:09:06.246599+00', '2022-06-09 07:09:06.246599+00', NULL, 22, '12884f8bb4852d02ad1f654daa7a2fc230c539e5a1d3dcd83a474eed139e1f7f', '0c000000020000000300000000010000000100023e8100000001000000000000066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546241, 'null');
INSERT INTO "public"."block_for_commit" VALUES (23, '2022-06-09 07:09:06.257099+00', '2022-06-09 07:09:06.257099+00', NULL, 23, '19ca2bf9cca9b55f61c3f2d352ab486ac7e529670b6af93171054cb8d82f4fee', '0d000000030000000200000000010000000100003e81000000000000000000000dde7a022857fec1b8ffa7664a937a250d3ae68f356061754d3531e2674103d80000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546252, 'null');
INSERT INTO "public"."block_for_commit" VALUES (24, '2022-06-09 07:09:06.270873+00', '2022-06-09 07:09:06.270873+00', NULL, 24, '137f5a5193ca65babef27c7f5be3ebc0eee3fd5c4de748d4cc0e736b80a99649', '0e00000002000000030000000000000200000000000000010000000000000000000000000000000000000004e200000000000000000019000000000100003e810000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546264, 'null');
INSERT INTO "public"."block_for_commit" VALUES (25, '2022-06-09 07:09:06.282652+00', '2022-06-09 07:09:06.282652+00', NULL, 25, '16033680a98409353095c6679b48d1fe06a03ec709b55e448a4c4a56e229393e', '0f000000020000010000000100023e810000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000', 1654758546277, 'null');
INSERT INTO "public"."block_for_commit" VALUES (26, '2022-06-09 07:09:06.293926+00', '2022-06-09 07:09:06.293926+00', NULL, 26, '278d08c3c1a50ed6e932abdfde1555b7843c43de10c1fded32f7cfc2987c9105', '1000000003000000020000000000000100010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000d5aa3b56a2e2139db315cdfe3b34149c8ed091710000000100003e81066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f10000000000000000000000000000000000000000000000000000000000000000214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45', 1654758546288, '[0]');

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
INSERT INTO "public"."l2_nft" VALUES (1, '2022-06-08 08:25:58.056967+00', '2022-06-09 07:09:06.168045+00', NULL, 0, 0, 0, '0', '0', '0', 0, 0);
INSERT INTO "public"."l2_nft" VALUES (2, '2022-06-09 05:59:35.3716+00', '2022-06-09 07:09:06.299699+00', NULL, 1, 0, 0, '0', '0', '0', 0, 0);

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
INSERT INTO "public"."l2_nft_exchange" VALUES (1, '2022-06-09 05:59:50.16996+00', '2022-06-09 05:59:50.16996+00', NULL, 3, 2, 1, 0, '10000');

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
INSERT INTO "public"."l2_nft_history" VALUES (1, '2022-06-09 07:09:06.149881+00', '2022-06-09 07:09:06.149881+00', NULL, 0, 0, 2, 'abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b', '0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1', '0', 0, 0, 0, 13);
INSERT INTO "public"."l2_nft_history" VALUES (2, '2022-06-09 07:09:06.170047+00', '2022-06-09 07:09:06.170047+00', NULL, 0, 0, 0, '0', '0', '0', 0, 0, 0, 15);
INSERT INTO "public"."l2_nft_history" VALUES (3, '2022-06-09 07:09:06.251098+00', '2022-06-09 07:09:06.251098+00', NULL, 1, 2, 3, '066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1', '0', '0', 0, 1, 0, 22);
INSERT INTO "public"."l2_nft_history" VALUES (4, '2022-06-09 07:09:06.26284+00', '2022-06-09 07:09:06.26284+00', NULL, 1, 2, 2, '066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1', '0', '0', 0, 1, 0, 23);
INSERT INTO "public"."l2_nft_history" VALUES (5, '2022-06-09 07:09:06.276395+00', '2022-06-09 07:09:06.276395+00', NULL, 1, 2, 3, '066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1', '0', '0', 0, 1, 0, 24);
INSERT INTO "public"."l2_nft_history" VALUES (6, '2022-06-09 07:09:06.300854+00', '2022-06-09 07:09:06.300854+00', NULL, 1, 0, 0, '0', '0', '0', 0, 0, 0, 26);

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
INSERT INTO "public"."l2_nft_withdraw_history" VALUES (1, '2022-06-09 07:09:06.167039+00', '2022-06-09 07:09:06.167039+00', NULL, 0, 0, 2, 'abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b', '0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1', '0', 0, 0);
INSERT INTO "public"."l2_nft_withdraw_history" VALUES (2, '2022-06-09 07:09:06.298804+00', '2022-06-09 07:09:06.298804+00', NULL, 1, 2, 3, '066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1', '0', '0', 0, 1);

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
INSERT INTO "public"."liquidity" VALUES (3, '2022-06-08 08:25:58.055259+00', '2022-06-09 07:09:06.130441+00', NULL, 2, 1, '0', 2, '0', '0', '0', 30, 0, 5);
INSERT INTO "public"."liquidity" VALUES (2, '2022-06-08 08:25:58.055259+00', '2022-06-09 07:09:06.137001+00', NULL, 1, 0, '0', 1, '0', '0', '0', 50, 0, 10);
INSERT INTO "public"."liquidity" VALUES (1, '2022-06-08 08:25:58.055259+00', '2022-06-09 07:09:06.230752+00', NULL, 0, 0, '99802', 2, '100000', '99900', '9980200000', 30, 0, 5);

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
INSERT INTO "public"."liquidity_history" VALUES (1, '2022-06-09 07:09:06.115785+00', '2022-06-09 07:09:06.115785+00', NULL, 0, 0, '0', 2, '0', '0', '0', 30, 0, 5, 9);
INSERT INTO "public"."liquidity_history" VALUES (2, '2022-06-09 07:09:06.124009+00', '2022-06-09 07:09:06.124009+00', NULL, 1, 0, '0', 1, '0', '0', '0', 30, 0, 5, 10);
INSERT INTO "public"."liquidity_history" VALUES (3, '2022-06-09 07:09:06.131379+00', '2022-06-09 07:09:06.131379+00', NULL, 2, 1, '0', 2, '0', '0', '0', 30, 0, 5, 11);
INSERT INTO "public"."liquidity_history" VALUES (4, '2022-06-09 07:09:06.137996+00', '2022-06-09 07:09:06.137996+00', NULL, 1, 0, '0', 1, '0', '0', '0', 50, 0, 10, 12);
INSERT INTO "public"."liquidity_history" VALUES (5, '2022-06-09 07:09:06.208003+00', '2022-06-09 07:09:06.208003+00', NULL, 0, 0, '100000', 2, '100000', '100000', '10000000000', 30, 0, 5, 18);
INSERT INTO "public"."liquidity_history" VALUES (6, '2022-06-09 07:09:06.219124+00', '2022-06-09 07:09:06.219124+00', NULL, 0, 0, '99901', 2, '100100', '100000', '10000000000', 30, 0, 5, 19);
INSERT INTO "public"."liquidity_history" VALUES (7, '2022-06-09 07:09:06.231662+00', '2022-06-09 07:09:06.231662+00', NULL, 0, 0, '99802', 2, '100000', '99900', '9980200000', 30, 0, 5, 20);

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
INSERT INTO "public"."mempool_tx" VALUES (2, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.058617+00', NULL, '9f50d170-e704-11ec-b6f3-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":1,"AccountName":"gas.legend","AccountNameHash":"CkjpiSpFoE0MWw8jWjrrB7khN7pxpZucRXd0uv3pWYM=","PubKey":"1ec94e497abe0fbb87f9ed2843e21163e17e3e97f6bbbae7a88399b826474f93"}', '', '', 1, 0, 0, 2, 1);
INSERT INTO "public"."mempool_tx" VALUES (3, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.067524+00', NULL, '9f50f093-e704-11ec-b6f3-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":2,"AccountName":"sher.legend","AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","PubKey":"b0b6f7466154578ec66d51a335ead65ffd6a7210567fad9e68b6df8a5ce5dd85"}', '', '', 2, 0, 0, 3, 1);
INSERT INTO "public"."mempool_tx" VALUES (4, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.074676+00', NULL, '9f510cb1-e704-11ec-b6f3-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":3,"AccountName":"gavin.legend","AccountNameHash":"HFTAnJj3renV7rpBJKx8kS5laZo/dvpl1x6vY1nZvOs=","PubKey":"0500ccea3ca064968f5292b850ac8d4d3ee48d499357351a5ebfa2f30bb6070e"}', '', '', 3, 0, 0, 4, 1);
INSERT INTO "public"."mempool_tx" VALUES (5, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.082495+00', NULL, '9f510cb1-e704-11ec-b6f4-988fe0603efa', 4, 0, '0', -1, -1, 0, '100000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 2, 0, 0, 5, 1);
INSERT INTO "public"."mempool_tx" VALUES (6, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.090802+00', NULL, '9f510cb1-e704-11ec-b6f5-988fe0603efa', 4, 0, '0', -1, -1, 0, '100000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":3,"AccountNameHash":"HFTAnJj3renV7rpBJKx8kS5laZo/dvpl1x6vY1nZvOs=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 3, 0, 0, 6, 1);
INSERT INTO "public"."mempool_tx" VALUES (7, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.097963+00', NULL, '9f510cb1-e704-11ec-b6f6-988fe0603efa', 4, 0, '0', -1, -1, 1, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, 7, 1);
INSERT INTO "public"."mempool_tx" VALUES (8, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.104587+00', NULL, '9f510cb1-e704-11ec-b6f7-988fe0603efa', 4, 0, '0', -1, -1, 2, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":2,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, 8, 1);
INSERT INTO "public"."mempool_tx" VALUES (9, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.112587+00', NULL, '9f510cb1-e704-11ec-b6f8-988fe0603efa', 2, 0, '0', -1, 0, 0, '0', '0', '{"TxType":2,"PairIndex":0,"AssetAId":0,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, 9, 1);
INSERT INTO "public"."mempool_tx" VALUES (10, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.121926+00', NULL, '9f510cb1-e704-11ec-b6f9-988fe0603efa', 2, 0, '0', -1, 1, 0, '0', '0', '{"TxType":2,"PairIndex":1,"AssetAId":0,"AssetBId":1,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, 10, 1);
INSERT INTO "public"."mempool_tx" VALUES (11, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.129367+00', NULL, '9f510cb1-e704-11ec-b6fa-988fe0603efa', 2, 0, '0', -1, 2, 0, '0', '0', '{"TxType":2,"PairIndex":2,"AssetAId":1,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0, 11, 1);
INSERT INTO "public"."mempool_tx" VALUES (12, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.135924+00', NULL, '9f510cb1-e704-11ec-b6fb-988fe0603efa', 3, 0, '0', -1, 1, 0, '0', '0', '{"TxType":3,"PairIndex":1,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', '', '', -1, 0, 0, 12, 1);
INSERT INTO "public"."mempool_tx" VALUES (13, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.145891+00', NULL, '9f51d005-e704-11ec-b6fb-988fe0603efa', 5, 0, '0', 0, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":5,"AccountIndex":2,"NftIndex":0,"NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftContentHash":"q9G2rnlQf3tKMqhKtklbyf7mdFDtMW27p2us6KPFGXs=","NftL1TokenId":0,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","CollectionId":0}', '', '', 2, 0, 0, 13, 1);
INSERT INTO "public"."mempool_tx" VALUES (14, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.155677+00', NULL, '9f51d005-e704-11ec-b6fc-988fe0603efa', 17, 0, '0', -1, -1, 1, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":17,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0, 14, 1);
INSERT INTO "public"."mempool_tx" VALUES (15, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.163659+00', NULL, '9f51d005-e704-11ec-b6fd-988fe0603efa', 18, 0, '0', 0, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":18,"AccountIndex":2,"CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftIndex":0,"CollectionId":0,"NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","CreatorAccountNameHash":"AA==","NftContentHash":"q9G2rnlQf3tKMqhKtklbyf7mdFDtMW27p2us6KPFGXs=","NftL1TokenId":0}', '', '', 2, 0, 0, 15, 1);
INSERT INTO "public"."mempool_tx" VALUES (17, '2022-06-09 05:59:00.037782+00', '2022-06-09 07:09:06.18983+00', NULL, '1d4024d1-a49a-4f6a-9221-b716fedbd4aa', 10, 2, '5000', -1, -1, 0, '10000000', '0x99AC8881834797ebC32f185ee27c2e96842e1a47', '{"FromAccountIndex":2,"AssetId":0,"AssetAmount":10000000,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ToAddress":"0x99AC8881834797ebC32f185ee27c2e96842e1a47","ExpiredAt":1654761540020,"Nonce":2,"Sig":"YLVT6d4HuMWX7zPw3gcsmqs//dqE0xnfqTbSyLxx3pMDN7IvgmErKDXGiZ/XKC75wf2I03R2dXpmZ6bercDVHA=="}', '', '', 2, 2, 1654761540020, 17, 1);
INSERT INTO "public"."mempool_tx" VALUES (18, '2022-06-09 05:59:08.01764+00', '2022-06-09 07:09:06.202731+00', NULL, 'd9f93c69-3be2-483c-b746-d98022a61ecb', 8, 2, '5000', -1, 0, 0, '100000', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":0,"AssetAAmount":100000,"AssetBId":2,"AssetBAmount":100000,"LpAmount":100000,"KLast":10000000000,"TreasuryAmount":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761547992,"Nonce":3,"Sig":"7x6AUCZwD+fmcXWDs0WERCMR+rIBRSDDlzSV3vrjggQEdNd8uj6ghy3uzqrM2oeqa/9gP8vRzPFHYmUvmCKsTQ=="}', '', '', 2, 3, 1654761547992, 18, 1);
INSERT INTO "public"."mempool_tx" VALUES (1, '2022-06-08 08:25:58.050504+00', '2022-06-09 07:09:06.048101+00', NULL, '9f5005a9-e704-11ec-b6f3-988fe0603efa', 1, 0, '0', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":0,"AccountName":"treasury.legend","AccountNameHash":"FnxTYwiKQKSDmRKocvQxZCcHQMfphuxVOXstWDMXq0o=","PubKey":"fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998"}', '', '', 0, 0, 0, 1, 1);
INSERT INTO "public"."mempool_tx" VALUES (16, '2022-06-09 05:58:53.463597+00', '2022-06-09 07:09:06.177934+00', NULL, '01de4078-304a-406d-9995-7c8550248f28', 6, 2, '5000', -1, -1, 0, '100000', '', '{"FromAccountIndex":2,"ToAccountIndex":3,"ToAccountNameHash":"1c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb","AssetId":0,"AssetAmount":100000,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"Memo":"transfer","CallData":"","CallDataHash":"Dd56AihX/sG4/6dmSpN6JQ065o81YGF1TTUx4mdBA9g=","ExpiredAt":1654761533445,"Nonce":1,"Sig":"epyzhZA39/F3mHPAvv8dz8NgPTtPWlYqPs9tEyDNsQgA8A4bi4ruGJe6evoUJ9BdWR49SJ1SCaJ+on1y2QyEFg=="}', '', 'transfer', 2, 1, 1654761533445, 16, 1);
INSERT INTO "public"."mempool_tx" VALUES (19, '2022-06-09 05:59:15.071544+00', '2022-06-09 07:09:06.214826+00', NULL, '8d60898f-ef87-4726-9322-1bda3fd22c2b', 7, 0, '5000', -1, 0, 0, '100', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":2,"AssetAAmount":100,"AssetBId":0,"AssetBMinAmount":98,"AssetBAmountDelta":99,"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"ExpiredAt":1654761555052,"Nonce":4,"Sig":"qCqSqikwaTLE/4VDURQHRYK+9gykmtanhYBv/ByGSoMGBJq+8D7z2b9yc0D8M1zNmfeD5YvpCpJkElsjarviGw=="}', '', '', 2, 4, 1654761555052, 19, 1);
INSERT INTO "public"."mempool_tx" VALUES (20, '2022-06-09 05:59:22.05211+00', '2022-06-09 07:09:06.226605+00', NULL, '4282c649-d9be-49fc-bc34-48dd4bab5f15', 9, 2, '5000', -1, 0, 0, '100', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":0,"AssetAMinAmount":98,"AssetBId":2,"AssetBMinAmount":99,"LpAmount":100,"AssetAAmountDelta":99,"AssetBAmountDelta":100,"KLast":9980200000,"TreasuryAmount":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761562031,"Nonce":5,"Sig":"A9Qrmz5Uj7mDA3WfUGRhFoNuWFMtO+RwxwfkbkmglBEAXouZOVBhU4iguZYAEtwO6xFsPYrGFI92A0z21KFk6Q=="}', '', '', 2, 5, 1654761562031, 20, 1);
INSERT INTO "public"."mempool_tx" VALUES (21, '2022-06-09 05:59:29.308896+00', '2022-06-09 07:09:06.237175+00', NULL, '0c3c72c1-ff2d-4ba3-b026-93b7cec8e6a1', 11, 2, '5000', -1, -1, 0, 'sher.legend', '0', '{"AccountIndex":2,"CollectionId":1,"Name":"Zecrey Collection","Introduction":"Wonderful zecrey!","GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761569287,"Nonce":6,"Sig":"HrXIp5Vauk06jV3Jj4Lke/+XYp3ThhQsVAx8QksB1aYFVYeEZOf5nmIML+U3TSatTwyLCGMmomQSjGMxVtouxw=="}', '', '', 2, 6, 1654761569287, 21, 1);
INSERT INTO "public"."mempool_tx" VALUES (22, '2022-06-09 05:59:35.369268+00', '2022-06-09 07:09:06.247097+00', NULL, '69b6e9bb-0f8b-4b20-af35-35c6945489aa', 12, 2, '5000', 1, -1, 0, '0', '', '{"CreatorAccountIndex":2,"ToAccountIndex":3,"ToAccountNameHash":"1c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb","NftIndex":1,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftCollectionId":1,"CreatorTreasuryRate":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761575344,"Nonce":7,"Sig":"KAC/ZwgC7PJo4KDqBpVI048lfSjzli9qLXp3I4CFBJIBZY/c301Cm46AgoxkfaC02p8M5W263VWSEOsW+YATRg=="}', '', '', 2, 7, 1654761575344, 22, 1);
INSERT INTO "public"."mempool_tx" VALUES (23, '2022-06-09 05:59:43.911324+00', '2022-06-09 07:09:06.257815+00', NULL, 'a512199e-4146-407c-9b13-0039c8796650', 13, 0, '5000', 1, -1, 0, '0', '', '{"FromAccountIndex":3,"ToAccountIndex":2,"ToAccountNameHash":"214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45","NftIndex":1,"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"CallData":"","CallDataHash":"Dd56AihX/sG4/6dmSpN6JQ065o81YGF1TTUx4mdBA9g=","ExpiredAt":1654761583894,"Nonce":1,"Sig":"G/YUzJOh4sp8ZF27Mz4s5hvLBC1P63uZbIVjMzHjqRkB0wTz30LsTSC5VIcxILqv7a/dCw4qd4Y3LYOtgLPW1w=="}', '', '', 3, 1, 1654761583894, 23, 1);
INSERT INTO "public"."mempool_tx" VALUES (24, '2022-06-09 05:59:50.167539+00', '2022-06-09 07:09:06.271098+00', NULL, '79245e26-0ec4-486e-9556-2fd477928380', 14, 0, '5000', 1, -1, 0, '10000', '', '{"AccountIndex":2,"BuyOffer":{"Type":0,"OfferId":0,"AccountIndex":3,"NftIndex":1,"AssetId":0,"AssetAmount":10000,"ListedAt":1654754390138,"ExpiredAt":1654761590138,"TreasuryRate":200,"Sig":"CrwNdL+oHhdWBgJ0j+O/IY5Ca5qnBw6kDkPyUWD4wywApriICAXoTooPa//9vP9QRDPEQsHu9C2vvfeNaXeUGA=="},"SellOffer":{"Type":1,"OfferId":0,"AccountIndex":2,"NftIndex":1,"AssetId":0,"AssetAmount":10000,"ListedAt":1654754390138,"ExpiredAt":1654761590138,"TreasuryRate":200,"Sig":"1k2/LHCg9jCQ2+S9qYW8hWRLFryR7xQU+32zmjlEAAICKm6Tlks2bqoUXLBiOe5VNwUIMJ5gwJxTKOlJbOExbA=="},"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"CreatorAmount":0,"TreasuryAmount":200,"Nonce":8,"ExpiredAt":1654761590138,"Sig":"Y/HMbBEAcLfdg5+eqAo/Gz+Nq8ZHdmLbm+SRUBZAB5cASst6Eo7UiL6O7+2IGNa7lBij3RgRvPn8bcBbxEO+Ww=="}', '', '', 2, 8, 1654761590138, 24, 1);
INSERT INTO "public"."mempool_tx" VALUES (25, '2022-06-09 06:00:01.732295+00', '2022-06-09 07:09:06.283098+00', NULL, 'fa2c5d73-deab-494a-a5c1-64938d1430aa', 15, 2, '5000', -1, -1, 0, 'sher.legend', '0', '{"AccountIndex":2,"OfferId":1,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761601706,"Nonce":9,"Sig":"XQyUjK2wFu2opRmOhCnmDtCsVFeyj0MDofWLqyQCBBQCRi4zpBIxphGhuSMkoDO1WiFWkxVaRXwINrpOKjfMug=="}', '', '', 2, 9, 1654761601706, 25, 1);
INSERT INTO "public"."mempool_tx" VALUES (26, '2022-06-09 06:00:11.597823+00', '2022-06-09 07:09:06.294273+00', NULL, '5c3021a0-15bc-42b3-a966-11b0afdb0c73', 16, 0, '5000', 1, -1, 0, '0', '', '{"AccountIndex":3,"CreatorAccountIndex":2,"CreatorAccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","CreatorTreasuryRate":0,"NftIndex":1,"NftContentHash":"Bmpl0+Q5etBfsuf1DqwWBkenSGws7bvqxkYkyL7qIvE=","NftL1Address":"0","NftL1TokenId":0,"CollectionId":1,"ToAddress":"0xd5Aa3B56a2E2139DB315CdFE3b34149c8ed09171","GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"ExpiredAt":1654761611582,"Nonce":2,"Sig":"sPJWFi9pTtfv6Z9zFI/QlQ2M9APJdOrWuPuq8wVnZ5kDbJkAuE2MoDA7o07rv5g/nz/cjloY2us88w1dkg+5sQ=="}', '', '', 3, 2, 1654761611582, 26, 1);

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
INSERT INTO "public"."mempool_tx_detail" VALUES (2, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 6, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (3, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 7, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (4, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 8, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (5, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 9, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (6, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 10, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (7, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 11, 2, 2, -1, '', '{"PairIndex":2,"AssetAId":1,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (8, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 12, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (9, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 13, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (11, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 14, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":null}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (12, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 15, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (13, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 15, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (14, '2022-06-09 05:58:53.464815+00', '2022-06-09 05:58:53.464815+00', NULL, 16, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (15, '2022-06-09 05:58:53.464815+00', '2022-06-09 05:58:53.464815+00', NULL, 16, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (16, '2022-06-09 05:58:53.464815+00', '2022-06-09 05:58:53.464815+00', NULL, 16, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (18, '2022-06-09 05:59:00.039002+00', '2022-06-09 05:59:00.039002+00', NULL, 17, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-10000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (19, '2022-06-09 05:59:00.039002+00', '2022-06-09 05:59:00.039002+00', NULL, 17, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (20, '2022-06-09 05:59:00.039002+00', '2022-06-09 05:59:00.039002+00', NULL, 17, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (21, '2022-06-09 05:59:08.018849+00', '2022-06-09 05:59:08.018849+00', NULL, 18, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (22, '2022-06-09 05:59:08.018849+00', '2022-06-09 05:59:08.018849+00', NULL, 18, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (23, '2022-06-09 05:59:08.018849+00', '2022-06-09 05:59:08.018849+00', NULL, 18, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (24, '2022-06-09 05:59:08.018849+00', '2022-06-09 05:59:08.018849+00', NULL, 18, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":100000,"OfferCanceledOrFinalized":0}', 3, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (28, '2022-06-09 05:59:15.072544+00', '2022-06-09 05:59:15.072544+00', NULL, 19, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-100,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (29, '2022-06-09 05:59:15.072544+00', '2022-06-09 05:59:15.072544+00', NULL, 19, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (30, '2022-06-09 05:59:15.072544+00', '2022-06-09 05:59:15.072544+00', NULL, 19, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (31, '2022-06-09 05:59:15.072544+00', '2022-06-09 05:59:15.072544+00', NULL, 19, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":-99,"AssetBId":2,"AssetB":100,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 3, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (32, '2022-06-09 05:59:15.072544+00', '2022-06-09 05:59:15.072544+00', NULL, 19, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 4, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (33, '2022-06-09 05:59:22.053572+00', '2022-06-09 05:59:22.053572+00', NULL, 20, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (34, '2022-06-09 05:59:22.053572+00', '2022-06-09 05:59:22.053572+00', NULL, 20, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":100,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (35, '2022-06-09 05:59:22.053572+00', '2022-06-09 05:59:22.053572+00', NULL, 20, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (36, '2022-06-09 05:59:22.053572+00', '2022-06-09 05:59:22.053572+00', NULL, 20, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":-100,"OfferCanceledOrFinalized":0}', 3, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (37, '2022-06-09 05:59:22.053572+00', '2022-06-09 05:59:22.053572+00', NULL, 20, 0, 1, 0, 'treasury.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 4, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (1, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 5, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (10, '2022-06-08 08:25:58.053003+00', '2022-06-08 08:25:58.053003+00', NULL, 13, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":2,"NftContentHash":"abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b","NftL1TokenId":"0","NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","CreatorTreasuryRate":0,"CollectionId":0}', 0, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (17, '2022-06-09 05:58:53.464815+00', '2022-06-09 05:58:53.464815+00', NULL, 16, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (25, '2022-06-09 05:59:08.018849+00', '2022-06-09 05:59:08.018849+00', NULL, 18, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":100000,"AssetBId":2,"AssetB":100000,"LpAmount":100000,"KLast":10000000000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 4, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (26, '2022-06-09 05:59:08.018849+00', '2022-06-09 05:59:08.018849+00', NULL, 18, 0, 1, 0, 'treasury.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 5, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (27, '2022-06-09 05:59:08.018849+00', '2022-06-09 05:59:08.018849+00', NULL, 18, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 6, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (38, '2022-06-09 05:59:22.053572+00', '2022-06-09 05:59:22.053572+00', NULL, 20, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":-99,"AssetBId":2,"AssetB":-100,"LpAmount":-100,"KLast":9980200000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 5, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (39, '2022-06-09 05:59:22.053572+00', '2022-06-09 05:59:22.053572+00', NULL, 20, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 6, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (41, '2022-06-09 05:59:29.310534+00', '2022-06-09 05:59:29.310534+00', NULL, 21, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (43, '2022-06-09 05:59:35.370489+00', '2022-06-09 05:59:35.370489+00', NULL, 22, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (44, '2022-06-09 05:59:35.370489+00', '2022-06-09 05:59:35.370489+00', NULL, 22, 2, 1, 3, 'gavin.legend', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (46, '2022-06-09 05:59:35.370489+00', '2022-06-09 05:59:35.370489+00', NULL, 22, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (47, '2022-06-09 05:59:43.912534+00', '2022-06-09 05:59:43.912534+00', NULL, 23, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (48, '2022-06-09 05:59:43.912534+00', '2022-06-09 05:59:43.912534+00', NULL, 23, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (49, '2022-06-09 05:59:43.912534+00', '2022-06-09 05:59:43.912534+00', NULL, 23, 1, 3, 2, 'sher.legend', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":2,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 2, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (50, '2022-06-09 05:59:43.912534+00', '2022-06-09 05:59:43.912534+00', NULL, 23, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (51, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (52, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":-10000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (53, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":1}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (54, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":9800,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (55, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":1}', 4, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (60, '2022-06-09 06:00:01.734295+00', '2022-06-09 06:00:01.734295+00', NULL, 25, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (61, '2022-06-09 06:00:01.734295+00', '2022-06-09 06:00:01.734295+00', NULL, 25, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":3}', 1, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (62, '2022-06-09 06:00:01.734295+00', '2022-06-09 06:00:01.734295+00', NULL, 25, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (63, '2022-06-09 06:00:11.599038+00', '2022-06-09 06:00:11.599038+00', NULL, 26, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (64, '2022-06-09 06:00:11.599038+00', '2022-06-09 06:00:11.599038+00', NULL, 26, 1, 3, -1, '', '{"NftIndex":1,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (65, '2022-06-09 06:00:11.599038+00', '2022-06-09 06:00:11.599038+00', NULL, 26, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (66, '2022-06-09 06:00:11.599038+00', '2022-06-09 06:00:11.599038+00', NULL, 26, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2);
INSERT INTO "public"."mempool_tx_detail" VALUES (40, '2022-06-09 05:59:29.310534+00', '2022-06-09 05:59:29.310534+00', NULL, 21, 0, 4, 2, 'sher.legend', '1', 0, 0);
INSERT INTO "public"."mempool_tx_detail" VALUES (42, '2022-06-09 05:59:29.310534+00', '2022-06-09 05:59:29.310534+00', NULL, 21, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1);
INSERT INTO "public"."mempool_tx_detail" VALUES (45, '2022-06-09 05:59:35.370489+00', '2022-06-09 05:59:35.370489+00', NULL, 22, 1, 3, 3, 'gavin.legend', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 2, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (56, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 5, 3);
INSERT INTO "public"."mempool_tx_detail" VALUES (57, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 1, 3, -1, '', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 6, -1);
INSERT INTO "public"."mempool_tx_detail" VALUES (58, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":200,"LpAmount":0,"OfferCanceledOrFinalized":0}', 7, 4);
INSERT INTO "public"."mempool_tx_detail" VALUES (59, '2022-06-09 05:59:50.16896+00', '2022-06-09 05:59:50.16896+00', NULL, 24, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 8, 4);

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
INSERT INTO "public"."proof_sender" VALUES (1, '2022-06-09 07:23:24.586408+00', '2022-06-09 07:23:24.586408+00', NULL, '{"A":[5357907446709718487941986733309225406048435776139431945762925093327186667256,17496725062937861203898650287725858704516594988593501934567880265799491463177],"B":[[14720521433584626624325963016940251896438503185693768206360980248724179708363,18375239315756568022095442647985200911749918229485560905531441891961308708856],[16363043350587558836741133002397756467217077564923596019135116682493418534006,5596519350005502017924021801867967131722550642201664380726442341779855698916]],"C":[1850693862752086122306022103343831655605374862562030081903491489803751696327,5443157994089927105172260006038197925773595247560597214530383901744905578760],"Inputs":[9450703979270269782239153154655868962158804821639950878322637986092829620047,15043264495212376832665268192414242291394558777525090122806455607283976407362,11963247688191561196873225082343613795306480307795283451691052361457074285021]}', 1, 0);
INSERT INTO "public"."proof_sender" VALUES (2, '2022-06-09 07:23:44.643141+00', '2022-06-09 07:23:44.643141+00', NULL, '{"A":[20813507163771042917946551071764721159056224597776408025052785981323338481154,10921825533225065035430794056222068611765401207937070688882451768008661226777],"B":[[15486517432908393408144952664765811446772407320369480943565640740700026795462,17674942133842931647379190043665571853160939366208023700909488416882045291640],[16199502416665520960377122180660905576906512074109147977536118367276226939751,12064251941308843097075711141155291440613675702020429411506723259048251240700]],"C":[16681275058732350549969434485232133225813187897612271968303104552461518982258,20050623136967140394695099894279884077307717504336685541180071165681973942195],"Inputs":[15043264495212376832665268192414242291394558777525090122806455607283976407362,12297177442334280409244260380119123763383333089941937037498066619533324895858,1945871703106592192445527630015428020428207196346331000444516874376214077870]}', 2, 0);
INSERT INTO "public"."proof_sender" VALUES (3, '2022-06-09 07:23:54.527389+00', '2022-06-09 07:23:54.527389+00', NULL, '{"A":[7559736663398274886813521875027328296080906357658551100428005329404128126355,4573291067759265048831985704695636350001464683771328909544384358380876318135],"B":[[19410483811700283141714932375622551750479122510836369528558895173524035900853,6795018017361741386612848268111959380245189784456180212424711313656215623940],[15202464355920779518305434970816636317493593488258678323070164540041171347404,14447803888624990866915689832929528497877374236470736218388819527564579233453]],"C":[17627580096248132224088972877089435327314530498914390280315000907016139621845,4249624274470976318203955964841436803380468133755247436840408256085294282023],"Inputs":[12297177442334280409244260380119123763383333089941937037498066619533324895858,11118933918917677552118113339774603267482948853529286163697696035860800912124,4141452066739870661498648949772962314122049597784115302113989740758522105342]}', 3, 0);
INSERT INTO "public"."proof_sender" VALUES (4, '2022-06-09 07:24:04.538222+00', '2022-06-09 07:24:04.538222+00', NULL, '{"A":[9110103407963974868366728464799245078858067689910507226822459220207506272499,21719474617242243481844770115660891794725263698349330338195525371390808817327],"B":[[21389161538241386313330361588428069577956957114833598623812545981430584522601,12971960139169430699066909962695468630217732692054862567443912236352393104033],[20509884633768299055224189396400437941908017883338859122321435747087159532373,15819007506379581663541742336647097354994209799108208341565219065761245952556]],"C":[19873323619447981526452307419853499863685799395091206487134024403814099312748,6119875941814807106675237921726989378323319043345040438007083815484556674579],"Inputs":[11118933918917677552118113339774603267482948853529286163697696035860800912124,3934520836078457484774962454656757098464038511289928405982572228040277672631,16357933347269012595297401294074469152155600812034184911699282662045711091062]}', 4, 0);
INSERT INTO "public"."proof_sender" VALUES (5, '2022-06-09 07:24:14.518489+00', '2022-06-09 07:24:14.518489+00', NULL, '{"A":[15533918622537399015518237774264923161077934783910189699836003437656124636309,5069299624683444574823466007839504469631626285727350111939705844716777430979],"B":[[3744919964186335507197012330680948825493812403368725511327036779440140104980,7492296653040452746904231660511773202247564846231760948945037281248294183181],[8965723521881953179017783714000090715251224020206188942678134208190811516624,7797531568507494197968822873157266008002521381654473396471317662493892575082]],"C":[6265171476735348143148882338652859440161392921916648421009933142497839856567,10921850148220087867261903715064576494902109547006503271526089466913338341096],"Inputs":[3934520836078457484774962454656757098464038511289928405982572228040277672631,16025607879873774270539617119489323551891610600919089473988379724010068742191,12961477835462357069095357285869702333916507490943518131419162928549401849131]}', 5, 0);
INSERT INTO "public"."proof_sender" VALUES (6, '2022-06-09 07:24:24.51569+00', '2022-06-09 07:24:24.51569+00', NULL, '{"A":[12775481537775540353198771706037410309403565868973758418831274624796609709031,12563220958406810787093974273390792396340258681839358919355603843534017576407],"B":[[466766940151359509587006629080093685334850865423711940572706407903099898485,8113475061863033439057383092505483824758758712632932303197768364205513767414],[8619352988342772426438347244630686188715154751181596645552301889409243178922,4871484385926229301046871359466334901001342722779358695691716172665669796423]],"C":[14671499948976426331719441470204431306391140342047201077417051680238410706345,20905717683481457094817392586169937372173832875383492533339292915438773370549],"Inputs":[16025607879873774270539617119489323551891610600919089473988379724010068742191,1182007653870860980188901032313239440165930770789334793344276798172967050151,10900582511275367572049526641256039768650071979656307171225251307845332413143]}', 6, 0);
INSERT INTO "public"."proof_sender" VALUES (7, '2022-06-09 07:24:34.61646+00', '2022-06-09 07:24:34.61646+00', NULL, '{"A":[19736957023120166357001005889444490685321584374642961893365372621327616694990,19392065716103988721648074939960091280107501528042542745581432276521263510412],"B":[[571492896427163448890807461782650204657343510215215625757504010828886836382,8001850526536950477450814216227647275767248232511233069549157099510638871718],[15399770455289783494841448938457475291997723025011314246777739774075988801854,6953533655578395070391182841774083704495115922140550198572691463822431464216]],"C":[12293068457310743600605512778002935015544221667513259377039012799901122256184,20161071113916839973112348786702810815193408038299265761391507217557745590211],"Inputs":[1182007653870860980188901032313239440165930770789334793344276798172967050151,17094011329777467206340521509783369268317290252369917384669054004578361939180,15017229726478937929116658928635635052098058456225427220340952831123077394838]}', 7, 0);
INSERT INTO "public"."proof_sender" VALUES (8, '2022-06-09 07:24:44.505085+00', '2022-06-09 07:24:44.505085+00', NULL, '{"A":[5980924307279874344680032233475410884184538889970060347530687502031738791447,21287811521893823900630817973519994854181255263477570592182999332351498328517],"B":[[529751275497961151962680821055254987754056664988591178253218172601907431388,5726880856402803987695595811737231271996887951427266623691512124984038042124],[7811601596705273184505367807438448214366325755942922992393197559446165768496,17344914781957304781850965204935074254591824007899403656719967166506658016449]],"C":[15814752059536845242306798590880089606594520429855000946339129876488798000392,17775757703161992670597128874999787610476786657426988862225264557367667309266],"Inputs":[17094011329777467206340521509783369268317290252369917384669054004578361939180,10689577469853096804426857094028199763247558905267837019735861787144318922388,16873122977431782423632070671545326080911017857820594834326133253531214104424]}', 8, 0);
INSERT INTO "public"."proof_sender" VALUES (9, '2022-06-09 07:24:54.49822+00', '2022-06-09 07:24:54.49822+00', NULL, '{"A":[17298168395729437977041039664220119339465364533281654677311938416691559070801,4271068355456674107335747387035117575005876702521403663596060266745310486890],"B":[[16270513231831194845343763638164526208133394145718804726859532012475995303773,17414153493171324446566093008114435376532165611209940832002789465557349791412],[785565747771497805408072558414309406441868066607847606639392630449005797289,7662969918737420327668916565397411820649711028656723319111596087480064056168]],"C":[11739174259408060548826038043849794049356355754716408122867437867125992732497,8893477733239534739317139667111081321007948604776245868588416718554518008063],"Inputs":[10689577469853096804426857094028199763247558905267837019735861787144318922388,6948952673628372168875450143833765144025527013713566967806168545755060380858,2381018844473592730594362683078647128924834795096405048545540899272736150108]}', 9, 0);
INSERT INTO "public"."proof_sender" VALUES (10, '2022-06-09 07:25:04.533155+00', '2022-06-09 07:25:04.533155+00', NULL, '{"A":[15521017184051992261896355549794298246534652628121152266063070688366730452162,13961508868962569273505124249185255794900260440581779157951474985385748174470],"B":[[17065657997897318674948795907310757449530732001037734380186702211180214299746,15677243625116285317847613474005417091647205554397557018353889845576538878865],[19390619060562936877588984907791381849976304337491405690373860956341962708934,5563341927983312698473895099760917317701486515061397527318804189948823286630]],"C":[16597160865679091783441187860804896114304069725353186870292748219295416554721,2394595445592421482288547744316437281444854255474176523427122295346372419127],"Inputs":[6948952673628372168875450143833765144025527013713566967806168545755060380858,4193345583120754934140829490350309912111035558299556830065423384594352669901,20271788291865604819762498640632971866147076040033239944791480616232093581018]}', 10, 0);
INSERT INTO "public"."proof_sender" VALUES (11, '2022-06-09 07:25:14.491108+00', '2022-06-09 07:25:14.491108+00', NULL, '{"A":[10690123035117347435583218686434724670922995232403377702074029183849931986601,11614324004353826540263410033076139666145259771313223397148950306079058192104],"B":[[2000851832757796106490606948503810002054532222084721787129197252803428227728,2979971793942668255925708494679646258534614415906908614190155356821720342782],[2311482297729564514731645320034205088354037661620700235056820462166900664425,17847514665886669859493612822919360619813712229216796895445963611376311242076]],"C":[3904059735216286784462585874468716885727990103665582715792194935387028735974,14307128667942802919098261629321726611748388900225190312728863743939670107572],"Inputs":[4193345583120754934140829490350309912111035558299556830065423384594352669901,10152064816703520911570552026983147930837787570518811408541967275193207832861,8586895846168170365040413310156165056878705110488484396408283666332873867086]}', 11, 0);
INSERT INTO "public"."proof_sender" VALUES (12, '2022-06-09 07:25:24.663412+00', '2022-06-09 07:25:24.663412+00', NULL, '{"A":[10159975002397651137232468947306723784663524120049685253062346380189043876645,11406588679264760619545135428596927996095766051967155205361292168064406762956],"B":[[21485535930335108169767673721139457748766431533328656866258396985007395593322,11483195628712409573411933921150045560230569903877254153789567973343214209489],[8377597436709839824938622408128856671090541846011514103291544481895408030905,12358524440356972678169391602672492180796154765740788494658455382769151527068]],"C":[16165101008492526683951381872986995534701186382159210474925499479976069214227,15432847146843294548990405680486435645749480783451120247215124966198786454914],"Inputs":[10152064816703520911570552026983147930837787570518811408541967275193207832861,4041848711751034178697814175798083696278503514951292272975039336585530128677,7813894203082824048002111901200503984007958306255276001456868296408697822876]}', 12, 0);
INSERT INTO "public"."proof_sender" VALUES (13, '2022-06-09 07:25:34.508794+00', '2022-06-09 07:25:34.508794+00', NULL, '{"A":[10261566542587671522158701555391897278151906206660783292567297598068776634404,14431967531693461723360728030346792766815705294739312195077940320409189664739],"B":[[15098950899127777846090369607864746915518927985834483464449391535208274309950,20643275506250556064271603846631212093790052706946095891782513042968936043553],[4514615009399575525815224706513983786443715249372435337001216796291604905574,20662716222622625024266833655160998686051845829662553482552997846267247812557]],"C":[14384476812401774943946787938840653353027237830566400231649510365669251514828,20074829254704778631937473514626755917209730062309287218455118244563346860026],"Inputs":[4041848711751034178697814175798083696278503514951292272975039336585530128677,16271226640539965147245273017495976697549012047434316200568775079391086367406,21140244431992199370190780936908877210928676351903003902465868132982527969590]}', 13, 0);
INSERT INTO "public"."proof_sender" VALUES (14, '2022-06-09 07:25:44.50934+00', '2022-06-09 07:25:44.50934+00', NULL, '{"A":[2210158445496352120899045035190168685146098850684767849659993025219098074105,10713476324707801711837733445879892899723197674209918769939423961418422183906],"B":[[810160846127514362971089065671508266995730849990827791074389192241995082479,17124714650811796689345365984747471876548722425821929813372000519204243123207],[11501093053962789887708697857007171022988320622370497060069208049036375154072,9988260441357861754140739934979132018062785935350555268885500845009532139197]],"C":[14073647615142636673801731120126325000171724717440383555614408416935944109915,19995541106659746044602934413206860303224000396734660878547845845329391686974],"Inputs":[16271226640539965147245273017495976697549012047434316200568775079391086367406,10639295657989775574086409957947829148199615559969022728688638593216982037730,14244661216982820488005781571188745917102390506429749184784599384196934352272]}', 14, 0);
INSERT INTO "public"."proof_sender" VALUES (15, '2022-06-09 07:25:54.499472+00', '2022-06-09 07:25:54.499472+00', NULL, '{"A":[17571712050418197449385769643123403538568730885831817467140695867950742576096,3090941610870382185102534174225411338274127812102697732643424632237105650634],"B":[[3369613931290973248116149070356946993596232435753033083434632596692438477339,18366942516192635738122891200923533303260759803093492328031426562456172291703],[16896288962244998074529166931355136611523654181929170060647399444920873404819,7119856041363768320082243762237331758563029093817707992004535650778247565643]],"C":[1840248247603955387301578228061851895876391391526584490102386947110025383239,1893584611121116190202568066637880678400976517808952365283486707485518827515],"Inputs":[10639295657989775574086409957947829148199615559969022728688638593216982037730,18544100231407746896690739092020187337464474124217572106403296132870486599844,7849209998090739933837234226634748060565161797241883646562575696184552731303]}', 15, 0);
INSERT INTO "public"."proof_sender" VALUES (16, '2022-06-09 07:26:04.528534+00', '2022-06-09 07:26:04.528534+00', NULL, '{"A":[2126836860157825719241372547050490994950916536601448821227821491730516016811,13089882072162797154796234962607920289216749510915300142391196912148875925630],"B":[[13659561965717366527017806818607480486672342922075516437810500485021637136675,2496538948713319851218392543472407434180517047990867205408257741031808477045],[15094687796192020436386343827450550386224173671744781873726481980329008843262,867732462846363625180283057195445738889659048593905680925877170955622438867]],"C":[9704168160272942752392608817731304970934678699143790116321686366666394590413,4122674166104500008511969464450669713016489254905378284799298012930487000827],"Inputs":[18544100231407746896690739092020187337464474124217572106403296132870486599844,5586088040550485664848597881510203165901247035169583122754084497094449251134,3625936899631428556066012573882141829798170239079403930433757352539709158032]}', 16, 0);
INSERT INTO "public"."proof_sender" VALUES (17, '2022-06-09 07:26:24.48907+00', '2022-06-09 07:26:24.48907+00', NULL, '{"A":[8021825240282321891602795145832694810174683334757724685843445828289183743411,9797806148552150218332209743807309164594090867572113559113072043118468089958],"B":[[13863111205755049631187565537917882894965629334513264946003524866530026042856,11209558927815688530042307249100566886043360494201459472397024212479453599141],[1458525839525311687666868325725678981703071909829436207947447624093353197195,14371955880321935916965723312134766118885105402963244880434574586878625571850]],"C":[5142659326844849799389371635098315968708084820094983091627390207520198145902,10975790742628854149436194753636462470319330706494819659909605694454074916861],"Inputs":[5586088040550485664848597881510203165901247035169583122754084497094449251134,20471331031958273670946341121221367175878128316941189710493194485442977281812,21503538493464361595391979114337605806935242050381727476274743294853677383658]}', 17, 0);
INSERT INTO "public"."proof_sender" VALUES (18, '2022-06-09 07:26:34.515311+00', '2022-06-09 07:26:34.515311+00', NULL, '{"A":[524446151924457502629782649096180450887763800184045018702064673837856878801,8645493695716812053638687555484455361153079383536366362474845175164849280430],"B":[[13214392869132535758338359438094175432538398792575925524242672252312289975848,17446421255848264570787336455552175374432005980647957336096459498263244110182],[8956035534231479808117633904629365467408603959678606431087745192681877660792,2096584104701727630376880684056131272718140728642981244842879176486428113166]],"C":[8356912430786065703682252072677821589280512372740410081300042583681741620653,17092389843809185642541566923425735468618940902636325750934483545031198522812],"Inputs":[20471331031958273670946341121221367175878128316941189710493194485442977281812,8173166197544277304186478691258384071873880483015113628785610445658683442517,14844503571774079792674155527073126612056383359258218655941264775591156150707]}', 18, 0);
INSERT INTO "public"."proof_sender" VALUES (19, '2022-06-09 07:26:44.506942+00', '2022-06-09 07:26:44.506942+00', NULL, '{"A":[16614803823762227265602542762562641617849651049582737327882059707161533543297,1989647283631880081800578290111135047905000283185640783386107664754726608866],"B":[[20517483818085564818237112031668130358941958900214193134365496704834545814495,19992039921266405329973970711495463926957582252208064666206277052770185140352],[20827926094190320228662724871873283173549730499119631069740628945187691395199,18605260650007427798312238830058849173731293343524854304739735661339975496477]],"C":[7135967049520761871891773095894762069023434730315873135828697605034781859080,10353166813353644128713769387067055957556937429813052188920135431420516271602],"Inputs":[8173166197544277304186478691258384071873880483015113628785610445658683442517,13442140803681527408895133679865362700517560901608420346744235869865161259829,6333711030315459658295063134508763993660996977628803990603795678616206379198]}', 19, 0);
INSERT INTO "public"."proof_sender" VALUES (20, '2022-06-09 07:26:54.490118+00', '2022-06-09 07:26:54.490118+00', NULL, '{"A":[6132310064061812917597796263388370286946492790233537399322889464402543655586,19501022468640836101914191464949093401739722695796254822615620304574394511496],"B":[[10564932849935093467524018260797185368528626335439755795496815498065708499361,12181575049670227291538072836469390186022690793637974386390835441070364460172],[17270157245668942413517738043529004891926380047387686288612710002861761976772,5020610668548115993849743727814694147202074122182742214407662069230058251452]],"C":[20111335642937765370660225974752385554013986571008863223283829441596809046862,17277203220361190408934792092857832556440285354519614185589057268311960712867],"Inputs":[13442140803681527408895133679865362700517560901608420346744235869865161259829,21047623044075927563049268030940086099536328266902531926068000182657135421142,780423291219507366819846711946964749186412840151552921123168758150621974449]}', 20, 0);
INSERT INTO "public"."proof_sender" VALUES (21, '2022-06-09 07:27:04.491762+00', '2022-06-09 07:27:04.491762+00', NULL, '{"A":[16204880395453769651183471645582504007136579741357249653408735404807902775575,11276361857025370970975160193182912988905786419776451433719369209273371611236],"B":[[1200179618287056260178250055792404941220317454194816388154861357457628691214,6143279746262085118120942236486502237997359991984043994710272891856626082483],[14580391975815040214423623849462454542400218822168596850723208790989682217681,19558548856321681450690854261761759117678761694861749630042902389894356441783]],"C":[18873465828122548482914840778283129050244062092889099188257531303887505408914,19901832500179328790914736663559185779975220202535263730181407432683217093314],"Inputs":[21047623044075927563049268030940086099536328266902531926068000182657135421142,9057099176725074620829751149253261324171759863292310461253640687639650844381,6696999952182686226580887420154672517510273548902301225211986252493422045748]}', 21, 0);
INSERT INTO "public"."proof_sender" VALUES (22, '2022-06-09 07:27:14.521571+00', '2022-06-09 07:27:14.521571+00', NULL, '{"A":[14562964723383369250614892264240850983218951872823483654716940872994934168222,18714665548629189509729781458902633894127532066976957075778926500601573130917],"B":[[10460335902891958301152056783471155320860960591164890155102139766864344954914,9702800026805007716668661710187751346083240230687670446136239080903120554346],[4675988947523719050292265611283630343218199784368931232743768167455679434561,7844147298526505729646700381208644021845085587734624963556951982625113881763]],"C":[18252077589966946687951281102899931594068265217916094749109049683522583514675,10549164743029520638807373574510828809421021652485838728206134173091971953915],"Inputs":[9057099176725074620829751149253261324171759863292310461253640687639650844381,8382471479713636303903675749137703248387104876156021657409718717386550288255,13792338193516504167605049995473338407598939251269101016710219761823104199464]}', 22, 0);
INSERT INTO "public"."proof_sender" VALUES (23, '2022-06-09 07:27:24.533483+00', '2022-06-09 07:27:24.533483+00', NULL, '{"A":[2926703581978961895068066663533236177576491453879253105032048888122012070666,7565050302237490214998916833549598732924889515843145342497594450511354526710],"B":[[20983553324602106415744502450751001321320010849139907626154686689590803824187,20952076097787923848866419557958010591451541910355446375784730028915647228249],[18566066011684678568885649123674957074249585518641554932716632277292855153842,6470374219420229210187114229231342604112241790502587950558902833071586584090]],"C":[18869937697941340802318991706403654013215397448685130150843326018726535552502,8399006894026466651718959278957840411670575275113677591162226392583051491877],"Inputs":[8382471479713636303903675749137703248387104876156021657409718717386550288255,11665027831340046981101462667750854285054296068534204360488168379139618721774,4408301714000765206938897260593154811260687353421837370346725156225419671552]}', 23, 0);
INSERT INTO "public"."proof_sender" VALUES (24, '2022-06-09 07:27:34.520581+00', '2022-06-09 07:27:34.520581+00', NULL, '{"A":[8613180370901318718535658826393657157718755207167640216722179502090122844596,7971365462724084651970567060770767223533461273856821699255777023502432582047],"B":[[11516845629720987684566185779404053777946172153484679065223422801568720688765,1054292701819868932735155683802379429757984484931516139527833884187119639764],[17222905044158934511657582945725149487047980273853254990095469965574557553059,17477985489841557074413411764295171684047570166779175602095250545750914557380]],"C":[15777708239550131531845517976436796320872537263437302331163260884736381117735,14019692207264213807601540972115716111720730295561640445545636361972264817784],"Inputs":[11665027831340046981101462667750854285054296068534204360488168379139618721774,8818957056799975707280962631258144156165571563895234601332200658646258521673,3643790213631949638864011697999856248793190006166832594111498243928238663307]}', 24, 0);
INSERT INTO "public"."proof_sender" VALUES (25, '2022-06-09 07:27:44.500311+00', '2022-06-09 07:27:44.500311+00', NULL, '{"A":[15907645387255906742293391976489550924924599442618279118513345514194139207749,14299252587697908068611513008159920094383486195374171300518689111245539354090],"B":[[1799064402919018012127201967533560600332495390927724231009113377096689031408,7411343843117199858564464170634859204350865750826298776609685467388249520344],[9779149591180700989360511827703341472063602823885947050689112328670212205416,9342713518381803847579955466085361551677908426489131536216915959174949180491]],"C":[16563916302308965019361984399780929675350645386842313591102501581696612381506,9966615749352168923007071815883740598284539694814305279346358061010472882966],"Inputs":[8818957056799975707280962631258144156165571563895234601332200658646258521673,9956559373054189521626153440752156380238465255941779671710647272705605843262,8240640732309855160099724147307196178791653968869763265416528684501512078130]}', 25, 0);
INSERT INTO "public"."proof_sender" VALUES (26, '2022-06-09 07:27:54.547263+00', '2022-06-09 07:27:54.547263+00', NULL, '{"A":[13278734818578729076788456409748139707332537534106790890065762058518357911317,732301948154512214339264190683228344624000693108116330381339967333728024523],"B":[[8062897158964664459698142749075207797521593608674749601607549387249478838369,6406497615395643218977142812598202091354173200192125929125964380656165127531],[13866921034158733047477149440298064321847924367044897483057198282082356920207,8558319630986355638688323043606199717274600482790105366148830850597179876078]],"C":[20413444276586054130791024748995999792808863624570698781301686567290319680246,556709541424462817359791800349120421572936831535638901907151249589940357489],"Inputs":[9956559373054189521626153440752156380238465255941779671710647272705605843262,17889387022434688957087371144479155250645584381568910126928831640401314877701,113415291948729229580264397657414532692480125788587862562577637314771872980]}', 26, 0);

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
INSERT INTO "public"."sys_config" VALUES (5, '2022-06-08 08:24:27.338662+00', '2022-06-08 08:24:27.338662+00', NULL, 'ZkbasContract', '0x39c6354FdB9009E15B4006205E5Aa4C08c558c35', 'string', 'Zecrey contract on BSC');
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
INSERT INTO "public"."tx" VALUES (1, '2022-06-09 07:09:06.044514+00', '2022-06-09 07:09:06.044514+00', NULL, '9f5005a9-e704-11ec-b6f3-988fe0603efa', 1, '0', 0, 1, 1, 2, '21422f9bebac15af8ddc504da0dbb88020c1a4de7e7b6722fe00acb0ed968942', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":0,"AccountName":"treasury.legend","AccountNameHash":"FnxTYwiKQKSDmRKocvQxZCcHQMfphuxVOXstWDMXq0o=","PubKey":"fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998"}', '', '', 0, 0, 0);
INSERT INTO "public"."tx" VALUES (2, '2022-06-09 07:09:06.057049+00', '2022-06-09 07:09:06.057049+00', NULL, '9f50d170-e704-11ec-b6f3-988fe0603efa', 1, '0', 0, 1, 2, 3, '1b2ff4ae0d507a971fb267849af6a28000b1d483865c5a610cc47db6f196c672', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":1,"AccountName":"gas.legend","AccountNameHash":"CkjpiSpFoE0MWw8jWjrrB7khN7pxpZucRXd0uv3pWYM=","PubKey":"1ec94e497abe0fbb87f9ed2843e21163e17e3e97f6bbbae7a88399b826474f93"}', '', '', 1, 0, 0);
INSERT INTO "public"."tx" VALUES (3, '2022-06-09 07:09:06.066023+00', '2022-06-09 07:09:06.066023+00', NULL, '9f50f093-e704-11ec-b6f3-988fe0603efa', 1, '0', 0, 1, 3, 4, '189517f4cfb59471e3539dae36b8f53cb1264d407daf6afbf86132917f1cbafc', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":2,"AccountName":"sher.legend","AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","PubKey":"b0b6f7466154578ec66d51a335ead65ffd6a7210567fad9e68b6df8a5ce5dd85"}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (4, '2022-06-09 07:09:06.073642+00', '2022-06-09 07:09:06.073642+00', NULL, '9f510cb1-e704-11ec-b6f3-988fe0603efa', 1, '0', 0, 1, 4, 5, '08b2dc20da16235e692de317d6134578159532d4f081827bd29a5fc783fcc2b7', -1, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":1,"AccountIndex":3,"AccountName":"gavin.legend","AccountNameHash":"HFTAnJj3renV7rpBJKx8kS5laZo/dvpl1x6vY1nZvOs=","PubKey":"0500ccea3ca064968f5292b850ac8d4d3ee48d499357351a5ebfa2f30bb6070e"}', '', '', 3, 0, 0);
INSERT INTO "public"."tx" VALUES (5, '2022-06-09 07:09:06.079888+00', '2022-06-09 07:09:06.079888+00', NULL, '9f510cb1-e704-11ec-b6f4-988fe0603efa', 4, '0', 0, 1, 5, 6, '236e2c312a52cfbe96fc14a0693ea0f26d59fae774b35d44ddcf7737d965902f', -1, -1, 0, '100000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (6, '2022-06-09 07:09:06.088848+00', '2022-06-09 07:09:06.088848+00', NULL, '9f510cb1-e704-11ec-b6f5-988fe0603efa', 4, '0', 0, 1, 6, 7, '029cfe1c99565d3722f32b6bdb4ee5740d4f4c78bf318968c366c9c7e82d9ba7', -1, -1, 0, '100000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":3,"AccountNameHash":"HFTAnJj3renV7rpBJKx8kS5laZo/dvpl1x6vY1nZvOs=","AssetId":0,"AssetAmount":100000000000000000}', '', '', 3, 0, 0);
INSERT INTO "public"."tx" VALUES (7, '2022-06-09 07:09:06.096384+00', '2022-06-09 07:09:06.096384+00', NULL, '9f510cb1-e704-11ec-b6f6-988fe0603efa', 4, '0', 0, 1, 7, 8, '25cade17a4affef4114a06b8ae6e8e18651a8c4aa0aa01e1c20abce23ad614ec', -1, -1, 1, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (8, '2022-06-09 07:09:06.103059+00', '2022-06-09 07:09:06.103059+00', NULL, '9f510cb1-e704-11ec-b6f7-988fe0603efa', 4, '0', 0, 1, 8, 9, '17a21620fe89a6ef610ceea7b2c6230dba84731020a11bd081b46ba23c1cae94', -1, -1, 2, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":4,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":2,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (9, '2022-06-09 07:09:06.111132+00', '2022-06-09 07:09:06.111132+00', NULL, '9f510cb1-e704-11ec-b6f8-988fe0603efa', 2, '0', 0, 1, 9, 10, '0f5cf7c3fa8452ccb12d87b99952cfde059999f3767ddbc032994d94f3fe24ba', -1, 0, 0, '0', '0', '{"TxType":2,"PairIndex":0,"AssetAId":0,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0);
INSERT INTO "public"."tx" VALUES (10, '2022-06-09 07:09:06.120087+00', '2022-06-09 07:09:06.120087+00', NULL, '9f510cb1-e704-11ec-b6f9-988fe0603efa', 2, '0', 0, 1, 10, 11, '0945597849e7df9b43bfade724068c4d5a9d6039da208e6b829feb530ce784cd', -1, 1, 0, '0', '0', '{"TxType":2,"PairIndex":1,"AssetAId":0,"AssetBId":1,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0);
INSERT INTO "public"."tx" VALUES (11, '2022-06-09 07:09:06.127598+00', '2022-06-09 07:09:06.127598+00', NULL, '9f510cb1-e704-11ec-b6fa-988fe0603efa', 2, '0', 0, 1, 11, 12, '1671dd749a5a522f18908e28512d1c6c10034740923bbe9bab5664585b87411d', -1, 2, 0, '0', '0', '{"TxType":2,"PairIndex":2,"AssetAId":1,"AssetBId":2,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '', '', -1, 0, 0);
INSERT INTO "public"."tx" VALUES (12, '2022-06-09 07:09:06.134315+00', '2022-06-09 07:09:06.134315+00', NULL, '9f510cb1-e704-11ec-b6fb-988fe0603efa', 3, '0', 0, 1, 12, 13, '08ef9af5048b3df61fe3bd025a8db3f47b591a0136281cb3325e7e89930f3925', -1, 1, 0, '0', '0', '{"TxType":3,"PairIndex":1,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', '', '', -1, 0, 0);
INSERT INTO "public"."tx" VALUES (13, '2022-06-09 07:09:06.143621+00', '2022-06-09 07:09:06.143621+00', NULL, '9f51d005-e704-11ec-b6fb-988fe0603efa', 5, '0', 0, 1, 13, 14, '23f9301b57dbde40b067fc04f2bb2e5241b58739845efa223de352a8a14dd2ae', 0, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":5,"AccountIndex":2,"NftIndex":0,"NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftContentHash":"q9G2rnlQf3tKMqhKtklbyf7mdFDtMW27p2us6KPFGXs=","NftL1TokenId":0,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","CollectionId":0}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (14, '2022-06-09 07:09:06.153872+00', '2022-06-09 07:09:06.153872+00', NULL, '9f51d005-e704-11ec-b6fc-988fe0603efa', 17, '0', 0, 1, 14, 15, '1785a0c0ef9c282c5dddde78ad80b9689d34cda4a59ed35fcf4f00966ff034e2', -1, -1, 1, '100000000000000000000', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":17,"AccountIndex":2,"AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","AssetId":1,"AssetAmount":100000000000000000000}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (15, '2022-06-09 07:09:06.161637+00', '2022-06-09 07:09:06.161637+00', NULL, '9f51d005-e704-11ec-b6fd-988fe0603efa', 18, '0', 0, 1, 15, 16, '28ff96ba5f7e023a7ed9d446cb412fc6965a6ed68d1439b357bb4014ec57a8a4', 0, -1, 0, '0', '0x56744Dc80a3a520F0cCABf083AC874a4bf6433F3', '{"TxType":18,"AccountIndex":2,"CreatorAccountIndex":0,"CreatorTreasuryRate":0,"NftIndex":0,"CollectionId":0,"NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","AccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","CreatorAccountNameHash":"AA==","NftContentHash":"q9G2rnlQf3tKMqhKtklbyf7mdFDtMW27p2us6KPFGXs=","NftL1TokenId":0}', '', '', 2, 0, 0);
INSERT INTO "public"."tx" VALUES (16, '2022-06-09 07:09:06.175353+00', '2022-06-09 07:09:06.175353+00', NULL, '01de4078-304a-406d-9995-7c8550248f28', 6, '5000', 2, 1, 16, 17, '0c599d212ed3641e0b6df735e8b04dd627accfdafbbfa38c173af5f38efb433e', -1, -1, 0, '100000', '', '{"FromAccountIndex":2,"ToAccountIndex":3,"ToAccountNameHash":"1c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb","AssetId":0,"AssetAmount":100000,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"Memo":"transfer","CallData":"","CallDataHash":"Dd56AihX/sG4/6dmSpN6JQ065o81YGF1TTUx4mdBA9g=","ExpiredAt":1654761533445,"Nonce":1,"Sig":"epyzhZA39/F3mHPAvv8dz8NgPTtPWlYqPs9tEyDNsQgA8A4bi4ruGJe6evoUJ9BdWR49SJ1SCaJ+on1y2QyEFg=="}', '', 'transfer', 2, 1, 1654761533445);
INSERT INTO "public"."tx" VALUES (17, '2022-06-09 07:09:06.187051+00', '2022-06-09 07:09:06.187051+00', NULL, '1d4024d1-a49a-4f6a-9221-b716fedbd4aa', 10, '5000', 2, 1, 17, 18, '2d425cddc3d5aaec5ed0dd1465e7038e7fbbcc679e4a4c6c620742134ef93714', -1, -1, 0, '10000000', '0x99AC8881834797ebC32f185ee27c2e96842e1a47', '{"FromAccountIndex":2,"AssetId":0,"AssetAmount":10000000,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ToAddress":"0x99AC8881834797ebC32f185ee27c2e96842e1a47","ExpiredAt":1654761540020,"Nonce":2,"Sig":"YLVT6d4HuMWX7zPw3gcsmqs//dqE0xnfqTbSyLxx3pMDN7IvgmErKDXGiZ/XKC75wf2I03R2dXpmZ6bercDVHA=="}', '', '', 2, 2, 1654761540020);
INSERT INTO "public"."tx" VALUES (18, '2022-06-09 07:09:06.200171+00', '2022-06-09 07:09:06.200171+00', NULL, 'd9f93c69-3be2-483c-b746-d98022a61ecb', 8, '5000', 2, 1, 18, 19, '1211d91f4e22bd2f1aa38daaec68431b35fd37f8a272d147ebe7ba3e73a58555', -1, 0, 0, '100000', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":0,"AssetAAmount":100000,"AssetBId":2,"AssetBAmount":100000,"LpAmount":100000,"KLast":10000000000,"TreasuryAmount":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761547992,"Nonce":3,"Sig":"7x6AUCZwD+fmcXWDs0WERCMR+rIBRSDDlzSV3vrjggQEdNd8uj6ghy3uzqrM2oeqa/9gP8vRzPFHYmUvmCKsTQ=="}', '', '', 2, 3, 1654761547992);
INSERT INTO "public"."tx" VALUES (19, '2022-06-09 07:09:06.212437+00', '2022-06-09 07:09:06.212437+00', NULL, '8d60898f-ef87-4726-9322-1bda3fd22c2b', 7, '5000', 0, 1, 19, 20, '1db7fb69796667194858edf7aea403110c42cddc0907b3953181e1184907fb35', -1, 0, 0, '100', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":2,"AssetAAmount":100,"AssetBId":0,"AssetBMinAmount":98,"AssetBAmountDelta":99,"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"ExpiredAt":1654761555052,"Nonce":4,"Sig":"qCqSqikwaTLE/4VDURQHRYK+9gykmtanhYBv/ByGSoMGBJq+8D7z2b9yc0D8M1zNmfeD5YvpCpJkElsjarviGw=="}', '', '', 2, 4, 1654761555052);
INSERT INTO "public"."tx" VALUES (20, '2022-06-09 07:09:06.22453+00', '2022-06-09 07:09:06.22453+00', NULL, '4282c649-d9be-49fc-bc34-48dd4bab5f15', 9, '5000', 2, 1, 20, 21, '2e888850863cf0c2dffa40c8a0c162749f1f93da6ddf225030a01648cdfc26d6', -1, 0, 0, '100', '', '{"FromAccountIndex":2,"PairIndex":0,"AssetAId":0,"AssetAMinAmount":98,"AssetBId":2,"AssetBMinAmount":99,"LpAmount":100,"AssetAAmountDelta":99,"AssetBAmountDelta":100,"KLast":9980200000,"TreasuryAmount":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761562031,"Nonce":5,"Sig":"A9Qrmz5Uj7mDA3WfUGRhFoNuWFMtO+RwxwfkbkmglBEAXouZOVBhU4iguZYAEtwO6xFsPYrGFI92A0z21KFk6Q=="}', '', '', 2, 5, 1654761562031);
INSERT INTO "public"."tx" VALUES (21, '2022-06-09 07:09:06.235545+00', '2022-06-09 07:09:06.235545+00', NULL, '0c3c72c1-ff2d-4ba3-b026-93b7cec8e6a1', 11, '5000', 2, 1, 21, 22, '140622efbca882cddada16ee07f9cc8718b69998a5d8d7922fa7591f2f533edd', -1, -1, 0, 'sher.legend', '0', '{"AccountIndex":2,"CollectionId":1,"Name":"Zecrey Collection","Introduction":"Wonderful zecrey!","GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761569287,"Nonce":6,"Sig":"HrXIp5Vauk06jV3Jj4Lke/+XYp3ThhQsVAx8QksB1aYFVYeEZOf5nmIML+U3TSatTwyLCGMmomQSjGMxVtouxw=="}', '', '', 2, 6, 1654761569287);
INSERT INTO "public"."tx" VALUES (22, '2022-06-09 07:09:06.245099+00', '2022-06-09 07:09:06.245099+00', NULL, '69b6e9bb-0f8b-4b20-af35-35c6945489aa', 12, '5000', 2, 1, 22, 23, '12884f8bb4852d02ad1f654daa7a2fc230c539e5a1d3dcd83a474eed139e1f7f', 1, -1, 0, '0', '', '{"CreatorAccountIndex":2,"ToAccountIndex":3,"ToAccountNameHash":"1c54c09c98f7ade9d5eeba4124ac7c912e65699a3f76fa65d71eaf6359d9bceb","NftIndex":1,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftCollectionId":1,"CreatorTreasuryRate":0,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761575344,"Nonce":7,"Sig":"KAC/ZwgC7PJo4KDqBpVI048lfSjzli9qLXp3I4CFBJIBZY/c301Cm46AgoxkfaC02p8M5W263VWSEOsW+YATRg=="}', '', '', 2, 7, 1654761575344);
INSERT INTO "public"."tx" VALUES (23, '2022-06-09 07:09:06.256098+00', '2022-06-09 07:09:06.256098+00', NULL, 'a512199e-4146-407c-9b13-0039c8796650', 13, '5000', 0, 1, 23, 24, '19ca2bf9cca9b55f61c3f2d352ab486ac7e529670b6af93171054cb8d82f4fee', 1, -1, 0, '0', '', '{"FromAccountIndex":3,"ToAccountIndex":2,"ToAccountNameHash":"214a2d7af2022dfaee49dadb8992d3d7c225d8ae36109b531c28406dd69aad45","NftIndex":1,"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"CallData":"","CallDataHash":"Dd56AihX/sG4/6dmSpN6JQ065o81YGF1TTUx4mdBA9g=","ExpiredAt":1654761583894,"Nonce":1,"Sig":"G/YUzJOh4sp8ZF27Mz4s5hvLBC1P63uZbIVjMzHjqRkB0wTz30LsTSC5VIcxILqv7a/dCw4qd4Y3LYOtgLPW1w=="}', '', '', 3, 1, 1654761583894);
INSERT INTO "public"."tx" VALUES (24, '2022-06-09 07:09:06.269189+00', '2022-06-09 07:09:06.269189+00', NULL, '79245e26-0ec4-486e-9556-2fd477928380', 14, '5000', 0, 1, 24, 25, '137f5a5193ca65babef27c7f5be3ebc0eee3fd5c4de748d4cc0e736b80a99649', 1, -1, 0, '10000', '', '{"AccountIndex":2,"BuyOffer":{"Type":0,"OfferId":0,"AccountIndex":3,"NftIndex":1,"AssetId":0,"AssetAmount":10000,"ListedAt":1654754390138,"ExpiredAt":1654761590138,"TreasuryRate":200,"Sig":"CrwNdL+oHhdWBgJ0j+O/IY5Ca5qnBw6kDkPyUWD4wywApriICAXoTooPa//9vP9QRDPEQsHu9C2vvfeNaXeUGA=="},"SellOffer":{"Type":1,"OfferId":0,"AccountIndex":2,"NftIndex":1,"AssetId":0,"AssetAmount":10000,"ListedAt":1654754390138,"ExpiredAt":1654761590138,"TreasuryRate":200,"Sig":"1k2/LHCg9jCQ2+S9qYW8hWRLFryR7xQU+32zmjlEAAICKm6Tlks2bqoUXLBiOe5VNwUIMJ5gwJxTKOlJbOExbA=="},"GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"CreatorAmount":0,"TreasuryAmount":200,"Nonce":8,"ExpiredAt":1654761590138,"Sig":"Y/HMbBEAcLfdg5+eqAo/Gz+Nq8ZHdmLbm+SRUBZAB5cASst6Eo7UiL6O7+2IGNa7lBij3RgRvPn8bcBbxEO+Ww=="}', '', '', 2, 8, 1654761590138);
INSERT INTO "public"."tx" VALUES (25, '2022-06-09 07:09:06.281436+00', '2022-06-09 07:09:06.281436+00', NULL, 'fa2c5d73-deab-494a-a5c1-64938d1430aa', 15, '5000', 2, 1, 25, 26, '16033680a98409353095c6679b48d1fe06a03ec709b55e448a4c4a56e229393e', -1, -1, 0, 'sher.legend', '0', '{"AccountIndex":2,"OfferId":1,"GasAccountIndex":1,"GasFeeAssetId":2,"GasFeeAssetAmount":5000,"ExpiredAt":1654761601706,"Nonce":9,"Sig":"XQyUjK2wFu2opRmOhCnmDtCsVFeyj0MDofWLqyQCBBQCRi4zpBIxphGhuSMkoDO1WiFWkxVaRXwINrpOKjfMug=="}', '', '', 2, 9, 1654761601706);
INSERT INTO "public"."tx" VALUES (26, '2022-06-09 07:09:06.292598+00', '2022-06-09 07:09:06.292598+00', NULL, '5c3021a0-15bc-42b3-a966-11b0afdb0c73', 16, '5000', 0, 1, 26, 27, '278d08c3c1a50ed6e932abdfde1555b7843c43de10c1fded32f7cfc2987c9105', 1, -1, 0, '0', '', '{"AccountIndex":3,"CreatorAccountIndex":2,"CreatorAccountNameHash":"IUotevICLfruSdrbiZLT18Il2K42EJtTHChAbdaarUU=","CreatorTreasuryRate":0,"NftIndex":1,"NftContentHash":"Bmpl0+Q5etBfsuf1DqwWBkenSGws7bvqxkYkyL7qIvE=","NftL1Address":"0","NftL1TokenId":0,"CollectionId":1,"ToAddress":"0xd5Aa3B56a2E2139DB315CdFE3b34149c8ed09171","GasAccountIndex":1,"GasFeeAssetId":0,"GasFeeAssetAmount":5000,"ExpiredAt":1654761611582,"Nonce":2,"Sig":"sPJWFi9pTtfv6Z9zFI/QlQ2M9APJdOrWuPuq8wVnZ5kDbJkAuE2MoDA7o07rv5g/nz/cjloY2us88w1dkg+5sQ=="}', '', '', 3, 2, 1654761611582);

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
INSERT INTO "public"."tx_detail" VALUES (1, '2022-06-09 07:09:06.08039+00', '2022-06-09 07:09:06.08039+00', NULL, 5, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (2, '2022-06-09 07:09:06.089081+00', '2022-06-09 07:09:06.089081+00', NULL, 6, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (3, '2022-06-09 07:09:06.096937+00', '2022-06-09 07:09:06.096937+00', NULL, 7, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (4, '2022-06-09 07:09:06.103587+00', '2022-06-09 07:09:06.103587+00', NULL, 8, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (5, '2022-06-09 07:09:06.111654+00', '2022-06-09 07:09:06.111654+00', NULL, 9, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":0,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":0,"TreasuryAccountIndex":0,"TreasuryRate":0}', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (6, '2022-06-09 07:09:06.120737+00', '2022-06-09 07:09:06.120737+00', NULL, 10, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":0,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":0,"TreasuryAccountIndex":0,"TreasuryRate":0}', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (7, '2022-06-09 07:09:06.128237+00', '2022-06-09 07:09:06.128237+00', NULL, 11, 2, 2, -1, '', '{"PairIndex":2,"AssetAId":0,"AssetA":0,"AssetBId":0,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":0,"TreasuryAccountIndex":0,"TreasuryRate":0}', '{"PairIndex":2,"AssetAId":1,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (8, '2022-06-09 07:09:06.134875+00', '2022-06-09 07:09:06.134875+00', NULL, 12, 1, 2, -1, '', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '{"PairIndex":1,"AssetAId":0,"AssetA":0,"AssetBId":1,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":50,"TreasuryAccountIndex":0,"TreasuryRate":10}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (9, '2022-06-09 07:09:06.144201+00', '2022-06-09 07:09:06.144201+00', NULL, 13, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (10, '2022-06-09 07:09:06.144201+00', '2022-06-09 07:09:06.144201+00', NULL, 13, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":2,"NftContentHash":"abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b","NftL1TokenId":"0","NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","CreatorTreasuryRate":0,"CollectionId":0}', 0, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (11, '2022-06-09 07:09:06.154429+00', '2022-06-09 07:09:06.154429+00', NULL, 14, 1, 1, 2, 'sher.legend', '{"AssetId":1,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":1,"Balance":-100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (12, '2022-06-09 07:09:06.162169+00', '2022-06-09 07:09:06.162169+00', NULL, 15, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (13, '2022-06-09 07:09:06.162169+00', '2022-06-09 07:09:06.162169+00', NULL, 15, 0, 3, 2, 'sher.legend', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":2,"NftContentHash":"abd1b6ae79507f7b4a32a84ab6495bc9fee67450ed316dbba76bace8a3c5197b","NftL1TokenId":"0","NftL1Address":"0xB7aD4A7E9459D0C1541Db2eEceceAcc7dBa803e1","CreatorTreasuryRate":0,"CollectionId":0}', '{"NftIndex":0,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (14, '2022-06-09 07:09:06.176047+00', '2022-06-09 07:09:06.176047+00', NULL, 16, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (15, '2022-06-09 07:09:06.176047+00', '2022-06-09 07:09:06.176047+00', NULL, 16, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":100000000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (16, '2022-06-09 07:09:06.176047+00', '2022-06-09 07:09:06.176047+00', NULL, 16, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (17, '2022-06-09 07:09:06.176047+00', '2022-06-09 07:09:06.176047+00', NULL, 16, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (18, '2022-06-09 07:09:06.187885+00', '2022-06-09 07:09:06.187885+00', NULL, 17, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999999900000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-10000000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (19, '2022-06-09 07:09:06.187885+00', '2022-06-09 07:09:06.187885+00', NULL, 17, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999995000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (20, '2022-06-09 07:09:06.187885+00', '2022-06-09 07:09:06.187885+00', NULL, 17, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (21, '2022-06-09 07:09:06.200912+00', '2022-06-09 07:09:06.200912+00', NULL, 18, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989900000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 2, 0);
INSERT INTO "public"."tx_detail" VALUES (22, '2022-06-09 07:09:06.200912+00', '2022-06-09 07:09:06.200912+00', NULL, 18, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999990000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 2, 0);
INSERT INTO "public"."tx_detail" VALUES (23, '2022-06-09 07:09:06.200912+00', '2022-06-09 07:09:06.200912+00', NULL, 18, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999890000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0, 2, 0);
INSERT INTO "public"."tx_detail" VALUES (24, '2022-06-09 07:09:06.200912+00', '2022-06-09 07:09:06.200912+00', NULL, 18, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989800000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":100000,"OfferCanceledOrFinalized":0}', 3, 0, 2, 0);
INSERT INTO "public"."tx_detail" VALUES (25, '2022-06-09 07:09:06.200912+00', '2022-06-09 07:09:06.200912+00', NULL, 18, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":0,"AssetBId":2,"AssetB":0,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '{"PairIndex":0,"AssetAId":0,"AssetA":100000,"AssetBId":2,"AssetB":100000,"LpAmount":100000,"KLast":10000000000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 4, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (26, '2022-06-09 07:09:06.200912+00', '2022-06-09 07:09:06.200912+00', NULL, 18, 0, 1, 0, 'treasury.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 5, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (27, '2022-06-09 07:09:06.200912+00', '2022-06-09 07:09:06.200912+00', NULL, 18, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":10000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 6, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (28, '2022-06-09 07:09:06.212936+00', '2022-06-09 07:09:06.212936+00', NULL, 19, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999885000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-100,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 3, 0);
INSERT INTO "public"."tx_detail" VALUES (29, '2022-06-09 07:09:06.212936+00', '2022-06-09 07:09:06.212936+00', NULL, 19, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989800000,"LpAmount":100000,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":99,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 3, 0);
INSERT INTO "public"."tx_detail" VALUES (30, '2022-06-09 07:09:06.212936+00', '2022-06-09 07:09:06.212936+00', NULL, 19, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989800099,"LpAmount":100000,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0, 3, 0);
INSERT INTO "public"."tx_detail" VALUES (31, '2022-06-09 07:09:06.212936+00', '2022-06-09 07:09:06.212936+00', NULL, 19, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":100000,"AssetBId":2,"AssetB":100000,"LpAmount":100000,"KLast":10000000000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '{"PairIndex":0,"AssetAId":0,"AssetA":-99,"AssetBId":2,"AssetB":100,"LpAmount":0,"KLast":0,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 3, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (32, '2022-06-09 07:09:06.212936+00', '2022-06-09 07:09:06.212936+00', NULL, 19, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 4, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (33, '2022-06-09 07:09:06.22503+00', '2022-06-09 07:09:06.22503+00', NULL, 20, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989795099,"LpAmount":100000,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":99,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 4, 0);
INSERT INTO "public"."tx_detail" VALUES (34, '2022-06-09 07:09:06.22503+00', '2022-06-09 07:09:06.22503+00', NULL, 20, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999884900,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":100,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 4, 0);
INSERT INTO "public"."tx_detail" VALUES (35, '2022-06-09 07:09:06.22503+00', '2022-06-09 07:09:06.22503+00', NULL, 20, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999885000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 0, 4, 0);
INSERT INTO "public"."tx_detail" VALUES (36, '2022-06-09 07:09:06.22503+00', '2022-06-09 07:09:06.22503+00', NULL, 20, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989795198,"LpAmount":100000,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":-100,"OfferCanceledOrFinalized":0}', 3, 0, 4, 0);
INSERT INTO "public"."tx_detail" VALUES (37, '2022-06-09 07:09:06.22503+00', '2022-06-09 07:09:06.22503+00', NULL, 20, 0, 1, 0, 'treasury.legend', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 4, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (38, '2022-06-09 07:09:06.22503+00', '2022-06-09 07:09:06.22503+00', NULL, 20, 0, 2, -1, '', '{"PairIndex":0,"AssetAId":0,"AssetA":99901,"AssetBId":2,"AssetB":100100,"LpAmount":100000,"KLast":10000000000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', '{"PairIndex":0,"AssetAId":0,"AssetA":-99,"AssetBId":2,"AssetB":-100,"LpAmount":-100,"KLast":9980200000,"FeeRate":30,"TreasuryAccountIndex":0,"TreasuryRate":5}', 5, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (39, '2022-06-09 07:09:06.22503+00', '2022-06-09 07:09:06.22503+00', NULL, 20, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":15000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 6, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (41, '2022-06-09 07:09:06.236063+00', '2022-06-09 07:09:06.236063+00', NULL, 21, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999880000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 0, 5, 0);
INSERT INTO "public"."tx_detail" VALUES (42, '2022-06-09 07:09:06.236063+00', '2022-06-09 07:09:06.236063+00', NULL, 21, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":20000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (43, '2022-06-09 07:09:06.246098+00', '2022-06-09 07:09:06.246098+00', NULL, 22, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999875000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 6, 1);
INSERT INTO "public"."tx_detail" VALUES (44, '2022-06-09 07:09:06.246098+00', '2022-06-09 07:09:06.246098+00', NULL, 22, 2, 1, 3, 'gavin.legend', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (45, '2022-06-09 07:09:06.246098+00', '2022-06-09 07:09:06.246098+00', NULL, 22, 1, 3, 3, 'gavin.legend', '{"NftIndex":1,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 2, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (46, '2022-06-09 07:09:06.246098+00', '2022-06-09 07:09:06.246098+00', NULL, 22, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":25000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (47, '2022-06-09 07:09:06.256597+00', '2022-06-09 07:09:06.256597+00', NULL, 23, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000100000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (48, '2022-06-09 07:09:06.256597+00', '2022-06-09 07:09:06.256597+00', NULL, 23, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (40, '2022-06-09 07:09:06.236063+00', '2022-06-09 07:09:06.236063+00', NULL, 21, 0, 4, 2, 'sher.legend', '0', '1', 0, 0, 5, 0);
INSERT INTO "public"."tx_detail" VALUES (49, '2022-06-09 07:09:06.256597+00', '2022-06-09 07:09:06.256597+00', NULL, 23, 1, 3, 2, 'sher.legend', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":2,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 2, -1, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (50, '2022-06-09 07:09:06.256597+00', '2022-06-09 07:09:06.256597+00', NULL, 23, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (51, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989795198,"LpAmount":99900,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (52, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000095000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":-10000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 1, 1, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (53, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000085000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":1}', 2, 1, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (54, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989790198,"LpAmount":99900,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":9800,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (55, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":1}', 4, 2, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (56, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":1}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 5, 3, 7, 1);
INSERT INTO "public"."tx_detail" VALUES (57, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 1, 3, -1, '', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":2,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', 6, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (58, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":10000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":200,"LpAmount":0,"OfferCanceledOrFinalized":0}', 7, 4, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (59, '2022-06-09 07:09:06.269598+00', '2022-06-09 07:09:06.269598+00', NULL, 24, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":10200,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 8, 4, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (60, '2022-06-09 07:09:06.282098+00', '2022-06-09 07:09:06.282098+00', NULL, 25, 2, 1, 2, 'sher.legend', '{"AssetId":2,"Balance":99999999999999870000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 8, 1);
INSERT INTO "public"."tx_detail" VALUES (61, '2022-06-09 07:09:06.282098+00', '2022-06-09 07:09:06.282098+00', NULL, 25, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":1}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":3}', 1, 0, 8, 1);
INSERT INTO "public"."tx_detail" VALUES (62, '2022-06-09 07:09:06.282098+00', '2022-06-09 07:09:06.282098+00', NULL, 25, 2, 1, 1, 'gas.legend', '{"AssetId":2,"Balance":30000,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":2,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (63, '2022-06-09 07:09:06.293268+00', '2022-06-09 07:09:06.293268+00', NULL, 26, 0, 1, 3, 'gavin.legend', '{"AssetId":0,"Balance":100000000000085000,"LpAmount":0,"OfferCanceledOrFinalized":1}', '{"AssetId":0,"Balance":-5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 0, 0, 1, 0);
INSERT INTO "public"."tx_detail" VALUES (64, '2022-06-09 07:09:06.293268+00', '2022-06-09 07:09:06.293268+00', NULL, 26, 1, 3, -1, '', '{"NftIndex":1,"CreatorAccountIndex":2,"OwnerAccountIndex":3,"NftContentHash":"066a65d3e4397ad05fb2e7f50eac160647a7486c2cedbbeac64624c8beea22f1","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":1}', '{"NftIndex":1,"CreatorAccountIndex":0,"OwnerAccountIndex":0,"NftContentHash":"0","NftL1TokenId":"0","NftL1Address":"0","CreatorTreasuryRate":0,"CollectionId":0}', 1, -1, 0, 0);
INSERT INTO "public"."tx_detail" VALUES (65, '2022-06-09 07:09:06.293268+00', '2022-06-09 07:09:06.293268+00', NULL, 26, 0, 1, 2, 'sher.legend', '{"AssetId":0,"Balance":99999999989799998,"LpAmount":99900,"OfferCanceledOrFinalized":3}', '{"AssetId":0,"Balance":0,"LpAmount":0,"OfferCanceledOrFinalized":0}', 2, 1, 9, 1);
INSERT INTO "public"."tx_detail" VALUES (66, '2022-06-09 07:09:06.293268+00', '2022-06-09 07:09:06.293268+00', NULL, 26, 0, 1, 1, 'gas.legend', '{"AssetId":0,"Balance":15200,"LpAmount":0,"OfferCanceledOrFinalized":0}', '{"AssetId":0,"Balance":5000,"LpAmount":0,"OfferCanceledOrFinalized":0}', 3, 2, 0, 0);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."account_history_id_seq"
OWNED BY "public"."account_history"."id";
SELECT setval('"public"."account_history_id_seq"', 40, true);

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
SELECT setval('"public"."block_for_commit_id_seq"', 26, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."block_id_seq"
OWNED BY "public"."block"."id";
SELECT setval('"public"."block_id_seq"', 27, true);

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
SELECT setval('"public"."l2_nft_exchange_id_seq"', 1, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_history_id_seq"
OWNED BY "public"."l2_nft_history"."id";
SELECT setval('"public"."l2_nft_history_id_seq"', 6, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_id_seq"
OWNED BY "public"."l2_nft"."id";
SELECT setval('"public"."l2_nft_id_seq"', 2, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."l2_nft_withdraw_history_id_seq"
OWNED BY "public"."l2_nft_withdraw_history"."id";
SELECT setval('"public"."l2_nft_withdraw_history_id_seq"', 2, true);

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
SELECT setval('"public"."liquidity_history_id_seq"', 7, true);

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
SELECT setval('"public"."mempool_tx_detail_id_seq"', 66, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."mempool_tx_id_seq"
OWNED BY "public"."mempool_tx"."id";
SELECT setval('"public"."mempool_tx_id_seq"', 26, true);

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
SELECT setval('"public"."proof_sender_id_seq"', 26, true);

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
SELECT setval('"public"."tx_detail_id_seq"', 66, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."tx_id_seq"
OWNED BY "public"."tx"."id";
SELECT setval('"public"."tx_id_seq"', 26, true);

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
