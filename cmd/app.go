package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	fan "github.com/joshmeranda/fan/pkg"
	"github.com/joshmeranda/fan/pkg/cache"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var (
	log *slog.Logger

	fanCache cache.Cache

	config Config
)

func setup(ctx *cli.Context) error {
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	switch configPath := ctx.String("config"); configPath {
	case "":
		config = DefaultConfig()
	default:
		data, err := os.ReadFile(configPath)

		if errors.Is(err, os.ErrNotExist) {
			config = DefaultConfig()
		} else if err != nil {
			return cli.Exit("failed to read config: "+err.Error(), 1)
		} else {
			if err := yaml.Unmarshal(data, &config); err != nil {
				return cli.Exit("failed to parse config: "+err.Error(), 1)
			}
		}
	}

	if config.CacheDir == "" {
		log.Debug("no cache specified, using noop cache")
		fanCache = cache.NewNoopCache()
	} else {
		fanCache = cache.NewDiskCache(config.CacheDir)
	}

	return nil
}

func actionRun(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return cli.Exit("no target specified", 1)
	}

	raw := ctx.Args().First()
	args := ctx.Args().Tail()

	url := raw

	if unaliased, found := config.Aliases[raw]; found {
		url = unaliased
	}

	_, executable, err := fanCache.GetTargetForUrl(url)
	if errors.Is(err, cache.ErrNotFound) {
		log.Debug("target not in cache, pulling...")

		target := fan.Target{
			Url:             url,
			InvalidateAfter: config.DefaultInvalidateAfter,
		}

		tmpExecutable, err := fan.Fetch(url)
		if err != nil {
			return cli.Exit("failed to fetch executable for target: "+err.Error(), 1)
		}

		if err := fanCache.AddTarget(target, tmpExecutable); err != nil {
			return cli.Exit("failed to add the target to the cache: "+err.Error(), 1)
		}

		if _, executable, err = fanCache.GetTargetForUrl(url); err != nil {
			return cli.Exit("failed to get new target from cache: "+err.Error(), 1)
		}
	} else if err != nil {
		return cli.Exit("failed to get target from cache: "+err.Error(), 1)
	}

	cmd := exec.CommandContext(ctx.Context, executable, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func actionCacheClean(ctx *cli.Context) error {
	if err := fanCache.Clean(); err != nil {
		return cli.Exit("failed to clean cache: "+err.Error(), 1)
	}

	return nil
}

func actionCacheInvalidate(ctx *cli.Context) error {
	all := ctx.Bool("all")

	if all {
		err := os.RemoveAll(config.CacheDir)
		if err != nil {
			return fmt.Errorf("failed to delete cached targets: %w", err)
		}

		return nil
	}

	if ctx.NArg() == 0 {
		return cli.Exit("no target specified", 1)
	}

	target := fan.Target{
		Url: ctx.Args().First(),
	}

	url := ctx.Args().First()
	if unaliased, found := config.Aliases[target.Url]; found {
		url = unaliased
	}

	if err := fanCache.InvalidateUrl(url); err != nil {
		return cli.Exit(fmt.Sprintf("could not invalidate '%s': %s", url, err), 1)
	}

	return nil
}

func actionAliasAdd(ctx *cli.Context) error {
	if ctx.NArg() != 2 {
		return cli.Exit("expected alais and url", 1)
	}

	alias := ctx.Args().First()
	url := ctx.Args().Get(1)

	config.Aliases[alias] = url

	if !ctx.Bool("force") {
		p, err := fan.Fetch(url)
		if err != nil {
			return fmt.Errorf("failed to fetch url '%s': %w", url, err)
		}
		defer os.Remove(p)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return cli.Exit("failed to marshal config: "+err.Error(), 1)
	}

	if err := os.WriteFile(ctx.String("config"), data, 0644); err != nil {
		return cli.Exit("failed to write config: "+err.Error(), 1)
	}

	return nil
}

func actionAliasList(ctx *cli.Context) error {
	maxAliasLen := 0
	for alias := range config.Aliases {
		if l := len(alias); l > maxAliasLen {
			maxAliasLen = l
		}
	}

	fmtString := fmt.Sprintf("%% %ds: %%s\n", maxAliasLen)

	for alias, url := range config.Aliases {
		fmt.Printf(fmtString, alias, url)
	}

	return nil
}

func actionAliasRemove(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return cli.Exit("expected at least 1 alias", 1)
	}

	for _, alias := range ctx.Args().Slice() {
		delete(config.Aliases, alias)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return cli.Exit("failed to marshal config: "+err.Error(), 1)
	}

	if err := os.WriteFile(ctx.String("config"), data, 0644); err != nil {
		return cli.Exit("failed to write config: "+err.Error(), 1)
	}

	return nil
}

func App() cli.App {
	return cli.App{
		Name:  "fan",
		Usage: "(F)etch (A)nd (R)un a script / executable",
		Commands: []*cli.Command{
			{
				Name:      "run",
				Usage:     "fetch and run a target",
				UsageText: "fan run <url|alias>",
				Before:    setup,
				Action:    actionRun,
			},
			{
				Name:   "cache",
				Before: setup,
				Subcommands: []*cli.Command{
					{
						Name:   "clean",
						Usage:  "check the cache for expired targets and remove them",
						Action: actionCacheClean,
					},
					{
						Name:      "invalidate",
						Usage:     "invalidate a target in the cache",
						UsageText: "fan cache invalidate --all\nfan cache invalidate <url|alias>...",
						Action:    actionCacheInvalidate,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "all",
								Usage: "invalidate all targets",
							},
						},
					},
				},
			},
			{
				Name:  "alias",
				Usage: "manage fan aliases",
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Usage:  "list all aliases",
						Before: setup,
						Action: actionAliasList,
					},
					{
						Name:   "remove",
						Usage:  "remove an alias",
						Before: setup,
						Action: actionAliasRemove,
					},
					{
						Name:   "add",
						Usage:  "add an alias",
						Before: setup,
						Action: actionAliasAdd,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "force",
								Usage: "do not fail if url cannot be reached",
							},
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: DefaultConfigPath(),
			},
		},
	}
}
