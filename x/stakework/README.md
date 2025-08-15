# Stakework Module

Stakework is a consensus module that implements the Bittensor epoch algorithm, focusing on implementing the complete Bittensor network consensus mechanism. This module is entirely based on data from the event module, providing independent parameter configuration and reward distribution for each subnet.

## Overview

The Stakework module implements the core consensus mechanism of the Bittensor network, including:

- **Epoch Algorithm**: Complete implementation of Bittensor epoch calculation logic
- **Weight Management**: Handles weight allocation and updates between validators
- **Consensus Calculation**: Consensus mechanism based on weighted median
- **Reward Distribution**: Distributes network rewards based on consensus results
- **Bonds System**: Implements exponential moving average of historical weights
- **Dynamic Alpha**: Supports dynamically adjustable EMA parameters

## Core Features

### 1. Epoch Execution Mechanism

The module checks whether an epoch needs to be run at each block based on the subnet's `tempo` parameter:

```go
func (k Keeper) shouldRunEpoch(ctx sdk.Context, netuid uint16, tempo uint64) bool {
    lastEpoch := k.getLastEpochBlock(ctx, netuid)
    currentBlock := uint64(ctx.BlockHeight())
    return (currentBlock - lastEpoch) >= tempo
}
```

### 2. Weight Normalization

Ensures all weights sum to 1:

```go
func (k Keeper) normalize(values []float64) []float64 {
    sum := 0.0
    for _, v := range values {
        sum += v
    }
    
    if sum == 0 {
        return make([]float64, len(values))
    }
    
    result := make([]float64, len(values))
    for i, v := range values {
        result[i] = v / sum
    }
    return result
}
```

### 3. Consensus Calculation

Uses weighted median to calculate consensus scores:

```go
func (k Keeper) weightedMedianCol(stake []float64, weights [][]float64, kappa float64) []float64 {
    // Implements weighted median algorithm
}
```

### 4. Bonds Calculation

Supports EMA calculation with both fixed and dynamic alpha:

```go
func (k Keeper) computeBonds(clippedWeights [][]float64, prevBonds [][]float64, alpha float64) [][]float64 {
    // Implements EMA calculation
}
```

## Parameter Configuration

Each subnet can independently configure the following parameters:

- **kappa**: Majority threshold (default 0.5)
- **alpha**: EMA parameter (default 0.1-0.9)
- **delta**: Weight clipping range (default 1.0)
- **tempo**: Epoch execution frequency
- **rho**: Incentive parameter
- **liquid_alpha_enabled**: Whether to enable dynamic alpha

## Data Structures

### EpochResult

```go
type EpochResult struct {
    Netuid    uint16      `json:"netuid"`
    Accounts  []string    `json:"accounts"`
    Emission  []uint64    `json:"emission"`
    Dividend  []uint64    `json:"dividend"`
    Incentive []uint64    `json:"incentive"`
    Bonds     [][]float64 `json:"bonds"`
    Consensus []float64   `json:"consensus"`
}
```

### EpochParams

```go
type EpochParams struct {
    Kappa                   float64 `json:"kappa"`
    Alpha                   float64 `json:"alpha"`
    Delta                   float64 `json:"delta"`
    ActivityCutoff          uint64  `json:"activity_cutoff"`
    ImmunityPeriod          uint64  `json:"immunity_period"`
    MaxWeightsLimit         uint64  `json:"max_weights_limit"`
    MinAllowedWeights       uint64  `json:"min_allowed_weights"`
    WeightsSetRateLimit     uint64  `json:"weights_set_rate_limit"`
    Tempo                   uint64  `json:"tempo"`
    BondsPenalty            float64 `json:"bonds_penalty"`
    BondsMovingAverage      float64 `json:"bonds_moving_average"`
    Rho                     float64 `json:"rho"`
    LiquidAlphaEnabled      bool    `json:"liquid_alpha_enabled"`
    AlphaSigmoidSteepness   float64 `json:"alpha_sigmoid_steepness"`
    AlphaLow                float64 `json:"alpha_low"`
    AlphaHigh               float64 `json:"alpha_high"`
}
```

## Usage

### 1. Run Epoch

```go
result, err := k.RunEpoch(ctx, netuid, ahetuEmission)
if err != nil {
    return err
}
```

### 2. Get Subnet Validators

```go
validators := k.getSubnetValidators(ctx, netuid)
```

### 3. Calculate Active Status

```go
active := k.calculateActive(ctx, netuid, validators, params)
```

## Dependencies

- **Event Module**: Provides subnet and validator data
- **Bank Module**: Handles token transfers and balance queries
- **Staking Module**: Gets staking-related information

## Configuration Example

```json
{
  "kappa": 0.5,
  "alpha": 0.1,
  "delta": 1.0,
  "activity_cutoff": 1000,
  "immunity_period": 100,
  "max_weights_limit": 100,
  "min_allowed_weights": 10,
  "weights_set_rate_limit": 1,
  "tempo": 100,
  "bonds_penalty": 0.1,
  "bonds_moving_average": 0.1,
  "rho": 0.1,
  "liquid_alpha_enabled": false,
  "alpha_sigmoid_steepness": 10.0,
  "alpha_low": 0.01,
  "alpha_high": 0.99
}
```

## Testing

The module includes comprehensive unit tests covering core algorithms and edge cases:

```bash
go test ./x/stakework/...
```

## Contributing

Issues and Pull Requests are welcome to improve this module.

## License

This project is licensed under the MIT License.