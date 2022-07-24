package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/datahearth/doggo-fetcher/pkg"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	logger = logrus.New()
	app    = &cli.App{
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
			useCmd,
			initCmd,
			uninstallCmd,
			lsCmd,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose mode"},
		},
	}
	useCmd = &cli.Command{
		Name:  "use",
		Usage: "Set a specific golang version",
		Description: `Use a specific golang version as primary golang binary for the user.
	If the version is not already downloaded, it'll downloaded and installed automatically.`,
		Flags: []cli.Flag{
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
			if ctx.Bool("verbose") {
				logger.Level = logrus.DebugLevel
			}

			var release string
			if ctx.NArg() == 0 {
				if !ctx.Bool("latest") {
					return errors.New("a release is required if \"--latest|--lts\" is not passed")
				}
				release = pkg.LTS
			} else {
				release = ctx.Args().First()
			}

			localFolder := strings.Replace(pkg.DGF_FOLDER, "~", os.Getenv("HOME"), 1)
			fi, err := os.Stat(localFolder)
			if err != nil {
				if !os.IsNotExist(err) {
					return err
				}

				if err := os.MkdirAll(localFolder, 0755); err != nil {
					return err
				}
			} else {
				if !fi.IsDir() {
					if err := os.RemoveAll(localFolder); err != nil {
						return err
					}
					if err := os.MkdirAll(localFolder, 0755); err != nil {
						return err
					}
				}
			}

			release, err = pkg.NewTags(release, ctx.Context).GetRelease(ctx.Bool("beta"), ctx.Bool("rc"))
			if err != nil {
				return err
			}

			hash, err := pkg.NewHash(localFolder)
			if err != nil {
				return err
			}
			r := pkg.NewRelease(release, filepath.Join(localFolder, release))

			if err := r.CheckReleaseExists(); err != nil {
				if err != pkg.ErrReleaseNotFound {
					return err
				}

				logger.Info("Release not found, downloading...")
				if err := r.DownloadRelease(); err != nil {
					return err
				}

				logger.Info("Release downloaded, installing...")
				if err := r.ExtractRelease(); err != nil {
					return err
				}

				logger.Debug("Release installed, generating hash...")
				h, err := hash.GetFolderHash(r.GetReleaseFolder())
				if err != nil {
					return err
				}
				if err := hash.AddHash(r.GetReleaseFolder(), h); err != nil {
					return err
				}
			} else {
				logger.Debug("Release found, checking hash...")
				h, err := hash.GetFolderHash(r.GetReleaseFolder())
				if err != nil {
					return err
				}

				if err := hash.CompareReleaseHash(r.GetReleaseFolder(), h); err != nil {
					if err == pkg.ErrHashNotFound {
						logger.Warnln("Hash not found in hash table, adding...")
						if err := hash.AddHash(r.GetReleaseFolder(), h); err != nil {
							return err
						}
					} else if err == pkg.ErrHashInvalid {
						logger.Warnln("Hash invalid, replacing...")
						if err := hash.ReplaceHash(r.GetReleaseFolder(), h); err != nil {
							return err
						}
					} else {
						return err
					}
				}
			}

			logger.Info("Setting golang binary")
			if err := pkg.UpdateRelease(r.GetReleaseFolder()); err != nil {
				return err
			}

			logger.Info("Everything done!")
			return nil
		},
	}
	initCmd = &cli.Command{
		Name:  "init",
		Usage: "Initialize doggofetcher",
		Description: `Initialize doggofetcher by adding a custom path inside your sheel configuration file.
It also creates a folder at "~/.local/doggofetcher" to store your custom golang binaries.`,
		Action: func(ctx *cli.Context) error {
			if ctx.Bool("verbose") {
				logger.Level = logrus.DebugLevel
			}

			logger.Infoln("Initializing doggofetcher...")
			return pkg.Init()
		},
	}
	uninstallCmd = &cli.Command{
		Name:  "uninstall",
		Usage: "Uninstall golang release",
		Description: `Uninstall current golang release.
It will remove the folder at "~/.local/doggofetcher/go"`,
		Action: func(ctx *cli.Context) error {
			return os.RemoveAll(filepath.Join(strings.Replace(pkg.DGF_FOLDER, "~", os.Getenv("HOME"), 1), "go"))
		},
	}
	lsCmd = &cli.Command{
		Name:  "ls",
		Usage: "List available releases",
		Description: `List available releases.
It will list all available releases in the "~/.local/doggofetcher/go*" folder.`,
		Action: func(ctx *cli.Context) error {
			if ctx.Bool("verbose") {
				logger.Level = logrus.DebugLevel
			}

			items, err := os.ReadDir(strings.Replace(pkg.DGF_FOLDER, "~", os.Getenv("HOME"), 1))
			if err != nil {
				return err
			}

			for _, i := range items {
				if i.IsDir() && i.Name() != "go" {
					fmt.Println(i.Name())
				}
			}

			return nil
		},
	}
)

// todo: ls (list downloaded-releases)
// todo: ls-remote (list remote releases)
// todo: exec (without setting release=
// todo: alias (set alias for a release)
// todo: run-bg (check each 10mins)
// todo: auto-use (automatically switch when changing directory)
// todo: remove (remove a release from local folder)

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error(err.Error())
	}
}
