package cmd

import (
	"errors"
	"log/slog"
	"os"

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

func actionCleanCache(ctx *cli.Context) error {
	if err := fanCache.Clean(); err != nil {
		return cli.Exit("failed to clean cache: "+err.Error(), 1)
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

func App() cli.App {
	return cli.App{
		Name:  "fan",
		Usage: "(F)etch (A)nd (R)un a script / executable",
		Commands: []*cli.Command{
			{
				Name:   "run",
				Before: setup,
				Action: actionRun,
			},
			{
				Name:   "clean-cache",
				Before: setup,
				Action: actionCleanCache,
			},
			{
				Name:   "alias",
				Before: setup,
				Action: actionAlias,
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
