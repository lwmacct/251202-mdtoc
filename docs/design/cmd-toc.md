# 命令行用法

> **状态**: ✅ 已完成 (Phase 1-2)

为 Markdown 文件自动生成符合规范的目录（Table of Contents）。

<!--TOC-->

## Table of Contents

- [命令行接口](#命令行接口) `:21+19`
- [功能特性](#功能特性) `:40+22`
- [输出格式](#输出格式) `:62+24`
- [TOC 标记规范](#toc-标记规范) `:86+15`
- [YAML Frontmatter 支持](#yaml-frontmatter-支持) `:101+22`
- [技术实现](#技术实现) `:123+14`
- [参考项目](#参考项目) `:137+7`

<!--TOC-->

## 命令行接口

```shell
mdtoc [options] <file>...
   fd -e md | mdtoc

Options:
  -m, --min-level    最小标题层级 (默认 1)
  -M, --max-level    最大标题层级 (默认 3)
  -i, --in-place     原地更新文件
  -f, --force        强制生成 TOC，即使文件中没有 <!--TOC--> 标记
  -d, --delete       删除文件中的 TOC 标记和内容
  -o, --ordered      有序列表
  -L, --line-number  显示行号范围 :start+count=end (默认启用)
  -p, --path         显示文件路径 path:start+count=end
  -g, --global       全局模式 (默认为章节模式)
  -a, --anchor       预览时显示锚点链接 [标题](#anchor)
```

## 功能特性

| 功能        | 说明                              | 状态      |
| ----------- | --------------------------------- | --------- |
| 标题解析    | 解析 ATX 风格标题 (`# ~ ######`)  | ✅ 已完成 |
| 锚点生成    | GitHub 规范 anchor link           | ✅ 已完成 |
| TOC 标记    | 支持 `<!--TOC-->` 标记定位        | ✅ 已完成 |
| 标记检查    | 默认只在有标记时更新，`-f` 强制   | ✅ 已完成 |
| 原地更新    | `-i` 直接修改文件                 | ✅ 已完成 |
| TOC 删除    | `-d` 删除文件中的 TOC             | ✅ 已完成 |
| 有序列表    | `-o` 生成 `1. 2. 3.` 格式         | ✅ 已完成 |
| 行号范围    | `-L` 显示 `:start+count=end`      | ✅ 已完成 |
| 文件路径    | `-p` 显示 `path:start+count=end`  | ✅ 已完成 |
| 锚点显示    | `-a` 预览时显示 `[标题](#anchor)` | ✅ 已完成 |
| 章节模式    | 默认：每个 H1 后生成独立子目录    | ✅ 已完成 |
| H2 检查     | 章节需至少包含一个 H2 才生成 TOC  | ✅ 已完成 |
| 多 H1 支持  | 单文档支持多个 H1 章节            | ✅ 已完成 |
| 全局模式    | `-g` 生成完整文档的单一目录       | ✅ 已完成 |
| 多文件处理  | 支持多文件和管道输入              | ✅ 已完成 |
| Frontmatter | 跳过 YAML frontmatter 区域        | ✅ 已完成 |
| 多框架支持  | VitePress、Hugo 等                | ✅ 已完成 |

## 输出格式

```shell
# 默认输出 (预览模式不显示锚点)
mdtoc README.md
# - [标题] `:1+10=11`

# 显示锚点链接
mdtoc -a README.md
# - [标题](#标题) `:1+10=11`

# 带文件路径
mdtoc -a -p README.md
# - [标题](#标题) `README.md:1+10=11`

# 禁用行号
mdtoc -a -L=false README.md
# - [标题](#标题)

# 写入文件时自动启用锚点链接
mdtoc -i README.md
# 文件内容: - [标题](#标题) `:1+10=11`
```

## TOC 标记规范

使用 HTML 注释作为标记，渲染后不可见：

```markdown

```

**更新逻辑**：

1. 查找第一个 `<!--TOC-->` 标记
2. 查找第二个 `<!--TOC-->` 标记（可选）
3. 替换两个标记之间的内容
4. 如果没有标记，默认跳过该文件；使用 `-f` 可强制在第一个标题后插入

## YAML Frontmatter 支持

TOC 工具会自动检测并跳过文件开头的 YAML frontmatter 区域。这确保了与 VitePress、Hugo 等静态站点生成器的兼容性。

**规则**：

- Frontmatter 必须从文件第一行开始，以 `---` 标记
- 结束标记可以是 `---` 或 `...`
- Frontmatter 内的 `#` 注释和 `<!--TOC-->` 标记会被忽略

```markdown
---
# 这是 YAML 注释，不会被识别为 Markdown 标题
title: 页面标题
layout: home
---

# 真正的 H1 标题

这里的内容会被正确处理...
```

## 技术实现

基于 [goldmark](https://github.com/yuin/goldmark) CommonMark 解析器。

**核心模块**：

| 文件           | 职责                         |
| -------------- | ---------------------------- |
| `types.go`     | Header/Options 类型定义      |
| `parser.go`    | 解析 Markdown，提取标题      |
| `anchor.go`    | GitHub 风格 anchor link 生成 |
| `generator.go` | TOC 字符串生成               |
| `marker.go`    | `<!--TOC-->` 标记处理        |

## 参考项目

| 项目         | 说明                 |
| ------------ | -------------------- |
| md-toc       | Python TOC 生成器    |
| gh-md-toc-go | Go GitHub TOC 生成器 |
| goldmark-toc | goldmark TOC 扩展    |
