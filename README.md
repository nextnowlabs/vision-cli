# vision-cli

**`vg`** — 多后端图像、视频与语音合成命令行工具，对接阿里云 DashScope 和火山方舟（Ark / TTS），单二进制、零运行时依赖。

```bash
vg gen -p "一只在霓虹晚霞中冲浪的微型宇航员" --ar 16:9 --res 2K --model seedream
```

## 为什么用 vg

各家云厂商的图像/视频生成 API 请求格式、轮询协议各不相同。`vg` 统一了这些差异：

- **一条 `vg gen`** 覆盖 5 个图像模型，根据需要选择速度、价格或文本渲染能力。
- **一条 `vg video gen`** 搞定 Seedance 2.0 的提交、轮询、下载全流程。
- **一条 `vg tts gen`** 完成文本到语音的合成，支持多种音色、语速、编码格式。
- **一条 `vg tts voices`** 获取所有可用音色列表。
- **一份历史记录** 追踪所有后端的 prompt、模型和输出。
- **单 Go 二进制分发**，`go build` 即可，零运行时依赖。

## 安装

要求 Go 1.22+。

```bash
git clone https://github.com/nextnowlabs/vision-cli
cd vision-cli
go build -o vg ./cmd/vg/
```

将 `vg` 放到 PATH 中即可使用。

也可以通过 `go install` 安装（需要已发布版本）：

```bash
go install github.com/nextnowlabs/vision-cli/cmd/vg@latest
```

## 配置 API 密钥

`vg` 优先读取配置文件，其次读取环境变量。仅配置你实际使用的后端。

配置和数据保存在标准用户配置目录：

| 平台 | 路径 |
|---|---|
| Linux | `~/.config/vision-cli/` |
| macOS | `~/Library/Application Support/vision-cli/` |
| Windows | `%APPDATA%\vision-cli\` |

```bash
# DashScope 阿里云 (https://dashscope.console.aliyun.com)
vg config set dashscope_api_key <key>
# 或 export DASHSCOPE_API_KEY=<key>

# Volcengine Ark 火山方舟 (https://console.volcengine.com/ark)
vg config set ark_api_key <key>
# 或 export ARK_API_KEY=<key>
```

### 火山方舟需要额外开通 ⚠️

仅配置 API 密钥还不够，需在 [Ark 控制台](https://console.volcengine.com/ark) 完成以下操作：

1. **开通模型服务**：系统管理 → 开通管理 → 视觉大模型 → 开通 Doubao-Seedream（生成视频还需开通 Seedance）
2. **（可选）创建推理接入点**：在线推理 → 创建推理接入点 → 复制 `ep-xxx`，然后 `vg config set ark_endpoint_id ep-xxx`
3. **生成 API 密钥**：API Key 管理 → 创建

如果跳过步骤 1，首次调用会报 `InvalidEndpointOrModel.NotFound`。

### 语音合成 TTS

TTS 功能使用火山引擎豆包语音，需要两组配置：

```bash
# 语音合成（必需）
vg config set tts_appid <appid>
vg config set tts_cluster <cluster>        # 如 volcano_tts
vg config set tts_token <token>
# 或 export TTS_TOKEN=<token>

# 音色列表查询（可选，需火山引擎 AK/SK）
vg config set tts_access_key <access_key>
vg config set tts_secret_key <secret_key>
```

在 [火山引擎语音控制台](https://console.volcengine.com/tts) 获取上述参数。

## 模型

### 图像 — `vg gen --model <别名>`

| 别名 | 后端 | 模型 | 备注 |
|---|---|---|---|
| `seedream`（默认） | 火山方舟 Ark | `doubao-seedream-4-5-251128` | 支持 4K，最多 14 张参考图，文本渲染最强 |
| `seedream-lite` | 火山方舟 Ark | `doubao-seedream-5.0-lite` | 最新 5.0，更快更便宜 |
| `seedream-legacy` | 火山方舟 Ark | `doubao-seedream-4-0-250828` | 4.0 回退版本 |
| `wan-pro` | DashScope 阿里云 | `wan2.7-image-pro` | 文生图支持 4K，最多 9 张参考图 |
| `wan` | DashScope 阿里云 | `wan2.7-image` | 最高 2K，速度更快 |

### 视频 — `vg video gen --model <别名>`

| 别名 | 后端 | 模型 | 备注 |
|---|---|---|---|
| `seedance`（默认） | 火山方舟 Ark | `doubao-seedance-2-0-260128` | 高质量，约 2-3 分钟/条 |
| `seedance-fast` | 火山方舟 Ark | `doubao-seedance-2-0-fast-260128` | 约 30-60 秒/条，价格更低 |

## 快速开始

### 图像生成

```bash
# 默认 Seedream
vg gen -p "一只图书馆里的猫，水彩风格"

