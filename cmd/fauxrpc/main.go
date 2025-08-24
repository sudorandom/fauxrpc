package main

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/alecthomas/kong"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
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
	Generate GenerateCmd `cmd:"generate" help:"Generate fake data"`
	Registry RegistryCmd `cmd:"" help:"Contains registry commands"`
	Curl     CurlCmd     `cmd:"" help:"Make requests with fake data"`
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
	switch cli.LogLevel {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "info":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		ctx.Fatalf("unknown log level: %s", cli.LogLevel)
	}
	ctx.FatalIfErrorf(ctx.Run(&cli.Globals))
}

func fullVersion() string {
	var b strings.Builder
	version, commit, date := getVersionInfo()
	b.WriteString(version)
	if commit != "" {
		b.WriteString(fmt.Sprintf(" (%s)", commit))
	}
	if date != "" {
		b.WriteString(fmt.Sprintf(" @%s", commit))
	}
	return b.String()
}

func getVersionInfo() (string, string, string) {
	currnetVersion := version
	currentCommit := commit
	currentDate := date

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			currnetVersion = info.Main.Version
		}

		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) >= 7 {
					currentCommit = setting.Value[:7] // Short commit hash
				} else {
					currentCommit = setting.Value
				}
			case "vcs.time":
				currentDate = setting.Value
			}
		}
	}

	return currnetVersion, currentCommit, currentDate
}
