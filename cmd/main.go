package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/datahearth/doggo-fetcher/pkg"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "dg",
	Usage: "I bring you your latest GoLang release with ease and efficiency (like a stick) !",
	Description: `Doggo-fetcher is a utility tool that manage for you your installed GoLang releases.
You can select a specific GoLang release or even set a specific one for directories.`,
	EnableBashCompletion: true,
	Authors: []*cli.Author{
		{
			Name:  "Antoine <DataHearth> Langlois",
			Email: "antoine.l@antoine-langlois.net",
		},
	},
	Suggest: true,
	Version: "0.1.0",
	Commands: []*cli.Command{
		{
			Name:        "install",
			Usage:       "Download a release",
			Description: "Download a release. If the given release is already stored, nothing will happen.",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "keep-release",
					Usage: "Downloaded release archive will be stored",
				},
				&cli.BoolFlag{
					Name:    "latest",
					Aliases: []string{"lts"},
					Usage:   "Use latest release",
				},
				&cli.BoolFlag{
					Name:  "rc",
					Usage: "Allow \"rc\" version to be fetched",
				},
				&cli.BoolFlag{
					Name:  "beta",
					Usage: "Allow \"beta\" version to be fetched",
				},
			},
			Action: func(ctx *cli.Context) error {
				var release string
				if ctx.NArg() == 0 {
					if !ctx.Bool("latest") {
						return errors.New("a release is required if \"--latest|--lts\" is not passed")
					}
					release = pkg.LTS
				} else {
					release = ctx.Args().First()
				}

				ghTags := pkg.NewTags(release, ctx.Context)
				release, err := ghTags.GetRelease(ctx.Bool("beta"), ctx.Bool("rc"))
				if err != nil {
					return err
				}

				dlRelease := pkg.NewDownload(release, ctx.Bool("keep-release"))
				releasePath, err := dlRelease.DownloadRelease()
				if err != nil {
					return err
				}

				extractedRelease, err := pkg.ExtractRelease(releasePath, release)
				if err != nil {
					return err
				}

				totalSum := sha256.New()
				err = filepath.Walk(extractedRelease, func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.Mode().IsRegular() || info.IsDir() {
						return nil
					}

					b, err := os.ReadFile(path)
					if err != nil {
						return err
					}
				})

				// todo: install release
				return nil
			},
		},
		{
			Name:  "use",
			Usage: "Set a specific golang version",
			Description: `Use a specific golang version as primary golang binary for the user.
	If the version is not already downloaded, it'll downloaded and installed automatically.`,
			Action: func(ctx *cli.Context) error {
				fmt.Println("to be implemented!")
				return nil
			},
		},
	},
	Action: func(c *cli.Context) error {
		fmt.Println("Hello friend!")
		return nil
	},
}

// todo: use (download if not present, install and set in path)
// todo: uninstall
// todo: ls (list installed-releases)
// todo: ls-remote (list remote releases)
// todo: exec (without setting release=
// todo: alias
// todo: run-bg (check each 10mins)
// todo: auto-use (automatically switch when changing directory)

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
