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
func GetMethods() (QngMethods, error) {
	if len(methods) > 0 {
		return methods, nil
	}
	const QngMethodsJson = `{
	  "methods": [
	    {
	      "name": "qng_getPeerInfo",
	      "call": "get_peer_info",
	      "desc": "Retrieves qng get Peer Info",
	      "params": 1
	    },
	    {
	      "name": "qng_getBlockWeight",
	      "call": "get_block_weight",
	      "desc": "Retrieves qng get Block Weight",
	      "params": 1
	    },
	    {
	      "name": "qng_getBlockByID",
	      "call": "get_block_by_id",
	      "desc": "Retrieves qng get Block By ID",
	      "params": 4
	    },
	    {
	      "name": "qng_getBlockByNum",
	      "call": "get_block_by_num",
	      "desc": "Retrieves qng get Block By Num",
	      "params": 4
	    },
	    {
	      "name": "qng_isBlue",
	      "call": "is_blue",
	      "desc": "Retrieves qng is Blue",
	      "params": 1
	    },
	    {
	      "name": "qng_getCoinbase",
	      "call": "get_coinbase",
	      "desc": "Retrieves qng get Coinbase",
	      "params": 2
	    },
	    {
	      "name": "qng_getFees",
	      "call": "get_fees",
	      "desc": "Retrieves qng get Fees",
	      "params": 1
	    },
	    {
	      "name": "qng_getMempool",
	      "call": "get_mempool",
	      "desc": "Retrieves qng get Mempool",
	      "params": 2
	    },
	    {
	      "name": "qng_estimateFee",
	      "call": "estimate_fee",
	      "desc": "Retrieves qng estimate Fee",
	      "params": 1
	    },
	    {
	      "name": "qng_getBlockTemplate",
	      "call": "get_block_template",
	      "desc": "Retrieves qng get Block Template",
	      "params": 2
	    },
	    {
	      "name": "qng_getRawTransaction",
	      "call": "get_raw_transaction",
	      "desc": "Retrieves qng get Raw Transaction",
	      "params": 2
	    },
	    {
	      "name": "qng_getUtxo",
	      "call": "get_utxo",
	      "desc": "Retrieves qng get Utxo",
	      "params": 3
	    },
	    {
	      "name": "qng_getRawTransactions",
	      "call": "get_raw_transactions",
	      "desc": "Retrieves qng get Raw Transactions",
	      "params": 7
	    },
	    {
	      "name": "qng_getRawTransactionByHash",
	      "call": "get_raw_transaction_by_hash",
	      "desc": "Retrieves qng get Raw Transaction By Hash",
	      "params": 2
	    },
	    {
	      "name": "qng_getNodeInfo",
	      "call": "get_node_info",
	      "desc": "Retrieves qng get Node Info",
	      "params": 0
	    },
	    {
	      "name": "qng_getRpcInfo",
	      "call": "get_rpc_info",
	      "desc": "Retrieves qng get Rpc Info",
	      "params": 0
	    },
	    {
	      "name": "qng_getTimeInfo",
	      "call": "get_time_info",
	      "desc": "Retrieves qng get Time Info",
	      "params": 0
	    },
	    {
	      "name": "qng_getNetworkInfo",
	      "call": "get_network_info",
	      "desc": "Retrieves qng get Network Info",
	      "params": 0
	    },
	    {
	      "name": "qng_getSubsidy",
	      "call": "get_subsidy",
	      "desc": "Retrieves qng get Subsidy",
	      "params": 0
	    },
	    {
	      "name": "qng_banlist",
	      "call": "banlist",
	      "desc": "Retrieves qng ban list",
	      "params": 0
	    },
	    {
	      "name": "qng_getBestBlockHash",
	      "call": "get_best_block_hash",
	      "desc": "Retrieves qng get Best Block Hash",
	      "params": 0
	    },
	    {
	      "name": "qng_getBlockTotal",
	      "call": "get_block_total",
	      "desc": "Retrieves qng get Block Total",
	      "params": 0
	    },
	    {
	      "name": "qng_getMainChainHeight",
	      "call": "get_main_chain_height",
	      "desc": "Retrieves qng get Main Chain Height",
	      "params": 0
	    },
	    {
	      "name": "qng_getOrphansTotal",
	      "call": "get_orphans_total",
	      "desc": "Retrieves qng get Orphans Total",
	      "params": 0
	    },
	    {
	      "name": "qng_isCurrent",
	      "call": "is_current",
	      "desc": "Retrieves qng is Current",
	      "params": 0
	    },
	    {
	      "name": "qng_tips",
	      "call": "tips",
	      "desc": "Retrieves qng tips",
	      "params": 0
	    },
	    {
	      "name": "qng_getTokenInfo",
	      "call": "get_token_info",
	      "desc": "Retrieves qng get Token Info",
	      "params": 0
	    },
	    {
	      "name": "qng_getMempoolCount",
	      "call": "get_mempool_count",
	      "desc": "Retrieves qng get Mempool Count",
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
