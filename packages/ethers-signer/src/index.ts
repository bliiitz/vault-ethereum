import { getAddress } from "@ethersproject/address";
import { Provider, TransactionRequest } from "@ethersproject/abstract-provider";
import { Signer, TypedDataDomain, TypedDataField, TypedDataSigner } from "@ethersproject/abstract-signer";
import { Bytes, SignatureLike } from "@ethersproject/bytes";
import { hashMessage, _TypedDataEncoder } from "@ethersproject/hash";
import { defineReadOnly, resolveProperties } from "@ethersproject/properties";
import { recoverAddress } from "@ethersproject/transactions";
import { Logger } from "@ethersproject/logger";
import { VaultEthereumAccountConfig, VaultEthereumAuthResponse, VaultEthereumConfig, VaultEthereumKubernetesAuth, VaultEthereumTokenAuth } from "./interfaces";
import * as fs from 'fs'


const logger = new Logger("1.0.0")

export { VaultEthereumAuthResponse, VaultEthereumConfig, VaultEthereumKubernetesAuth, VaultEthereumTokenAuth }

export class VaultEthereumSigner extends Signer implements TypedDataSigner {
    readonly address: string;
    public provider: Provider;

    readonly account: string
    readonly config: VaultEthereumConfig
    readonly authMethod: string
    readonly auth: VaultEthereumTokenAuth | VaultEthereumKubernetesAuth

    private authStatus: VaultEthereumAuthResponse;
    private vaultToken: string

    constructor(account: string, config: VaultEthereumConfig, provider?: Provider) {
        super();

        /* istanbul ignore if */
        if (provider && !Provider.isProvider(provider)) {
            logger.throwArgumentError("invalid provider", "provider", provider);
        }

        if(!account)
            throw Error("No walletId specified")

        const configuration: VaultEthereumConfig = {
            endpoint: config.endpoint || "http://localhost:8200",
            pluginPath: config.pluginPath || "vault-ethereum"
        }
        
        defineReadOnly(this, "account", account);
        defineReadOnly(this, "config", configuration);
        this.provider = provider
    }

    async authenticate(method: string, auth: VaultEthereumTokenAuth | VaultEthereumKubernetesAuth) {
        
        this.vaultToken = await VaultEthereumSigner.getTokenFromAuthConfig(method, this.config, auth)
        let accountRequest = await fetch(`${this.config.endpoint}/v1/${this.config.pluginPath}/accounts/${this.account}`, {
            headers: {
                Authorization: `Bearer ${this.vaultToken}`
            }
        })

        let account: {data: {address: string}} = await accountRequest.json()
        defineReadOnly(this, "authMethod", method);
        defineReadOnly(this, "auth", auth);
        defineReadOnly(this, "address", account.data.address);
    }

    static async getTokenFromAuthConfig(method: string, config: VaultEthereumConfig, auth: VaultEthereumTokenAuth | VaultEthereumKubernetesAuth): Promise<string> {
        let vaultToken: string
        let authConfig: VaultEthereumTokenAuth | VaultEthereumKubernetesAuth
        switch(method) {
            case "token":
                authConfig = <VaultEthereumTokenAuth>auth
                vaultToken = authConfig.token
                break
            case "kubernetes":
                authConfig = <VaultEthereumKubernetesAuth>auth
                let jwt= fs.readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token').toString()
                let role = authConfig.role

                let loginRequest = await fetch(`${config.endpoint}/v1/auth/${authConfig.kubeAuthPluginPath}/login`, {
                    method: "POST",
                    body: JSON.stringify({ jwt, role })
                })

                let response = await loginRequest.json()
                let authStatus = response.auth
                vaultToken = authStatus.client_token
                break

            default:
                throw Error("Vault authentication method does't exists")
        }

        if(!vaultToken)
            throw Error("Vault authentication method failed")

        return vaultToken
    }

    static async listVaultsAccounts(
        authMethod: string,
        config: VaultEthereumConfig, 
        auth: VaultEthereumTokenAuth | VaultEthereumKubernetesAuth, 
        accountsPath: string = ""
    ): Promise<string[]> {
        let token = await VaultEthereumSigner.getTokenFromAuthConfig(authMethod, config, auth)
        let listAccounts = await fetch(`${config.endpoint}/v1/${config.pluginPath}/accounts/${accountsPath}`, {
            method: "LIST",
            headers: {
                Authorization: `Bearer ${token}`
            }
        })

        let result: { data: {keys: string[]} } = await listAccounts.json()
        return result.data.keys;
    }
    
    async getAddress(): Promise<string> {
        if(!this.address)
            throw Error("Vault authentication not authenticated")

        return this.address;
    }

    async connectAccount(account: string): Promise<VaultEthereumSigner> {
        let signer = new VaultEthereumSigner(account, this.config, this.provider)
        await signer.authenticate(this.authMethod, this.auth)
        return signer
    }

    connect(provider: Provider): VaultEthereumSigner {
        this.provider = provider
        return this
    }

    async signTransaction(transaction: TransactionRequest): Promise<string> {
        return resolveProperties(transaction).then(async (tx) => {
            if (tx.from != null) {
                if (tx.from.toLowerCase() !== this.address.toLowerCase()) {
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
            
            let signTxRequest = await fetch(`${this.config.endpoint}/v1/${this.config.pluginPath}/accounts/${this.account}/${txEndpoint}`, {
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
        let signTxRequest = await fetch(`${this.config.endpoint}/v1/${this.config.pluginPath}/accounts/${this.account}/sign`, {
            method: "POST",
            headers: {
                Authorization: `Bearer ${this.vaultToken}`
            },
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