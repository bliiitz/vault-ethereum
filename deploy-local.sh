#!/bin/bash
export VAULT_ADDR="http://127.0.0.1:8200"

vault secrets disable vault-ethereum

SHA256=$(cat ./local/SHA256)
echo $SHA256
vault write sys/plugins/catalog/secret/vault-ethereum sha_256=$SHA256 command="vault-ethereum"
vault secrets enable -path=vault-ethereum -plugin-name=vault-ethereum plugin

vault write -format=json vault-ethereum/accounts/test mnemonic='test test test test test test test test test test test junk' index=0
vault write -force -format=json vault-ethereum/accounts/aleph