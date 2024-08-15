package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

type Globals struct {
	LogLevel string      `short:"l" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info"`
	Version  VersionFlag `name:"version" help:"Print version information and quit"`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type CLI struct {
	Globals

	Run RunCmd `cmd:"" help:"Run the FauxRPC server"`
}

func main() {
	cli := CLI{
		Globals: Globals{
			Version: VersionFlag("0.0.1"),
		},
	}

	ctx := kong.Parse(&cli,
		kong.Name("fauxrpc"),
		kong.Description("A fake gRPC/gRPC-Web/Connect/REST server powered by protobuf."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": "0.0.1",
		})
	ctx.FatalIfErrorf(ctx.Run(&cli.Globals))
}

type staticNames struct {
	names []string
}

func (n *staticNames) Names() []string {
	return n.names
}
