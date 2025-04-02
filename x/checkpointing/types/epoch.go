package types

// FirstBlockInEpoch checks if the given height is the first block in the epoch
func FirstBlockInEpoch(height int64, EpochWindows int64) bool {
	return (height-1)%EpochWindows == 0
}

// LastBlockInEpoch checks if the given height is the last block in the epoch
func LastBlockInEpoch(height int64, EpochWindows int64) bool {
	return height%EpochWindows == 0
}

// CurrentEpochNumber returns the current epoch number for the given height and EpochWindows
func CurrentEpochNumber(height int64, EpochWindows int64) uint64 {
	return uint64((height-1)/EpochWindows + 1)
}
