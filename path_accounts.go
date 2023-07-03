// Copyright © 2018 Immutability, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	bip44 "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	// DerivationPath is the root in a BIP44 hdwallet
	DerivationPath string = "m/44'/60'/0'/0/%d"
	// Empty is the empty string
	Empty string = ""
)

// AccountJSON is what we store for an Ethereum account
type AccountJSON struct {
	Index    int    `json:"index"`
	Mnemonic string `json:"mnemonic"`
}

func accountPaths(b *vaultEthereumBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: QualifiedPath("accounts/?"),
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathAccountsList,
			},
			HelpSynopsis: "List all the Ethereum accounts at a path",
			HelpDescription: `
			All the Ethereum accounts will be listed.
			`,
		},
		{
			Pattern:      QualifiedPath("accounts/" + framework.GenericNameRegex("name")),
			HelpSynopsis: "Create an Ethereum account using a generated or provided passphrase.",
			HelpDescription: `

Creates (or updates) an Ethereum account: an account controlled by a private key. Also
The generator produces a high-entropy passphrase with the provided length and requirements.

`,
			Fields: map[string]*framework.FieldSchema{
				"name": {Type: framework.TypeString},
				"mnemonic": {
					Type:        framework.TypeString,
					Default:     Empty,
					Description: "The mnemonic to use to create the account. If not provided, one is generated.",
				},
				"index": {
					Type:        framework.TypeInt,
					Description: "The index used in BIP-44.",
					Default:     0,
				},
			},
			ExistenceCheck: pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ReadOperation:   b.pathAccountsRead,
				logical.CreateOperation: b.pathAccountsCreate,
				logical.DeleteOperation: b.pathAccountsDelete,
			},
		},
		{
			Pattern:      QualifiedPath("accounts/" + framework.GenericNameRegex("name") + "/sign-1559-tx"),
			HelpSynopsis: "Sign a transaction.",
			HelpDescription: `

Sign an EIP 1559 transaction.

`,
			Fields: map[string]*framework.FieldSchema{
				"name":    {Type: framework.TypeString},
				"address": {Type: framework.TypeString},
				"chain_id": {
					Type:        framework.TypeInt64,
					Description: "The chain ID of the tx to sign.",
				},
				"to": {
					Type:        framework.TypeString,
					Description: "The address of the wallet to send ETH to.",
				},
				"data": {
					Type:        framework.TypeString,
					Description: "The data to sign.",
				},
				"value": {
					Type:        framework.TypeString,
					Description: "Value of ETH (in wei).",
				},
				"nonce": {
					Type:        framework.TypeInt64,
					Description: "The transaction nonce.",
				},
				"gas_limit": {
					Type:        framework.TypeString,
					Description: "The gas limit for the transaction - defaults to 21000.",
					Default:     "21000",
				},
				"max_priority_fee_per_gas": {
					Type:        framework.TypeString,
					Description: "The gas price for the transaction in wei.",
					Default:     "0",
				},
				"max_fee_per_gas": {
					Type:        framework.TypeString,
					Description: "The gas price for the transaction in wei.",
					Default:     "0",
				},
			},
			ExistenceCheck: pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathSignEIP1559Tx,
				logical.UpdateOperation: b.pathSignEIP1559Tx,
			},
		},
		{
			Pattern:      QualifiedPath("accounts/" + framework.GenericNameRegex("name") + "/sign-tx"),
			HelpSynopsis: "Sign a transaction.",
			HelpDescription: `

Sign a transaction.

`,
			Fields: map[string]*framework.FieldSchema{
				"name":    {Type: framework.TypeString},
				"address": {Type: framework.TypeString},
				"chain_id": {
					Type:        framework.TypeInt64,
					Description: "The chain ID of the tx to sign.",
				},
				"to": {
					Type:        framework.TypeString,
					Description: "The address of the wallet to send ETH to.",
				},
				"data": {
					Type:        framework.TypeString,
					Description: "The data to sign.",
				},
				"value": {
					Type:        framework.TypeString,
					Description: "Value of ETH (in wei).",
				},
				"nonce": {
					Type:        framework.TypeInt64,
					Description: "The transaction nonce.",
				},
				"gas_limit": {
					Type:        framework.TypeString,
					Description: "The gas limit for the transaction - defaults to 21000.",
					Default:     "21000",
				},
				"gas_price": {
					Type:        framework.TypeString,
					Description: "The gas price for the transaction in wei.",
					Default:     "0",
				},
			},
			ExistenceCheck: pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathSignTx,
				logical.UpdateOperation: b.pathSignTx,
			},
		},
		{
			Pattern:      QualifiedPath("accounts/" + framework.GenericNameRegex("name") + "/sign"),
			HelpSynopsis: "Sign a message",
			HelpDescription: `

Sign calculates an ECDSA signature for:
keccack256("\x19Ethereum Signed Message:\n" + len(message) + message).

https://eth.wiki/json-rpc/API#eth_sign

		`,
			Fields: map[string]*framework.FieldSchema{
				"name": {Type: framework.TypeString},
				"message": {
					Type:        framework.TypeString,
					Description: "Message to sign.",
				},
			},
			ExistenceCheck: pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.CreateOperation: b.pathSignMessage,
				logical.UpdateOperation: b.pathSignMessage,
			},
		},
	}
}

