
export interface VaultEthereumConfig {
    endpoint: string
    walletId: string
    pluginPath?: string
    kubePath?: string
}

export interface VaultEthereumTokenAuth {
    token: string
}

export interface VaultEthereumKubernetesAuth {
    role: string, 
}

export interface VaultEthereumAuthResponse {
    client_token: string
    accessor: string
    policies: string[]
    metadata: any
    lease_duration: number
    renewable: boolean
}
