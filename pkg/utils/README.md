# TLS 工具包

提供 TLS 证书生成和管理工具函数，位于 `pkg/utils/tls.go`。

## 功能

- **GenerateSelfSignedCert** — 生成自签名 TLS 证书（开发/测试用）

## 使用方法

```go
import "terminalog/pkg/utils"

// 生成自签名证书到指定路径
err := utils.GenerateSelfSignedCert("resources/https.crt", "resources/https.key", "localhost")
if err != nil {
    log.Fatal(err)
}
```

## 设计决策

| 决策 | 原因 |
|------|------|
| 使用 ECDSA P-256 | 比 RSA 更快、密钥更短，现代浏览器/TLS 库广泛支持 |
| O_EXCL 创建文件 | 防止意外覆盖已有证书 |
| 私钥权限 0600 | 遵循最小权限原则 |
| 默认 365 天有效期 | 与 openssl 自签名证书惯例一致 |
| 包含 127.0.0.1 和 ::1 | 支持本地 localhost 开发 |
| 主机名默认 localhost | 最常见的开发场景 |

## 注意事项

- ⚠️ 自签名证书**仅用于开发/测试**，切勿用于生产环境
- 浏览器会显示安全警告，这是预期行为
- 生产环境请使用 Let's Encrypt 或 CA 签发的证书
