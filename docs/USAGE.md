# SuperRay-TUI 使用说明

## 配置

### 环境变量

复制配置文件并修改：

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```bash
# 订阅地址 (必需)
SUPERRAY_SUB_URL=https://your-subscription-url

# 本地代理端口 (SOCKS5)
SUPERRAY_LOCAL_PORT=10808

# GeoIP 数据目录
SUPERRAY_GEO_PATH=./geoip

# 日志文件
ACCESS_LOG=access.log
ERROR_LOG=error.log

# 直连国家/地区 (逗号分隔，小写)
DIRECT_COUNTRIES=cn
```

### 代理端口

- SOCKS5: `SUPERRAY_LOCAL_PORT` (默认 10808)
- HTTP: `SUPERRAY_LOCAL_PORT + 1` (默认 10809)

## 启动

```bash
# 直接运行
./superray-tui

# 或指定配置
SUPERRAY_SUB_URL=https://xxx ./superray-tui
```

## 界面操作

### 主界面布局

```
┌────────────────────────────────────────────────────┐
│  服务器列表                                         │
│  ├─ [1] 香港节点 - 50ms                            │
│  ├─ [2] 日本节点 - 80ms                            │
│  └─ [3] 美国节点 - 150ms                           │
├────────────────────────────────────────────────────┤
│  状态: 已连接 | 流量: ↑1.2MB ↓5.6MB               │
│  当前节点: 香港节点                                 │
├────────────────────────────────────────────────────┤
│  [q]退出 [Enter]连接 [t]测速 [r]刷新 [?]帮助      │
└────────────────────────────────────────────────────┘
```

### 快捷键

| 按键 | 功能 |
|------|------|
| `↑` / `k` | 向上选择 |
| `↓` / `j` | 向下选择 |
| `Enter` / `c` | 连接选中服务器 |
| `d` | 断开当前连接 |
| `t` | 测试所有服务器延迟 (测试后自动排序) |
| `r` | 刷新订阅 |
| `s` | 设置订阅地址 |
| `u` | 切换 TUN/代理模式 |
| `f` | 刷新界面 |
| `q` | 退出程序 |

## 功能说明

### 1. 服务器列表

- 显示所有可用服务器
- 带延迟显示 (绿色 <100ms, 黄色 <300ms, 红色 >300ms)
- 当前连接的服务器会高亮显示

### 2. 延迟测试

按 `t` 测试所有服务器延迟：
- 并行测试所有节点
- 实时更新延迟数值
- 测试完成后自动排序

### 3. 订阅更新

按 `r` 刷新订阅：
- 从订阅地址获取最新服务器列表
- 自动解析各种格式 (V2Ray, Shadowsocks, Trojan 等)

### 4. 流量统计

- 实时显示上传/下载流量
- 显示当前连接状态
- 速度图表 (如果支持)

## TUN 模式

TUN 模式可以实现系统全局透明代理，无需配置系统代理设置。

### 启用 TUN 模式

TUN 模式需要管理员权限才能创建虚拟网卡。

**macOS**:
```bash
sudo ./superray-tui
```

**Linux**:
```bash
sudo ./superray-tui

# 或使用 capabilities 避免每次输入密码
sudo setcap cap_net_admin=eip ./superray-tui
./superray-tui
```

**Windows**:
- 右键点击 `superray-tui.exe`
- 选择「以管理员身份运行」

### TUN 模式特点

- 系统全局代理，所有流量自动走代理
- 无需配置系统代理或浏览器插件
- 支持 UDP 流量
- 基于 GeoIP 智能分流

### 注意事项

- TUN 模式会修改系统路由表
- 退出程序时会自动恢复网络设置
- 如异常退出导致网络问题，重启网络服务或重启电脑

## 代理模式 (非 TUN)

如果不使用 TUN 模式，需要手动配置系统或应用使用代理。

### 系统代理

**macOS**:
- 系统偏好设置 → 网络 → 高级 → 代理
- SOCKS 代理: 127.0.0.1:10808

**Linux**:
```bash
export http_proxy=http://127.0.0.1:10809
export https_proxy=http://127.0.0.1:10809
export all_proxy=socks5://127.0.0.1:10808
```

**Windows**:
- 设置 → 网络和 Internet → 代理
- 手动设置代理: 127.0.0.1:10809

### 浏览器代理

推荐使用 [SwitchyOmega](https://github.com/nicholascw/proxy-switchyomega-rules) 扩展管理代理。

### 终端代理

```bash
# 临时使用
all_proxy=socks5://127.0.0.1:10808 curl https://google.com

# Git 代理
git config --global http.proxy socks5://127.0.0.1:10808
```

## 故障排除

### 连接失败

1. 检查订阅地址是否正确
2. 检查网络连接
3. 查看 `error.log` 获取详细错误信息

### 延迟显示超时

- 服务器可能不可用
- 网络阻断
- 尝试其他服务器

### 动态库加载失败

确保 `lib/` 目录与可执行文件在同一目录下：
```
./superray-tui
./lib/libsuperray.dylib  # macOS
./lib/libsuperray.so     # Linux
```
