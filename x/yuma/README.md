# Yuma 共识模块

## 概述

Yuma 是基于 Bittensor 的去中心化机器学习共识机制的 Cosmos SDK 模块实现。该模块实现了完整的 Yuma 共识算法，包括神经元网络、权重设置、bonds 计算、以及激励分配等核心功能。

## 核心特性

### 1. Yuma 共识算法
- **Epoch 运行**: 定期执行共识计算，更新神经元状态
- **权重验证**: 验证神经元权重设置的合法性
- **Bonds 计算**: 动态计算神经元间的 bonds 矩阵
- **激励分配**: 基于共识结果分配激励和分红

### 2. 参数管理
- **活跃性检查**: 使用 `ActivityCutoff` 参数判断神经元活跃状态
- **免疫期保护**: 使用 `ImmunityPeriod` 保护新注册的神经元
- **权重限制**: 通过 `MaxWeightsLimit` 和 `MinAllowedWeights` 限制权重数量
- **验证器管理**: 使用 `MaxAllowedValidators` 等参数管理验证器

### 3. 存储管理
- **子网信息**: 存储和管理子网配置
- **神经元状态**: 维护神经元的完整状态信息
- **权重数据**: 存储神经元间的权重关系
- **Bonds 数据**: 存储计算得出的 bonds 矩阵

## 模块结构

```
src/yuma/
├── types/
│   ├── keys.go          # 存储键定义
│   ├── params.go        # 参数定义和验证
│   ├── subnet.go        # 子网和神经元数据结构
│   └── genesis.go       # 创世状态定义
├── keeper/
│   ├── keeper.go        # 核心 Keeper 实现
│   ├── epoch.go         # Epoch 运行逻辑
│   ├── bonds.go         # Bonds 计算逻辑
│   └── validator.go     # 验证器管理逻辑
├── module.go            # 模块定义和生命周期
└── README.md           # 本文档
```

## 核心参数

### 共识相关参数
- `Rho`: 控制 bonds 稀疏性的参数
- `Kappa`: 控制 bonds 饱和性的参数
- `AdjustmentInterval`: Epoch 执行间隔
- `ActivityCutoff`: 神经元活跃性判断阈值
- `ImmunityPeriod`: 新神经元免疫期长度

### 权重相关参数
- `MaxWeightsLimit`: 单个神经元最大权重数量
- `MinAllowedWeights`: 单个神经元最小权重数量
- `WeightsSetRateLimit`: 权重设置频率限制

### 验证器相关参数
- `MaxAllowedValidators`: 最大验证器数量
- `ValidatorSequenceLength`: 验证器序列长度
- `ValidatorEpochLength`: 验证器 epoch 长度
- `ValidatorLogitsDivergence`: 验证器 logits 分歧阈值

### Alpha 机制参数
- `LiquidAlphaEnabled`: 是否启用动态 alpha
- `AlphaHigh`: 高 alpha 阈值
- `AlphaLow`: 低 alpha 阈值
- `BondsPenalty`: Bonds 惩罚系数

## 核心算法

### 1. Epoch 执行流程

```
1. 检查执行条件 (AdjustmentInterval)
2. 获取神经元信息并应用 MaxAllowedUids 限制
3. 收集和验证权重 (MaxWeightsLimit, MinAllowedWeights)
4. 计算活跃性 (ActivityCutoff)
5. 应用免疫期保护 (ImmunityPeriod)
6. 更新 bonds 矩阵
7. 执行 Yuma 共识计算:
   - 计算 consensus 分数
   - 计算 trust 分数
   - 计算 ranks 分数
   - 计算 incentives 分数
   - 计算 dividends 分数
   - 计算最终 emission
8. 更新神经元状态
9. 更新子网统计信息
```

### 2. Bonds 更新算法

```
1. 初始化 bonds 矩阵
2. 计算激励变化 (incentive variance)
3. 计算 alpha 值 (liquid alpha 机制)
4. 应用 bonds 更新公式: B_new = α * W + (1-α) * B_old
5. 归一化 bonds 行
6. 应用 bonds 惩罚
7. 应用 rho 和 kappa 变换
```

### 3. 验证器管理

```
1. 基于 trust 分数选择验证器
2. 验证验证器序列长度
3. 计算 logits 分歧
4. 应用验证器重置机制
```

## 与 Event 模块集成

Yuma 模块通过 Event 模块获取外部数据:
- 神经元注册信息通过合约完成，Event 模块监听相关事件
- 子网参数更新通过 Event 模块同步
- 支持从 Event 模块更新子网配置

## 使用方法

### 1. 模块初始化

在 `app.go` 中添加 Yuma 模块:

```go
import "github.com/evmos/evmos/v14/x/yuma"

// 在 NewApp 函数中
app.YumaKeeper = yumakeeper.NewKeeper(
    appCodec,
    keys[yumatypes.StoreKey],
    keys[yumatypes.MemStoreKey],
    app.GetSubspace(yumatypes.ModuleName),
    app.BankKeeper,
    app.StakingKeeper,
    app.EventKeeper,
)

yumaModule := yuma.NewAppModule(appCodec, app.YumaKeeper, app.AccountKeeper, app.BankKeeper)
```

### 2. 创世状态配置

在创世文件中配置初始参数:

```json
{
  "yuma": {
    "params": {
      "rho": "10",
      "kappa": "32767",
      "max_allowed_uids": 4096,
      "immunity_period": 4096,
      "activity_cutoff": 5000,
      "adjustment_interval": 112
    },
    "subnets": [],
    "neurons": []
  }
}
```

### 3. 运行时操作

模块会在每个区块的 `BeginBlock` 阶段自动检查是否需要运行 epoch:

```go
func (am AppModule) BeginBlock(ctx sdk.Context, _ abci.RequestBeginBlock) {
    am.runEpochCheck(ctx)
}
```

## 关键实现细节

### 1. 数值计算精度
- 使用 `sdk.Dec` 进行高精度十进制计算
- 权重和 bonds 数据存储为 `uint16` 以节省空间
- 提供归一化函数确保数值稳定性

### 2. 存储优化
- 使用前缀键管理不同类型的数据
- 支持按子网分组存储神经元信息
- 提供迭代器支持批量操作

### 3. 参数验证
- 所有参数都有对应的验证函数
- 支持运行时参数更新
- 确保参数值在合理范围内

## 未来扩展

1. **性能优化**: 大规模网络的计算优化
2. **存储优化**: 历史数据压缩和归档
3. **监控集成**: 添加更多监控指标
4. **治理集成**: 支持链上治理参数调整

## 注意事项

1. **计算复杂度**: Epoch 运行的计算复杂度为 O(n²)，其中 n 是神经元数量
2. **存储空间**: bonds 矩阵需要 O(n²) 的存储空间
3. **参数调优**: 需要根据实际网络规模调整各项参数
4. **兼容性**: 与 Event 模块的集成需要确保事件格式兼容