# vision-cli

**`vg`** — 多后端图像、视频与语音合成命令行工具，对接火山方舟（Ark / TTS），单二进制、零运行时依赖。

```bash
vg gen -p "一只在霓虹晚霞中冲浪的微型宇航员" --ar 16:9 --res 2K --model seedream
```

## 为什么用 vg

各家云厂商的图像/视频生成 API 请求格式、轮询协议各不相同。`vg` 统一了这些差异：

- **一条 `vg gen`** 覆盖 3 个图像模型，根据需要选择速度、价格或文本渲染能力。
- **一条 `vg video gen`** 搞定 Seedance 2.0 的提交、轮询、下载全流程。
- **一条 `vg tts gen`** 完成文本到语音的合成，支持多种音色、语速、编码格式。
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

TTS 功能使用火山引擎豆包语音大模型 V3 接口：

```bash
# 语音合成（必需）
vg config set tts_api_key <api-key>          # 新版控制台的 API Key
vg config set tts_resource_id seed-tts-2.0   # 模型版本（默认 seed-tts-2.0）
# 或 export TTS_API_KEY=<api-key>
```

在 [火山引擎语音控制台](https://console.volcengine.com/speech/new) → API Key 管理 获取 API Key，在 [资源管理](https://console.volcengine.com/speech/new) 开通语音合成资源包。

## 模型

### 图像 — `vg gen --model <别名>`

| 别名 | 后端 | 模型 | 备注 |
|---|---|---|---|
| `seedream`（默认） | 火山方舟 Ark | `doubao-seedream-4-5-251128` | 支持 4K，最多 14 张参考图，文本渲染最强 |
| `seedream-lite` | 火山方舟 Ark | `doubao-seedream-5.0-lite` | 最新 5.0，更快更便宜 |
| `seedream-legacy` | 火山方舟 Ark | `doubao-seedream-4-0-250828` | 4.0 回退版本 |

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

# 多参考图融合（最多 14 张）
vg gen -p "服装来自图1，姿势来自图2" -i a.png -i b.png --model seedream

# 并发生成 4 张变体
vg gen -p "一间安静的书房" --repeat 4 -o studies.png
# 输出：studies_1.png, studies_2.png, studies_3.png, studies_4.png

# 从文件读取 prompt
vg gen -p @prompt.txt --model seedream --res 4K
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

### 语音合成（Volcengine TTS V3）

```bash
# 基本合成
vg tts gen -p "你好，欢迎使用语音合成功能"

# 指定音色和编码
vg tts gen -p "春眠不觉晓，处处闻啼鸟" --voice-type zh_female_shuangkuaisisi_uranus_bigtts --encoding mp3

# 调整语速和音量
vg tts gen -p "这是一段快速播报的内容" --speed 1.5 --volume 1.2

# 情感和情绪强度
vg tts gen -p "今天真是太开心了" --emotion happy --emotion-scale 4

# 音高调整（半音，-12 到 12）
vg tts gen -p "低沉的声音" --pitch -3

# 指定输出文件
vg tts gen -p "静夜思" -o poem.mp3
```

支持的编码格式：`mp3`（默认）、`pcm`、`ogg_opus`。
语速/音量范围 0.1-3.0，对应 V3 内部的 -50 至 100 标度。

#### 内置音色参考

`--voice-type` 需使用 `_uranus_bigtts` 后缀的 2.0 音色：

| 音色 | voice_type |
|---|---|
| Vivi 2.0 | `zh_female_vv_uranus_bigtts` |
| 小何 2.0 | `zh_female_xiaohe_uranus_bigtts` |
| 云舟 2.0 | `zh_male_m191_uranus_bigtts` |
| 小天 2.0 | `zh_male_taocheng_uranus_bigtts` |
| 刘飞 2.0 | `zh_male_liufei_uranus_bigtts` |
| 爽快思思 2.0 | `zh_female_shuangkuaisisi_uranus_bigtts` |
| 魅力苏菲 2.0 | `zh_female_sophie_uranus_bigtts` |
| 清新女声 2.0 | `zh_female_qingxinnvsheng_uranus_bigtts` |
| 甜美小源 2.0 | `zh_female_tianmeixiaoyuan_uranus_bigtts` |
| 甜美桃子 2.0 | `zh_female_tianmeitaozi_uranus_bigtts` |
| 邻家女孩 2.0 | `zh_female_linjianvhai_uranus_bigtts` |
| 温暖阿虎 2.0 | `zh_male_wennuanahu_uranus_bigtts` |
| 少年梓辛 2.0 | `zh_male_shaonianzixin_uranus_bigtts` |
| 猴哥 2.0 | `zh_male_sunwukong_uranus_bigtts` |
| 四郎 2.0 | `zh_male_silang_uranus_bigtts` |
| 儒雅青年 2.0 | `zh_male_ruyaqingnian_uranus_bigtts` |
| 擎苍 2.0 | `zh_male_qingcang_uranus_bigtts` |
| 佩奇猪 2.0 | `zh_female_peiqi_uranus_bigtts` |
| 熊二 2.0 | `zh_male_xionger_uranus_bigtts` |
| 樱桃丸子 2.0 | `zh_female_yingtaowanzi_uranus_bigtts` |
| 奶气萌娃 2.0 | `zh_male_naiqimengwa_uranus_bigtts` |
| 婆婆 2.0 | `zh_female_popo_uranus_bigtts` |
| 林潇 2.0 | `zh_female_linxiao_uranus_bigtts` |
| 霸道总裁 2.0 | `zh_male_aojiaobazong_uranus_bigtts` |
| 高冷御姐 2.0 | `zh_female_gaolengyujie_uranus_bigtts` |
| 暖阳女声 2.0 | `zh_female_kefunvsheng_uranus_bigtts` |
| 鸡汤女 2.0 | `zh_female_jitangnv_uranus_bigtts` |
| 悬疑解说 2.0 | `zh_male_xuanyijieshuo_uranus_bigtts` |
| 霸气青叔 2.0 | `zh_male_baqiqingshu_uranus_bigtts` |
| 磁性解说男声 2.0 | `zh_male_cixingjieshuonan_uranus_bigtts` |
| Tim (美式英语) | `en_male_tim_uranus_bigtts` |
| Dacey (美式英语) | `en_female_dacey_uranus_bigtts` |
| Stokie (美式英语) | `en_female_stokie_uranus_bigtts` |
| Tina老师 2.0 (中英) | `zh_female_yingyujiaoxue_uranus_bigtts` |

完整列表见 https://www.volcengine.com/docs/6561/1257544

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

- **参考图数量限制**：火山方舟最多 14 张
- **4K 分辨率**：seedream 全系支持
- **负面提示词**：Seedream 追加到 prompt 末尾
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