func (b *vaultEthereumBackend) pathAccountsList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vals, err := req.Storage.List(ctx, QualifiedPath("accounts/"))
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(vals), nil
}

func readAccount(ctx context.Context, req *logical.Request, name string) (*AccountJSON, error) {
	path := QualifiedPath(fmt.Sprintf("accounts/%s", name))
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	var accountJSON AccountJSON
	err = entry.DecodeJSON(&accountJSON)

	if entry == nil {
		return nil, fmt.Errorf("failed to deserialize account at %s", path)
	}
	return &accountJSON, nil
}

func (b *vaultEthereumBackend) pathAccountsRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	name := data.Get("name").(string)
	accountJSON, err := readAccount(ctx, req, name)
	if err != nil {
		return nil, err
	}

	_, account, err := getWalletAndAccount(*accountJSON)
	if err != nil {
		return nil, err
	}
	if err != nil || accountJSON == nil {
		return nil, fmt.Errorf("Error reading account")
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"address": account.Address.Hex(),
		},
	}, nil
}

func (b *vaultEthereumBackend) pathAccountsDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	name := data.Get("name").(string)
	_, err := readAccount(ctx, req, name)
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Delete(ctx, req.Path); err != nil {
		return nil, err
	}
	return nil, nil
}

func getWalletAndAccount(accountJSON AccountJSON) (*bip44.Wallet, *accounts.Account, error) {
	hdwallet, err := bip44.NewFromMnemonic(accountJSON.Mnemonic)
	if err != nil {
		return nil, nil, err
	}
	derivationPath := fmt.Sprintf(DerivationPath, accountJSON.Index)
	path := bip44.MustParseDerivationPath(derivationPath)
	account, err := hdwallet.Derive(path, true)
	if err != nil {
		return nil, nil, err
	}
	return hdwallet, &account, nil
}

func (b *vaultEthereumBackend) pathAccountsCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	name := data.Get("name").(string)
	index := data.Get("index").(int)
	mnemonic := data.Get("mnemonic").(string)

	if mnemonic == Empty {
		entropy, err := bip39.NewEntropy(128)
		if err != nil {
			return nil, err
		}

		mnemonic, err = bip39.NewMnemonic(entropy)
	}

	accountJSON := &AccountJSON{
		Index:    index,
		Mnemonic: mnemonic,
	}
	_, account, err := getWalletAndAccount(*accountJSON)
	if err != nil {
		return nil, err
	}

	err = b.updateAccount(ctx, req, name, accountJSON)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"address": account.Address.Hex(),
		},
	}, nil
}

