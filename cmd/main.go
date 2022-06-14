package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/datahearth/doggo-fetcher/pkg"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:  "dg",
	Usage: "I bring you your latest Golang version with ease and efficiency (like a stick) !",
	Description: `Doggo-fetcher is a utility tool that manage for you your Golang version.
You can select a specific go version or even set version for directories.`,
	EnableBashCompletion: true,
	Authors: []*cli.Author{{
		Name:  "Antoine <DataHearth> Langlois",
		Email: "antoine.l@antoine-langlois.net",
	}},
	Suggest: true,
	Version: "0.1.0",
	Commands: []*cli.Command{
		{
			Name:        "install",
			Usage:       "Download a given release",
			Description: "Download a given release and set it as first release to be use",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "keep-release",
					Usage: "Downloaded release archive will be stored",
				},
				&cli.BoolFlag{
					Name:    "latest",
					Aliases: []string{"lts"},
					Usage:   "Download and install the latest release",
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
						return errors.New("a release is required if \"--latest\" is not passed")
					}
					release = "lts"
				} else {
					release = ctx.Args().First()
				}

				ghTags := pkg.NewTags(release, ctx.Context)
				release, err := ghTags.CheckReleaseExists(ctx.Bool("beta"), ctx.Bool("rc"))
				if err != nil {
					return err
				}

				fmt.Printf("release: %v\n", release)

				return nil
			},
		},
	},
	Action: func(c *cli.Context) error {
		fmt.Println("Hello friend!")
		return nil
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
