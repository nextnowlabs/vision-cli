# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- 初始版本，Go 重写的多后端图像与视频生成 CLI 工具 `vg`
- 图像生成 (`vg gen`) 支持 DashScope（`wan-pro`、`wan`）和火山方舟 Ark（`seedream`、`seedream-lite`、`seedream-legacy`）共 5 个模型
- 视频生成 (`vg video gen`) 基于 Seedance 2.0，支持 `seedance` 和 `seedance-fast`
- 视频任务查询 (`vg video status`) 和下载 (`vg video download`)
- 并发生成（`--repeat`，最多 4 个 goroutine 并行）
- 多参考图输入（DashScope ≤9，Seedream ≤14）
- prompt 文件读取（`-p @file.txt`）
- 配置管理 (`vg config show` / `vg config set`)，密钥支持环境变量回退
- 生成历史记录 (`vg history`)，支持分页和关键词搜索
- 用量统计 (`vg stats`)，包含按月/天统计
- 数据存储遵循 XDG 规范（Linux `~/.config/vision-cli/`）
- Makefile 构建，版本号由 git tag 或 commit SHA 自动注入

[Unreleased]: https://github.com/nextnowlabs/vision-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/nextnowlabs/vision-cli/releases/tag/v0.1.0
