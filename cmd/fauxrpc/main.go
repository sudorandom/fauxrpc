package main

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/alecthomas/kong"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Globals struct {
	LogLevel string      `short:"l" help:"Set the logging level (debug|info|warn|error)" default:"info"`
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

	Run      RunCmd      `cmd:"" help:"Run the FauxRPC server"`
	Stub     StubCmd     `cmd:"" help:"Contains stub commands"`
	Registry RegistryCmd `cmd:"" help:"Contains registry commands"`
}

func main() {
	version := fullVersion()
	cli := CLI{
		Globals: Globals{
			Version: VersionFlag(version),
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
			"version": version,
		})
	switch cli.Globals.LogLevel {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "info":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		ctx.Fatalf("unknown log level: %s", cli.Globals.LogLevel)
	}
	ctx.FatalIfErrorf(ctx.Run(&cli.Globals))
}

func fullVersion() string {
	return fmt.Sprintf("%s (%s) @ %s; %s", version, commit, date, runtime.Version())
}
