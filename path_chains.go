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
	"context"
	"fmt"

	"github.com/immutability-io/vault-ethereum/util"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/cidrutil"
	"github.com/hashicorp/vault/sdk/logical"
)

// ChainJSON contains the configuration for each mount
type ChainJSON struct {
	BoundCIDRList []string `json:"bound_cidr_list_list" structs:"bound_cidr_list" mapstructure:"bound_cidr_list"`
	Inclusions    []string `json:"inclusions"`
	Exclusions    []string `json:"exclusions"`
	RPC           string   `json:"rpc_url"`
	ChainID       string   `json:"chain_id"`
}

// ValidAddress returns an error if the address is not included or if it is excluded
func (config *ChainJSON) ValidAddress(toAddress *common.Address) error {
	if util.Contains(config.Exclusions, toAddress.Hex()) {
		return fmt.Errorf("%s is excludeded by this mount", toAddress.Hex())
	}

	if len(config.Inclusions) > 0 && !util.Contains(config.Inclusions, toAddress.Hex()) {
		return fmt.Errorf("%s is not in the set of inclusions of this mount", toAddress.Hex())
	}
	return nil
}

func chainPaths(b *PluginBackend) []*framework.Path {
	return []*framework.Path{
		{
			Pattern: QualifiedPath("chains/?"),
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ListOperation: b.pathChainsList,
			},
			HelpSynopsis: "List all the RPC configs at a path",
			HelpDescription: `
			All the RPC configs will be listed.
			`,
		},
		{
			Pattern: QualifiedPath("chains/" + framework.GenericNameRegex("name")),
			ExistenceCheck: pathExistenceCheck,
			Callbacks: map[logical.Operation]framework.OperationFunc{
				logical.ReadOperation:   b.pathChainsRead,
				logical.CreateOperation: b.pathChainsCreate,
				logical.UpdateOperation: b.pathChainsUpdate,
				logical.DeleteOperation: b.pathChainsDelete,
			},
			HelpSynopsis: "Configure chains and rpc for the Vault Ethereum plugin.",
			HelpDescription: `
			Configure the Vault Ethereum plugin.
			`,
			Fields: map[string]*framework.FieldSchema{
				"name": {Type: framework.TypeString},
				"chain_id": {
					Type:        framework.TypeString,
					Default:     "8545",
					Description: "The chainId of the Ethereum network",
				},
				"inclusions": {
					Type:        framework.TypeCommaStringSlice,
					Description: "Only these accounts may be transaction with",
				},
				"exclusions": {
					Type:        framework.TypeCommaStringSlice,
					Description: "These accounts can never be transacted with",
				},
				"bound_cidr_list": {
					Type: framework.TypeCommaStringSlice,
					Description: `Comma separated string or list of CIDR blocks.
If set, specifies the blocks of IPs which can perform the login operation;
if unset, there are no IP restrictions.`,
				},
			},
		},
	}
}

func (config *ChainJSON) getRPCURL() string {
	return config.RPC
}

func (b *PluginBackend) pathChainsList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vals, err := req.Storage.List(ctx, QualifiedPath("chains/"))
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(vals), nil
}

func (b *PluginBackend) pathChainsCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	name := data.Get("name").(string)

	rpcURL := data.Get("rpc_url").(string)
	chainId := data.Get("chain_id").(string)
	var boundCIDRList []string
	if boundCIDRListRaw, ok := data.GetOk("bound_cidr_list"); ok {
		boundCIDRList = boundCIDRListRaw.([]string)
	}
	var inclusions []string
	if inclusionsRaw, ok := data.GetOk("inclusions"); ok {
		inclusions = inclusionsRaw.([]string)
	}
	var exclusions []string
	if exclusionsRaw, ok := data.GetOk("exclusions"); ok {
		exclusions = exclusionsRaw.([]string)
	}

	chainBundle := &ChainJSON{
		BoundCIDRList: boundCIDRList,
		Inclusions:    inclusions,
		Exclusions:    exclusions,
		RPC:           rpcURL,
		ChainID:       chainId,
	}
	
	err := writeChain(ctx, req, name, chainBundle)
	if err != nil {
		return nil, err
	}

	// Return the secret
	return &logical.Response{
		Data: map[string]interface{}{
			"bound_cidr_list": chainBundle.BoundCIDRList,
			"inclusions":      chainBundle.Inclusions,
			"exclusions":      chainBundle.Exclusions,
			"rpc_url":         chainBundle.RPC,
			"chain_id":         chainBundle.ChainID,
		},
	}, nil
}

