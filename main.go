package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/akedrou/textdiff"
	"github.com/piger/tailscale-acl/internal/config"
	"github.com/tailscale/hujson"
	"github.com/tailscale/tailscale-client-go/tailscale"
)

var (
	flagPreview = flag.Bool("preview", true, "Preview (diff) the changes to the ACL")
	flagCommit  = flag.Bool("commit", false, "Commit the ACL changes")
)

func run(filename string) error {
	cfg, err := config.Read("config.yml")
	if err != nil {
		return err
	}

	aclContents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	client, err := tailscale.NewClient(cfg.APIKey, cfg.Tailnet)
	if err != nil {
		return err
	}

	ctx := context.Background()

	rawAcl, err := client.RawACL(ctx)
	if err != nil {
		return err
	}

	aclNewFmt, err := hujson.Format(aclContents)
	if err != nil {
		return err
	}

	diffo := textdiff.Unified("old", "new", rawAcl, string(aclNewFmt))
	if diffo == "" {
		fmt.Println("ACL unchanged")
		return nil
	}

	if *flagPreview {
		fmt.Println(diffo)
	}

	if *flagCommit {
		aclNew := string(aclContents)
		if err := client.ValidateACL(ctx, aclNew); err != nil {
			return fmt.Errorf("invalid ACL: %w", err)
		}

		if err := client.SetACL(ctx, aclNew); err != nil {
			return fmt.Errorf("error setting ACL: %w", err)
		}
	}

	return nil
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] <filename>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		usage()
		os.Exit(1)
	}

	if err := run(filename); err != nil {
		slog.Error("error", "err", err)
		os.Exit(1)
	}
}
