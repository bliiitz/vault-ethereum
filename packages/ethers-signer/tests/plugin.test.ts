import { BigNumber, ethers, providers } from "ethers";
import { VaultEthereumSigner } from "../src/index"

describe('Test with local vault', function() {

    it('Make a simple ethereum transfer', async function() {
        let provider = new providers.JsonRpcProvider("http://127.0.0.1:8545")
        let signer = new VaultEthereumSigner({
            endpoint: "http://127.0.0.1:8200",
            walletId: "test",
            pluginPath: "vault-ethereum"
        }, provider)

        await signer.authenticate("token", { token: "root"})

        const tx = {
            type: 0,
            to: "0xe74b28c2eAe8679e3cCc3a94d5d0dE83CCB84705",
            value: ethers.utils.parseEther("1"),
            nonce: await provider.getTransactionCount(signer.address, "latest"),
            gasLimit: BigNumber.from("1000000"), // 100000
        }
        await signer.sendTransaction(tx)
        
    }); 

    it('Make a contract call', async function() {
        let provider = new providers.JsonRpcProvider("http://127.0.0.1:8545")
        let signer = new VaultEthereumSigner({
            endpoint: "http://127.0.0.1:8200",
            walletId: "test",
            pluginPath: "vault-ethereum"
        }, provider)

        await signer.authenticate("token", { token: "root"})

        const abi = [
            "function deposit() payable external",
            "function withdraw(uint256) external",
            "function transfer(address _to, uint256 _amount) external returns (bool)",
            "function transferFrom(address _from, address _to, uint256 _amount) external",
            "function approve(address _to, uint256 _amount) external",
            "function balanceOf(address _address) external view returns (uint256)",
            "function nonces(address _owner) external view returns (uint256)",
            "function name() external view returns (string memory)",
        ]
    
        const WETHContract = new ethers.Contract("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", abi, signer)
        await WETHContract.deposit({ value: ethers.utils.parseEther("1"), type: 0 }) //mint 1000 USDT

        
    }); 

});