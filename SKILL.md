---
name: vision-cli
description: Multi-backend Image + Video + TTS CLI (vg command). Image: DashScope 阿里云百炼 (wan-pro=wan2.7-image-pro, wan=wan2.7-image), Volcengine Ark 火山方舟 字节 Seedream (seedream=doubao-seedream-4-5-251128, seedream-lite=doubao-seedream-5.0-lite, seedream-legacy=doubao-seedream-4-0-250828). Video: Volcengine Ark Seedance 2.0 (seedance=doubao-seedance-2-0-260128, seedance-fast=doubao-seedance-2-0-fast-260128). TTS: Volcengine TTS V3 火山引擎豆包语音 (tts gen). Use when user wants to generate images, videos, or speech, check generation history/stats, or manage API configuration. Triggers on "vg", "vision-cli", "生成图片", "出图", "万相", "wan", "百炼出图", "seedream", "字节出图", "豆包出图", "火山方舟", "ark", "生成视频", "出视频", "seedance", "视频生成", "语音合成", "文字转语音", "TTS", "tts", "音色", "voice".
---

# vision-cli — Multi-backend Image, Video & TTS CLI

CLI tool `vg` wraps two image generation backends. Pick the backend via `--model <alias>`:

| Alias | Backend | Model ID | Notes |
|-------|---------|----------|-------|
| `seedream` | Volcengine Ark 火山方舟 | `doubao-seedream-4-5-251128` | 字节 Seedream 4.5，4K，最多 14 张参考图，文字渲染极强 (default) |
| `seedream-lite` | Volcengine Ark 火山方舟 | `doubao-seedream-5.0-lite` | 字节 Seedream 5.0 Lite，最新 5.0 系列，更快更便宜 |
| `seedream-legacy` | Volcengine Ark 火山方舟 | `doubao-seedream-4-0-250828` | 字节 Seedream 4.0（已被 4.5 覆盖，留作回退） |
| `wan-pro` | DashScope 阿里云 | `wan2.7-image-pro` | 文生图支持 4K；编辑最高 2K；最多 9 张参考图 |
| `wan` | DashScope 阿里云 | `wan2.7-image` | 最高 2K，速度更快；最多 9 张参考图 |

**Backend-specific caveats:**
- 参考图数量上限：DashScope `wan*` 9、Seedream 14
- 4K：Seedream 全系、`wan-pro` 文生图都原生支持；`wan` 4K 会被降档到 2K
- `--negative-prompt`：DashScope 原生支持；Seedream 自动折叠进 prompt
- **Volcengine Ark / Seedream 接入门槛**：仅有 API key 不足以调用，必须先在控制台手动开通服务（见 Prerequisites）

## Prerequisites

- Install: `go install github.com/nextnowlabs/vision-cli/cmd/vg@latest` or clone and `go build -o vg ./cmd/vg/`
- Entry point: `vg`
- API keys:
  - DashScope 国内: `vg config set dashscope_api_key <key>` 或 `export DASHSCOPE_API_KEY=<key>`
  - Volcengine Ark 火山方舟（Seedream）：`vg config set ark_api_key <key>` 或 `export ARK_API_KEY=<key>`
  - Volcengine TTS 火山引擎豆包语音：详见 TTS 章节
- Data stored in standard user config directory:
  - Linux: `~/.config/vision-cli/`
  - macOS: `~/Library/Application Support/vision-cli/`
  - Windows: `%APPDATA%\vision-cli\`

### Volcengine Ark 接入前置（必做）

只配置 `ark_api_key` 不够。在 [火山方舟控制台](https://console.volcengine.com/ark) 完成：

1. **开通服务**：系统管理 → 开通管理 → 视觉大模型 → 找到 Doubao-Seedream（4.0 / 4.5 / 5.0-lite）→ 开通
2. **(可选) 创建推理接入点**：在线推理 → 创建推理接入点 → 选已开通的 Seedream 模型，得到 `ep-xxx` ID。复制后 `vg config set ark_endpoint_id ep-xxx`（设了就用 endpoint，不设就直接走 model name）
3. **API Key**：API Key 管理 → 创建 → `vg config set ark_api_key <key>`

## Commands

### 1. Generate — `vg gen`

```bash
# Basic (Seedream)
vg gen -p "描述"

