--
-- PostgreSQL database dump
--

-- Dumped from database version 13.7
-- Dumped by pg_dump version 15.2 (Debian 15.2-1.pgdg110+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

DROP DATABASE IF EXISTS "zkbnb-perf";
--
-- Name: zkbnb-perf; Type: DATABASE; Schema: -; Owner: cloud
--

CREATE DATABASE "zkbnb-perf" WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.UTF-8';


ALTER DATABASE "zkbnb-perf" OWNER TO cloud;

\connect -reuse-previous=on "dbname='zkbnb-perf'"

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: cloud
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO cloud;

--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: cloud
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: account; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.account (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_index bigint,
    public_key text,
    l1_address text,
    nonce bigint,
    collection_nonce bigint,
    asset_info text,
    asset_root text,
    l2_block_height bigint,
    status bigint
);


ALTER TABLE public.account OWNER TO cloud;

--
-- Name: account_history; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.account_history (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    account_index bigint,
    public_key text,
    l1_address text,
    nonce bigint,
    collection_nonce bigint,
    asset_info text,
    asset_root text,
    l2_block_height bigint,
    status bigint
);


ALTER TABLE public.account_history OWNER TO cloud;

--
-- Name: account_history_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.account_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.account_history_id_seq OWNER TO cloud;

--
-- Name: account_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.account_history_id_seq OWNED BY public.account_history.id;


--
-- Name: account_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.account_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.account_id_seq OWNER TO cloud;

--
-- Name: account_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.account_id_seq OWNED BY public.account.id;


--
-- Name: asset; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.asset (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    asset_id bigint,
    asset_name text,
    asset_symbol text,
    l1_address text,
    decimals bigint,
    status bigint,
    is_gas_asset bigint
);


ALTER TABLE public.asset OWNER TO cloud;

--
-- Name: asset_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.asset_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.asset_id_seq OWNER TO cloud;

--
-- Name: asset_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.asset_id_seq OWNED BY public.asset.id;


--
-- Name: block; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.block (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    block_size integer,
    block_commitment text,
    block_height bigint,
    state_root text,
    priority_operations bigint,
    pending_on_chain_operations_hash text,
    pending_on_chain_operations_pub_data text,
    committed_tx_hash text,
    committed_at bigint,
    verified_tx_hash text,
    verified_at bigint,
    block_status bigint,
    account_indexes text,
    nft_indexes text
);


ALTER TABLE public.block OWNER TO cloud;

--
-- Name: block_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.block_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.block_id_seq OWNER TO cloud;

--
-- Name: block_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.block_id_seq OWNED BY public.block.id;


--
-- Name: block_witness; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.block_witness (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    height bigint,
    witness_data text,
    block_size integer,
    status bigint
);


ALTER TABLE public.block_witness OWNER TO cloud;

--
-- Name: block_witness_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.block_witness_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.block_witness_id_seq OWNER TO cloud;

--
-- Name: block_witness_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.block_witness_id_seq OWNED BY public.block_witness.id;


--
-- Name: compressed_block; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.compressed_block (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    block_size integer,
    block_height bigint,
    state_root text,
    public_data text,
    "timestamp" bigint,
    public_data_offsets text,
    real_block_size integer
);


ALTER TABLE public.compressed_block OWNER TO cloud;

--
-- Name: compressed_block_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.compressed_block_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.compressed_block_id_seq OWNER TO cloud;

--
-- Name: compressed_block_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.compressed_block_id_seq OWNED BY public.compressed_block.id;


--
-- Name: l1_rollup_tx; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.l1_rollup_tx (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    l1_tx_hash text,
    tx_status bigint,
    tx_type smallint,
    l2_block_height bigint,
    gas_price bigint,
    gas_used bigint,
    l1_nonce bigint
);


ALTER TABLE public.l1_rollup_tx OWNER TO cloud;

--
-- Name: l1_rollup_tx_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.l1_rollup_tx_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.l1_rollup_tx_id_seq OWNER TO cloud;

--
-- Name: l1_rollup_tx_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.l1_rollup_tx_id_seq OWNED BY public.l1_rollup_tx.id;


--
-- Name: l1_synced_block; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.l1_synced_block (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    l1_block_height bigint,
    block_info text,
    type bigint
);


ALTER TABLE public.l1_synced_block OWNER TO cloud;

--
-- Name: l1_synced_block_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.l1_synced_block_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.l1_synced_block_id_seq OWNER TO cloud;

--
-- Name: l1_synced_block_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.l1_synced_block_id_seq OWNED BY public.l1_synced_block.id;


--
-- Name: l2_nft; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.l2_nft (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    nft_index bigint,
    creator_account_index bigint,
    owner_account_index bigint,
    nft_content_hash text,
    nft_content_type bigint,
    royalty_rate bigint,
    collection_id bigint,
    l2_block_height bigint
);


ALTER TABLE public.l2_nft OWNER TO cloud;

--
-- Name: l2_nft_history; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.l2_nft_history (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    nft_index bigint,
    creator_account_index bigint,
    owner_account_index bigint,
    nft_content_hash text,
    nft_content_type bigint,
    royalty_rate bigint,
    collection_id bigint,
    status bigint,
    l2_block_height bigint
);


ALTER TABLE public.l2_nft_history OWNER TO cloud;

--
-- Name: l2_nft_history_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.l2_nft_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.l2_nft_history_id_seq OWNER TO cloud;

--
-- Name: l2_nft_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.l2_nft_history_id_seq OWNED BY public.l2_nft_history.id;


--
-- Name: l2_nft_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.l2_nft_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.l2_nft_id_seq OWNER TO cloud;

--
-- Name: l2_nft_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.l2_nft_id_seq OWNED BY public.l2_nft.id;


--
-- Name: l2_nft_metadata_history; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.l2_nft_metadata_history (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    nonce bigint,
    nft_index bigint,
    tx_hash text,
    ipfs_cid text,
    ipns_cid text,
    ipns_name text,
    ipns_id text,
    metadata text,
    mutable text,
    status bigint
);


ALTER TABLE public.l2_nft_metadata_history OWNER TO cloud;

--
-- Name: l2_nft_metadata_history_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.l2_nft_metadata_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.l2_nft_metadata_history_id_seq OWNER TO cloud;

--
-- Name: l2_nft_metadata_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.l2_nft_metadata_history_id_seq OWNED BY public.l2_nft_metadata_history.id;


--
-- Name: pool_tx; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.pool_tx (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tx_hash text,
    tx_type bigint,
    tx_info text,
    account_index bigint,
    nonce bigint,
    from_account_index bigint,
    to_account_index bigint,
    expired_at bigint,
    gas_fee text,
    gas_fee_asset_id bigint,
    nft_index bigint,
    collection_id bigint,
    asset_id bigint,
    tx_amount text,
    memo text,
    extra_info text,
    native_address text,
    is_create_account boolean,
    tx_index bigint,
    channel_name text,
    block_height bigint,
    block_id bigint,
    tx_status bigint,
    rollback boolean,
    l1_request_id bigint
);


ALTER TABLE public.pool_tx OWNER TO cloud;

--
-- Name: pool_tx_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.pool_tx_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.pool_tx_id_seq OWNER TO cloud;

--
-- Name: pool_tx_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.pool_tx_id_seq OWNED BY public.pool_tx.id;


--
-- Name: priority_request; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.priority_request (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    l1_tx_hash text,
    l1_block_height bigint,
    sender_address text,
    request_id bigint,
    tx_type bigint,
    pubdata text,
    expiration_block bigint,
    status bigint,
    l2_tx_hash text
);


ALTER TABLE public.priority_request OWNER TO cloud;

--
-- Name: priority_request_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.priority_request_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.priority_request_id_seq OWNER TO cloud;

--
-- Name: priority_request_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.priority_request_id_seq OWNED BY public.priority_request.id;


--
-- Name: proof; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.proof (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    proof_info text,
    block_number bigint,
    status bigint
);


ALTER TABLE public.proof OWNER TO cloud;

--
-- Name: proof_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.proof_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.proof_id_seq OWNER TO cloud;

--
-- Name: proof_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.proof_id_seq OWNED BY public.proof.id;


--
-- Name: rollback; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.rollback (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    from_block_height bigint,
    from_pool_tx_id bigint,
    from_tx_hash text,
    pool_tx_ids text,
    block_heights text,
    account_indexes text,
    nft_indexes text
);


ALTER TABLE public.rollback OWNER TO cloud;

--
-- Name: rollback_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.rollback_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.rollback_id_seq OWNER TO cloud;

--
-- Name: rollback_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.rollback_id_seq OWNED BY public.rollback.id;


--
-- Name: sys_config; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.sys_config (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    value text,
    value_type text,
    comment text
);


ALTER TABLE public.sys_config OWNER TO cloud;

--
-- Name: sys_config_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.sys_config_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.sys_config_id_seq OWNER TO cloud;

--
-- Name: sys_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.sys_config_id_seq OWNED BY public.sys_config.id;


--
-- Name: tx; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.tx (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tx_hash text,
    tx_type bigint,
    tx_info text,
    account_index bigint,
    nonce bigint,
    from_account_index bigint,
    to_account_index bigint,
    expired_at bigint,
    gas_fee text,
    gas_fee_asset_id bigint,
    nft_index bigint,
    collection_id bigint,
    asset_id bigint,
    tx_amount text,
    memo text,
    extra_info text,
    native_address text,
    is_create_account boolean,
    tx_index bigint,
    channel_name text,
    block_height bigint,
    block_id bigint,
    tx_status bigint,
    pool_tx_id bigint,
    verify_at timestamp with time zone
);


ALTER TABLE public.tx OWNER TO cloud;

--
-- Name: tx_detail; Type: TABLE; Schema: public; Owner: cloud
--

CREATE TABLE public.tx_detail (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    pool_tx_id bigint,
    asset_id bigint,
    asset_type bigint,
    account_index bigint,
    l1_address text,
    balance text,
    balance_delta text,
    "order" bigint,
    account_order bigint,
    nonce bigint,
    collection_nonce bigint,
    is_gas boolean DEFAULT false,
    public_key text,
    block_height bigint
);


ALTER TABLE public.tx_detail OWNER TO cloud;

--
-- Name: tx_detail_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.tx_detail_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tx_detail_id_seq OWNER TO cloud;

--
-- Name: tx_detail_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.tx_detail_id_seq OWNED BY public.tx_detail.id;


--
-- Name: tx_id_seq; Type: SEQUENCE; Schema: public; Owner: cloud
--

CREATE SEQUENCE public.tx_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.tx_id_seq OWNER TO cloud;

--
-- Name: tx_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: cloud
--

ALTER SEQUENCE public.tx_id_seq OWNED BY public.tx.id;


--
-- Name: account id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.account ALTER COLUMN id SET DEFAULT nextval('public.account_id_seq'::regclass);


--
-- Name: account_history id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.account_history ALTER COLUMN id SET DEFAULT nextval('public.account_history_id_seq'::regclass);


--
-- Name: asset id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.asset ALTER COLUMN id SET DEFAULT nextval('public.asset_id_seq'::regclass);


--
-- Name: block id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.block ALTER COLUMN id SET DEFAULT nextval('public.block_id_seq'::regclass);


--
-- Name: block_witness id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.block_witness ALTER COLUMN id SET DEFAULT nextval('public.block_witness_id_seq'::regclass);


--
-- Name: compressed_block id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.compressed_block ALTER COLUMN id SET DEFAULT nextval('public.compressed_block_id_seq'::regclass);


--
-- Name: l1_rollup_tx id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l1_rollup_tx ALTER COLUMN id SET DEFAULT nextval('public.l1_rollup_tx_id_seq'::regclass);


--
-- Name: l1_synced_block id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l1_synced_block ALTER COLUMN id SET DEFAULT nextval('public.l1_synced_block_id_seq'::regclass);


--
-- Name: l2_nft id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l2_nft ALTER COLUMN id SET DEFAULT nextval('public.l2_nft_id_seq'::regclass);


--
-- Name: l2_nft_history id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l2_nft_history ALTER COLUMN id SET DEFAULT nextval('public.l2_nft_history_id_seq'::regclass);


--
-- Name: l2_nft_metadata_history id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l2_nft_metadata_history ALTER COLUMN id SET DEFAULT nextval('public.l2_nft_metadata_history_id_seq'::regclass);


--
-- Name: pool_tx id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.pool_tx ALTER COLUMN id SET DEFAULT nextval('public.pool_tx_id_seq'::regclass);


--
-- Name: priority_request id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.priority_request ALTER COLUMN id SET DEFAULT nextval('public.priority_request_id_seq'::regclass);


--
-- Name: proof id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.proof ALTER COLUMN id SET DEFAULT nextval('public.proof_id_seq'::regclass);


--
-- Name: rollback id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.rollback ALTER COLUMN id SET DEFAULT nextval('public.rollback_id_seq'::regclass);


--
-- Name: sys_config id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.sys_config ALTER COLUMN id SET DEFAULT nextval('public.sys_config_id_seq'::regclass);


--
-- Name: tx id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.tx ALTER COLUMN id SET DEFAULT nextval('public.tx_id_seq'::regclass);


--
-- Name: tx_detail id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.tx_detail ALTER COLUMN id SET DEFAULT nextval('public.tx_detail_id_seq'::regclass);


--
-- Name: account_history account_history_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.account_history
    ADD CONSTRAINT account_history_pkey PRIMARY KEY (id);


--
-- Name: account account_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_pkey PRIMARY KEY (id);


--
-- Name: asset asset_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.asset
    ADD CONSTRAINT asset_pkey PRIMARY KEY (id);


--
-- Name: block block_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.block
    ADD CONSTRAINT block_pkey PRIMARY KEY (id);


--
-- Name: block_witness block_witness_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.block_witness
    ADD CONSTRAINT block_witness_pkey PRIMARY KEY (id);


--
-- Name: compressed_block compressed_block_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.compressed_block
    ADD CONSTRAINT compressed_block_pkey PRIMARY KEY (id);


--
-- Name: l1_rollup_tx l1_rollup_tx_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l1_rollup_tx
    ADD CONSTRAINT l1_rollup_tx_pkey PRIMARY KEY (id);


--
-- Name: l1_synced_block l1_synced_block_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l1_synced_block
    ADD CONSTRAINT l1_synced_block_pkey PRIMARY KEY (id);


--
-- Name: l2_nft_history l2_nft_history_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l2_nft_history
    ADD CONSTRAINT l2_nft_history_pkey PRIMARY KEY (id);


--
-- Name: l2_nft_metadata_history l2_nft_metadata_history_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l2_nft_metadata_history
    ADD CONSTRAINT l2_nft_metadata_history_pkey PRIMARY KEY (id);


--
-- Name: l2_nft l2_nft_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.l2_nft
    ADD CONSTRAINT l2_nft_pkey PRIMARY KEY (id);


--
-- Name: pool_tx pool_tx_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.pool_tx
    ADD CONSTRAINT pool_tx_pkey PRIMARY KEY (id);


--
-- Name: priority_request priority_request_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.priority_request
    ADD CONSTRAINT priority_request_pkey PRIMARY KEY (id);


--
-- Name: proof proof_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.proof
    ADD CONSTRAINT proof_pkey PRIMARY KEY (id);


--
-- Name: rollback rollback_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.rollback
    ADD CONSTRAINT rollback_pkey PRIMARY KEY (id);


--
-- Name: sys_config sys_config_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.sys_config
    ADD CONSTRAINT sys_config_pkey PRIMARY KEY (id);


--
-- Name: tx_detail tx_detail_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.tx_detail
    ADD CONSTRAINT tx_detail_pkey PRIMARY KEY (id);


--
-- Name: tx tx_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.tx
    ADD CONSTRAINT tx_pkey PRIMARY KEY (id);


--
-- Name: idx_account_account_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_account_account_index ON public.account USING btree (account_index);


--
-- Name: idx_account_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_account_deleted_at ON public.account USING btree (deleted_at);


--
-- Name: idx_account_history_account_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_account_history_account_index ON public.account_history USING btree (account_index);


--
-- Name: idx_account_history_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_account_history_deleted_at ON public.account_history USING btree (deleted_at);


--
-- Name: idx_account_history_l1_address; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_account_history_l1_address ON public.account_history USING btree (l1_address);


--
-- Name: idx_account_history_l2_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_account_history_l2_block_height ON public.account_history USING btree (l2_block_height);


--
-- Name: idx_account_history_public_key; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_account_history_public_key ON public.account_history USING btree (public_key);


--
-- Name: idx_account_l1_address; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_account_l1_address ON public.account USING btree (l1_address);


--
-- Name: idx_account_l2_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_account_l2_block_height ON public.account USING btree (l2_block_height);


--
-- Name: idx_account_public_key; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_account_public_key ON public.account USING btree (public_key);


--
-- Name: idx_asset_asset_id; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_asset_asset_id ON public.asset USING btree (asset_id);


--
-- Name: idx_asset_asset_symbol; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_asset_asset_symbol ON public.asset USING btree (asset_symbol);


--
-- Name: idx_asset_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_asset_deleted_at ON public.asset USING btree (deleted_at);


--
-- Name: idx_asset_is_gas_asset; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_asset_is_gas_asset ON public.asset USING btree (is_gas_asset);


--
-- Name: idx_asset_l1_address; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_asset_l1_address ON public.asset USING btree (l1_address);


--
-- Name: idx_block_block_commitment; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_block_block_commitment ON public.block USING btree (block_commitment);


--
-- Name: idx_block_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_block_block_height ON public.block USING btree (block_height);


--
-- Name: idx_block_block_status; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_block_block_status ON public.block USING btree (block_status);


--
-- Name: idx_block_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_block_deleted_at ON public.block USING btree (deleted_at);


--
-- Name: idx_block_witness_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_block_witness_deleted_at ON public.block_witness USING btree (deleted_at);


--
-- Name: idx_block_witness_status; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_block_witness_status ON public.block_witness USING btree (status);


--
-- Name: idx_compressed_block_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_compressed_block_block_height ON public.compressed_block USING btree (block_height);


--
-- Name: idx_compressed_block_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_compressed_block_deleted_at ON public.compressed_block USING btree (deleted_at);


--
-- Name: idx_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_height ON public.block_witness USING btree (height);


--
-- Name: idx_l1_nonce; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l1_nonce ON public.l1_rollup_tx USING btree (l1_nonce);


--
-- Name: idx_l1_rollup_tx_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l1_rollup_tx_deleted_at ON public.l1_rollup_tx USING btree (deleted_at);


--
-- Name: idx_l1_rollup_tx_l1_tx_hash; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l1_rollup_tx_l1_tx_hash ON public.l1_rollup_tx USING btree (l1_tx_hash);


--
-- Name: idx_l1_synced_block_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l1_synced_block_deleted_at ON public.l1_synced_block USING btree (deleted_at);


--
-- Name: idx_l1_synced_block_l1_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l1_synced_block_l1_block_height ON public.l1_synced_block USING btree (l1_block_height);


--
-- Name: idx_l1_synced_block_type; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l1_synced_block_type ON public.l1_synced_block USING btree (type);


--
-- Name: idx_l2_nft_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l2_nft_deleted_at ON public.l2_nft USING btree (deleted_at);


--
-- Name: idx_l2_nft_history_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l2_nft_history_deleted_at ON public.l2_nft_history USING btree (deleted_at);


--
-- Name: idx_l2_nft_metadata_history_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l2_nft_metadata_history_deleted_at ON public.l2_nft_metadata_history USING btree (deleted_at);


--
-- Name: idx_l2_nft_metadata_history_nft_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l2_nft_metadata_history_nft_index ON public.l2_nft_metadata_history USING btree (nft_index);


--
-- Name: idx_l2_nft_metadata_history_status; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l2_nft_metadata_history_status ON public.l2_nft_metadata_history USING btree (status);


--
-- Name: idx_l2_nft_metadata_history_tx_hash; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_l2_nft_metadata_history_tx_hash ON public.l2_nft_metadata_history USING btree (tx_hash);


--
-- Name: idx_l2_nft_nft_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_l2_nft_nft_index ON public.l2_nft USING btree (nft_index);


--
-- Name: idx_nft_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_nft_index ON public.l2_nft USING btree (l2_block_height);


--
-- Name: idx_number; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_number ON public.proof USING btree (block_number);


--
-- Name: idx_owner_account_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_owner_account_index ON public.l2_nft USING btree (owner_account_index, nft_content_hash);


--
-- Name: idx_pool_tx_account_index_nonce; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_pool_tx_account_index_nonce ON public.pool_tx USING btree (account_index, nonce);


--
-- Name: idx_pool_tx_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_pool_tx_block_height ON public.pool_tx USING btree (block_height);


--
-- Name: idx_pool_tx_block_id; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_pool_tx_block_id ON public.pool_tx USING btree (block_id);


--
-- Name: idx_pool_tx_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_pool_tx_deleted_at ON public.pool_tx USING btree (deleted_at);


--
-- Name: idx_pool_tx_from_account_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_pool_tx_from_account_index ON public.pool_tx USING btree (from_account_index);


--
-- Name: idx_pool_tx_to_account_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_pool_tx_to_account_index ON public.pool_tx USING btree (to_account_index);


--
-- Name: idx_pool_tx_tx_hash; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_pool_tx_tx_hash ON public.pool_tx USING btree (tx_hash);


--
-- Name: idx_pool_tx_tx_status; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_pool_tx_tx_status ON public.pool_tx USING btree (tx_status);


--
-- Name: idx_priority_request_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_priority_request_deleted_at ON public.priority_request USING btree (deleted_at);


--
-- Name: idx_priority_request_l2_tx_hash; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_priority_request_l2_tx_hash ON public.priority_request USING btree (l2_tx_hash);


--
-- Name: idx_proof_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_proof_deleted_at ON public.proof USING btree (deleted_at);


--
-- Name: idx_rollback_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_rollback_deleted_at ON public.rollback USING btree (deleted_at);


--
-- Name: idx_rollback_from_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_rollback_from_block_height ON public.rollback USING btree (from_block_height);


--
-- Name: idx_rollback_from_pool_tx_id; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_rollback_from_pool_tx_id ON public.rollback USING btree (from_pool_tx_id);


--
-- Name: idx_sys_config_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_sys_config_deleted_at ON public.sys_config USING btree (deleted_at);


--
-- Name: idx_sys_config_name; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_sys_config_name ON public.sys_config USING btree (name);


--
-- Name: idx_tx_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_block_height ON public.tx USING btree (block_height);


--
-- Name: idx_tx_block_id; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_block_id ON public.tx USING btree (block_id);


--
-- Name: idx_tx_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_deleted_at ON public.tx USING btree (deleted_at);


--
-- Name: idx_tx_detail_account_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_detail_account_index ON public.tx_detail USING btree (account_index);


--
-- Name: idx_tx_detail_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_detail_block_height ON public.tx_detail USING btree (block_height);


--
-- Name: idx_tx_detail_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_detail_deleted_at ON public.tx_detail USING btree (deleted_at);


--
-- Name: idx_tx_detail_pool_tx_id; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_detail_pool_tx_id ON public.tx_detail USING btree (pool_tx_id);


--
-- Name: idx_tx_from_account_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_from_account_index ON public.tx USING btree (from_account_index);


--
-- Name: idx_tx_pool_tx_id; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_tx_pool_tx_id ON public.tx USING btree (pool_tx_id);


--
-- Name: idx_tx_status; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_status ON public.l1_rollup_tx USING btree (tx_status, tx_type);


--
-- Name: idx_tx_to_account_index; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_to_account_index ON public.tx USING btree (to_account_index);


--
-- Name: idx_tx_tx_hash; Type: INDEX; Schema: public; Owner: cloud
--

CREATE UNIQUE INDEX idx_tx_tx_hash ON public.tx USING btree (tx_hash);


--
-- Name: idx_tx_tx_status; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_tx_tx_status ON public.tx USING btree (tx_status);


--
-- Name: l2_block_height; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX l2_block_height ON public.l1_rollup_tx USING btree (l2_block_height);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: cloud
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

