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

DROP DATABASE IF EXISTS "zkbnb";
--
-- Name: zkbnb; Type: DATABASE; Schema: -; Owner: cloud
--

CREATE DATABASE "zkbnb" WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.UTF-8';


ALTER DATABASE "zkbnb" OWNER TO cloud;

\connect -reuse-previous=on "dbname='zkbnb'"

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

SET default_tablespace = '';

SET default_table_access_method = heap;

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
-- Name: sys_config id; Type: DEFAULT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.sys_config ALTER COLUMN id SET DEFAULT nextval('public.sys_config_id_seq'::regclass);


--
-- Data for Name: sys_config; Type: TABLE DATA; Schema: public; Owner: cloud
--

COPY public.sys_config (id, created_at, updated_at, deleted_at, name, value, value_type, comment) FROM stdin;
1	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	SysGasFee	{"0":{"1":10000000000000,"10":12000000000000,"11":20000000000000,"4":10000000000000,"5":20000000000000,"6":10000000000000,"7":10000000000000,"8":12000000000000,"9":18000000000000},"1":{"1":10000000000000,"10":12000000000000,"11":20000000000000,"4":10000000000000,"5":20000000000000,"6":10000000000000,"7":10000000000000,"8":12000000000000,"9":18000000000000}}	string	based on BNB
2	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	ProtocolRate	200	int	protocol rate
3	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	ProtocolAccountIndex	0	int	protocol index
4	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	GasAccountIndex	1	int	gas index
5	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	ZkBNBContract	0xBd012395D9D85499Fc4BF60d7F024d34fD3a88FF	string	ZkBNB contract on BSC
6	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	CommitAddress	0x83a1f1BaBF815056fa56586f752F116B2A14D26b	string	ZkBNB commit on BSC
7	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	VerifyAddress	0xc785309fee44Fa66848135b58BfDdBb74d75b38D	string	ZkBNB verify on BSC
8	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	GovernanceContract	0xB933CD36D937EB2430D4508DbC4470308Bb28813	string	Governance contract on BSC
9	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	BscTestNetworkRpc	https://bsc-testnet.nodereal.io/v1/a1cee760ac744f449416a711f20d99dd	string	BSC network rpc
10	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	LocalTestNetworkRpc	http://127.0.0.1:8545/	string	Local network rpc
11	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	ZnsPriceOracle	0x67611D3E0fbB56C016C2B44d428Bb588B1943e9d	string	Zns Price Oracle
12	2023-04-22 13:10:19.340905+00	2023-04-22 13:10:19.340905+00	\N	DefaultNftFactory	0xDA8c0929ec116C81a85280cAaf73218553848e9D	string	ZkBNB default nft factory contract on BSC
13	2023-04-22 13:13:28.631977+00	2023-04-22 13:13:28.631977+00	\N	Governor	0x35888CD920AFbE82c7D4cDac9896e720E0aa8cb5	string	governor
14	2023-04-22 13:13:28.631977+00	2023-04-22 13:13:28.631977+00	\N	AssetGovernanceContract	0x80A3D3eDA8cCD58DC40b76c6004dc8388fa2475A	string	asset governance contract
15	2023-04-22 13:13:28.631977+00	2023-04-22 13:13:28.631977+00	\N	Validators	{"0x35888CD920AFbE82c7D4cDac9896e720E0aa8cb5":{"Address":"0x35888CD920AFbE82c7D4cDac9896e720E0aa8cb5","IsActive":true},"0x83a1f1BaBF815056fa56586f752F116B2A14D26b":{"Address":"0x83a1f1BaBF815056fa56586f752F116B2A14D26b","IsActive":true},"0xc785309fee44Fa66848135b58BfDdBb74d75b38D":{"Address":"0xc785309fee44Fa66848135b58BfDdBb74d75b38D","IsActive":true}}	map[string]*ValidatorInfo	validator info
\.


--
-- Name: sys_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: cloud
--

SELECT pg_catalog.setval('public.sys_config_id_seq', 15, true);


--
-- Name: sys_config sys_config_pkey; Type: CONSTRAINT; Schema: public; Owner: cloud
--

ALTER TABLE ONLY public.sys_config
    ADD CONSTRAINT sys_config_pkey PRIMARY KEY (id);


--
-- Name: idx_sys_config_deleted_at; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_sys_config_deleted_at ON public.sys_config USING btree (deleted_at);


--
-- Name: idx_sys_config_name; Type: INDEX; Schema: public; Owner: cloud
--

CREATE INDEX idx_sys_config_name ON public.sys_config USING btree (name);


--
-- PostgreSQL database dump complete
--

