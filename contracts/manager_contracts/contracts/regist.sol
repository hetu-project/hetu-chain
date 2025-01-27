// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract CohortManager {
    struct Cohort {
        string serviceAddr;
        address owner;
    }

    // Mapping from cohort name to Cohort struct
    mapping(string => Cohort) private cohorts;

    // Array to store all registered cohort names
    string[] private cohortNames;

    // Events
    event CohortRegistered(string indexed cohortName, string serviceAddr, address indexed owner);
    event CohortUnregistered(string indexed cohortName, address indexed owner);
    event ServiceAddrUpdated(string indexed cohortName, string newServiceAddr);

    // Register a new cohort
    function registerCohort(string memory cohortName, string memory serviceAddr) public {
        require(bytes(cohortName).length > 0, "Cohort name cannot be empty");
        require(bytes(serviceAddr).length > 0, "Service address cannot be empty");
        require(cohorts[cohortName].owner == address(0), "Cohort is already registered");

        cohorts[cohortName] = Cohort({
            serviceAddr: serviceAddr,
            owner: msg.sender
        });

        cohortNames.push(cohortName);

        emit CohortRegistered(cohortName, serviceAddr, msg.sender);
    }

    // Unregister a cohort
    function unregisterCohort(string memory cohortName) public {
        require(cohorts[cohortName].owner == msg.sender, "Only the owner can unregister the cohort");

        // Find and remove the cohort from the cohortNames array
        for (uint i = 0; i < cohortNames.length; i++) {
            if (keccak256(bytes(cohortNames[i])) == keccak256(bytes(cohortName))) {
                cohortNames[i] = cohortNames[cohortNames.length - 1];
                cohortNames.pop();
                break;
            }
        }

        delete cohorts[cohortName];

        emit CohortUnregistered(cohortName, msg.sender);
    }

    // Get cohort information
    function getCohort(string memory cohortName) public view returns (string memory serviceAddr, address owner) {
        require(cohorts[cohortName].owner != address(0), "Cohort is not registered");

        Cohort memory cohort = cohorts[cohortName];
        return (cohort.serviceAddr, cohort.owner);
    }

    // Transfer cohort ownership
    function transferCohortOwnership(string memory cohortName, address newOwner) public {
        require(cohorts[cohortName].owner == msg.sender, "Only the owner can transfer the cohort");

        cohorts[cohortName].owner = newOwner;
    }

    // Update service address for a cohort
    function updateServiceAddr(string memory cohortName, string memory newServiceAddr) public {
        require(cohorts[cohortName].owner == msg.sender, "Only the owner can update the service address");
        require(bytes(newServiceAddr).length > 0, "New service address cannot be empty");

        cohorts[cohortName].serviceAddr = newServiceAddr;

        emit ServiceAddrUpdated(cohortName, newServiceAddr);
    }

    // Get all registered cohort names
    function getAllCohorts() public view returns (string[] memory) {
        return cohortNames;
    }
}