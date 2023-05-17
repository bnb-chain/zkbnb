create table if not exists sys_config
(
    id         bigserial
    primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name       text,
    value      text,
    value_type text,
    comment    text
);

-- alter table sys_config
--     owner to postgres;

create index if not exists idx_sys_config_name
    on sys_config (name);

create index if not exists idx_sys_config_deleted_at
    on sys_config (deleted_at);

create table if not exists account
(
    id               bigserial
    primary key,
    created_at       timestamp with time zone,
    updated_at       timestamp with time zone,
    deleted_at       timestamp with time zone,
    account_index    bigint,
    public_key       text,
    l1_address       text,
    nonce            bigint,
    collection_nonce bigint,
    asset_info       text,
    asset_root       text,
    l2_block_height  bigint,
    status           bigint
);

-- alter table account
--     owner to postgres;

create index if not exists idx_account_l2_block_height
    on account (l2_block_height);

create unique index if not exists idx_account_l1_address
    on account (l1_address);

create unique index if not exists idx_account_account_index
    on account (account_index);

create index if not exists idx_account_deleted_at
    on account (deleted_at);

create table if not exists account_history
(
    id               bigserial
    primary key,
    created_at       timestamp with time zone,
    updated_at       timestamp with time zone,
    deleted_at       timestamp with time zone,
    account_index    bigint,
    public_key       text,
    l1_address       text,
    nonce            bigint,
    collection_nonce bigint,
    asset_info       text,
    asset_root       text,
    l2_block_height  bigint,
    status           bigint
);

-- alter table account_history
--     owner to postgres;

create index if not exists idx_account_history_l2_block_height
    on account_history (l2_block_height);

create index if not exists idx_account_history_l1_address
    on account_history (l1_address);

create index if not exists idx_account_history_account_index
    on account_history (account_index);

create index if not exists idx_account_history_deleted_at
    on account_history (deleted_at);

create table if not exists asset
(
    id           bigserial
    primary key,
    created_at   timestamp with time zone,
    updated_at   timestamp with time zone,
    deleted_at   timestamp with time zone,
    asset_id     bigint,
    asset_name   text,
    asset_symbol text,
    l1_address   text,
    decimals     bigint,
    status       bigint,
    is_gas_asset bigint
);

-- alter table asset
--     owner to postgres;

create unique index if not exists idx_asset_asset_id
    on asset (asset_id);

create index if not exists idx_asset_deleted_at
    on asset (deleted_at);

create index if not exists idx_asset_is_gas_asset
    on asset (is_gas_asset);

create index if not exists idx_asset_l1_address
    on asset (l1_address);

create index if not exists idx_asset_asset_symbol
    on asset (asset_symbol);

create table if not exists pool_tx
(
    id                 bigserial
    primary key,
    created_at         timestamp with time zone,
    updated_at         timestamp with time zone,
    deleted_at         timestamp with time zone,
    tx_hash            text,
    tx_type            bigint,
    tx_info            text,
    account_index      bigint,
    nonce              bigint,
    from_account_index bigint,
    to_account_index   bigint,
    expired_at         bigint,
    gas_fee            text,
    gas_fee_asset_id   bigint,
    nft_index          bigint,
    collection_id      bigint,
    asset_id           bigint,
    tx_amount          text,
    memo               text,
    extra_info         text,
    native_address     text,
    is_create_account  boolean,
    tx_index           bigint,
    channel_name       text,
    block_height       bigint,
    block_id           bigint,
    tx_status          bigint,
    rollback           boolean,
    l1_request_id      bigint
);

-- alter table pool_tx
--     owner to postgres;

create index if not exists idx_pool_tx_tx_status
    on pool_tx (tx_status);

create index if not exists idx_pool_tx_block_height
    on pool_tx (block_height);

create index if not exists idx_pool_tx_tx_type
    on pool_tx (tx_type);

create unique index if not exists idx_pool_tx_tx_hash
    on pool_tx (tx_hash);

create index if not exists idx_pool_tx_deleted_at
    on pool_tx (deleted_at);

create index if not exists idx_pool_tx_nft_index
    on pool_tx (nft_index);

create index if not exists idx_pool_tx_to_account_index
    on pool_tx (to_account_index);

create index if not exists idx_pool_tx_from_account_index
    on pool_tx (from_account_index);

create index if not exists idx_pool_tx_account_index_nonce
    on pool_tx (account_index, nonce);

create table if not exists block
(
    id                                   bigserial
    primary key,
    created_at                           timestamp with time zone,
    updated_at                           timestamp with time zone,
    deleted_at                           timestamp with time zone,
    block_size                           integer,
    block_commitment                     text,
    block_height                         bigint,
    state_root                           text,
    priority_operations                  bigint,
    pending_on_chain_operations_hash     text,
    pending_on_chain_operations_pub_data text,
    committed_tx_hash                    text,
    committed_at                         bigint,
    verified_tx_hash                     text,
    verified_at                          bigint,
    block_status                         bigint,
    account_indexes                      text,
    nft_indexes                          text
);

