package cli

import (
	"flag"
	"fmt"

	"github.com/N-Hoque/static-file-server/internal/cli/help"
	"github.com/N-Hoque/static-file-server/internal/cli/version"
	"github.com/N-Hoque/static-file-server/internal/config"
	"github.com/N-Hoque/static-file-server/internal/server"
)

var option struct {
	configFile  string
	helpFlag    bool
	versionFlag bool
}

// Assignments used to simplify testing.
var (
	selectRoutine   = selectionRoutine
	unknownArgsFunc = unknownArgs
	runServerFunc   = server.Run
	runHelpFunc     = help.Run
	runVersionFunc  = version.Run
	loadConfig      = config.Load
)

func init() {
	setupFlags()
}

func setupFlags() {
	flag.StringVar(&option.configFile, "config", "", "")
	flag.StringVar(&option.configFile, "c", "", "")
	flag.BoolVar(&option.helpFlag, "help", false, "")
	flag.BoolVar(&option.helpFlag, "h", false, "")
	flag.BoolVar(&option.versionFlag, "version", false, "")
	flag.BoolVar(&option.versionFlag, "v", false, "")
}

// Execute CLI arguments.
func Execute() (err error) {
	// Parse flag options, then parse commands arguments.
	flag.Parse()
	args := Parse(flag.Args())

	job := selectRoutine(args)
	return job()
}

func selectionRoutine(args Args) func() error {
	switch {

	// serve help
	// serve --help
	// serve -h
	case args.Matches("help") || option.helpFlag:
		return runHelpFunc

	// serve version
	// serve --version
	// serve -v
	case args.Matches("version") || option.versionFlag:
		return runVersionFunc

	// serve
	case args.Matches():
		return withConfig(runServerFunc)

	// Unknown arguments.
	default:
		return unknownArgsFunc(args)
	}
}

func unknownArgs(args Args) func() error {
	return func() error {
		return fmt.Errorf(
			"unknown arguments provided [%v], try: 'help'",
			args,
		)
	}
}

func withConfig(routine func(*config.Config) error) func() error {
	return func() error {
		config, err := loadConfig(option.configFile)
		if err != nil {
			return err
		}
		return routine(config)
	}
}