# Reference image + aspect ratio + resolution
vg gen -p "描述" -i ref.png --ar 3:4 --res 2K -o output.png

# 万相 wan2.7-image-pro 4K 文生图
vg gen -p "描述" --model wan-pro --ar 16:9 --res 4K

# 万相带参考图做图像编辑（最多 9 张）
vg gen -p "把这只狗换成宇航员" -i dog.png --model wan --ar 1:1 --res 1K

# Seedream 4.5 文字渲染（海报/PPT/电商主图）
vg gen -p "中文海报，主标题『春日上新』，副标题小字" --model seedream --ar 3:4 --res 4K

# Seedream 多参考图融合（最多 14 张）
vg gen -p "保留参考图1的服装，融合参考图2的姿势" -i a.png -i b.png --model seedream

# Seedream 5.0 Lite，更快更便宜
vg gen -p "概念图" --model seedream-lite --ar 16:9 --res 2K

# DashScope 负向提示词
vg gen -p "描述" --model wan --negative-prompt "低分辨率, 畸形, 模糊"

# Prompt from file
vg gen -p @prompt.txt

# Repeat same prompt N times (1-8), outputs auto-numbered
vg gen -p "描述" --repeat 4 -o out.png
# → out_1.png, out_2.png, out_3.png, out_4.png

# Multiple reference images
vg gen -p "描述" -i a.png -i b.png -i c.png
```

**Options:**
| Flag | Description |
|------|-------------|
| `-p, --prompt` | Required. Text or `@file.txt` |
| `-o, --output` | Output path (default: auto timestamp) |
| `-i, --input` | Reference image, repeatable (上限：wan 系列 9, Seedream 14) |
| `--ar` | Aspect ratio: `1:1, 2:3, 3:2, 3:4, 4:3, 4:5, 5:4, 9:16, 16:9, 21:9` |
| `--res` | Resolution: `1K, 2K, 4K` (`wan` 4K 会降档到 2K；Seedream 4K 原生) |
| `--repeat` | Generate N copies, 1-8 |
| `--model` | `seedream, seedream-lite, seedream-legacy, wan-pro, wan` (default from config or seedream) |
| `--negative-prompt` | Negative prompt（DashScope 原生；Seedream 折叠进 prompt） |

**像素尺寸换算**：DashScope 和 Seedream 都把 `--ar` × `--res` 自动映射到像素 size（如 `--ar 16:9 --res 4K` → `4096*2304` / `4096x2304`），按模型能力封顶。

### 2. Video — `vg video` (Seedance 2.0 on Volcengine Ark)

**Video models:**
| Alias | Model ID | 定位 |
|---|---|---|
| `seedance` | `doubao-seedance-2-0-260128` | 字节 Seedance 2.0 标准版（高质量），约 2-3 分钟出片 |
| `seedance-fast` | `doubao-seedance-2-0-fast-260128` | Fast 版，便宜约 36%，30-60 秒出片 |

复用 `ark_api_key` —— 但需在控制台**额外开通 Seedance 视频服务**（开通管理 → 视频生成）。

```bash
# 文生视频（默认 seedance、720p、5s、无声）
vg video gen -p "一只柴犬在樱花树下慢镜头转身" --ar 16:9

# 图生视频（首帧/参考图）
vg video gen -p "镜头从产品平移到背景" -i product.png --duration 8

# 高质量 1080p + 同步音频 + 自动下载
vg video gen -p "电影级人物对白镜头" --model seedance --res 1080p --duration 10 --audio -o out.mp4

# Fast 版快速迭代
vg video gen -p "概念草图" --model seedance-fast --duration 4

# 异步：只提交，不等
vg video gen -p "..." --no-poll

# 查询任务状态
vg video status <task_id>

