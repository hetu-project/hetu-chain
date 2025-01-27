// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/math/Math.sol";

/**
 * @title Staking Contract for PoS blockchain
 * @notice This contract allows users to stake tokens and earn rewards
 */
contract Staking is ReentrancyGuard, Ownable {
    using Math for uint256;

    // Staking details for each user
    struct Stake {
        uint256 amount;        // Amount of tokens staked
        uint256 since;         // Timestamp when stake was made
        uint256 claimedRewards; // Already claimed rewards
    }
    
    // Mapping of user address to their stake info
    mapping(address => Stake) public stakes;
    
    // Annual reward rate in percentage (e.g. 10 = 10%)
    uint256 public rewardRate;
    
    // Minimum staking amount
    uint256 public minimumStake;
    
    // Total staked amount
    uint256 public totalStaked;

    // Events
    event Staked(address indexed user, uint256 amount);
    event Unstaked(address indexed user, uint256 amount);
    event RewardsClaimed(address indexed user, uint256 amount);

    /**
     * @dev Constructor
     * @param _rewardRate Annual reward rate in percentage
     * @param _minimumStake Minimum amount that can be staked
     */
    constructor(
        uint256 _rewardRate,
        uint256 _minimumStake
    ) Ownable(msg.sender) {
        rewardRate = _rewardRate;
        minimumStake = _minimumStake;
    }

    /**
     * @dev Stakes tokens
     */
    function stake() external payable nonReentrant {
        require(msg.value >= minimumStake, "Amount below minimum stake");

        if (stakes[msg.sender].amount > 0) {
            claimRewards();
        }

        stakes[msg.sender].amount += msg.value;
        stakes[msg.sender].since = block.timestamp;
        totalStaked += msg.value;

        emit Staked(msg.sender, msg.value);
    }

    /**
     * @dev Calculates pending rewards for a user
     * @param _user Address of the user
     * @return Pending reward amount
     */
    function calculateRewards(address _user) public view returns (uint256) {
        Stake memory userStake = stakes[_user];
        if (userStake.amount == 0) return 0;

        uint256 timeStaked = block.timestamp - userStake.since;
        uint256 rewards = userStake.amount
            * rewardRate
            * timeStaked
            / 365 days
            / 100;

        return rewards - userStake.claimedRewards;
    }

    /**
     * @dev Claims pending rewards
     */
    function claimRewards() public nonReentrant {
        uint256 rewards = calculateRewards(msg.sender);
        require(rewards > 0, "No rewards to claim");
        require(address(this).balance >= rewards, "Insufficient contract balance for rewards");

        stakes[msg.sender].claimedRewards += rewards;
        (bool success, ) = payable(msg.sender).call{value: rewards}("");
        require(success, "Transfer failed");

        emit RewardsClaimed(msg.sender, rewards);
    }

    /**
     * @dev Unstakes tokens
     * @param _amount Amount of tokens to unstake
     */
    function unstake(uint256 _amount) external nonReentrant {
        require(stakes[msg.sender].amount >= _amount, "Insufficient staked amount");

        // directly handle rewards here
        uint256 rewards = calculateRewards(msg.sender);
        if (rewards > 0) {
            require(address(this).balance >= rewards, "Insufficient contract balance for rewards");
            stakes[msg.sender].claimedRewards += rewards;
            (bool rewardSuccess, ) = payable(msg.sender).call{value: rewards}("");
            require(rewardSuccess, "Reward transfer failed");
            emit RewardsClaimed(msg.sender, rewards);
        }

        stakes[msg.sender].amount -= _amount;
        totalStaked -= _amount;

        (bool success, ) = payable(msg.sender).call{value: _amount}("");
        require(success, "Transfer failed");

        emit Unstaked(msg.sender, _amount);
    }

    /**
     * @dev Updates the reward rate
     * @param _newRate New annual reward rate in percentage
     */
    function setRewardRate(uint256 _newRate) external onlyOwner {
        rewardRate = _newRate;
    }

    /**
     * @dev Updates the minimum stake amount
     * @param _newMinimum New minimum stake amount
     */
    function setMinimumStake(uint256 _newMinimum) external onlyOwner {
        minimumStake = _newMinimum;
    }

    // contract need to receive ETH
    receive() external payable {}
}
