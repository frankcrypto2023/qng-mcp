package main

import (
	"encoding/json"
	"fmt"
)

// Method represents a method or property with its details.
type Method struct {
	Name   string `json:"name"`
	Call   string `json:"call"`
	Desc   string `json:"desc"`
	Params int    `json:"params"`
}
type QngMethods []Method

func (m *QngMethods) FindName(call string) (Method, error) {
	for _, v := range *m {
		if v.Call == call {
			return v, nil
		}
	}
	return Method{}, fmt.Errorf("method not found")
}

var methods QngMethods

// GetMethods parses the QngJs JSON and returns a slice of Method structs.
// These methods are organized into categories for better AI model understanding:
// - Block Operations: get_block_by_id, get_block_by_num, get_block_weight, etc.
// - Transaction Operations: get_raw_transaction, get_raw_transactions, get_utxo, etc.
// - Network Operations: get_peer_info, get_network_info, banlist, etc.
// - Node Operations: get_node_info, get_rpc_info, get_time_info, etc.
// - Mining Operations: get_block_template, get_subsidy, etc.
// - Chain Operations: get_best_block_hash, get_block_total, is_current, etc.
func GetMethods() (QngMethods, error) {
	if len(methods) > 0 {
		return methods, nil
	}
	const QngMethodsJson = `{
	  "methods": [
	    {
	      "name": "qng_getPeerInfo",
	      "call": "get_peer_info",
	      "desc": "Retrieves detailed peer connection information from the QNG network. Returns data about connected peers including their addresses, connection status, and network statistics.",
	      "params": 1
	    },
	    {
	      "name": "qng_getBlockWeight",
	      "call": "get_block_weight",
	      "desc": "Retrieves the weight (difficulty) of a specific block in the QNG blockchain. Block weight is used in consensus algorithms to determine the chain with the most accumulated work.",
	      "params": 1
	    },
	    {
	      "name": "qng_getBlockByID",
	      "call": "get_block_by_id",
	      "desc": "Retrieves complete block information using the block's unique identifier (block hash). Returns full block data including header, transactions, and metadata.",
	      "params": 4
	    },
	    {
	      "name": "qng_getBlockByNum",
	      "call": "get_block_by_num",
	      "desc": "Retrieves complete block information using the block number (height). Returns full block data including header, transactions, and metadata for the specified block height.",
	      "params": 4
	    },
	    {
	      "name": "qng_isBlue",
	      "call": "is_blue",
	      "desc": "Checks if a specific block is considered 'blue' in the QNG consensus algorithm. Blue blocks are part of the main chain and have been confirmed by the network.",
	      "params": 1
	    },
	    {
	      "name": "qng_getCoinbase",
	      "call": "get_coinbase",
	      "desc": "Retrieves coinbase transaction information for a specific block. The coinbase transaction is the first transaction in each block and contains the block reward.",
	      "params": 2
	    },
	    {
	      "name": "qng_getFees",
	      "call": "get_fees",
	      "desc": "Retrieves current network fee information including recommended transaction fees, fee rates, and fee estimation data for optimal transaction processing.",
	      "params": 1
	    },
	    {
	      "name": "qng_getMempool",
	      "call": "get_mempool",
	      "desc": "Retrieves information about transactions currently in the memory pool (mempool). Returns pending transactions waiting to be included in the next block.",
	      "params": 2
	    },
	    {
	      "name": "qng_estimateFee",
	      "call": "estimate_fee",
	      "desc": "Estimates the appropriate transaction fee for a given transaction size or priority level. Helps users set optimal fees for timely transaction confirmation.",
	      "params": 1
	    },
	    {
	      "name": "qng_getBlockTemplate",
	      "call": "get_block_template",
	      "desc": "Retrieves a block template for mining operations. Returns the structure and data needed to construct a new block, including transaction selection and header information.",
	      "params": 2
	    },
	    {
	      "name": "qng_getRawTransaction",
	      "call": "get_raw_transaction",
	      "desc": "Retrieves raw transaction data by transaction hash. Returns the complete transaction in its serialized format as it appears on the blockchain.",
	      "params": 2
	    },
	    {
	      "name": "qng_getUtxo",
	      "call": "get_utxo",
	      "desc": "Retrieves Unspent Transaction Output (UTXO) information for a specific address or transaction. UTXOs represent available funds that can be spent in new transactions.",
	      "params": 3
	    },
	    {
	      "name": "qng_getRawTransactions",
	      "call": "get_raw_transactions",
	      "desc": "Retrieves multiple raw transactions based on various filtering criteria. Returns serialized transaction data for multiple transactions matching the specified parameters.",
	      "params": 7
	    },
	    {
	      "name": "qng_getRawTransactionByHash",
	      "call": "get_raw_transaction_by_hash",
	      "desc": "Retrieves raw transaction data using the transaction hash as the lookup key. Returns the complete serialized transaction data for the specified transaction.",
	      "params": 2
	    },
	    {
	      "name": "qng_getNodeInfo",
	      "call": "get_node_info",
	      "desc": "Retrieves comprehensive information about the current QNG node including version, build information, network status, and configuration details.",
	      "params": 0
	    },
	    {
	      "name": "qng_getRpcInfo",
	      "call": "get_rpc_info",
	      "desc": "Retrieves information about the RPC server configuration and status including available methods, connection details, and server statistics.",
	      "params": 0
	    },
	    {
	      "name": "qng_getTimeInfo",
	      "call": "get_time_info",
	      "desc": "Retrieves time-related information from the QNG node including current blockchain time, synchronization status, and time offset data.",
	      "params": 0
	    },
	    {
	      "name": "qng_getNetworkInfo",
	      "call": "get_network_info",
	      "desc": "Retrieves comprehensive network information including peer connections, network topology, bandwidth statistics, and network health metrics.",
	      "params": 0
	    },
	    {
	      "name": "qng_getSubsidy",
	      "call": "get_subsidy",
	      "desc": "Retrieves current block subsidy information including mining rewards, emission rates, and subsidy schedule for the QNG blockchain.",
	      "params": 0
	    },
	    {
	      "name": "qng_banlist",
	      "call": "banlist",
	      "desc": "Retrieves the list of banned or blocked network peers. Returns information about peers that have been temporarily or permanently blocked from connecting to the node.",
	      "params": 0
	    },
	    {
	      "name": "qng_getBestBlockHash",
	      "call": "get_best_block_hash",
	      "desc": "Retrieves the hash of the current best (highest) block in the blockchain. This represents the tip of the main chain and the most recent confirmed block.",
	      "params": 0
	    },
	    {
	      "name": "qng_getBlockTotal",
	      "call": "get_block_total",
	      "desc": "Retrieves the total number of blocks in the blockchain. Returns the current block count (height) representing the total blocks mined since genesis.",
	      "params": 0
	    },
	    {
	      "name": "qng_getMainChainHeight",
	      "call": "get_main_chain_height",
	      "desc": "Retrieves the height of the main blockchain. Returns the number of blocks in the longest valid chain, representing the current blockchain length.",
	      "params": 0
	    },
	    {
	      "name": "qng_getOrphansTotal",
	      "call": "get_orphans_total",
	      "desc": "Retrieves the total number of orphaned blocks in the blockchain. Orphaned blocks are valid blocks that are not part of the main chain due to chain reorganization.",
	      "params": 0
	    },
	    {
	      "name": "qng_isCurrent",
	      "call": "is_current",
	      "desc": "Checks if the node is currently synchronized with the network. Returns whether the local blockchain is up-to-date with the latest blocks from the network.",
	      "params": 0
	    },
	    {
	      "name": "qng_tips",
	      "call": "tips",
	      "desc": "Retrieves information about blockchain tips (multiple potential chain heads). Returns data about competing chain branches and their respective weights.",
	      "params": 0
	    },
	    {
	      "name": "qng_getTokenInfo",
	      "call": "get_token_info",
	      "desc": "Retrieves information about tokens and assets on the QNG blockchain including token metadata, supply information, and token contract details.",
	      "params": 0
	    },
	    {
	      "name": "qng_getMempoolCount",
	      "call": "get_mempool_count",
	      "desc": "Retrieves the current count of transactions in the memory pool. Returns the number of pending transactions waiting to be included in the next block.",
	      "params": 0
	    }
	  ]
	}`

	var data struct {
		Methods []Method `json:"methods"`
	}

	err := json.Unmarshal([]byte(QngMethodsJson), &data)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}
	methods = data.Methods
	return methods, nil
}