func (b *vaultEthereumBackend) updateAccount(ctx context.Context, req *logical.Request, name string, accountJSON *AccountJSON) error {
	path := QualifiedPath(fmt.Sprintf("accounts/%s", name))

	entry, err := logical.StorageEntryJSON(path, accountJSON)
	if err != nil {
		return err
	}

	err = req.Storage.Put(ctx, entry)
	if err != nil {
		return err
	}
	return nil
}

func pathExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, fmt.Errorf("existence check failed: %v", err)
	}

	return out != nil, nil
}

// returns (nonce, toAddress, amount, gasPrice, gasLimit, error)

func (b *vaultEthereumBackend) pathSignEIP1559Tx(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	var err error

	name := data.Get("name").(string)

	accountJSON, err := readAccount(ctx, req, name)
	if err != nil {
		return nil, err
	}

	wallet, account, err := getWalletAndAccount(*accountJSON)
	if err != nil {
		return nil, err
	}

	tx, err := getEIP1559TransactionData(data)
	if err != nil {
		return nil, err
	}

	signedTx, err := wallet.SignTx(*account, tx, tx.ChainId())
	if err != nil {
		return nil, err
	}

	var signedTxBuff bytes.Buffer
	signedTx.EncodeRLP(&signedTxBuff)

	return &logical.Response{
		Data: map[string]interface{}{
			"chainId":           tx.ChainId(),
			"signedTransaction": signedTx,
			"rlpSignature":      hexutil.Encode(signedTxBuff.Bytes()),
		},
	}, nil
}

func (b *vaultEthereumBackend) pathSignTx(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	var err error

	name := data.Get("name").(string)

	accountJSON, err := readAccount(ctx, req, name)
	if err != nil {
		return nil, err
	}

	wallet, account, err := getWalletAndAccount(*accountJSON)
	if err != nil {
		return nil, err
	}

	tx, err := getTransactionData(data)
	if err != nil {
		return nil, err
	}

	chainId := data.Get("chain_id").(int64)
	bigChainID := new(big.Int).SetInt64(chainId)
	signedTx, err := wallet.SignTxEIP155(*account, tx, bigChainID)
	if err != nil {
		return nil, err
	}

	var signedTxBuff bytes.Buffer
	signedTx.EncodeRLP(&signedTxBuff)

	return &logical.Response{
		Data: map[string]interface{}{
			"chainId":           bigChainID,
			"signedTransaction": signedTx,
			"rlpSignature":      hexutil.Encode(signedTxBuff.Bytes()),
		},
	}, nil
}

// LogTx is for debugging
func (b *vaultEthereumBackend) LogTx(tx *types.Transaction) {
	b.Logger().Info(fmt.Sprintf("\nTX DATA: %s\nGAS: %d\nGAS PRICE: %d\nVALUE: %d\nNONCE: %d\nTO: %s\nCHAINID: %d\n", hexutil.Encode(tx.Data()), tx.Gas(), tx.GasPrice(), tx.Value(), tx.Nonce(), tx.To(), tx.ChainId()))
}

func (b *vaultEthereumBackend) pathSignMessage(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	message := data.Get("message").(string)
	name := data.Get("name").(string)

	accountJSON, err := readAccount(ctx, req, name)
	if err != nil {
		return nil, err
	}

	wallet, account, err := getWalletAndAccount(*accountJSON)
	if err != nil {
		return nil, err
	}

	hashedMessage, _ := accounts.TextAndHash([]byte(message))

	signedMessage, err := wallet.SignHash(*account, []byte(hashedMessage))
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"signature":     hexutil.Encode(signedMessage),
			"address":       account.Address,
			"hashedMessage": hexutil.Encode(hashedMessage),
		},
	}, nil
}
