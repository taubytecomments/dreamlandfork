package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Relative
	"github.com/pterm/pterm"
	"github.com/taubyte/dreamland/cli/common"
	inject "github.com/taubyte/dreamland/cli/inject"
	"github.com/taubyte/dreamland/cli/kill"
	"github.com/taubyte/dreamland/cli/new"
	"github.com/taubyte/dreamland/cli/status"

	// Actual imports

	client "github.com/taubyte/dreamland/service"
	"github.com/taubyte/tau/libdream/services"
	"github.com/urfave/cli/v2"

	// Empty imports for initializing fixtures, and client/service run methods"
	_ "github.com/taubyte/tau/clients/p2p/auth"
	_ "github.com/taubyte/tau/clients/p2p/hoarder"
	_ "github.com/taubyte/tau/clients/p2p/monkey"
	_ "github.com/taubyte/tau/clients/p2p/patrick"
	_ "github.com/taubyte/tau/clients/p2p/seer"
	_ "github.com/taubyte/tau/clients/p2p/tns"
	_ "github.com/taubyte/tau/libdream/common/fixtures"
	_ "github.com/taubyte/tau/protocols/auth"
	_ "github.com/taubyte/tau/protocols/hoarder"
	_ "github.com/taubyte/tau/protocols/monkey"
	_ "github.com/taubyte/tau/protocols/monkey/fixtures/compile"
	_ "github.com/taubyte/tau/protocols/patrick"
	_ "github.com/taubyte/tau/protocols/seer"
	_ "github.com/taubyte/tau/protocols/substrate"
	_ "github.com/taubyte/tau/protocols/tns"
)

func main() {
	ctx, ctxC := context.WithCancel(context.Background())

	// buffered channel for synchronous os signal control - AMB 14SEP2023 
	signals := make(chan os.Signal, 1)

	// first signal for kick-off/startup - AMB 14SEP2023 
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signals
		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			pterm.Info.Println("Received signal... Shutting down.")
			ctxC()
		}
	}()

	defer func() {
		if common.DoDaemon {
			ctxC()
			services.Zeno()
		}
	}()

	// client service confiration - AMB 14SEP2023 
	ops := []client.Option{client.URL(common.DefaultDreamlandURL), client.Timeout(300 * time.Second)}

	// client service startup - AMB 14SEP2023 
	multiverse, err := client.New(ctx, ops...)

	// fatal service exception handler - AMB 14SEP2023 
	if err != nil {
		log.Fatalf("Starting new dreamland client failed with: %s", err.Error())
	}

	// CLI configuration - AMB 14SEP2023 
	err = defineCLI(&common.Context{Ctx: ctx, Multiverse: multiverse}).RunContext(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func defineCLI(ctx *common.Context) *(cli.App) {
	app := &cli.App{
		UseShortOptionHandling: true,
		Commands: []*cli.Command{
			new.Command(ctx),
			inject.Command(ctx),
			kill.Command(ctx),
			status.Command(ctx),
		},
		Suggest:              true,
		EnableBashCompletion: true,
	}

	return app
}
