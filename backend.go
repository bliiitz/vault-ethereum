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
	"sync"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

type vaultEthereumBackend struct {
	*framework.Backend
	lock sync.RWMutex
}

// Factory returns the backend
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

// FactoryType returns the factory
func FactoryType(backendType logical.BackendType) logical.Factory {
	return func(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
		b := backend()
		b.BackendType = backendType
		if err := b.Setup(ctx, conf); err != nil {
			return nil, err
		}
		return b, nil
	}
}

func backend() *vaultEthereumBackend {
	var b vaultEthereumBackend
	b.Backend = &framework.Backend{
		Help: "",
		Paths: framework.PathAppend(
			accountPaths(&b),
		),
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				"accounts/",
			},
		},
		Secrets:     []*framework.Secret{},
		BackendType: logical.TypeLogical,
	}
	return &b
}

// QualifiedPath prepends the token symbol to the path
func QualifiedPath(subpath string) string {
	return subpath
}

// SealWrappedPaths returns the paths that are seal wrapped
func SealWrappedPaths(b *vaultEthereumBackend) []string {
	return []string{
		QualifiedPath("accounts/"),
	}
}
