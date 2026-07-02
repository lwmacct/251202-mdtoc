package root

import (
	"github.com/lwmacct/251202-mdtoc/internal/config"
	"github.com/lwmacct/251207-go-pkg-cfgm/pkg/cfgm"
	"github.com/lwmacct/251207-go-pkg-version/pkg/version"
	"github.com/urfave/cli/v3"
)

var (
	defaults = config.DefaultConfig()
	usage    = cfgm.Schema(defaults).Command()
)

// Command 返回 mdtoc 主命令
var Command = &cli.Command{
	Name:            "mdtoc",
	Usage:           "生成和查看 Markdown 文档的大纲 (TOC)",
	Version:         version.AppVersion,
	Commands:        []*cli.Command{version.Command},
	HideHelpCommand: true,
	UsageText: `mdtoc [options] <file>...
fd -e md | mdtoc`,
	Flags:  commandFlags(),
	Action: action,
}

func commandFlags() []cli.Flag {
	return []cli.Flag{
		cfgm.ConfigFlag(),
		&cli.IntFlag{
			Name:    "min-level",
			Aliases: []string{"m"},
			Value:   defaults.MinLevel,
			Usage:   usage.MustUsage("min-level"),
		},
		&cli.IntFlag{
			Name:    "max-level",
			Aliases: []string{"M"},
			Value:   defaults.MaxLevel,
			Usage:   usage.MustUsage("max-level"),
		},
		&cli.BoolFlag{
			Name:    "in-place",
			Aliases: []string{"i"},
			Usage:   "原地更新文件 (在 <!--TOC--> 标记处插入)",
		},
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Value:   defaults.Force,
			Usage:   usage.MustUsage("force"),
		},
		&cli.BoolFlag{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "删除文件中的 TOC 标记和内容",
		},
		&cli.BoolFlag{
			Name:    "ordered",
			Aliases: []string{"o"},
			Value:   defaults.Ordered,
			Usage:   usage.MustUsage("ordered"),
		},
		&cli.BoolFlag{
			Name:    "line-number",
			Aliases: []string{"L"},
			Value:   defaults.LineNumber,
			Usage:   usage.MustUsage("line-number"),
		},
		&cli.BoolFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Value:   defaults.ShowPath,
			Usage:   usage.MustUsage("path"),
		},
		&cli.BoolFlag{
			Name:    "global",
			Aliases: []string{"g"},
			Value:   defaults.Global,
			Usage:   usage.MustUsage("global"),
		},
		&cli.BoolFlag{
			Name:    "anchor",
			Aliases: []string{"a"},
			Value:   defaults.Anchor,
			Usage:   usage.MustUsage("anchor"),
		},
		&cli.StringFlag{
			Name:    "toc-title",
			Aliases: []string{"T"},
			Value:   defaults.TOCTitle,
			Usage:   usage.MustUsage("toc-title"),
		},
	}
}
