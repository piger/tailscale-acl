package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/akedrou/textdiff"
	"github.com/piger/tailscale-acl/internal/config"
	"github.com/tailscale/hujson"
	"tailscale.com/client/tailscale/v2"
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

	client := &tailscale.Client{
		Tailnet: cfg.Tailnet,
		APIKey:  cfg.APIKey,
	}

	ctx := context.Background()

	policyFile := client.PolicyFile()

	rawACL, err := policyFile.Raw(ctx)
	if err != nil {
		return err
	}

	aclNewFmt, err := hujson.Format(aclContents)
	if err != nil {
		return err
	}

	diffo := textdiff.Unified("old", "new", rawACL.HuJSON, string(aclNewFmt))
	if diffo == "" {
		fmt.Println("ACL unchanged")
		return nil
	}

	if *flagPreview {
		fmt.Println(diffo)
	}

	if *flagCommit {
		aclNew := string(aclContents)
		if err := policyFile.Validate(ctx, aclNew); err != nil {
			return fmt.Errorf("invalid ACL: %w", err)
		}

		if err := policyFile.Set(ctx, aclNew, ""); err != nil {
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
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}