-- alter table block
--     owner to postgres;

create index if not exists idx_block_block_status
    on block (block_status);

create unique index if not exists idx_block_block_height
    on block (block_height);

create index if not exists idx_block_block_commitment
    on block (block_commitment);

create index if not exists idx_block_deleted_at
    on block (deleted_at);

create table if not exists tx
(
    id                 bigserial
    primary key,
    created_at         timestamp with time zone,
    updated_at         timestamp with time zone,
    deleted_at         timestamp with time zone,
    tx_hash            text,
    tx_type            bigint,
    tx_info            text,
    account_index      bigint,
    nonce              bigint,
    from_account_index bigint,
    to_account_index   bigint,
    expired_at         bigint,
    gas_fee            text,
    gas_fee_asset_id   bigint,
    nft_index          bigint,
    collection_id      bigint,
    asset_id           bigint,
    tx_amount          text,
    memo               text,
    extra_info         text,
    native_address     text,
    is_create_account  boolean,
    tx_index           bigint,
    channel_name       text,
    block_height       bigint,
    block_id           bigint,
    tx_status          bigint,
    pool_tx_id         bigint,
    verify_at          timestamp with time zone
);

-- alter table tx
--     owner to postgres;

create index if not exists idx_tx_block_height
    on tx (block_height);

create unique index if not exists idx_tx_pool_tx_id
    on tx (pool_tx_id);

create index if not exists idx_tx_nft_index
    on tx (nft_index);

create index if not exists idx_tx_to_account_index
    on tx (to_account_index);

create index if not exists idx_tx_from_account_index
    on tx (from_account_index);

create index if not exists idx_tx_tx_type
    on tx (tx_type);

create unique index if not exists idx_tx_tx_hash
    on tx (tx_hash);

create index if not exists idx_tx_deleted_at
    on tx (deleted_at);

create index if not exists idx_tx_tx_status
    on tx (tx_status);

create table if not exists tx_detail
(
    id               bigserial
    primary key,
    created_at       timestamp with time zone,
    updated_at       timestamp with time zone,
    deleted_at       timestamp with time zone,
    pool_tx_id       bigint,
    asset_id         bigint,
    asset_type       bigint,
    account_index    bigint,
    l1_address       text,
    balance          text,
    balance_delta    text,
    "order"          bigint,
    account_order    bigint,
    nonce            bigint,
    collection_nonce bigint,
    is_gas           boolean default false,
    public_key       text,
    block_height     bigint
);

-- alter table tx_detail
--     owner to postgres;

create index if not exists idx_tx_detail_block_height
    on tx_detail (block_height);

create index if not exists idx_tx_detail_account_index
    on tx_detail (account_index);

create index if not exists idx_tx_detail_pool_tx_id
    on tx_detail (pool_tx_id);

create index if not exists idx_tx_detail_deleted_at
    on tx_detail (deleted_at);

create table if not exists compressed_block
(
    id                  bigserial
    primary key,
    created_at          timestamp with time zone,
    updated_at          timestamp with time zone,
    deleted_at          timestamp with time zone,
    block_size          integer,
    block_height        bigint,
    state_root          text,
    public_data         text,
    timestamp           bigint,
    public_data_offsets text,
    real_block_size     integer
);

-- alter table compressed_block
--     owner to postgres;

create index if not exists idx_compressed_block_block_height
    on compressed_block (block_height);

create index if not exists idx_compressed_block_deleted_at
    on compressed_block (deleted_at);

create table if not exists block_witness
(
    id           bigserial
    primary key,
    created_at   timestamp with time zone,
    updated_at   timestamp with time zone,
    deleted_at   timestamp with time zone,
    height       bigint,
    witness_data text,
    block_size   integer,
    status       bigint
);

-- alter table block_witness
--     owner to postgres;

create index if not exists idx_block_witness_status
    on block_witness (status);

create unique index if not exists idx_height
    on block_witness (height);

create index if not exists idx_block_witness_deleted_at
    on block_witness (deleted_at);

create table if not exists proof
(
    id           bigserial
    primary key,
    created_at   timestamp with time zone,
    updated_at   timestamp with time zone,
    deleted_at   timestamp with time zone,
    proof_info   text,
    block_number bigint,
    status       bigint
);

-- alter table proof
--     owner to postgres;

create index if not exists idx_proof_status
    on proof (status);

create unique index if not exists idx_number
    on proof (block_number);

create index if not exists idx_proof_deleted_at
    on proof (deleted_at);

create table if not exists l1_synced_block
(
    id              bigserial
    primary key,
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    deleted_at      timestamp with time zone,
    l1_block_height bigint,
    block_info      text,
    type            bigint
);

-- alter table l1_synced_block
--     owner to postgres;

create index if not exists idx_l1_synced_block_type
    on l1_synced_block (type);

create index if not exists idx_l1_synced_block_l1_block_height
    on l1_synced_block (l1_block_height);

