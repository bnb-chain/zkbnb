/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
CREATE TABLE `l1_asset_info` (
                                 `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                 `is_deleted` tinyint(1) DEFAULT 0 COMMENT 'is deleted?: 1 for yes, 0 for no',
                                 `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                                 `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                                 `chain_id` smallint NOT NULL,
                                 `asset_id` bigint  unsigned NOT NULL,
                                 `l2_asset_pk` bigint  unsigned NOT NULL,
                                 `asset_name` varchar(50) DEFAULT NULL,
                                 `asset_symbol` varchar(50) DEFAULT NULL,
                                 `asset_address` varchar(100) DEFAULT NULL,
                                 `decimals` tinyint unsigned NOT NULL,
                                 PRIMARY KEY (`id`),
                                 FOREIGN KEY (`l2_asset_pk`) REFERENCES l2_asset_info(`id`),
                                 KEY `idx_l1_asset_info_is_deleted` (`is_deleted`),
                                 KEY `idx_l1_asset_info_l2_asset_pk` (`l2_asset_pk`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;