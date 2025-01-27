# Cohort Manager Contracts

## Overview

A smart contract project built with Hardhat framework for managing cohort-related operations. This project provides a complete development environment including:

- Smart contracts for cohort management
- Comprehensive test suite
- Deployment scripts for multiple networks
- Local development chain support

The project utilizes Hardhat's robust development tools and testing capabilities, making it ideal for both local development and production deployments. You can easily run tests, deploy contracts, and interact with them using the provided scripts.

Key Features:
- Hardhat integration
- Gas reporting capabilities
- Multi-network deployment support
- Local chain testing environment

## Run
Try running some of the following tasks:

```shell
npx hardhat help
npx hardhat test
REPORT_GAS=true npx hardhat test
npx hardhat node
npx hardhat run scripts/deploy.js
# Or local chain
npx hardhat run scripts/deploy.js --network local
```