package deadweight

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/theo303/deadweight/lsp"
)

type lspClient struct {
	cmd *exec.Cmd

	wg sync.WaitGroup

	pipeIn io.Writer
	ready  chan struct{}

	pendingMessages sync.Map // map[int32]messageHandler
	idCounter       atomic.Int32

	root string
}

func NewLSPClient(ctx context.Context, root string) (*lspClient, error) {
	cmd := exec.CommandContext(ctx, "gopls", "-vv")

	lc := &lspClient{
		cmd:             cmd,
		wg:              sync.WaitGroup{},
		ready:           make(chan struct{}, 1),
		pendingMessages: sync.Map{},
		root:            root,
	}

	pipeOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	pipeErr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	lc.pipeIn, err = cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	lc.readStdOut(pipeOut)
	lc.logStrErr(pipeErr)

	return lc, nil
}

func (lc *lspClient) RunAndInitialize(ctx context.Context) error {
	if err := lc.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}
	slog.Info("lsp client running")

	if err := lc.sendCommand("initialize",
		map[string]any{
			"processId": nil,
			"rootUri":   lc.root,
			"capabilities": map[string]any{
				"textDocument": map[string]any{
					"documentSymbol": map[string]any{
						"hierarchicalDocumentSymbolSupport": true,
					},
				},
			},
			"trace": "off",
		},
		initializeResponse(lc.ready),
	); err != nil {
		return fmt.Errorf("failed to send intialize command: %w", err)
	}

	select {
	case <-lc.ready:
		if err := lc.sendCommand("initialized", map[string]any{}, nil); err != nil {
			return fmt.Errorf("failed to send intialized command: %w", err)
		}
	case <-ctx.Done():
		return nil
	}

	return nil
}

func (lc *lspClient) Wait() {
	_ = lc.cmd.Wait()
	lc.wg.Wait()
	slog.Info("lsp client exited")
}

func (lc *lspClient) ListDocumentSymbols(filePath string, wg *sync.WaitGroup, symbols *SymbolMap) error {

	if err := lc.sendCommand("textDocument/documentSymbol",
		map[string]any{
			"textDocument": map[string]any{
				"uri": lc.root + "/" + filePath,
			},
		},
		lc.documentSymbolResponse(wg, symbols, filePath),
	); err != nil {
		wg.Done()
		return fmt.Errorf("failed to send workspace/symbol command: %w", err)
	}
	return nil
}

func (lc *lspClient) ReferencesSymbols(allSymbols *SymbolMap) (*ReferenceMap, error) {
	defer allSymbols.Unlock()

	wg := &sync.WaitGroup{}
	references := NewReferenceMap()

	allSymbols.Lock()
	for filePath, symbols := range allSymbols.m {
		for _, symbol := range symbols {
			wg.Add(1)
			if err := lc.references(
				wg,
				references,
				filePath,
				symbol,
			); err != nil {
				return nil, err
			}
		}
	}
	wg.Wait()

	return references, nil
}

func (lc *lspClient) references(
	wg *sync.WaitGroup,
	references *ReferenceMap,
	filePath string,
	symbol Symbol,
) error {

	if err := lc.sendCommand("textDocument/references",
		map[string]any{
			"textDocument": map[string]any{
				"uri": lc.root + "/" + filePath,
			},
			"position": map[string]any{
				"line":      symbol.Position.Line,
				"character": symbol.Position.Character,
			},
			"context": map[string]any{
				"includeDeclaration": false,
			},
		},
		referencesResponse(wg, references, filePath, symbol),
	); err != nil {
		wg.Done()
		return fmt.Errorf("failed to send textDocument/references command: %w", err)
	}
	return nil
}

func (lc *lspClient) isEmbedded(filePath string, documentSymbol lsp.DocumentSymbol) (bool, error) {
	if documentSymbol.Kind != lsp.SymbolKindField {
		return false, nil
	}
	detailSplit := strings.Split(documentSymbol.Detail, ".")
	if detailSplit[len(detailSplit)-1] != documentSymbol.Name {
		return false, nil
	}

	hasParentType := make(chan bool)
	defer close(hasParentType)

	pos := documentSymbol.SelectionRange.Start
	if err := lc.PrepareTypeHierarchy(filePath, pos, hasParentType); err != nil {
		return false, err
	}
	return <-hasParentType, nil
}

func (lc *lspClient) PrepareTypeHierarchy(filePath string, position lsp.Position, hasParentType chan bool) error {

	if err := lc.sendCommand("textDocument/prepareTypeHierarchy", map[string]any{
		"textDocument": map[string]any{
			"uri": lc.root + "/" + filePath,
		},
		"position": map[string]any{
			"line":      position.Line,
			"character": position.Character,
		},
	},
		prepareTypeHierarchyResponse(hasParentType),
	); err != nil {
		return fmt.Errorf("failed to send textDocument/prepareTypeHierarchy command: %w", err)
	}

	return nil
}

type command struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      int32          `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
}

func (lc *lspClient) sendCommand(method string, params map[string]any, handler messageHandler) error {
	id := lc.idCounter.Add(1)
	if handler != nil {
		lc.pendingMessages.Store(id, handler)
	}

	payload, err := json.Marshal(command{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}
	msg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(payload), payload)
	if _, err := io.WriteString(lc.pipeIn, msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

func (lc *lspClient) readStdOut(r io.Reader) {
	reader := bufio.NewReader(r)
	lc.wg.Go(func() {
		for {
			// Read headers
			var contentLength int
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if !errors.Is(err, io.EOF) {
						slog.Error("failed to read headers from STDOUT", slog.Any("error", err))
					}
					return
				}
				line = strings.TrimSpace(line)
				if line == "" {
					break
				}
				if strings.HasPrefix(line, "Content-Length:") {
					_, _ = fmt.Sscanf(line, "Content-Length: %d", &contentLength)
				}
			}
			if contentLength == 0 {
				slog.Warn("received empty response from gopls")
				continue
			}
			body := make([]byte, contentLength)
			if _, err := io.ReadFull(reader, body); err != nil {
				slog.Error("failed to read body from STDOUT", slog.Any("error", err))
				continue
			}

			var msg lsp.Message
			if err := json.Unmarshal(body, &msg); err != nil {
				slog.Error("failed to unmarshal message", slog.Any("error", err))
				continue
			}

			slog.Debug("message received", slog.String("body", string(body)), slog.Any("id", msg.ID))
			if msg.ID == 0 {
				continue
			}

			value, ok := lc.pendingMessages.LoadAndDelete(msg.ID)
			if !ok {
				slog.Debug("id not in map, ignored", slog.Any("id", msg.ID))
				continue
			}
			go value.(messageHandler)(msg)
		}
	})
}

func (lc *lspClient) logStrErr(r io.Reader) {
	reader := bufio.NewReader(r)
	lc.wg.Go(func() {
		for {
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if !errors.Is(err, io.EOF) {
						slog.Error("failed to read from STDERR", slog.Any("error", err))
					}
					return
				}
				line = strings.TrimSpace(line)
				if line == "" {
					break
				}
				slog.Error("error from gopls", slog.Any("error", line))
			}
		}
	})
}
