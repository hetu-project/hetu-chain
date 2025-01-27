const { ethers } = require("hardhat");
const { fromBech32 } = require('@cosmjs/encoding');

function convertEvmosToEthereumAddress(evmosAddress) {
    const { prefix, data } = fromBech32(evmosAddress);
    if (prefix !== 'evmos') {
        throw new Error('Invalid Evmos address');
    }

    // convert Bech32 words to hex
    return '0x' + Buffer.from(data).toString('hex');
}

// Function to query balance
async function queryBalance(tokenContract, address, decimals) {
    const balance = await tokenContract.balanceOf(address);
    console.log(`${address} Balance: ${ethers.formatUnits(balance, decimals)} HET`);
}

async function main() {
    const myTokenAddress = "0x480cBFD2D9b0dAA01Af7Ddd74178aFA1bC4F7544";
    const MyToken = await ethers.getContractAt("MyToken", myTokenAddress);

    const evmosAddress = "evmos1rkutcsdq23pg356pejhdld2ek56q2j5vupl3sk";
    const ethereumAddress = convertEvmosToEthereumAddress(evmosAddress);
    await queryBalance(MyToken, ethereumAddress, 4);

    await queryBalance(MyToken, "0xddC3Cd21222358BC893895CCd43B4F8BE46F254E", 4);
    await MyToken.transfer("0xddC3Cd21222358BC893895CCd43B4F8BE46F254E", ethers.parseUnits("3000", 4));
    await new Promise(resolve => setTimeout(resolve, 8000));
    await queryBalance(MyToken, "0xddC3Cd21222358BC893895CCd43B4F8BE46F254E", 4);
}

main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exit(1);
    });
