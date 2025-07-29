package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/peter-evans/patience"
	"github.com/piger/tailscale-acl/internal/config"
	"github.com/tailscale/hujson"
	"tailscale.com/client/tailscale/v2"
)

var (
	flagPreview = flag.Bool("preview", true, "Preview (diff) the changes to the ACL")
	flagCommit  = flag.Bool("commit", false, "Commit the ACL changes")
)

func diff(old, new string) string {
	a := strings.Split(old, "\n")
	b := strings.Split(new, "\n")
	diffs := patience.Diff(a, b)
	return patience.UnifiedDiffText(diffs)
}

func readACLFile(filename string) (string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("reading ACL file: %w", err)
	}

	bf, err := hujson.Format(b)
	if err != nil {
		return "", fmt.Errorf("formatting ACL as HuJSON: %w", err)
	}

	return string(bf), nil
}

func run(filename string) error {
	cfg, err := config.Read("config.yml")
	if err != nil {
		return err
	}

	newACL, err := readACLFile(filename)
	if err != nil {
		return err
	}

	ctx := context.Background()
	client := &tailscale.Client{
		Tailnet: cfg.Tailnet,
		APIKey:  cfg.APIKey,
	}
	policyFile := client.PolicyFile()

	rawACL, err := policyFile.Raw(ctx)
	if err != nil {
		return err
	}

	diffo := diff(rawACL.HuJSON, newACL)
	if diffo == "" {
		fmt.Println("ACL unchanged")
		return nil
	}

	if *flagPreview {
		fmt.Println(diffo)
	}

	if *flagCommit {
		if err := policyFile.Validate(ctx, newACL); err != nil {
			return fmt.Errorf("invalid ACL: %w", err)
		}

		if err := policyFile.Set(ctx, newACL, rawACL.ETag); err != nil {
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
