# Vault Ethereum Wallet Signer for ethers.js

The Vault Ethereum Cold Wallet Signer for ethers.js is a signer implementation that integrates with the Vault Ethereum Cold Wallet Plugin. It allows you to securely sign Ethereum transactions using private keys stored in the Vault cold wallet.

## Features

- Securely sign Ethereum transactions using the Vault Ethereum Wallet Plugin
- Seamless integration with ethers.js library
- Supports multiple Ethereum chains

## Prerequisites

Before using the Vault Ethereum Wallet Signer, ensure that you have the following:

- A running instance of HashiCorp Vault with the Ethereum Wallet Plugin installed
- An Ethereum account stored in the Vault wallet

## Installation

```shell
npm install @your-org/vault-ethereum-wallet-signer
```

## Usage

1. Import the necessary packages:

   ```javascript
   const ethers = require('ethers');
   const { VaultColdWalletSigner } = require('@your-org/vault-eth-cold-wallet-signer');
   ```

2. Create a provider instance:

   ```javascript
   const provider = new ethers.providers.JsonRpcProvider('https://your-rpc-url');
   ```

3. Initialize the signer with the Vault Ethereum Cold Wallet Plugin details:

   ```javascript
   const signer = new VaultColdWalletSigner({
     accountId: 'my-wallet',
     vaultUrl: 'https://your-vault-url',
     vaultToken: 'your-vault-token',
   });
   ```

4. Set the provider for the signer:

   ```javascript
   signer.setProvider(provider);
   ```

5. Use the signer to sign transactions:

   ```javascript
   const tx = {
     to: '0x123...',
     value: ethers.utils.parseEther('1.0'),
   };

   const signedTx = await signer.signTransaction(tx);
   ```

For more detailed usage examples and available methods, please refer to the [Vault Ethereum Cold Wallet Signer Documentation](https://your-docs-url.com).

## Security

The Vault Ethereum Wallet Signer ensures the security of private keys by leveraging the Vault Ethereum Wallet Plugin's secure storage capabilities. It protects the private keys with HashiCorp Vault's access control policies and authentication mechanisms.

Please ensure that you have properly configured and secured your HashiCorp Vault instance. Refer to the [Vault Documentation](https://www.vaultproject.io/docs) for more information on securing and configuring HashiCorp Vault.

## Contributing

Contributions to the Vault Ethereum Wallet Signer for ethers.js are welcome! If you find any issues or have suggestions for new features, please open an issue or submit a pull request. Make sure to follow the [Contributing Guidelines](CONTRIBUTING.md) when making contributions.

## License

The Vault Ethereum Wallet Signer for ethers.js is licensed under the [MIT License](LICENSE). Please see the [LICENSE](LICENSE) file for more information.