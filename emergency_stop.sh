#!/bin/bash

echo "=== 紧急停止 Hetu 网络 ==="
echo "正在停止所有节点..."

# 停止本地节点
pkill hetud || true

# 停止远程节点（如果有的话）
# 请根据你的实际部署情况修改 IP 地址
for ip in "1.2.3.4" "5.6.7.8" "9.10.11.12" "13.14.15.16" "17.18.19.20"; do
    echo "停止节点 $ip..."
    ssh root@$ip "pkill hetud || true" 2>/dev/null || echo "无法连接到 $ip"
done

echo "所有节点已停止"
echo "请立即检查节点状态确认停止成功" 