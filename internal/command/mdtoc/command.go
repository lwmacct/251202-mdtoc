package mdtoc

import (
	"github.com/lwmacct/251207-go-pkg-version/pkg/version"
	"github.com/urfave/cli/v3"
)

// Command 返回 mc-mdtoc 主命令
var Command = &cli.Command{
	Name:     "mc-mdtoc",
	Usage:    "生成和查看 Markdown 文档的大纲 (TOC)",
	Commands: []*cli.Command{version.Command},
	UsageText: `mc-mdtoc [options] <file>...
fd -e md | mc-mdtoc`,
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "min-level",
			Aliases: []string{"m"},
			Value:   1,
			Usage:   "最小标题层级 (1-6)",
		},
		&cli.IntFlag{
			Name:    "max-level",
			Aliases: []string{"M"},
			Value:   3,
			Usage:   "最大标题层级 (1-6)",
		},
		&cli.BoolFlag{
			Name:    "in-place",
			Aliases: []string{"i"},
			Usage:   "原地更新文件 (在 <!--TOC--> 标记处插入)",
		},
		&cli.BoolFlag{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "删除文件中的 TOC 标记和内容",
		},
		&cli.BoolFlag{
			Name:    "ordered",
			Aliases: []string{"o"},
			Usage:   "使用有序列表 (1. 2. 3.)",
		},
		&cli.BoolFlag{
			Name:    "line-number",
			Aliases: []string{"L"},
			Value:   true,
			Usage:   "显示行号范围 (:start:end)",
		},
		&cli.BoolFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "显示文件路径 (path:start:end)",
		},
		&cli.BoolFlag{
			Name:    "global",
			Aliases: []string{"g"},
			Usage:   "全局模式: 生成完整文档的单一目录 (默认为章节模式)",
		},
		&cli.BoolFlag{
			Name:    "anchor",
			Aliases: []string{"a"},
			Usage:   "预览时显示锚点链接 [标题](#anchor)",
		},
		&cli.StringFlag{
			Name:    "toc-title",
			Aliases: []string{"T"},
			Usage:   "TOC 标题 (如 '文档目录'，将在 TOC 内生成 ## 文档目录)",
		},
	},
	Action: action,
}
