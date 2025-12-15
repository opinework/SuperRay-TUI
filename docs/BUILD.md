# SuperRay-TUI 编译指南

## 环境要求

### 必需
- Go 1.21 或更高版本
- CGO 支持 (CGO_ENABLED=1)

### 交叉编译 (可选)
- [Zig](https://ziglang.org/) - 用于 Linux/Windows 交叉编译

## 目录结构

```
SuperRay-TUI/
├── main.go                    # 主程序
├── pkg/superray/              # CGO 绑定
├── third_party/superray/      # SuperRay 库
│   ├── include/               # 头文件
│   ├── lib/                   # 动态库
│   │   ├── macos/             # macOS (.dylib)
│   │   ├── linux/             # Linux (.so)
│   │   └── windows/           # Windows (.dll)
│   └── geoip/                 # GeoIP 数据
├── build/                     # 编译输出
└── dist/                      # 打包输出
```

## 编译命令

### 快速编译 (当前平台)

```bash
make build
# 或
make native
```

### macOS 编译

```bash
# Apple Silicon (M1/M2/M3)
make darwin-arm64

# Intel
make darwin-amd64

# Universal (同时支持两种架构)
make darwin-universal

# 所有 macOS 版本
make darwin-all
```

### Linux 编译 (需要 zig)

```bash
# x86_64
make linux-amd64

# ARM64
make linux-arm64
```

### Windows 编译 (需要 zig)

```bash
make windows-amd64
```

### 全平台编译

```bash
make package-all
```

## 编译输出

编译后的文件位于 `build/<平台>/` 目录：

```
build/darwin-arm64/
├── superray-tui           # 可执行文件
├── lib/
│   └── libsuperray.dylib  # 动态库
├── geoip/
│   ├── geoip.dat          # GeoIP 数据
│   └── geosite.dat
└── .env.example           # 配置示例
```

## 打包命令

```bash
# 打包当前平台
make package

# 打包 macOS
make package-darwin

# 打包全平台
make package-all
```

打包文件位于 `dist/` 目录：
- `superray-tui-1.0.0-darwin-arm64.tar.gz`
- `superray-tui-1.0.0-darwin-amd64.tar.gz`
- `superray-tui-1.0.0-darwin-universal.tar.gz`
- `superray-tui-1.0.0-linux-amd64.tar.gz`
- `superray-tui-1.0.0-linux-arm64.tar.gz`
- `superray-tui-1.0.0-windows-amd64.zip`

## 其他命令

```bash
# 清理编译文件
make clean

# 查看版本
make version

# 检查二进制依赖
make check

# 查看帮助
make help
```

## 常见问题

### Q: 交叉编译报错找不到编译器
A: 安装 zig: `brew install zig`

### Q: macOS 编译出现 duplicate -rpath 警告
A: 这是 Go 自动添加的 rpath 导致的，不影响使用。

### Q: 找不到 libsuperray
A: 确保 `third_party/superray/lib/` 下有对应平台的库文件。
