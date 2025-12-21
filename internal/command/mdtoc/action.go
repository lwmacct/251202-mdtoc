package mdtoc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/lwmacct/251202-mc-mdtoc/internal/config"
	"github.com/lwmacct/251202-mc-mdtoc/internal/mdtoc"
	"github.com/lwmacct/251207-go-pkg-cfgm/pkg/cfgm"
	"github.com/lwmacct/251207-go-pkg-version/pkg/version"
	"github.com/urfave/cli/v3"
)

func action(ctx context.Context, cmd *cli.Command) error {
	// 加载配置 (配置文件 → 环境变量 → CLI flags)
	cfg := *cfgm.MustLoadCmd(cmd, config.DefaultConfig(), version.GetAppRawName())

	// 操作模式 (不在配置文件中)
	inPlace := cmd.Bool("in-place")
	deleteMode := cmd.Bool("delete")

	// 验证层级参数
	if cfg.MinLevel < 1 || cfg.MinLevel > 6 {
		return errors.New("min-level 必须在 1-6 之间")
	}
	if cfg.MaxLevel < 1 || cfg.MaxLevel > 6 {
		return errors.New("max-level 必须在 1-6 之间")
	}
	if cfg.MinLevel > cfg.MaxLevel {
		return errors.New("min-level 不能大于 max-level")
	}

	// 收集要处理的文件
	files := collectFiles(cmd.Args().Slice())
	if len(files) == 0 {
		// 无文件时显示帮助
		return cli.ShowSubcommandHelp(cmd)
	}

	// 创建基础选项
	// 默认启用章节模式 (SectionTOC=true)，只有指定 --global 才使用全局模式
	baseOpts := mdtoc.Options{
		MinLevel:   cfg.MinLevel,
		MaxLevel:   cfg.MaxLevel,
		Ordered:    cfg.Ordered,
		LineNumber: cfg.LineNumber,
		ShowPath:   cfg.ShowPath,
		SectionTOC: !cfg.Global,
		ShowAnchor: cfg.Anchor,
		TOCTitle:   cfg.TOCTitle,
	}

	// 根据模式执行不同操作
	switch {
	case deleteMode:
		return processDelete(mdtoc.New(baseOpts), files)
	case inPlace:
		// inPlace 模式强制启用 ShowAnchor（写入文件必须有链接）
		writeOpts := baseOpts
		writeOpts.ShowAnchor = true
		return processInPlace(mdtoc.New(writeOpts), files)
	default:
		return processStdout(baseOpts, files)
	}
}

// collectFiles 收集要处理的文件列表
// 优先从命令行参数获取，如果没有则尝试从 stdin 读取
func collectFiles(args []string) []string {
	var files []string

	// 从命令行参数收集
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg != "" {
			files = append(files, arg)
		}
	}

	// 如果没有参数，尝试从 stdin 读取
	if len(files) == 0 && !isTerminal(os.Stdin) {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				files = append(files, line)
			}
		}
	}

	return files
}

// isTerminal 检查文件是否是终端
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return true
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// processDelete 删除模式 - 删除文件中的 TOC
func processDelete(toc *mdtoc.TOC, files []string) error {
	var errors []string

	for _, file := range files {
		if err := checkFileExists(file); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))
			continue
		}

		deleted, err := toc.DeleteTOC(file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))
			continue
		}

		if deleted {
			_, _ = os.Stdout.WriteString(file + ": TOC 已删除\n")
		} else {
			_, _ = os.Stdout.WriteString(file + ": 无 TOC 标记\n")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分文件处理失败:\n%s", strings.Join(errors, "\n"))
	}
	return nil
}

// processInPlace 原地更新模式
// 如果文件没有 TOC 标记，会自动在第一个标题后插入
func processInPlace(toc *mdtoc.TOC, files []string) error {
	var errors []string

	for _, file := range files {
		if err := checkFileExists(file); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))
			continue
		}

		hasMarker, _ := toc.HasMarker(file)

		if err := toc.UpdateFile(file); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))
			continue
		}

		if hasMarker {
			_, _ = os.Stdout.WriteString(file + ": 已更新\n")
		} else {
			_, _ = os.Stdout.WriteString(file + ": 已插入 (在第一个标题后)\n")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分文件处理失败:\n%s", strings.Join(errors, "\n"))
	}
	return nil
}

// processStdout 输出到 stdout 模式
func processStdout(baseOpts mdtoc.Options, files []string) error {
	for i, file := range files {
		if err := checkFileExists(file); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			continue
		}

		// 为每个文件创建带有文件路径的 TOC 实例
		opts := baseOpts
		opts.FilePath = file
		toc := mdtoc.New(opts)

		var tocStr string
		var err error

		if opts.SectionTOC {
			// 章节模式：预览每个 H1 的子目录
			content, readErr := os.ReadFile(file) //nolint:gosec // G304: file path from user input is intentional
			if readErr != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", file, readErr)
				continue
			}
			tocStr, err = toc.GenerateSectionTOCsPreview(content)
		} else {
			tocStr, err = toc.GenerateFromFile(file)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			continue
		}

		// 跳过空的 TOC
		if strings.TrimSpace(tocStr) == "" {
			continue
		}

		// 多文件时添加文件名标题
		if len(files) > 1 {
			_, _ = os.Stdout.WriteString("## " + file + "\n\n")
		}

		_, _ = os.Stdout.WriteString(tocStr + "\n")

		// 多文件时添加分隔
		if len(files) > 1 && i < len(files)-1 {
			_, _ = os.Stdout.WriteString("\n")
		}
	}

	return nil
}

// checkFileExists 检查文件是否存在
func checkFileExists(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return errors.New("文件不存在")
	}
	return nil
}
