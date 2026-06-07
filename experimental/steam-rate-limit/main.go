package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, "配置错误:", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	report, err := runExperiment(ctx, cfg)
	if err != nil && report.RunID == "" {
		fmt.Fprintln(os.Stderr, "实验失败:", err)
		os.Exit(1)
	}

	generated, writeErr := writeOutputs(report, cfg.OutputDir)
	if writeErr != nil {
		fmt.Fprintln(os.Stderr, "写出报告失败:", writeErr)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println("实验已中断，已写出 partial 报告:", err)
	} else {
		fmt.Println("实验完成")
	}
	for _, path := range generated {
		fmt.Println("-", path)
	}
	fmt.Println("建议:", report.Summary.Recommendation)
	if err != nil {
		os.Exit(1)
	}
}
