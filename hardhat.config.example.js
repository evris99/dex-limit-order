/**
 * @type import('hardhat/config').HardhatUserConfig
 */
module.exports = {
  solidity: "0.7.3",
  defaultNetwork: "hardhat",
  networks: {
    hardhat: {
      chainId: 1337,
      loggingEnabled: true,
      forking: {
        enabled: true,
        url: "<ARCHIVE_NODE_URL>",
        blockNumber: 11163381
      },
    }
  }
};
