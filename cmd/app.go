package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"

	fan "github.com/joshmeranda/fan/pkg"
	"github.com/joshmeranda/fan/pkg/cache"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var (
	log *slog.Logger

	fanCache cache.Cache
	config   Config
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

	target, err := fanCache.GetTargetForUrl(url)
	if errors.Is(err, cache.ErrNotFound) {
		log.Debug("target not in cache", "url", url)

		target = fan.Target{
			Url:             url,
			InvalidateAfter: config.DefaultInvalidateAfter,
		}

		target.Path, err = fan.Fetch(url)
		if err != nil {
			return cli.Exit("failed to fetch target: "+err.Error(), 1)
		}
		defer os.Remove(target.Path)

		if err := fanCache.AddTarget(target); err != nil {
			return cli.Exit("failed to add target to cache: "+err.Error(), 1)
		}

		target, err = fanCache.GetTargetForUrl(url)
		if err != nil {
			return cli.Exit("failed to get target from cache: "+err.Error(), 1)
		}
	} else if err != nil {
		return cli.Exit("failed to check cache for target: "+err.Error(), 1)
	}

	if err := target.Run(ctx.Context, args); err != nil {
		return cli.Exit(err, 1)
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

	if unaliased, found := config.Aliases[target.Url]; found {
		target.Url = unaliased
	}

	targetPath := path.Join(config.CacheDir, fmt.Sprintf("%d", target.Hash()))

	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("failed to delete cached target: %w", err)
	}

	return nil
}

func actionAlias(ctx *cli.Context) error {
	if ctx.NArg() != 2 {
		return cli.Exit("expected alais and url", 1)
	}

	alias := ctx.Args().First()
	url := ctx.Args().Get(1)

	config.Aliases[alias] = url

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
				Name:   "alias",
				Usage:  "add an alias for a target url",
				Before: setup,
				Action: actionAlias,
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
