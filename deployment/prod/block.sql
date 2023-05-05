INSERT INTO public.block
(created_at, updated_at, deleted_at, block_size, block_commitment, block_height, state_root, priority_operations, pending_on_chain_operations_hash, pending_on_chain_operations_pub_data, committed_tx_hash, committed_at, verified_tx_hash, verified_at, block_status, account_indexes, nft_indexes)
VALUES(now(), now(), NULL, 0, '0000000000000000000000000000000000000000000000000000000000000000', 0, '1bb54bd4586b34192cd80ca2b19d3579b68509c2a9302405fa8758ba905765c4', 0, 'c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470', '', '', 0, '', 0, 5, '', '');
