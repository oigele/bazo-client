package cli

import (
	"github.com/oigele/bazo-client/REST"
	"github.com/oigele/bazo-client/client"
	"github.com/urfave/cli"
)

func GetRestCommand() cli.Command {
	return cli.Command {
		Name:	"rest",
		Usage:	"start the REST service",
		Action:	func(c *cli.Context) error {
			client.Sync()
			REST.Init()
			return nil
		},
	}
}