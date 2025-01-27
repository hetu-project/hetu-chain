const { expect } = require("chai");

describe("CohortManager", function () {
    let CohortManager;
    let cohortManager;
    let owner;
    let addr1;
    let addr2;

    beforeEach(async function () {
        CohortManager = await ethers.getContractFactory("CohortManager");
        [owner, addr1, addr2] = await ethers.getSigners();

        cohortManager = await CohortManager.deploy();
        console.log(await cohortManager.getAddress());
    });

    describe("Cohort Registration", function () {
        it("should register a new cohort", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            const cohort = await cohortManager.getCohort("example.com");

            expect(cohort.serviceAddr).to.equal("0x1234567890abcdef1234567890abcdef12345678");
            expect(cohort.owner).to.equal(owner.address);
        });

        it("should emit CohortRegistered event on successful registration", async function () {
            await expect(cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678"))
                .to.emit(cohortManager, "CohortRegistered")
                .withArgs("example.com", "0x1234567890abcdef1234567890abcdef12345678", owner.address);
        });

        it("should not allow registering the same cohort twice", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await expect(cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678"))
                .to.be.revertedWith("Cohort is already registered");
        });
    });

    describe("Cohort Unregistration", function () {
        it("should unregister a cohort", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await cohortManager.unregisterCohort("example.com");

            await expect(cohortManager.getCohort("example.com")).to.be.revertedWith("Cohort is not registered");
        });

        it("should emit CohortUnregistered event on successful unregistration", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await expect(cohortManager.unregisterCohort("example.com"))
                .to.emit(cohortManager, "CohortUnregistered")
                .withArgs("example.com", owner.address);
        });

        it("should not allow non-owners to unregister a cohort", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await expect(cohortManager.connect(addr1).unregisterCohort("example.com"))
                .to.be.revertedWith("Only the owner can unregister the cohort");
        });
    });

    describe("Service Address Updates", function () {
        it("should update the service address", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await cohortManager.updateServiceAddr("example.com", "0xabcdef1234567890abcdef1234567890abcdef12");

            const cohort = await cohortManager.getCohort("example.com");
            expect(cohort.serviceAddr).to.equal("0xabcdef1234567890abcdef1234567890abcdef12");
        });

        it("should emit ServiceAddrUpdated event on successful update", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await expect(cohortManager.updateServiceAddr("example.com", "0xabcdef1234567890abcdef1234567890abcdef12"))
                .to.emit(cohortManager, "ServiceAddrUpdated")
                .withArgs("example.com", "0xabcdef1234567890abcdef1234567890abcdef12");
        });

        it("should not allow non-owners to update the service address", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await expect(cohortManager.connect(addr1).updateServiceAddr("example.com", "0xabcdef1234567890abcdef1234567890abcdef12"))
                .to.be.revertedWith("Only the owner can update the service address");
        });
    });

    describe("Cohort Ownership Transfer", function () {
        it("should transfer cohort ownership", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await cohortManager.transferCohortOwnership("example.com", addr1.address);

            const cohort = await cohortManager.getCohort("example.com");
            expect(cohort.owner).to.equal(addr1.address);
        });

        it("should not allow non-owners to transfer the cohort", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await expect(cohortManager.connect(addr1).transferCohortOwnership("example.com", addr2.address))
                .to.be.revertedWith("Only the owner can transfer the cohort");
        });
    });

    describe("Get Registered Cohorts", function () {
        it("should return all registered cohort names", async function () {
            await cohortManager.registerCohort("example.com", "0x1234567890abcdef1234567890abcdef12345678");
            await cohortManager.registerCohort("test.com", "0xabcdef1234567890abcdef1234567890abcdef12");

            const cohorts = await cohortManager.getAllCohorts();
            console.log(cohorts);
            const cohortArray = [...cohorts];
            expect(cohortArray).to.have.members(["example.com", "test.com"]);
        });
    });
});