# 单独下载已完成的视频（task_id 必须 status=succeeded）
vg video download <task_id> -o out.mp4
```

**Options (`vg video gen`)：**
| Flag | Description |
|------|-------------|
| `-p, --prompt` | Required. Text or `@file.txt` |
| `-o, --output` | Output MP4 path (default: auto timestamp) |
| `-i, --input` | Reference image (repeatable，本地文件自动 base64) |
| `--ar` | `16:9, 9:16, 4:3, 3:4, 1:1, 21:9, adaptive` |
| `--res` | `480p, 720p, 1080p, 2K` (默认 720p) |
| `--duration` | 4-15 秒（默认 5） |
| `--audio / --no-audio` | 是否生成原生同步音频（默认关） |
| `--seed` | 随机种子 |
| `--model` | `seedance, seedance-fast`（默认 seedance） |
| `--poll / --no-poll` | 默认轮询并自动下载；`--no-poll` 仅提交 |

**重要约束**：生成的视频 URL **24 小时后过期**。`--poll` 模式会立即下载到本地。如果用 `--no-poll`，请尽快执行 `vg video download <task_id>`。

### 3. TTS — `vg tts` (Volcengine TTS V3 火山引擎豆包语音)

```bash
# 基本合成（默认 mp3）
vg tts gen -p "你好，欢迎使用语音合成"

# 指定音色、编码和语速
vg tts gen -p "春雨惊春清谷天" --voice-type zh_female_shuangkuaisisi_uranus_bigtts --encoding mp3 --speed 1.2

# 调整音量、情感
vg tts gen -p "播报内容" --volume 1.5 --emotion happy --emotion-scale 4

# 音高调整（半音，-12 到 12）
vg tts gen -p "低沉的声音" --pitch -3

