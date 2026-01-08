package main

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/theo303/deadweight"
)

var debugMode = false

func files(current string) []string {
	if len(os.Args) > 1 {
		return os.Args[1:]
	}

	var goFiles []string
	if err := filepath.WalkDir(current, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || strings.Contains(name, "mock") {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			if !strings.HasSuffix(path, "_test.go") {
				goFiles = append(goFiles, strings.TrimPrefix(path, current+"/"))
			}
		}

		return nil
	}); err != nil {
		slog.Error("failed to walk directory", slog.Any("error", err))
		os.Exit(1)
	}
	return goFiles
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	if debugMode {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	current, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	lc, err := deadweight.NewLSPClient(ctx, "file://"+current)
	if err != nil {
		slog.Error("failed to initialize LSP client", slog.Any("error", err))
		os.Exit(1)
	}

	if err := lc.RunAndInitialize(ctx); err != nil {
		slog.Error("failed to run LSP client", slog.Any("error", err))
		os.Exit(1)
	}

	allSymbols := deadweight.NewSymbolMap()
	wg := &sync.WaitGroup{}

	files := files(current)

	for _, file := range files {
		wg.Add(1)
		go func() {
			if err := lc.ListDocumentSymbols(file, wg, allSymbols); err != nil {
				slog.Error("failed to list workspace symbols", slog.Any("error", err))
				os.Exit(1)
			}
		}()
	}
	wg.Wait()

	if debugMode {
		allSymbols.Print()
	}

	references, err := lc.ReferencesSymbols(allSymbols)
	if err != nil {
		slog.Error("failed to reference symbols", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("unused symbols found")
	references.GetUnusedSymbols().Print()

	stop()
	lc.Wait()
}
