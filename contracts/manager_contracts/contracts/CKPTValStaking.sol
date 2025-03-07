// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/utils/math/Math.sol";

contract CKPTValStaking is Ownable {
    using Math for uint256;

    // Staking token (could be the native token or an ERC20)
    IERC20 public stakingToken;

    // Reward rate (tokens per second)
    uint256 public rewardRate;

    // Unstaking lock period in seconds
    uint256 public lockPeriod = 7 days;

    // Minimum stake amount
    uint256 public minimumStake;

    // Cursor for validator rotation in getTopValidators
    uint256 private validatorCursor = 0;

    // Validator structure
    struct Validator {
        uint256 stakedAmount;
        uint256 lastRewardTime;
        uint256 pendingRewards;
        uint256 unstakeTime; // 0 means not unstaking
        string dispatcherURL;
        bool isActive;
    }

    // Mapping from address to validator info
    mapping(address => Validator) public validators;

    // Array of validator addresses for enumeration
    address[] public validatorAddresses;

    // Events
    event Staked(address indexed validator, uint256 amount);
    event UnstakeInitiated(
        address indexed validator,
        uint256 amount,
        uint256 unlockTime
    );
    event Unstaked(address indexed validator, uint256 amount);
    event RewardsClaimed(address indexed validator, uint256 amount);
    event ValidatorRegistered(address indexed validator, string dispatcherURL);
    event ValidatorUpdated(address indexed validator, string dispatcherURL);

    constructor(
        address _stakingToken,
        uint256 _rewardRate,
        uint256 _minimumStake
    ) Ownable(msg.sender) {
        stakingToken = IERC20(_stakingToken);
        rewardRate = _rewardRate;
        minimumStake = _minimumStake;
    }

    // Set reward rate (only owner)
    function setRewardRate(uint256 _rewardRate) external onlyOwner {
        rewardRate = _rewardRate;
    }

    // Set lock period (only owner)
    function setLockPeriod(uint256 _lockPeriod) external onlyOwner {
        lockPeriod = _lockPeriod;
    }

    // Set minimum stake (only owner)
    function setMinimumStake(uint256 _minimumStake) external onlyOwner {
        minimumStake = _minimumStake;
    }

    // Register as a validator with dispatcher URL
    function registerValidator(string calldata _dispatcherURL) external {
        require(validators[msg.sender].stakedAmount == 0, "Already registered");

        validators[msg.sender] = Validator({
            stakedAmount: 0,
            lastRewardTime: block.timestamp,
            pendingRewards: 0,
            unstakeTime: 0,
            dispatcherURL: _dispatcherURL,
            isActive: false
        });

        validatorAddresses.push(msg.sender);

        emit ValidatorRegistered(msg.sender, _dispatcherURL);
    }

    // Update validator's dispatcher URL
    function updateDispatcherURL(string calldata _dispatcherURL) external {
        require(validators[msg.sender].stakedAmount > 0, "Not a validator");

        validators[msg.sender].dispatcherURL = _dispatcherURL;

        emit ValidatorUpdated(msg.sender, _dispatcherURL);
    }

    // Stake tokens to become a validator
    function stake(uint256 _amount) external {
        require(_amount == minimumStake, "Must stake exact minimum amount");
        require(validators[msg.sender].stakedAmount == 0, "Already staked");
        require(
            validators[msg.sender].unstakeTime == 0,
            "Unstaking in progress"
        );

        // Update rewards before changing stake
        _updateRewards(msg.sender);

        // Transfer tokens from sender to contract
        stakingToken.transferFrom(msg.sender, address(this), _amount);

        // Update validator's staked amount
        validators[msg.sender].stakedAmount =
            validators[msg.sender].stakedAmount +
            _amount;

        // Activate validator if not already active
        if (!validators[msg.sender].isActive) {
            validators[msg.sender].isActive = true;
            validatorAddresses.push(msg.sender); // Add new validator address
        }

        emit Staked(msg.sender, _amount);
    }

    // Initiate unstaking process
    function initiateUnstake(uint256 _amount) external {
        require(_amount > 0, "Cannot unstake 0");
        require(
            validators[msg.sender].stakedAmount >= _amount,
            "Not enough staked"
        );
        require(
            validators[msg.sender].unstakeTime == 0,
            "Unstaking already in progress"
        );

        // Update rewards before changing stake
        _updateRewards(msg.sender);

        // Set unstake time
        validators[msg.sender].unstakeTime = block.timestamp + lockPeriod;

        // Update validator's staked amount
        validators[msg.sender].stakedAmount =
            validators[msg.sender].stakedAmount -
            _amount;

        // Deactivate validator if fully unstaking
        if (validators[msg.sender].stakedAmount == 0) {
            validators[msg.sender].isActive = false;
        }

        emit UnstakeInitiated(
            msg.sender,
            _amount,
            validators[msg.sender].unstakeTime
        );
    }

    // Complete unstaking after lock period
    function completeUnstake() external {
        require(
            validators[msg.sender].unstakeTime > 0,
            "No unstaking in progress"
        );
        require(
            block.timestamp >= validators[msg.sender].unstakeTime,
            "Still in lock period"
        );

        uint256 unstakeAmount = validators[msg.sender].stakedAmount;

        // Reset unstake time
        validators[msg.sender].unstakeTime = 0;

        // Transfer tokens back to validator
        stakingToken.transfer(msg.sender, unstakeAmount);

        emit Unstaked(msg.sender, unstakeAmount);
    }

    // Claim accumulated rewards
    function claimRewards() external {
        _updateRewards(msg.sender);

        uint256 rewards = validators[msg.sender].pendingRewards;
        require(rewards > 0, "No rewards to claim");

        validators[msg.sender].pendingRewards = 0;

        // Transfer rewards to validator
        stakingToken.transfer(msg.sender, rewards);

        emit RewardsClaimed(msg.sender, rewards);
    }

    // Get validator info
    function getValidator(
        address _validator
    )
        external
        view
        returns (
            uint256 stakedAmount,
            uint256 pendingRewards,
            uint256 unstakeTime,
            string memory dispatcherURL,
            bool isActive
        )
    {
        Validator storage validator = validators[_validator];

        // Calculate current rewards
        uint256 currentRewards = validator.pendingRewards;
        if (validator.isActive && validator.stakedAmount > 0) {
            uint256 timeElapsed = block.timestamp - validator.lastRewardTime;
            currentRewards =
                currentRewards +
                ((timeElapsed * rewardRate * validator.stakedAmount) / 1e18);
        }

        return (
            validator.stakedAmount,
            currentRewards,
            validator.unstakeTime,
            validator.dispatcherURL,
            validator.isActive
        );
    }

    // Get top N validators by rotation (view function)
    function getTopValidators(
        uint256 _count
    )
        external
        view
        returns (
            address[] memory addresses,
            uint256[] memory stakes,
            string[] memory dispatcherURLs
        )
    {
        uint256 validatorCount = validatorAddresses.length;

        // If no validators, return empty arrays
        if (validatorCount == 0) {
            return (new address[](0), new uint256[](0), new string[](0));
        }

        // Determine how many validators to return
        uint256 returnCount = _count < validatorCount ? _count : validatorCount;

        addresses = new address[](returnCount);
        stakes = new uint256[](returnCount);
        dispatcherURLs = new string[](returnCount);

        uint256 validFound = 0;
        uint256 startCursor = validatorCursor;

        // Loop through validators starting from cursor
        for (
            uint256 i = 0;
            i < validatorCount && validFound < returnCount;
            i++
        ) {
            uint256 index = (startCursor + i) % validatorCount;
            address validatorAddr = validatorAddresses[index];
            Validator storage validator = validators[validatorAddr];

            if (validator.isActive && validator.stakedAmount > 0) {
                addresses[validFound] = validatorAddr;
                stakes[validFound] = validator.stakedAmount;
                dispatcherURLs[validFound] = validator.dispatcherURL;
                validFound++;
            }
        }

        // If we found fewer validators than requested, resize arrays
        if (validFound < returnCount) {
            assembly {
                mstore(addresses, validFound)
                mstore(stakes, validFound)
                mstore(dispatcherURLs, validFound)
            }
        }

        return (addresses, stakes, dispatcherURLs);
    }

    // Update cursor for validator rotation (non-view function)
    function updateValidatorCursor(uint256 _count) external {
        uint256 validatorCount = validatorAddresses.length;
        require(validatorCount > 0, "No validators available");
        uint256 startCursor = validatorCursor;
        uint256 returnCount = _count < validatorCount ? _count : validatorCount;
        validatorCursor = (startCursor + returnCount) % validatorCount;
    }

    // Get total number of validators
    function getValidatorCount() external view returns (uint256) {
        return validatorAddresses.length;
    }

    // Get stake amount for a specific validator
    function getStake(address _validator) external view returns (uint256) {
        return validators[_validator].stakedAmount;
    }

    // Internal function to update rewards
    function _updateRewards(address _validator) internal {
        Validator storage validator = validators[_validator];

        if (validator.isActive && validator.stakedAmount > 0) {
            uint256 timeElapsed = block.timestamp - validator.lastRewardTime;
            if (timeElapsed > 0) {
                uint256 rewards = (timeElapsed *
                    rewardRate *
                    validator.stakedAmount) / 1e18;
                validator.pendingRewards = validator.pendingRewards + rewards;
                validator.lastRewardTime = block.timestamp;
            }
        }
    }
}