# 指定输出文件路径
vg tts gen -p "静夜思" -o poem.mp3
```

**TTS 配置：**

```bash
# 语音合成必需
vg config set tts_api_key <api-key>          # 新版控制台的 API Key
vg config set tts_resource_id seed-tts-2.0   # 模型版本（默认 seed-tts-2.0）
# 或 export TTS_API_KEY=<api-key>
```

**Options (`vg tts gen`)：**
| Flag | Description |
|------|-------------|
| `-p, --prompt` | Required. Text to synthesize or `@file.txt` |
| `-o, --output` | Output audio path (default: auto timestamp) |
| `--voice-type` | Voice type / speaker ID (default: `zh_female_shuangkuaisisi_uranus_bigtts`) |
| `--encoding` | `mp3`, `pcm`, `ogg_opus` (default: `mp3`) |
| `--rate` | Sample rate: `8000`, `16000`, `22050`, `24000`, `32000`, `44100`, `48000` (default: `24000`) |
| `--speed` | Speech speed ratio 0.1-3.0 (default: `1.0`) |
| `--volume` | Volume ratio 0.1-3.0 (default: `1.0`) |
| `--pitch` | Pitch shift in semitones -12 to 12 (default: `0`) |
| `--language` | Language code, e.g. `zh-cn`, `en`, `ja` |
| `--emotion` | Emotion/style, e.g. `happy`, `sad` |
| `--emotion-scale` | Emotion intensity 1-5 (default: `4` when emotion is set) |

**TTS 配置参数获取**：前往 [火山引擎语音控制台](https://console.volcengine.com/speech/new) → API Key 管理 获取 API Key。

**内置音色参考**（全部为 2.0 音色，`_uranus_bigtts` 后缀）：

常用音色：`zh_female_vv_uranus_bigtts` (Vivi), `zh_female_xiaohe_uranus_bigtts` (小何), `zh_male_m191_uranus_bigtts` (云舟), `zh_male_taocheng_uranus_bigtts` (小天), `zh_female_shuangkuaisisi_uranus_bigtts` (爽快思思), `zh_male_liufei_uranus_bigtts` (刘飞), `zh_female_sophie_uranus_bigtts` (魅力苏菲), `zh_female_qingxinnvsheng_uranus_bigtts` (清新女声), `zh_female_linjianvhai_uranus_bigtts` (邻家女孩), `zh_male_wennuanahu_uranus_bigtts` (温暖阿虎), `zh_male_sunwukong_uranus_bigtts` (猴哥), `zh_male_silang_uranus_bigtts` (四郎), `zh_male_ruyaqingnian_uranus_bigtts` (儒雅青年), `zh_male_qingcang_uranus_bigtts` (擎苍), `zh_female_peiqi_uranus_bigtts` (佩奇猪), `zh_male_xionger_uranus_bigtts` (熊二), `zh_female_yingtaowanzi_uranus_bigtts` (樱桃丸子), `zh_male_naiqimengwa_uranus_bigtts` (奶气萌娃), `zh_female_popo_uranus_bigtts` (婆婆), `zh_female_gaolengyujie_uranus_bigtts` (高冷御姐), `zh_male_aojiaobazong_uranus_bigtts` (傲娇霸总), `zh_female_kefunvsheng_uranus_bigtts` (暖阳女声), `zh_female_jitangnv_uranus_bigtts` (鸡汤女), `zh_male_xuanyijieshuo_uranus_bigtts` (悬疑解说), `zh_male_baqiqingshu_uranus_bigtts` (霸气青叔), `en_male_tim_uranus_bigtts` (Tim), `en_female_dacey_uranus_bigtts` (Dacey), `en_female_stokie_uranus_bigtts` (Stokie)。

完整列表见 https://www.volcengine.com/docs/6561/1257544

### 3. History — `vg history`

```bash
vg history                    # Recent 20 records
vg history -n 50              # Last 50
vg history -s "keyword"       # Search by prompt
vg history <record_id>        # Full detail (JSON)
```

History records include a `backend` field (`dashscope` / `volcengine_ark`) alongside the resolved `model` id.

### 4. Config — `vg config`

```bash
vg config show
vg config set dashscope_api_key <key>  # DashScope 国内
vg config set ark_api_key <key>        # Volcengine Ark 火山方舟
vg config set ark_endpoint_id ep-xxx   # 可选，传了就用 endpoint
vg config set tts_api_key <key>           # TTS 语音合成 API Key
vg config set tts_resource_id <id>         # TTS 资源 ID（默认 seed-tts-2.0）
vg config set default_ar 3:4
vg config set default_res 2K
vg config set default_model seedream-lite  # seedream | seedream-lite | seedream-legacy | wan-pro | wan
vg config set poll_interval 20             # Poll seconds
vg config set output_dir ./out             # Default output directory
```

API key priority:
- DashScope: config `dashscope_api_key` > `DASHSCOPE_API_KEY` env var
- Volcengine Ark: config `ark_api_key` > `ARK_API_KEY` env var
- Volcengine TTS: config `tts_api_key` > `TTS_API_KEY` env var

### 5. Stats — `vg stats`

```bash
vg stats    # Total calls, success/fail, direct/batch, monthly/daily breakdown
```

## When helping the user

- **默认模型**：用户说"出一张图" / "生成图片"，默认 Seedream `seedream`
- **国内场景 / 不能翻墙**：用 DashScope (`wan` / `wan-pro`) 或 Volcengine Ark (`seedream*`)
- **中文文字渲染 / 海报 / PPT / 电商主图**：首选 `seedream`（Seedream 4.5 文字最强）
- **多参考图融合（>9 张）**：只能用 `seedream` (≤14)
- **4K 输出**：Seedream 全系、`wan-pro` 文生图都原生；`wan` 会降档
- **追新 / 想试 5.0**：用 `seedream-lite`（5.0 系列目前公开只有 lite 版）
- **Volcengine 报错 InvalidEndpointOrModel.NotFound**：99% 是没在控制台开通对应模型服务，去开通管理里加一下
- 相同 prompt 要多张用 `--repeat`
- Always confirm aspect ratio and resolution before generating
- For xhs-generator workflow, default to `--ar 3:4 --res 2K`
- **视频生成**：用户说"出视频" / "生成视频"时走 `vg video gen`（Seedance 2.0），默认 `seedance` + 720p + 5s + 无音频
- 视频迭代/草稿用 `seedance-fast`，最终成片用 `seedance`
- 视频结果 24 小时过期，默认 `--poll` 会自动下载；`--no-poll` 时务必提醒用户尽快 `vg video download`
- **语音合成**：用户说"文字转语音" / "语音合成" / "TTS" / "朗读"时走 `vg tts gen`，默认 mp3 + 2.0 音色
- 用户想查看可用音色时说"看看有哪些音色" / "音色列表"，展示内置 2.0 音色参考表
- TTS 合成需要 `tts_api_key` 配置
