# Yuma Module

Yuma 是一个实现 Bittensor epoch 算法的共识模块，专注于实现完整的 Bittensor 网络共识机制。该模块完全基于 event 模块的数据，为每个子网提供独立的参数配置和奖励分配。

## 概述

Yuma 模块实现了 Bittensor 网络的核心共识机制，包括：

- **加权中位数共识**：基于验证者权重和质押的共识计算
- **动态 Alpha 计算**：支持 LiquidAlpha 的动态权重更新
- **权重裁剪**：防止异常权重的机制
- **激励计算**：基于 rho 参数的激励分配
- **奖励分配**：结合 incentive 和 dividends 的 token 分配

## 核心特性

### 1. 完整的 Bittensor Epoch 算法
- 支持动态 alpha 计算（LiquidAlpha）
- 支持固定 alpha 计算（传统模式）
- 实现激励机制（incentive）
- 实现分红机制（dividends）

### 2. 每个子网独立参数
每个子网可以配置自己的参数：
```json
{
  "kappa": "0.5",                    // 多数阈值
  "alpha": "0.1",                    // EMA 参数
  "delta": "1.0",                    // 权重裁剪范围
  "activity_cutoff": "5000",         // 活跃截止时间
  "immunity_period": "4096",         // 免疫期
  "max_weights_limit": "1000",       // 最大权重数量
  "min_allowed_weights": "8",        // 最小权重数量
  "weights_set_rate_limit": "100",   // 权重设置速率限制
  "tempo": "100",                    // epoch 运行频率
  "bonds_penalty": "0.1",            // bonds 惩罚
  "bonds_moving_average": "0.9",     // bonds 移动平均
  "rho": "0.5",                      // 激励参数
  "liquid_alpha_enabled": "false",   // 是否启用动态 alpha
  "alpha_sigmoid_steepness": "10.0", // sigmoid 陡峭度
  "alpha_low": "0.01",               // alpha 下限
  "alpha_high": "0.99"               // alpha 上限
}
```

### 3. 动态 Alpha 计算
当启用 `liquid_alpha_enabled` 时：
- 使用 sigmoid 函数计算动态 alpha
- 基于权重与共识的差异调整更新速度
- 支持买入/卖出差异的智能调整

### 4. 激励机制
- 使用 rho 参数控制激励强度
- 结合权重矩阵和质押权重计算激励
- 与分红机制结合进行最终分配

## 算法流程

### Epoch 执行步骤

1. **获取子网数据**
   - 从 event 模块获取子网信息
   - 解析子网参数

2. **检查执行条件**
   - 验证是否到达执行时间（tempo）
   - 检查子网是否存在

3. **收集验证者数据**
   - 获取所有验证者的质押信息
   - 获取权重矩阵

4. **计算活跃状态**
   - 基于活跃截止时间判断
   - 应用免疫期保护

5. **执行核心算法**
   - 归一化质押权重
   - 计算加权中位数共识
   - 裁剪异常权重
   - 计算动态或固定 alpha
   - 更新 EMA bonds

6. **计算激励和分红**
   - 计算激励（incentive）
   - 计算分红（dividends）
   - 归一化分配

7. **分配奖励**
   - 结合激励和分红
   - 分配 emission

## 核心函数

### RunEpoch
```go
func (k Keeper) RunEpoch(ctx sdk.Context, netuid uint16, raoEmission uint64) (*types.EpochResult, error)
```

主要的 epoch 执行函数，返回：
- 账户地址列表
- 每个账户的 emission 分配
- 每个账户的 dividend 分配
- 每个账户的 incentive 分配
- bonds 矩阵
- 共识分数

### 新增的辅助函数

- `computeLiquidAlphaValues`: 动态 alpha 矩阵计算
- `alphaSigmoid`: sigmoid alpha 计算
- `computeBondsWithDynamicAlpha`: 动态 alpha bonds 更新
- `computeIncentive`: 激励计算
- `matMul`: 矩阵乘法
- `normalize`: 数组归一化
- `clamp`: 值范围限制

## 模块结构

```
x/yuma/
├── types/
│   ├── epoch.go          # epoch 相关类型定义
│   ├── interfaces.go     # 接口定义
│   └── module.go         # 模块基本类型
├── keeper/
│   ├── keeper.go         # 简化的 keeper
│   ├── epoch.go          # 核心算法实现
│   └── epoch_test.go     # 测试文件
├── module.go             # 模块定义
├── abci.go              # ABCI 接口
└── README.md            # 本文档
```

## 使用示例

### 运行 Epoch
```go
// 运行子网 1 的 epoch，分配 1M rao
result, err := keeper.RunEpoch(ctx, 1, 1000000)
if err != nil {
    // 处理错误
}

// 查看结果
fmt.Printf("Epoch result: %+v\n", result)
fmt.Printf("Incentive: %v\n", result.Incentive)
fmt.Printf("Dividend: %v\n", result.Dividend)
```

### 配置子网参数
```go
// 在 event 模块中设置子网参数
subnet.Params = map[string]string{
    "kappa": "0.5",
    "alpha": "0.1",
    "delta": "1.0",
    "tempo": "100",
    "rho": "0.5",
    "liquid_alpha_enabled": "true",
    "alpha_sigmoid_steepness": "10.0",
}
```

## 依赖关系

- **Event 模块**: 提供所有数据源
  - 子网信息 (`Subnet`)
  - 验证者质押 (`ValidatorStake`)
  - 权重矩阵 (`ValidatorWeight`)

## 注意事项

1. **参数验证**: 所有参数都有合理的默认值和验证
2. **错误处理**: 模块会优雅地处理各种错误情况
3. **性能优化**: 算法经过优化，支持大规模验证者网络
4. **可扩展性**: 模块设计支持未来功能扩展
5. **动态 Alpha**: 需要谨慎配置 sigmoid 参数以获得最佳效果

## 开发计划

- [ ] 实现真正的活跃性检查逻辑
- [ ] 添加 bonds 持久化存储
- [ ] 实现更多的参数验证
- [ ] 优化算法性能
- [ ] 添加更完整的测试覆盖
- [ ] 实现 RPC 查询接口

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个模块。