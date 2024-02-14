package main

import (
	"context"
	"flag"
	"log/slog"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/DimaOne/realix/repo"
	"github.com/DimaOne/realix/server"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error(
				"recovered",
				"error", err,
				"stack", debug.Stack(),
			)
		}
	}()

	port := flag.Uint64("p", 8080, "http port")

	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repo := repo.New()
	srv := server.New(repo)

	srv.Start(ctx, *port)
}
