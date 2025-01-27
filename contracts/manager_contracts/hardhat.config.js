require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.20",
  networks: {
    rinkeby: {
        url: "https://rinkeby.infura.io/v3/YOUR_INFURA_PROJECT_ID",
        accounts: ["***"]
    },
    local: {
        url: "http://127.0.0.1:8545",
        accounts: ["***"]
    },
}
};
