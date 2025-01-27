const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("Staking", function () {
  let Staking;
  let staking;
  let owner;
  let addr1;
  let addr2;

  beforeEach(async function () {
    // deploy contract
    Staking = await ethers.getContractFactory("Staking");
    [owner, addr1, addr2] = await ethers.getSigners();
    staking = await Staking.deploy(
        10,                   // 10% annual reward rate
        ethers.parseEther("1") // minimum stake amount is 1 ETH
    );
    await staking.waitForDeployment();
    console.log(await staking.getAddress());
    
    // send some ETH to the contract for rewards
    await owner.sendTransaction({
      to: await staking.getAddress(),
      value: ethers.parseEther("10.0") // 10 ETH as reward pool
    });
  });

  describe("stake", function () {
    it("should allow user to stake ETH", async function () {
      const stakeAmount = ethers.parseEther("1.0");
      await staking.connect(addr1).stake({ value: stakeAmount });
      
      const userStake = await staking.stakes(addr1.address);
      expect(userStake.amount).to.equal(stakeAmount);
    });

    it("should correctly record stake time", async function () {
      const stakeAmount = ethers.parseEther("1.0");
      await staking.connect(addr1).stake({ value: stakeAmount });
      
      const userStake = await staking.stakes(addr1.address);
      const stakeTime = userStake.since;
      expect(stakeTime).to.be.above(0);
    });
  });

  describe("withdraw", function () {
    it("should allow user to withdraw staked tokens", async function () {
      const stakeAmount = ethers.parseEther("1.0");
      await staking.connect(addr1).stake({ value: stakeAmount });
      
      await staking.connect(addr1).unstake(stakeAmount);
      
      const userStake = await staking.stakes(addr1.address);
      expect(userStake.amount).to.equal(0);
    });
  });

  describe("calculateReward", function () {
    it("should correctly calculate stake reward", async function () {
      const stakeAmount = ethers.parseEther("1.0");
      await staking.connect(addr1).stake({ value: stakeAmount });
      
      // simulate time passing
      await network.provider.send("evm_increaseTime", [86400]); // increase 1 day
      await network.provider.send("evm_mine");
      
      const reward = await staking.calculateRewards(addr1.address);
      expect(reward).to.be.above(0);
    });
  });
});
