package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alextacho/forge/internal/guide"
	"github.com/alextacho/forge/internal/mcp"
	"github.com/alextacho/forge/internal/project"
	"github.com/alextacho/forge/internal/version"
	"golang.org/x/term"
)

type Environment struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Getwd      func() (string, error)
	IsTerminal func(io.Reader) bool
}

type options struct {
	command   string
	name      string
	yes       bool
	noClobber bool
}

func Run(args []string, env Environment) int {
	opts, err := parse(args)
	if err != nil {
		fmt.Fprintln(env.Stderr, "Error:", err)
		return 2
	}
	if opts.command == "help" {
		printUsage(env.Stdout)
		return 0
	}
	if opts.command == "version" {
		fmt.Fprintln(env.Stdout, version.String())
		return 0
	}
	if opts.command == "instructions" {
		fmt.Fprintln(env.Stdout, guide.Instructions)
		return 0
	}
	if opts.command == "mcp" {
		return mcp.Serve(mcp.Server{
			Stdin:  env.Stdin,
			Stdout: env.Stdout,
			Stderr: env.Stderr,
			Getwd:  env.Getwd,
		})
	}
	workingDirectory, err := env.Getwd()
	if err != nil {
		fmt.Fprintln(env.Stderr, "Error:", err)
		return 1
	}
	found, err := project.Find(workingDirectory)
	if err != nil {
		fmt.Fprintln(env.Stderr, "Error:", err)
		return 1
	}
	switch opts.command {
	case "save":
		err = runSave(found, opts, env)
	case "load":
		err = runLoad(found, opts, env)
	case "reset":
		err = runReset(found, opts, env)
	case "status":
		err = runStatus(found, env)
	}
	if err != nil {
		fmt.Fprintln(env.Stderr, "Error:", err)
		return 1
	}
	return 0
}

func parse(args []string) (options, error) {
	if len(args) == 0 {
		return options{command: "help"}, nil
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		return options{command: "help"}, nil
	}
	if args[0] == "version" || args[0] == "--version" || args[0] == "-v" {
		if len(args) != 1 {
			return options{}, fmt.Errorf("version accepts no arguments")
		}
		return options{command: "version"}, nil
	}
	if args[0] == "instructions" {
		if len(args) != 1 {
			return options{}, fmt.Errorf("instructions accepts no arguments")
		}
		return options{command: "instructions"}, nil
	}
	if args[0] == "mcp" {
		if len(args) != 1 {
			return options{}, fmt.Errorf("mcp accepts no arguments")
		}
		return options{command: "mcp"}, nil
	}

	switch args[0] {
	case "status":
		flags, positional, err := parseCommandArguments(args[1:], false)
		if err != nil {
			return options{}, err
		}
		if flags.yes {
			return options{}, fmt.Errorf("status does not accept --yes")
		}
		if len(positional) != 0 {
			return options{}, fmt.Errorf("status accepts no arguments")
		}
		return options{command: "status"}, nil
	case "save":
		flags, positional, err := parseCommandArguments(args[1:], true)
		if err != nil {
			return options{}, err
		}
		if len(positional) > 1 {
			return options{}, fmt.Errorf("save accepts at most one snapshot name")
		}
		name := "default"
		if len(positional) == 1 {
			name = positional[0]
		}
		return options{command: "save", name: name, yes: flags.yes, noClobber: flags.noClobber}, nil
	case "load":
		flags, positional, err := parseCommandArguments(args[1:], false)
		if err != nil {
			return options{}, err
		}
		if len(positional) > 1 {
			return options{}, fmt.Errorf("load accepts at most one snapshot name")
		}
		name := "default"
		if len(positional) == 1 {
			name = positional[0]
		}
		return options{command: "load", name: name, yes: flags.yes}, nil
	case "reset":
		flags, positional, err := parseCommandArguments(args[1:], false)
		if err != nil {
			return options{}, err
		}
		if len(positional) != 0 {
			return options{}, fmt.Errorf("reset accepts no arguments")
		}
		return options{command: "reset", yes: flags.yes}, nil
	default:
		return options{}, fmt.Errorf("unknown command %q", args[0])
	}
}

func parseCommandArguments(args []string, allowNoClobber bool) (options, []string, error) {
	var flags options
	var positional []string
	for _, arg := range args {
		switch arg {
		case "--yes":
			flags.yes = true
		case "--no-clobber":
			if !allowNoClobber {
				return options{}, nil, fmt.Errorf("flag --no-clobber is only valid for save")
			}
			flags.noClobber = true
		default:
			if strings.HasPrefix(arg, "-") {
				return options{}, nil, fmt.Errorf("unknown flag %q", arg)
			}
			positional = append(positional, arg)
		}
	}
	return flags, positional, nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, guide.Usage)
}

func approve(opts options, env Environment) error {
	if opts.yes {
		return nil
	}
	isTerminal := env.IsTerminal
	if isTerminal == nil {
		isTerminal = defaultIsTerminal
	}
	if !isTerminal(env.Stdin) {
		return fmt.Errorf("confirmation requires an interactive terminal; use --yes to proceed")
	}
	fmt.Fprint(env.Stdout, "Continue? [y/N] ")
	line, err := bufio.NewReader(env.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("read confirmation: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return nil
	default:
		return errorsDeclined
	}
}

var errorsDeclined = fmt.Errorf("operation cancelled")

func defaultIsTerminal(reader io.Reader) bool {
	file, ok := reader.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}
