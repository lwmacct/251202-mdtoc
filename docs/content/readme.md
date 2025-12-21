# mc-mdtoc

<!--TOC-->

- [功能特性](#功能特性) `:19+8`
- [安装](#安装) `:27+10`
- [使用示例](#使用示例) `:37+26`
- [命令选项](#命令选项) `:63+14`
- [开发](#开发) `:77+18`
  - [环境准备](#环境准备) `:79+10`
  - [构建](#构建) `:89+6`
- [参考项目](#参考项目) `:95+7`
- [相关链接](#相关链接) `:102+4`

<!--TOC-->

Markdown TOC 生成工具，为 Markdown 文件自动生成符合规范的目录（Table of Contents）。

## 功能特性

- 生成 GitHub 风格的 Table of Contents
- 支持 `<!--TOC-->` 标记定位，原地更新文件
- 章节模式：每个 H1 后生成独立子目录
- 支持 YAML Frontmatter（VitePress、Hugo 等）
- 多文件批量处理，支持管道输入

## 安装

```shell
# 从 GitHub 安装
go install github.com/lwmacct/251202-mc-mdtoc/cmd/mc-mdtoc@latest

# 本地构建安装
go install ./cmd/mc-mdtoc
```

## 使用示例

```shell
# 查看帮助
mc-mdtoc --help

# 生成 TOC 到 stdout
mc-mdtoc README.md

# 原地更新文件 (在 <!--TOC--> 标记处插入)
mc-mdtoc -i README.md

# 显示文件路径 + 行号范围
mc-mdtoc -p README.md
# 输出: - [标题](#标题) `README.md:1+10`

# 使用有序列表 + 指定层级
mc-mdtoc -o -m 2 -M 4 README.md

# 多文件处理
mc-mdtoc -i docs/*.md

# 管道输入 (从 stdin 读取文件列表)
fd -e md | mc-mdtoc -i
```

## 命令选项

| 选项            | 短选项 | 说明                                   |
| --------------- | ------ | -------------------------------------- |
| `--min-level`   | `-m`   | 最小标题层级 (默认 1)                  |
| `--max-level`   | `-M`   | 最大标题层级 (默认 3)                  |
| `--in-place`    | `-i`   | 原地更新文件                           |
| `--delete`      | `-d`   | 删除文件中的 TOC                       |
| `--ordered`     | `-o`   | 使用有序列表                           |
| `--line-number` | `-L`   | 显示行号范围 `:start+count` (默认启用) |
| `--path`        | `-p`   | 显示文件路径 `path:start+count`        |
| `--global`      | `-g`   | 全局模式 (默认为章节模式)              |
| `--anchor`      | `-a`   | 预览时显示锚点链接                     |

## 开发

### 环境准备

```shell
# 安装 pre-commit hooks
pre-commit install

# 查看可用任务
task -a
```

### 构建

```shell
go build ./cmd/mc-mdtoc/
```

## 参考项目

| 项目                                         | 语言   | 说明              |
| -------------------------------------------- | ------ | ----------------- |
| [md-toc](https://github.com/frnmst/md-toc)   | Python | TOC 生成          |
| [goldmark](https://github.com/yuin/goldmark) | Go     | CommonMark 解析器 |

## 相关链接

- [CommonMark Spec](https://spec.commonmark.org/0.31.2/) - Markdown 规范
- [Taskfile](https://taskfile.dev) - 任务管理
