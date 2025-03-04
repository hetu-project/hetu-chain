package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)


const EpochWindows int64 = 500

func TestFirstBlockInEpoch(t *testing.T) {
	require.True(t, FirstBlockInEpoch(1, EpochWindows))
	require.True(t, FirstBlockInEpoch(501, EpochWindows))
	require.True(t, FirstBlockInEpoch(1001, EpochWindows))
	require.False(t, FirstBlockInEpoch(2, EpochWindows))
	require.False(t, FirstBlockInEpoch(500, EpochWindows))
	require.False(t, FirstBlockInEpoch(1000, EpochWindows))
}

func TestLastBlockInEpoch(t *testing.T) {
	require.True(t, LastBlockInEpoch(500, EpochWindows))
	require.True(t, LastBlockInEpoch(1000, EpochWindows))
	require.True(t, LastBlockInEpoch(1500, EpochWindows))
	require.False(t, LastBlockInEpoch(1, EpochWindows))
	require.False(t, LastBlockInEpoch(501, EpochWindows))
	require.False(t, LastBlockInEpoch(1001, EpochWindows))
}

func TestCurrentEpochNumber(t *testing.T) {
	require.Equal(t, uint64(1), CurrentEpochNumber(1, EpochWindows))
	require.Equal(t, uint64(1), CurrentEpochNumber(500, EpochWindows))
	require.Equal(t, uint64(2), CurrentEpochNumber(501, EpochWindows))
	require.Equal(t, uint64(2), CurrentEpochNumber(1000, EpochWindows))
	require.Equal(t, uint64(3), CurrentEpochNumber(1001, EpochWindows))
	require.Equal(t, uint64(3), CurrentEpochNumber(1500, EpochWindows))
	require.Equal(t, uint64(4), CurrentEpochNumber(1501, EpochWindows))
}