# Vault Ethereum Wallet Plugin

The Vault Ethereum Wallet Plugin is a HashiCorp Vault plugin that provides secure management and operations for a software wallet on the Ethereum blockchain. It leverages HashiCorp Vault's robust security features to protect private keys and provide cryptographic signing capabilities.

## Features

- Generation and management of Ethereum accounts
- Importing existing Ethereum accounts using mnemonic phrases and derivation paths
- Secure storage of private keys using HashiCorp Vault's key management capabilities
- Cryptographic signing of Ethereum transactions
- Support for multiple Ethereum chains

## Installation

1. Clone the repository:

   ```shell
   git clone https://github.com/bliiitz/vault-ethereum.git
   ```

2. Build the plugin binary:

   ```shell
   ./build.sh
   ```

3. Start dev vault
   ```shell
   ./start-dev-vault.sh
   ```

4. Deploy plugin:
    ```shell
   ./deploy-local.sh
   ```
   

## Usage

Once the Vault Ethereum Cold Wallet Plugin is installed and enabled, you can interact with it using the Vault CLI or the Vault API. The following are some example operations:

- **Create a new Ethereum account:**

  ```shell
  vault write vault-ethereum/accounts/my-wallet
  ```

- **Import an existing Ethereum account:**

  ```shell
  vault write vault-ethereum/accounts/my-wallet \
    mnemonic="..." \
    derivation_path_index=0
  ```

- **Sign a message:**

  ```shell
  vault write vault-ethereum/accounts/my-wallet/sign message="Hello, Ethereum!"
  ```

- **Sign and send a transaction:**

  ```shell
  vault write vault-ethereum/accounts/my-wallet/sign-tx \
    to="0x123..." \
    value="1000000000000000000" \
    data="..." \
    chain="1" \
    gas_price="20000000000" \
    gas_limit="21000"
  ```

For more detailed information on available operations and usage examples, please refer to the [Vault Ethereum Cold Wallet Plugin Documentation](https://your-docs-url.com).

## Security

The Vault Ethereum Cold Wallet Plugin ensures the security of private keys by leveraging HashiCorp Vault's robust security features, including:

- Encryption of private keys at rest
- Protection of private keys with access control policies and authentication mechanisms
- Auditing and logging of all operations performed on the plugin
- Integration with HashiCorp Vault's High Availability and Disaster Recovery capabilities

Please refer to the [Vault Documentation](https://www.vaultproject.io/docs) for more information on securing and configuring HashiCorp Vault.

## Contributing

Contributions to the Vault Ethereum Cold Wallet Plugin are welcome! If you find any issues or have suggestions for new features, please open an issue or submit a pull request. Make sure to follow the [Contributing Guidelines](CONTRIBUTING.md) when making contributions.

## License

The Vault Ethereum Cold Wallet Plugin is licensed under the [MIT License](LICENSE). Please see the [LICENSE](LICENSE) file for more information.