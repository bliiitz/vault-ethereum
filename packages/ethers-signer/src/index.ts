import { getAddress } from "@ethersproject/address";
import { Provider, TransactionRequest } from "@ethersproject/abstract-provider";
import { Signer, TypedDataDomain, TypedDataField, TypedDataSigner } from "@ethersproject/abstract-signer";
import { Bytes, SignatureLike } from "@ethersproject/bytes";
import { hashMessage, _TypedDataEncoder } from "@ethersproject/hash";
import { defineReadOnly, resolveProperties } from "@ethersproject/properties";
import { recoverAddress, serialize, UnsignedTransaction } from "@ethersproject/transactions";
import { Logger } from "@ethersproject/logger";
import { VaultEthereumAuthResponse, VaultEthereumConfig, VaultEthereumKubernetesAuth, VaultEthereumTokenAuth } from "./interfaces";
import * as fs from 'fs'
import { BigNumber, Wallet } from "ethers";

const logger = new Logger("1.0.0")

export class VaultEthereumSigner extends Signer implements TypedDataSigner {
        readonly address: string;
        public provider: Provider;

        readonly config: VaultEthereumConfig
    
        private authStatus: VaultEthereumAuthResponse;
        private vaultToken: string
    
        constructor(config: VaultEthereumConfig, provider?: Provider) {
            super();
    
            /* istanbul ignore if */
            if (provider && !Provider.isProvider(provider)) {
                logger.throwArgumentError("invalid provider", "provider", provider);
            }

            if(!config.walletId)
                throw Error("No walletId specified")

            const vaultConfig: VaultEthereumConfig = {
                endpoint: config.endpoint || "http://localhost:8200",
                pluginPath: config.pluginPath || "vault-ethereum",
                kubePath: config.kubePath || "kubernetes",
                walletId: config.walletId,
            }
            
            defineReadOnly(this, "config", vaultConfig);
            this.provider = provider
        }

        async authenticate(method: string, config: any) {
            let authConfig: VaultEthereumTokenAuth | VaultEthereumKubernetesAuth
            switch(method) {
                case "token":
                    authConfig = <VaultEthereumTokenAuth>config
                    this.vaultToken = authConfig.token
                    break
                case "kubernetes":
                    authConfig = <VaultEthereumKubernetesAuth>config
                    let jwt= fs.readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token').toString()
                    let role = authConfig.role

                    let loginRequest = await fetch(`${this.config.endpoint}/v1/auth/${this.config.kubePath}/login`, {
                        method: "POST",
                        body: JSON.stringify({ jwt, role })
                    })

                    let response = await loginRequest.json()
                    this.authStatus = response.auth
                    this.vaultToken = this.authStatus.client_token
                    break

                default:
                    throw Error("Vault authentication method does't exists")
            }

            if(!this.vaultToken)
                throw Error("Vault authentication method failed")

            let accountRequest = await fetch(`${this.config.endpoint}/v1/${this.config.pluginPath}/accounts/${this.config.walletId}`, {
                headers: {
                    Authorization: `Bearer ${this.vaultToken}`
                }
            })
            let account: {data: {address: string}} = await accountRequest.json()
            defineReadOnly(this, "address", account.data.address);
                  
        }
        
        async getAddress(): Promise<string> {
            if(!this.address)
                throw Error("Vault authentication not authenticated")

            return this.address;
        }
    
        connect(provider: Provider): VaultEthereumSigner {
            this.provider = provider
            return this
        }
    
        async signTransaction(transaction: TransactionRequest): Promise<string> {
            return resolveProperties(transaction).then(async (tx) => {
                if (tx.from != null) {
                    if (getAddress(tx.from) !== this.address) {
                        logger.throwArgumentError("transaction from address mismatch", "transaction.from", transaction.from);
                    }
                    delete tx.from;
                }


                let network = await this.provider.getNetwork()

                let data = <string>transaction.data
                let hexData: string = ""
                if(data && data.length > 2) {
                    hexData = data.slice(2)
                }

                let txEndpoint: string
                let txReq: any
                if(transaction.type == 2) {
                    txEndpoint = "sign-1559-tx"
                    txReq = {
                        chain_id: network.chainId,
                        to: transaction.to.toString(),
                        data: hexData,
                        value: transaction.value.toString(),
                        nonce: transaction.nonce,
                        gas_limit: transaction.gasLimit.toString(),
                        max_priority_fee_per_gas: transaction.maxPriorityFeePerGas.toString(),
                        max_fee_per_gas: transaction.maxFeePerGas.toString(),
                    }
                } else {
                    txEndpoint = "sign-tx"
                    txReq = {
                        chain_id: network.chainId,
                        to: transaction.to.toString(),
                        data: hexData,
                        value: transaction.value.toString(),
                        nonce: transaction.nonce,
                        gas_limit: transaction.gasLimit.toString(),
                        gas_price: transaction.gasPrice.toString() || (await this.provider.getGasPrice()).toString()
                    }
                }
                
                let signTxRequest = await fetch(`${this.config.endpoint}/v1/${this.config.pluginPath}/accounts/${this.config.walletId}/${txEndpoint}`, {
                    method: "POST",
                    headers: {
                        Authorization: `Bearer ${this.vaultToken}`
                    },
                    body: JSON.stringify(txReq)
                })
                let signTxResponse = await signTxRequest.json()
                return signTxResponse.data.rlpSignature;
            });
        }
    
        async signMessage(message: Bytes | string): Promise<string> {
            let signTxRequest = await fetch(`${this.config.endpoint}/v1/${this.config.pluginPath}/accounts/${this.config.walletId}/sign`, {
                method: "POST",
                body: JSON.stringify({
                    message,
                })
            })
            let signTxResponse: {data: {signature: string}} = await signTxRequest.json()
            return signTxResponse.data.signature;
        }
    
        async _signTypedData(domain: TypedDataDomain, types: Record<string, Array<TypedDataField>>, value: Record<string, any>): Promise<string> {
            throw Error("_signTypedData not implemented")
            // Populate any ENS names
            // const populated = await _TypedDataEncoder.resolveNames(domain, types, value, (name: string) => {
            //     if (this.provider == null) {
            //         logger.throwError("cannot resolve ENS names without a provider", Logger.errors.UNSUPPORTED_OPERATION, {
            //             operation: "resolveName",
            //             value: name
            //         });
            //     }
            //     return this.provider.resolveName(name);
            // });
    
            // return joinSignature(this._signingKey().signDigest(_TypedDataEncoder.hash(populated.domain, types, populated.value)));
        }
    
    }
    
    export function verifyMessage(message: Bytes | string, signature: SignatureLike): string {
        return recoverAddress(hashMessage(message), signature);
    }
    
    export function verifyTypedData(domain: TypedDataDomain, types: Record<string, Array<TypedDataField>>, value: Record<string, any>, signature: SignatureLike): string {
        return recoverAddress(_TypedDataEncoder.hash(domain, types, value), signature);
    }