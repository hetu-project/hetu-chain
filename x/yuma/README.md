# Yuma Module

Yuma 是一个简化的共识模块，专注于实现 Bittensor 的 epoch 算法。该模块完全基于 event 模块的数据，为每个子网提供独立的参数配置和奖励分配。

## 概述

Yuma 模块实现了 Bittensor 网络的核心共识机制，包括：

- **加权中位数共识**：基于验证者权重和质押的共识计算
- **EMA Bonds 更新**：历史权重的指数移动平均
- **权重裁剪**：防止异常权重的机制
- **奖励分配**：基于共识分数的 token 分配

## 核心特性

### 1. 完全依赖 Event 模块
- 所有数据都从 event 模块获取
- 子网参数通过 `Subnet.Params` 配置
- 验证者信息和权重从 event 模块读取

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
  "bonds_moving_average": "0.9"      // bonds 移动平均
}
```

### 3. 简化的存储
- 最小化状态存储
- 主要存储 bonds 历史数据
- 简单的 epoch 时间记录

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
   - 更新 EMA bonds

6. **分配奖励**
   - 计算分红（dividends）
   - 归一化分红
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
- bonds 矩阵
- 共识分数

### 辅助函数

- `weightedMedianCol`: 加权中位数计算
- `clipWeights`: 权重裁剪
- `computeBonds`: EMA bonds 更新
- `computeDividends`: 分红计算
- `distributeEmission`: 奖励分配

## 模块结构

```
x/yuma/
├── types/
│   ├── epoch.go          # epoch 相关类型定义
│   ├── interfaces.go     # 接口定义
│   └── module.go         # 模块基本类型
├── keeper/
│   ├── keeper.go         # 简化的 keeper
│   └── epoch.go          # 核心算法实现
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
```

### 配置子网参数
```go
// 在 event 模块中设置子网参数
subnet.Params = map[string]string{
    "kappa": "0.5",
    "alpha": "0.1",
    "delta": "1.0",
    "tempo": "100",
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

## 开发计划

- [ ] 添加更复杂的活跃性检查
- [ ] 实现 bonds 持久化存储
- [ ] 添加更多的参数验证
- [ ] 优化算法性能
- [ ] 添加测试覆盖

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个模块。