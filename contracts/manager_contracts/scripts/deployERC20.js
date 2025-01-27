const { ethers } = require("hardhat");

async function main() {
    const initialSupply = ethers.parseUnits("10000000000", 4); // 10 billion HET，4 decimals
    const MyToken = await ethers.getContractFactory("MyToken");
    const myToken = await MyToken.deploy(initialSupply);

    console.log(`MyToken deployed to: ${myToken.target}`);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