func (b *PluginBackend) pathChainsUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	name := data.Get("name").(string)

	chainJSON, err := readChain(ctx, req, name)
	if err != nil {
		return nil, err
	}

	rpcURL := data.Get("rpc_url").(string)
	chainId := data.Get("chain_id").(string)
	var boundCIDRList []string
	if boundCIDRListRaw, ok := data.GetOk("bound_cidr_list"); ok {
		boundCIDRList = boundCIDRListRaw.([]string)
	}
	var inclusions []string
	if inclusionsRaw, ok := data.GetOk("inclusions"); ok {
		inclusions = inclusionsRaw.([]string)
	}
	var exclusions []string
	if exclusionsRaw, ok := data.GetOk("exclusions"); ok {
		exclusions = exclusionsRaw.([]string)
	}

	chainJSON.BoundCIDRList = boundCIDRList
	chainJSON.Inclusions = inclusions
	chainJSON.Exclusions = exclusions
	chainJSON.RPC = rpcURL
	chainJSON.ChainID = chainId

	err = writeChain(ctx, req, name, chainJSON)
	if err != nil {
		return nil, err
	}

	// Return the secret
	return &logical.Response{
		Data: map[string]interface{}{
			"bound_cidr_list": chainJSON.BoundCIDRList,
			"inclusions":      chainJSON.Inclusions,
			"exclusions":      chainJSON.Exclusions,
			"rpc_url":         chainJSON.RPC,
			"chain_id":        chainJSON.ChainID,
		},
	}, nil
}

func (b *PluginBackend) pathChainsRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	
	name := data.Get("name").(string)

	chainBundle, err := readChain(ctx, req, name)
	if err != nil {
		return nil, err
	}

	if chainBundle == nil {
		return nil, nil
	}

	// Return the secret
	return &logical.Response{
		Data: map[string]interface{}{
			"bound_cidr_list": chainBundle.BoundCIDRList,
			"inclusions":      chainBundle.Inclusions,
			"exclusions":      chainBundle.Exclusions,
			"rpc_url":         chainBundle.RPC,
			"chain_id":         chainBundle.ChainID,
		},
	}, nil
}

func (b *PluginBackend) pathChainsDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {

	name := data.Get("name").(string)

	_, err := readChain(ctx, req, name)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Delete(ctx, req.Path); err != nil {
		return nil, err
	}
	return nil, nil
}

// Config returns the configuration for this PluginBackend.
func readChain(ctx context.Context, req *logical.Request, name string) (*ChainJSON, error) {
	path := QualifiedPath(fmt.Sprintf("chains/%s", name))
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, fmt.Errorf("the plugin has not been configured yet")
	}

	var result ChainJSON
	if entry != nil {
		if err := entry.DecodeJSON(&result); err != nil {
			return nil, fmt.Errorf("error reading configuration: %s", err)
		}
	}

	return &result, nil
}

func writeChain(ctx context.Context, req *logical.Request, name string, chainJSON *ChainJSON) error {
	path := QualifiedPath(fmt.Sprintf("chains/%s", name))

	entry, err := logical.StorageEntryJSON(path, chainJSON)
	if err != nil {
		return err
	}

	err = req.Storage.Put(ctx, entry)
	if err != nil {
		return err
	}
	return nil
}

func (b *PluginBackend) configured_chain(ctx context.Context, req *logical.Request, name string) (*ChainJSON, error) {
	chain, err := readChain(ctx, req, name)
	if err != nil {
		return nil, err
	}
	if validConnection, err := b.validIPConstraints(chain, req); !validConnection {
		return nil, err
	}

	return chain, nil
}

func (b *PluginBackend) validIPConstraints(config *ChainJSON, req *logical.Request) (bool, error) {
	if len(config.BoundCIDRList) != 0 {
		if req.Connection == nil || req.Connection.RemoteAddr == "" {
			return false, fmt.Errorf("failed to get connection information")
		}

		belongs, err := cidrutil.IPBelongsToCIDRBlocksSlice(req.Connection.RemoteAddr, config.BoundCIDRList)
		if err != nil {
			return false, errwrap.Wrapf("failed to verify the CIDR restrictions set on the role: {{err}}", err)
		}
		if !belongs {
			return false, fmt.Errorf("source address %q unauthorized through CIDR restrictions on the role", req.Connection.RemoteAddr)
		}
	}
	return true, nil
}
