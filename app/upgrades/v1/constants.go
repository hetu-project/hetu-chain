package v1

const (
	// UpgradeName is the shared upgrade plan name for mainnet
	UpgradeName = "v1.0.1"

	// TestnetUpgradeHeight defines the Evmos testnet block height on which the upgrade will take place
	TestnetUpgradeHeight = 800

	// UpgradeInfo defines the binaries that will be used for the upgrade
	// UpgradeInfo = `'{"binaries":{"darwin/amd64":"https://....tar.gz","darwin/x86_64":"https://....tar.gz","linux/arm64":"https://....tar.gz","linux/amd64":"https://....tar.gz","windows/x86_64":"https://....zip"}}'`
	UpgradeInfo = `'{"binaries":{"darwin/amd64":"https://github.com/evmos/evmos/releases/download/v19.2.0/evmos_19.2.0_Darwin_arm64.tar.gz","darwin/x86_64":"https://github.com/evmos/evmos/releases/download/v19.2.0/evmos_19.2.0_Darwin_x86_64.tar.gz","linux/arm64":"https://github.com/evmos/evmos/releases/download/v19.2.0/evmos_19.2.0_Linux_arm64.tar.gz","linux/amd64":"https://github.com/evmos/evmos/releases/download/v19.2.0/evmos_19.2.0_Linux_amd64.tar.gz","windows/x86_64":"https://github.com/evmos/evmos/releases/download/v19.2.0/evmos_19.2.0_Windows_x86_64.zip"}}'`
)
