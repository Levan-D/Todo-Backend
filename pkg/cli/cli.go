package cli

import (
	"fmt"
	"github.com/Levan-D/Todo-Backend/pkg/config"
	"github.com/Levan-D/Todo-Backend/pkg/database/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

func Run(runServer func()) {
	app := &cli.App{
		Name:                   "todo",
		Usage:                  "usage description",
		UseShortOptionHandling: false,
		Action: func(context *cli.Context) error {
			runServer()
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "print current version",
				Action: func(c *cli.Context) error {
					fmt.Println(fmt.Sprintf("Todo platform version %s (compiled %s)", config.BuildVersion, config.BuildTime))
					fmt.Println("Email: todo@todo.lan Web: https://todo.lan")
					return nil
				},
			},
			{
				Name:    "migrate",
				Aliases: []string{"m"},
				Usage:   "database migration",
				Subcommands: []*cli.Command{
					{
						Name:  "up",
						Usage: "migrate up",
						Action: func(c *cli.Context) error {
							result := postgres.MigrationUp()
							fmt.Println(result)
							return nil
						},
					},
					{
						Name:  "down",
						Usage: "migrate down",
						Action: func(c *cli.Context) error {
							result := postgres.MigrationDown()
							fmt.Println(result)
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
