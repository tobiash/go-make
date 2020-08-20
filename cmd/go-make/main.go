package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/tobiash/go-make/pkg/mk"
	"github.com/tobiash/go-make/pkg/mk/frontends/yamlfe"
	"github.com/tobiash/go-make/pkg/mk/shell"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func main() {
	wd, _ := os.Getwd()
	app := &cli.App{
		Name: "go-make",
		Action: func(c *cli.Context) error {
			mkfile, err := Makefile(c)
			if err != nil {
				return err
			}
			rules, err := mkfile.BuildRules()
			if err != nil {
				return err
			}
			m := mk.Make{
				Sum:   &mk.YamlSumStorageFile{Path: filepath.Join(c.Path("directory"), c.Path("sumfile")), Perm: 0644},
				Rules: rules,
			}
			targets := make([]mk.Target, c.NArg())
			for i := 0; i < c.Args().Len(); i++ {
				targets[i] = &mk.FileTarget{Dir: c.Path("directory"), Path: c.Args().Get(i)}
			}

			ctx := log.Logger.WithContext(context.Background())

			return m.Make(&shell.ShellExecutor{
				Dir: c.Path("directory"),
			}, ctx, targets...)
		},
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "directory",
				Aliases: []string{"C"},
				Value:   wd,
			},
			&cli.PathFlag{
				Name:    "file",
				Aliases: []string{"makefile", "f"},
				Value:   "go-make.yaml",
			},
			&cli.PathFlag{
				Name:    "sumfile",
				Aliases: []string{"s"},
				Value:   "go-make.sum",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("application failed")
	}
}

func Makefile(c *cli.Context) (*yamlfe.Makefile, error) {
	mkfile := yamlfe.Makefile{}
	f, err := os.Open(filepath.Join(c.Path("directory"), c.Path("file")))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	if err = mkfile.Parse(f); err != nil {
		return nil, err
	}
	return &mkfile, nil
}
