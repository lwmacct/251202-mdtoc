# ImageMagick 图标生成技巧

<!--TOC-->

- [基础用法](#基础用法) `:19+15`
- [参数说明](#参数说明) `:34+12`
- [常用变体](#常用变体) `:46+32`
  - [圆角图标](#圆角图标) `:48+10`
  - [渐变背景](#渐变背景) `:58+9`
  - [带边框](#带边框) `:67+11`
- [查看可用字体](#查看可用字体) `:78+6`
- [适用场景](#适用场景) `:84+7`
- [参考](#参考) `:91+4`

<!--TOC-->

使用 ImageMagick 快速生成简单的项目图标，无需图形设计软件。

## 基础用法

```bash
# 安装 ImageMagick
apt-get install -y imagemagick

# 生成 128x128 蓝底白字图标
convert -size 128x128 xc:'#4a90d9' \
  -fill white -font DejaVu-Sans-Bold -pointsize 28 -gravity center \
  -draw "text 0,-15 'TOC'" \
  -fill white -font DejaVu-Sans -pointsize 14 -gravity center \
  -draw "text 0,20 '≡ ≡ ≡'" \
  icon.png
```

## 参数说明

| 参数                       | 说明                                      |
| -------------------------- | ----------------------------------------- |
| `-size 128x128`            | 图片尺寸                                  |
| `xc:'#4a90d9'`             | 背景色（X Color）                         |
| `-fill white`              | 填充色（文字颜色）                        |
| `-font DejaVu-Sans-Bold`   | 字体                                      |
| `-pointsize 28`            | 字号                                      |
| `-gravity center`          | 对齐方式                                  |
| `-draw "text 0,-15 'TOC'"` | 绘制文字，`0,-15` 是相对于 gravity 的偏移 |

## 常用变体

### 圆角图标

```bash
convert -size 128x128 xc:none \
  -fill '#4a90d9' -draw "roundrectangle 0,0 127,127 16,16" \
  -fill white -font DejaVu-Sans-Bold -pointsize 32 -gravity center \
  -draw "text 0,0 'MD'" \
  icon.png
```

### 渐变背景

```bash
convert -size 128x128 gradient:'#667eea'-'#764ba2' \
  -fill white -font DejaVu-Sans-Bold -pointsize 28 -gravity center \
  -draw "text 0,0 'TOC'" \
  icon.png
```

### 带边框

```bash
convert -size 128x128 xc:'#4a90d9' \
  -fill none -stroke white -strokewidth 4 \
  -draw "roundrectangle 8,8 119,119 12,12" \
  -fill white -font DejaVu-Sans-Bold -pointsize 28 -gravity center \
  -draw "text 0,0 'TOC'" \
  icon.png
```

## 查看可用字体

```bash
convert -list font | grep -i "Font:"
```

## 适用场景

- VSCode 插件图标（128x128 PNG）
- npm 包图标
- CLI 工具 Logo
- 快速原型设计

## 参考

- [ImageMagick 官方文档](https://imagemagick.org/script/command-line-processing.php)
- [Drawing Primitives](https://imagemagick.org/script/magick-vector-graphics.php)
