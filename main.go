package main

import (
	"log/slog"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

var (
	log *slog.Logger

	cache  *Cache
	config Config
)

func loadCache(ctx *cli.Context) error {
	return nil
}

func cleanCache(ctx *cli.Context) error {
	return nil
}

func run(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.Exit("no target specified", 1)
	}

	name := ctx.Args().First()
	args := ctx.Args().Tail()

	target, ok := config.GetTargetForAlias(name)
	if !ok {
		target = Target{
			Url:             name, // cmd is assumed to be a valid URL if no alias is assigned
			InvalidateAfter: config.DefaultInvalidateAfter,
		}
	}

	path, err := cache.GetTargetPath(target)
	if err != nil {
		return cli.Exit("failed to check cache for target: "+err.Error(), 1)
	}

	if path == "" {
		log.Debug("target not in cache, fetching")

		path, err = FetchToPath(target.Url)
		if err != nil {
			return cli.Exit("failed to fetch target: "+err.Error(), 1)
		}
		defer os.Remove(path)
	}

	cmd := exec.CommandContext(ctx.Context, path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	app := cli.App{
		Name:  "fan",
		Usage: "(F)etch (A)nd (R)un a script / executable",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Before: loadCache,
				Action: run,
				After:  cleanCache,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
	}
}