create index if not exists idx_l1_synced_block_deleted_at
    on l1_synced_block (deleted_at);

create table if not exists priority_request
(
    id               bigserial
    primary key,
    created_at       timestamp with time zone,
    updated_at       timestamp with time zone,
    deleted_at       timestamp with time zone,
    l1_tx_hash       text,
    l1_block_height  bigint,
    sender_address   text,
    request_id       bigint,
    tx_type          bigint,
    pubdata          text,
    expiration_block bigint,
    status           bigint,
    l2_tx_hash       text
);

-- alter table priority_request
--     owner to postgres;

create index if not exists idx_priority_request_l1_block_height
    on priority_request (l1_block_height);

create index if not exists idx_priority_request_deleted_at
    on priority_request (deleted_at);

create index if not exists idx_priority_request_l2_tx_hash
    on priority_request (l2_tx_hash);

create index if not exists idx_priority_request_status
    on priority_request (status);

create index if not exists idx_priority_request_request_id
    on priority_request (request_id);

create table if not exists l1_rollup_tx
(
    id              bigserial
    primary key,
    created_at      timestamp with time zone,
    updated_at      timestamp with time zone,
    deleted_at      timestamp with time zone,
    l1_tx_hash      text,
    tx_status       bigint,
    tx_type         smallint,
    l2_block_height bigint,
    gas_price       bigint,
    gas_used        bigint,
    l1_nonce        bigint
);

-- alter table l1_rollup_tx
--     owner to postgres;

create index if not exists idx_tx_status
    on l1_rollup_tx (tx_status, tx_type);

create index if not exists idx_l1_rollup_tx_l1_tx_hash
    on l1_rollup_tx (l1_tx_hash);

create index if not exists idx_l1_rollup_tx_deleted_at
    on l1_rollup_tx (deleted_at);

create index if not exists idx_l1_nonce
    on l1_rollup_tx (l1_nonce);

create index if not exists l2_block_height
    on l1_rollup_tx (l2_block_height);

create table if not exists l2_nft
(
    id                    bigserial
    primary key,
    created_at            timestamp with time zone,
    updated_at            timestamp with time zone,
    deleted_at            timestamp with time zone,
    nft_index             bigint,
    creator_account_index bigint,
    owner_account_index   bigint,
    nft_content_hash      text,
    nft_content_type      bigint,
    royalty_rate          bigint,
    collection_id         bigint,
    l2_block_height       bigint
);

-- alter table l2_nft
--     owner to postgres;

create index if not exists idx_nft_index
    on l2_nft (l2_block_height);

create index if not exists idx_owner_account_index
    on l2_nft (owner_account_index, nft_content_hash);

create unique index if not exists idx_l2_nft_nft_index
    on l2_nft (nft_index);

create index if not exists idx_l2_nft_deleted_at
    on l2_nft (deleted_at);

create table if not exists l2_nft_history
(
    id                    bigserial
    primary key,
    created_at            timestamp with time zone,
    updated_at            timestamp with time zone,
    deleted_at            timestamp with time zone,
    nft_index             bigint,
    creator_account_index bigint,
    owner_account_index   bigint,
    nft_content_hash      text,
    nft_content_type      bigint,
    royalty_rate          bigint,
    collection_id         bigint,
    status                bigint,
    l2_block_height       bigint
);

-- alter table l2_nft_history
--     owner to postgres;

create index if not exists idx_l2_nft_history_deleted_at
    on l2_nft_history (deleted_at);

create table if not exists rollback
(
    id                bigserial
    primary key,
    created_at        timestamp with time zone,
    updated_at        timestamp with time zone,
    deleted_at        timestamp with time zone,
    from_block_height bigint,
    from_pool_tx_id   bigint,
    from_tx_hash      text,
    pool_tx_ids       text,
    block_heights     text,
    account_indexes   text,
    nft_indexes       text
);

-- alter table rollback
--     owner to postgres;

create index if not exists idx_rollback_from_pool_tx_id
    on rollback (from_pool_tx_id);

create index if not exists idx_rollback_from_block_height
    on rollback (from_block_height);

create index if not exists idx_rollback_deleted_at
    on rollback (deleted_at);

create table if not exists l2_nft_metadata_history
(
    id         bigserial
    primary key,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    nonce      bigint,
    nft_index  bigint,
    tx_hash    text,
    ipfs_cid   text,
    ipns_cid   text,
    ipns_name  text,
    ipns_id    text,
    metadata   text,
    mutable    text,
    status     bigint
);

-- alter table l2_nft_metadata_history
--     owner to postgres;

create index if not exists idx_l2_nft_metadata_history_status
    on l2_nft_metadata_history (status);

create index if not exists idx_l2_nft_metadata_history_tx_hash
    on l2_nft_metadata_history (tx_hash);

create index if not exists idx_l2_nft_metadata_history_nft_index
    on l2_nft_metadata_history (nft_index);

create index if not exists idx_l2_nft_metadata_history_deleted_at
    on l2_nft_metadata_history (deleted_at);

