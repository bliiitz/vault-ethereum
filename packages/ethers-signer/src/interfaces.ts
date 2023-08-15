
export interface VaultEthereumConfig {
    endpoint: string
    pluginPath?: string
}

export interface VaultEthereumAccountConfig extends VaultEthereumConfig {
    walletId: string
}

export interface VaultEthereumTokenAuth {
    token: string
}

export interface VaultEthereumKubernetesAuth {
    
    role: string,
    kubeAuthPluginPath?: string
}

export interface VaultEthereumAuthResponse {
    client_token: string
    accessor: string
    policies: string[]
    metadata: any
    lease_duration: number
    renewable: boolean
}
