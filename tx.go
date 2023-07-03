package main

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/bliiitz/vault-ethereum/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/vault/sdk/framework"
)

// TransactionParams are typical parameters for a transaction
type TransactionParams struct {
	Nonce    uint64          `json:"nonce"`
	To       *common.Address `json:"address"`
	Value    *big.Int        `json:"value"`
	GasPrice *big.Int        `json:"gas_price"`
	GasLimit uint64          `json:"gas_limit"`
}

func getEIP1559TransactionData(data *framework.FieldData) (*types.Transaction, error) {
	var err error
	var to common.Address
	var nonce uint64
	var value *big.Int
	var gasLimit uint64
	var chainId int64
	var feeCap *big.Int
	var tip *big.Int

	dataOrFile := data.Get("data").(string)

	var txDataToSign []byte = []byte("")
	if len(dataOrFile) > 0 {
		txDataToSign, err = util.Decode([]byte(dataOrFile))
		if err != nil {
			return nil, err
		}
	}

	_, ok := data.GetOk("value")
	if ok {
		value = util.ValidNumber(data.Get("value").(string))
		if value == nil {
			return nil, errors.New("invalid value")
		}
	} else {
		value = util.ValidNumber("0")
	}

	_, ok = data.GetOk("nonce")
	if ok {
		uintNonce := data.Get("nonce").(int64)
		nonce = uint64(uintNonce)
	} else {
		return nil, errors.New("Nonce not specified")
	}

	_, ok = data.GetOk("gas_limit")
	if ok {
		uintGasLimit := data.Get("gas_limit").(string)
		gasLimit, _ = strconv.ParseUint(uintGasLimit, 10, 64)
	} else {
		return nil, errors.New("Gas limit not specified")
	}
	_, ok = data.GetOk("max_priority_fee_per_gas")
	if ok {
		uintTip := data.Get("max_priority_fee_per_gas").(string)
		tip, _ = new(big.Int).SetString(uintTip, 10)
	} else {
		return nil, errors.New("Gas limit not specified")
	}
	_, ok = data.GetOk("max_fee_per_gas")
	if ok {
		uintFeeCap := data.Get("max_fee_per_gas").(string)
		feeCap, _ = new(big.Int).SetString(uintFeeCap, 10)
	} else {
		return nil, errors.New("Gas limit not specified")
	}

	_, ok = data.GetOk("to")
	if ok {
		to = common.HexToAddress(data.Get("to").(string))
	} else {
		return nil, errors.New("To address not specified")
	}

	bigChainID := new(big.Int).SetInt64(chainId)
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   bigChainID,
		Nonce:     nonce,
		GasFeeCap: feeCap,
		GasTipCap: tip,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      txDataToSign,
	})

	return tx, nil
}

func getTransactionData(data *framework.FieldData) (*types.Transaction, error) {
	var err error
	var to common.Address
	var nonce uint64
	var value *big.Int
	var gasLimit uint64
	var chainId int64
	var gasPrice *big.Int
	dataOrFile := data.Get("data").(string)

	var txDataToSign []byte = []byte("")
	if len(dataOrFile) > 0 {
		txDataToSign, err = util.Decode([]byte(dataOrFile))
		if err != nil {
			return nil, err
		}
	}

	_, ok := data.GetOk("chain_id")
	if ok {
		chainId = data.Get("chain_id").(int64)
		if chainId == 0 {
			return nil, errors.New("invalid chain id")
		}
	} else {
		return nil, errors.New("Chain ID not specified")
	}

	_, ok = data.GetOk("value")
	if ok {
		value = util.ValidNumber(data.Get("value").(string))
		if value == nil {
			return nil, errors.New("invalid value")
		}
	} else {
		value = util.ValidNumber("0")
	}

	_, ok = data.GetOk("nonce")
	if ok {
		uintNonce := data.Get("nonce").(int64)
		nonce = uint64(uintNonce)
	} else {
		return nil, errors.New("Nonce not specified")
	}

	_, ok = data.GetOk("gas_limit")
	if ok {
		uintGasLimit := data.Get("gas_limit").(string)
		gasLimit, _ = strconv.ParseUint(uintGasLimit, 10, 64)
	} else {
		return nil, errors.New("Gas limit not specified")
	}

	_, ok = data.GetOk("gas_price")
	if ok {
		uintGasPrice := data.Get("gas_price").(string)
		gasPrice, _ = new(big.Int).SetString(uintGasPrice, 10)
	} else {
		return nil, errors.New("Gas price not specified")
	}

	_, ok = data.GetOk("to")
	if ok {
		to = common.HexToAddress(data.Get("to").(string))
	} else {
		return nil, errors.New("To address not specified")
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		To:       &to,
		Value:    value,
		Data:     txDataToSign,
	})

	return tx, nil
}