# 参考图 + 宽高比 + 分辨率
vg gen -p "把这张图变成日落时分" -i photo.jpg --ar 16:9 --res 2K -o sunset.png

# Seedream 4.5 — 适合带中文排版的海报
vg gen -p "促销海报，标题 春日上新，柔和粉彩" --model seedream --ar 3:4 --res 4K

# Wan 万相
vg gen -p "宁静的山水风景" --model wan-pro --ar 16:9 --res 4K

# 多参考图融合（Seedream 最多 14 张）
vg gen -p "服装来自图1，姿势来自图2" -i a.png -i b.png --model seedream

# 并发生成 4 张变体
vg gen -p "一间安静的书房" --repeat 4 -o studies.png
# 输出：studies_1.png, studies_2.png, studies_3.png, studies_4.png

# 从文件读取 prompt
vg gen -p @prompt.txt --model wan-pro --res 4K
```

### 视频生成（Seedance 2.0）

```bash
# 文生视频，5 秒，720p，无声（默认）
vg video gen -p "一只柴犬在樱花树下转身，慢动作" --ar 16:9

# 图生视频，以图片为首帧
vg video gen -p "镜头推近后环绕产品旋转" -i product.png --duration 8

# 电影级 1080p + 同步音频
vg video gen -p "咖啡师在晨光中制作浓缩咖啡" --res 1080p --duration 10 --audio

# 快速草稿
vg video gen -p "概念初稿" --model seedance-fast --duration 4

# 仅提交任务（不等结果），稍后手动查看
vg video gen -p "..." --poll=false
vg video status <task_id>
vg video download <task_id> -o final.mp4
```

⚠️ 生成的视频在完成后 24 小时内有效。使用 `--poll`（默认行为）会在完成时自动下载；使用 `--poll=false` 需在有效期内执行 `vg video download` 下载。

### 语音合成（Volcengine TTS）

```bash
# 基本合成
vg tts gen -p "你好，欢迎使用语音合成功能"

# 指定音色和编码
vg tts gen -p "春眠不觉晓，处处闻啼鸟" --voice-type BV700_streaming --encoding mp3

# 调整语速和音量
vg tts gen -p "这是一段快速播报的内容" --speed 1.5 --volume 1.2

# 指定输出文件
vg tts gen -p "静夜思" -o poem.mp3

# 查看可用音色列表（需 AK/SK 配置）
vg tts voices
```

支持的编码格式：`mp3`（默认）、`wav`、`pcm`、`ogg_opus`。
语速范围 0.2-3.0，音量和音高范围 0.1-3.0。

### 历史记录与统计

```bash
vg history                    # 最近 20 条
vg history -s "海报"           # 按 prompt 关键词搜索
vg history -n 50              # 显示最近 50 条
vg history <record_id>        # 查看单条完整 JSON
vg stats                      # 总量、成功/失败、按月/按天统计
```

## 命令参考

| 命令 | 说明 |
|---|---|
| `vg gen` | 根据文本 prompt 生成图像 |
| `vg video gen` | 提交视频生成任务（提交 + 轮询 + 下载） |
| `vg video status <id>` | 查询 Seedance 任务状态 |
| `vg video download <id>` | 下载已完成的 Seedance 视频 |
| `vg tts gen` | 文本转语音合成 |
| `vg tts voices` | 获取可用音色列表 |
| `vg history` | 浏览历史生成记录 |
| `vg config show` | 查看当前配置 |
| `vg config set <key> <value>` | 设置配置项 |
| `vg stats` | 查看用量统计 |

运行 `vg <command> --help` 查看完整选项。

## 可预设的默认值

```bash
vg config set default_model seedream-lite   # 不传 --model 时生效
vg config set default_ar 3:4
vg config set default_res 2K
vg config set output_dir ./out
vg config set poll_interval 20              # 视频轮询间隔（秒）
```

## 后端特性速览

- **参考图数量限制**：DashScope 最多 9 张，火山方舟最多 14 张
- **4K 分辨率**：seedream 全系支持，wan-pro 仅文生图支持；wan 4K 降级为 2K
- **负面提示词**：DashScope 原生支持，Seedream 追加到 prompt 末尾
- **火山方舟 Ark**：仅配 API 密钥不够 — 必须先在控制台开通对应模型服务

## 开发

```bash
git clone https://github.com/nextnowlabs/vision-cli
cd vision-cli

make build               # 编译（版本号自动从 git 获取）
VERSION=v0.1.0 make build  # 编译（指定版本号）
make test                # 运行测试
make lint                # 代码检查
make clean               # 清理构建产物
```

仅依赖 [cobra](https://github.com/spf13/cobra)，其余全部使用 Go 标准库。

## License

[MIT](LICENSE)
