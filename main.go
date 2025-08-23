package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/konradmalik/efm-langserver/core"
	"github.com/konradmalik/efm-langserver/logs"
	"github.com/konradmalik/efm-langserver/lsp"
)

const (
	name    = "efm-langserver"
	version = "0.0.54"
)

var revision = "HEAD"

func main() {
	var logfile string
	var loglevel int
	var showVersion bool
	var usage bool

	flag.StringVar(&logfile, "logfile", "", "File to save logs into. If provided stderr won't be used anymore.")
	flag.IntVar(&loglevel, "loglevel", 2, "Set the log level. Max is 3 (debug), min is 0 (error). Higher number logs less. Set <0 for no logs.")
	flag.BoolVar(&showVersion, "v", false, "Print the version")
	flag.BoolVar(&usage, "h", false, "Show help")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return
	}

	if usage || flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}

	config := core.NewConfig()
	logs.InitializeLogger(logfile, logs.LogLevel(max(loglevel, -1)))
	logs.Log.Logln(logs.Info, "reading on stdin, writing on stdout")

	var f *os.File
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	internalHandler := core.NewHandler(config)
	handler := lsp.NewHandler(internalHandler)
	<-jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(handler.Handle),
		jsonrpc2.LogMessages(logs.Log)).DisconnectNotify()

	logs.Log.Logln(logs.Info, "efm-langserver: connections closed")
}

type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (c stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (c stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
