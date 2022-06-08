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
CREATE TABLE `block` (
                                 `id` bigint unsigned NOT NULL AUTO_INCREMENT,
                                 `is_deleted` tinyint(1) DEFAULT 0 COMMENT 'is deleted?: 1 for yes, 0 for no',
                                 `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                                 `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                                 `block_commitment` varchar(255) NOT NULL,
                                 `block_height` bigint unsigned NOT NULL,
                                 `block_status` tinyint unsigned NOT NULL,
                                 `account_root` varchar(100) DEFAULT NULL,
                                 `verified_tx_hash` varchar(200) DEFAULT NULL,
                                 `verified_at` int NULL DEFAULT NULL,
                                 `committed_tx_hash` varchar(200) DEFAULT NULL,
                                 `committed_at` int NULL DEFAULT NULL,
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `idx_block_block_commitment` (`block_commitment`),
                                 KEY `idx_block_l2_block_height` (`block_height`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;