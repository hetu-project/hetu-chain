// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/utils/math/Math.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract CKPTValStaking is Ownable, AccessControl {
    using Math for uint256;

    // Define a new role for validators
    bytes32 public constant VALIDATOR_ROLE = keccak256("VALIDATOR_ROLE");

    // Define new roles
    bytes32 public constant DISPATCHER_ROLE = keccak256("DISPATCHER_ROLE");
    bytes32 public constant DISTRIBUTER_ROLE = keccak256("DISTRIBUTER_ROLE");

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

    // Stake lock period in seconds
    uint256 public stakeLockPeriod = 1 days;

    // Mapping to track when a stake becomes active
    mapping(address => uint256) public stakeActivationTime;

    // Validator structure
    struct Validator {
        uint256 stakedAmount;
        uint256 lastRewardTime;
        uint256 pendingRewards;
        uint256 unstakeTime; // 0 means not unstaking
        string dispatcherURL;
        string blsPublicKey;
        bool isActive;
        uint256 index; // Index in the validatorAddresses array
    }

    // Checkpoint structure
    struct Checkpoint {
        uint64 epochNum; // Epoch number
        bytes32 blockHash; // Hash of the latest block
        bytes bitmap; // Bitmap indicating BLS signers
        bytes blsMultiSig; // Aggregated BLS multi-signature
        bytes blsAggrPk; // Aggregated BLS public key
        uint64 powerSum; // Accumulated voting power
    }

    // Mapping from epoch to checkpoint
    mapping(uint64 => Checkpoint) public epochToCheckpoint;

    // Mapping from address to validator info
    mapping(address => Validator) public validators;

    // Array of validator addresses for enumeration
    address[] public validatorAddresses;

    // Mapping to track distributed epochs
    mapping(uint64 => bool) public distributedEpochs;

    // Events
    event Staked(address indexed validator, uint256 amount);
    event UnstakeInitiated(
        address indexed validator,
        uint256 amount,
        uint256 unlockTime
    );
    event Unstaked(address indexed validator, uint256 amount);
    event RewardsClaimed(address indexed validator, uint256 amount);
    event ValidatorRegistered(
        address indexed validator,
        string dispatcherURL,
        string blsPublicKey
    );
    event ValidatorUpdated(
        address indexed validator,
        string dispatcherURL,
        string blsPublicKey
    );
    event CheckpointSubmitted(
        uint64 indexed epochNum,
        bytes32 blockHash,
        uint64 powerSum
    );

    constructor(
        address _stakingToken,
        uint256 _rewardRate,
        uint256 _minimumStake
    ) Ownable(msg.sender) {
        stakingToken = IERC20(_stakingToken);
        rewardRate = _rewardRate;
        minimumStake = _minimumStake;

        // Grant the owner the default admin role and validator role
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(VALIDATOR_ROLE, msg.sender);
        _grantRole(DISPATCHER_ROLE, msg.sender);
        _grantRole(DISTRIBUTER_ROLE, msg.sender);
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

    // Set stake lock period (only owner)
    function setStakeLockPeriod(uint256 _stakeLockPeriod) external onlyOwner {
        stakeLockPeriod = _stakeLockPeriod;
    }

    // Register as a validator with dispatcher URL and BLS public key
    function registerValidator(
        string calldata _dispatcherURL,
        string calldata _blsPublicKey
    ) external {
        require(validators[msg.sender].stakedAmount == 0, "Already registered");
        require(
            hasRole(VALIDATOR_ROLE, msg.sender),
            "CKPTValStaking: must have validator role to stake"
        );

        validators[msg.sender] = Validator({
            stakedAmount: 0,
            lastRewardTime: block.timestamp,
            pendingRewards: 0,
            unstakeTime: 0,
            dispatcherURL: _dispatcherURL,
            blsPublicKey: _blsPublicKey,
            isActive: false,
            index: validatorAddresses.length
        });

        validatorAddresses.push(msg.sender);

        emit ValidatorRegistered(msg.sender, _dispatcherURL, _blsPublicKey);
    }

    // Update validator's dispatcher URL and BLS public key
    function updateValidatorInfo(
        string calldata _dispatcherURL,
        string calldata _blsPublicKey
    ) external {
        require(validators[msg.sender].stakedAmount > 0, "Not a validator");
        require(
            hasRole(VALIDATOR_ROLE, msg.sender),
            "CKPTValStaking: must have validator role to stake"
        );

        validators[msg.sender].dispatcherURL = _dispatcherURL;
        validators[msg.sender].blsPublicKey = _blsPublicKey;

        emit ValidatorUpdated(msg.sender, _dispatcherURL, _blsPublicKey);
    }

    // Grant validator role (only callable by owner or admin)
    function grantValidatorRole(address account) external onlyOwner {
        grantRole(VALIDATOR_ROLE, account);
    }

    // Revoke validator role (only callable by owner or admin)
    function revokeValidatorRole(address account) external onlyOwner {
        revokeRole(VALIDATOR_ROLE, account);
    }

    // Grant dispatcher role (only callable by owner)
    function grantDispatcherRole(address account) external onlyOwner {
        grantRole(DISPATCHER_ROLE, account);
    }

    // Revoke dispatcher role (only callable by owner)
    function revokeDispatcherRole(address account) external onlyOwner {
        revokeRole(DISPATCHER_ROLE, account);
    }

    // Grant distributer role (only callable by owner)
    function grantDistributerRole(address account) external onlyOwner {
        grantRole(DISTRIBUTER_ROLE, account);
    }

    // Revoke distributer role (only callable by owner)
    function revokeDistributerRole(address account) external onlyOwner {
        revokeRole(DISTRIBUTER_ROLE, account);
    }

    // Stake tokens to become a validator
    function stake(uint256 _amount) external {
        require(
            hasRole(VALIDATOR_ROLE, msg.sender),
            "CKPTValStaking: must have validator role to stake"
        );
        require(
            _amount >= minimumStake,
            "Must stake exact bigger than minimum amount"
        );
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
            validators[msg.sender].index = validatorAddresses.length;
            validatorAddresses.push(msg.sender); // Add new validator address
        }

        // Set stake activation time
        stakeActivationTime[msg.sender] = block.timestamp + stakeLockPeriod;

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

    // Submit a checkpoint (only callable by dispatcher)
    function submitCheckpoint(
        uint64 _epochNum,
        bytes32 _blockHash,
        bytes calldata _bitmap,
        bytes calldata _blsMultiSig,
        bytes calldata _blsAggrPk,
        uint64 _powerSum
    ) external {
        require(
            hasRole(DISPATCHER_ROLE, msg.sender),
            "CKPTValStaking: must have dispatcher role to submit checkpoint"
        );
        require(
            epochToCheckpoint[_epochNum].epochNum == 0,
            "Checkpoint already exists for this epoch"
        );

        epochToCheckpoint[_epochNum] = Checkpoint({
            epochNum: _epochNum,
            blockHash: _blockHash,
            bitmap: _bitmap,
            blsMultiSig: _blsMultiSig,
            blsAggrPk: _blsAggrPk,
            powerSum: _powerSum
        });

        emit CheckpointSubmitted(_epochNum, _blockHash, _powerSum);
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
            string memory blsPublicKey,
            bool isActive,
            uint256 index,
            uint256 activationTime
        )
    {
        Validator storage validator = validators[_validator];

        // Calculate current rewards
        uint256 currentRewards = validator.pendingRewards;
        if (
            validator.isActive &&
            validator.stakedAmount > 0 &&
            block.timestamp >= stakeActivationTime[_validator]
        ) {
            uint256 timeElapsed = block.timestamp - validator.lastRewardTime;
            currentRewards =
                currentRewards +
                ((timeElapsed * rewardRate * validator.stakedAmount) / 1e18);
        }

        activationTime = stakeActivationTime[_validator];

        return (
            validator.stakedAmount,
            currentRewards,
            validator.unstakeTime,
            validator.dispatcherURL,
            validator.blsPublicKey,
            validator.isActive,
            validator.index,
            activationTime
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
            string[] memory dispatcherURLs,
            string[] memory blsPublicKeys
        )
    {
        uint256 validatorCount = validatorAddresses.length;

        // If no validators, return empty arrays
        if (validatorCount == 0) {
            return (
                new address[](0),
                new uint256[](0),
                new string[](0),
                new string[](0)
            );
        }

        // Determine how many validators to return
        uint256 returnCount = _count < validatorCount ? _count : validatorCount;

        addresses = new address[](returnCount);
        stakes = new uint256[](returnCount);
        dispatcherURLs = new string[](returnCount);
        blsPublicKeys = new string[](returnCount);

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

            if (
                validator.isActive &&
                validator.stakedAmount > 0 &&
                block.timestamp >= stakeActivationTime[validatorAddr]
            ) {
                addresses[validFound] = validatorAddr;
                stakes[validFound] = validator.stakedAmount;
                dispatcherURLs[validFound] = validator.dispatcherURL;
                blsPublicKeys[validFound] = validator.blsPublicKey;
                validFound++;
            }
        }

        // If we found fewer validators than requested, resize arrays
        if (validFound < returnCount) {
            assembly {
                mstore(addresses, validFound)
                mstore(stakes, validFound)
                mstore(dispatcherURLs, validFound)
                mstore(blsPublicKeys, validFound)
            }
        }

        return (addresses, stakes, dispatcherURLs, blsPublicKeys);
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

        if (
            validator.isActive &&
            validator.stakedAmount > 0 &&
            block.timestamp >= stakeActivationTime[_validator]
        ) {
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

    // Distribute checkpoint rewards (only callable by distributer)
    function distributeCheckpointRewards(uint64 _epochNum) external {
        require(
            hasRole(DISTRIBUTER_ROLE, msg.sender),
            "CKPTValStaking: must have distributer role to distribute rewards"
        );
        Checkpoint storage checkpoint = epochToCheckpoint[_epochNum];
        require(
            checkpoint.epochNum != 0,
            "Checkpoint does not exist for this epoch"
        );
        require(
            !distributedEpochs[_epochNum],
            "Rewards already distributed for this epoch"
        );

        // Get sorted validator set (max 512 validators)
        address[] memory sortedValidators = _getSortedValidators();

        uint256 validatorCount = sortedValidators.length;
        require(validatorCount > 0, "No validators available for rewards");

        uint256 rewardPerValidator = checkpoint.powerSum / validatorCount;

        for (uint256 i = 0; i < validatorCount; i++) {
            address validator = sortedValidators[i];
            uint256 validatorIndex = i;

            // Check if the validator participated in the checkpoint
            if (
                validatorIndex < 512 &&
                _isBitmapSet(checkpoint.bitmap, validatorIndex)
            ) {
                validators[validator].pendingRewards += rewardPerValidator;
            }
        }

        // Mark the epoch as distributed
        distributedEpochs[_epochNum] = true;
    }

    // Internal function to get sorted validator set (max 512 validators)
    function _getSortedValidators() internal view returns (address[] memory) {
        uint256 validatorCount = validatorAddresses.length;
        uint256 count = validatorCount > 512 ? 512 : validatorCount;

        require(count > 0, "No validators available");

        address[] memory sortedValidators = new address[](count);
        for (uint256 i = 0; i < count; i++) {
            sortedValidators[i] = validatorAddresses[i];
        }

        // Sort validators by address
        for (uint256 i = 0; i < count - 1; i++) {
            for (uint256 j = 0; j < count - i - 1; j++) {
                if (
                    _compareAddresses(
                        sortedValidators[j],
                        sortedValidators[j + 1]
                    )
                ) {
                    (sortedValidators[j], sortedValidators[j + 1]) = (
                        sortedValidators[j + 1],
                        sortedValidators[j]
                    );
                }
            }
        }

        return sortedValidators;
    }

    // Internal function to compare two addresses (BigEndian order)
    function _compareAddresses(
        address a,
        address b
    ) internal pure returns (bool) {
        return uint256(uint160(a)) > uint256(uint160(b));
    }

    // Internal function to check if a bit is set in the bitmap
    function _isBitmapSet(
        bytes memory bitmap,
        uint256 index
    ) internal pure returns (bool) {
        uint256 byteIndex = index / 8;
        uint256 bitIndex = index % 8;
        if (byteIndex >= bitmap.length) return false;
        return (uint8(bitmap[byteIndex]) & (1 << bitIndex)) != 0;
    }
}