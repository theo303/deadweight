package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"flag"

	yaml "github.com/goccy/go-yaml"
	"github.com/theo303/deadweight"
)

var debugFlag = flag.Bool("d", false, "debug mode")
var configFlag = flag.String("c", "", "config file")

func files(current string) []string {
	if len(flag.Args()) > 0 {
		fmt.Println(os.Args[1:])
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

func loadConfig(current string) (deadweight.Rules, error) {
	var configFile string
	if configFlag != nil && *configFlag != "" {
		configFile = *configFlag
	} else {
		configFile = filepath.Join(current, ".deadweight.yaml")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			configFile = ""
		}
	}
	var rules deadweight.Rules
	if configFile != "" {
		content, err := os.ReadFile(configFile)
		if err != nil {
			return deadweight.Rules{}, fmt.Errorf("reading config file %s: %w", configFile, err)
		}
		if err := yaml.Unmarshal(content, &rules); err != nil {
			return deadweight.Rules{}, fmt.Errorf("unmarshaling config file %s: %w", configFile, err)
		}
	}
	return rules, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	flag.Parse()

	debugMode := debugFlag != nil && *debugFlag
	if debugMode {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	current, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	rules, err := loadConfig(current)
	if err != nil {
		panic(fmt.Errorf("loading config:%w", err))
	}

	lc, err := deadweight.NewLSPClient(ctx, "file://"+current, rules)
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

	unusedSymbols := references.GetUnusedSymbols()
	if unusedSymbols.Len() > 0 {
		slog.Info("unused symbols found:")
		unusedSymbols.Print()
	} else {
		slog.Info("no unused symbols found")
	}

	stop()
	lc.Wait()
}
