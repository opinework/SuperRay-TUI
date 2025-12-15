# SuperRay-TUI 安装指南

## 下载

从 `dist/` 目录下载对应平台的安装包：

| 平台 | 文件 |
|------|------|
| macOS Apple Silicon | `superray-tui-1.0.0-darwin-arm64.tar.gz` |
| macOS Intel | `superray-tui-1.0.0-darwin-amd64.tar.gz` |
| macOS 通用 | `superray-tui-1.0.0-darwin-universal.tar.gz` |
| Linux x86_64 | `superray-tui-1.0.0-linux-amd64.tar.gz` |
| Linux ARM64 | `superray-tui-1.0.0-linux-arm64.tar.gz` |
| Windows x64 | `superray-tui-1.0.0-windows-amd64.zip` |

## macOS 安装

### 方式一：解压直接运行

```bash
# 解压
tar -xzf superray-tui-1.0.0-darwin-arm64.tar.gz -C ~/Applications/superray-tui

# 运行
cd ~/Applications/superray-tui
./superray-tui
```

### 方式二：安装到系统目录

```bash
# 使用 Makefile 安装
make install

# 安装位置：
# - /usr/local/bin/superray-tui
# - /usr/local/share/superray/geoip/
```

### 卸载

```bash
make uninstall
```

## Linux 安装

```bash
# 解压
tar -xzf superray-tui-1.0.0-linux-amd64.tar.gz -C /opt/superray-tui

# 添加到 PATH (可选)
echo 'export PATH=$PATH:/opt/superray-tui' >> ~/.bashrc
source ~/.bashrc

# 运行
cd /opt/superray-tui
./superray-tui
```

## Windows 安装

1. 解压 `superray-tui-1.0.0-windows-amd64.zip`
2. 将解压后的文件夹移动到合适位置
3. 双击 `superray-tui.exe` 运行

## 目录说明

解压后的目录结构：

```
superray-tui/
├── superray-tui          # 主程序 (Windows 为 .exe)
├── lib/                  # 动态库目录
│   └── libsuperray.*     # .dylib (macOS) / .so (Linux)
├── geoip/                # GeoIP 数据
│   ├── geoip.dat
│   └── geosite.dat
└── .env.example          # 配置文件示例
```

**重要**：
- 可执行文件和 `lib/` 目录必须保持相对位置
- 运行时程序会自动从 `./lib/` 加载动态库
- Windows 版本的 `superray.dll` 在根目录下

## 验证安装

```bash
# 查看版本
./superray-tui --version

# 或直接运行查看界面
./superray-tui
```
