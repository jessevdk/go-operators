package main

import (
	"github.com/jessevdk/go-operators/types"
	"fmt"
	"github.com/jessevdk/go-flags"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
)

func getSources(args []string) (dirname string, files []string) {
	if len(args) == 0 {
		dirname, err := os.Getwd()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to obtain current working directory\n")
			os.Exit(1)
		}

		return dirname, nil
	}

	if len(args) == 1 {
		info, err := os.Stat(args[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check file `%s': %s\n", args[0], err)
			os.Exit(1)
		}

		if info.IsDir() {
			return args[0], nil
		} else {
			return path.Dir(args[0]), args
		}
	} else {
		dirname := ""

		for _, f := range args {
			info, err := os.Stat(f)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to check file `%s': %s\n", f, err)
				os.Exit(1)
			}

			if info.IsDir() {
				fmt.Fprintf(os.Stderr, "Specify files, not dirs...\n")
				os.Exit(1)
			}

			d := path.Dir(f)

			if len(dirname) == 0 {
				dirname = d
			} else if dirname != d {
				fmt.Fprintf(os.Stderr, "Cannot parse files in separate packages")
				os.Exit(1)
			}
		}

		return dirname, args
	}
}

func main() {
	var opts struct {
		Output  string `short:"o" long:"output" description:"Output directory (required)" required:"true"`
		Verbose bool   `short:"v" long:"verbose" description:"Enable verbose mode"`
		Package string `short:"p" long:"package" description:"Package name" default:"main"`
	}

	fp := flags.NewParser(&opts, flags.Default)
	fp.Usage = "--output OUTPUT_DIR [OPTIONS] SOURCE_DIR"

	args, err := fp.Parse()

	if err != nil {
		os.Exit(1)
	}

	if err := os.MkdirAll(opts.Output, 0766); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %s\n", err)
		os.Exit(1)
	}

	dirname, files := getSources(args)
	parseFiles := make([]string, 0, len(files))

	filesmap := make(map[string]bool)

	if len(dirname) != 0 {
		p, err := build.ImportDir(dirname, 0)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while importing build: %s\n", err)
			os.Exit(1)
		}

		for _, f := range p.GoFiles {
			parseFiles = append(parseFiles, path.Join(dirname, f))
			filesmap[f] = true
		}
	}

	for _, f := range files {
		if _, ok := filesmap[f]; !ok {
			filesmap[f] = true
			parseFiles = append(parseFiles, f)
		}
	}

	fs := token.NewFileSet()
	afs := make([]*ast.File, 0, len(parseFiles))

	afsmap := make(map[string]*ast.File)

	for _, f := range parseFiles {
		af, err := parser.ParseFile(fs, f, nil, 0)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while parsing AST: %s\n", err)
			os.Exit(1)
		}

		afsmap[f] = af

		afs = append(afs, af)
	}

	pkgname := path.Base(dirname)

	if len(pkgname) == 0 || len(dirname) == 0 {
		pkgname = opts.Package
	}

	pp, err := types.Check(pkgname, fs, afs)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while parsing: %s\n", err)
		os.Exit(1)
	}

	overloads := pp.Overloads()

	for i, f := range files {
		ff := afsmap[f]

		ff = replace(func(node ast.Node) ast.Node {
			expr, ok := node.(ast.Expr)

			if !ok {
				return node
			}

			info, ok := overloads[expr]

			if !ok {
				return node
			}

			sel := &ast.SelectorExpr{
				X:   info.Recv,
				Sel: ast.NewIdent(info.Func.Name()),
			}

			args := []ast.Expr{}

			if info.Oper != nil {
				args = append(args, info.Oper)
			}

			// Create function call expression
			call := &ast.CallExpr{
				Fun:  sel,
				Args: args,
			}

			return call
		}, ff).(*ast.File)

		fn := path.Join(opts.Output, f)
		err := os.MkdirAll(path.Join(opts.Output, path.Dir(files[i])), 0766)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create output directory: %s\n", err)
			os.Exit(1)
		}

		of, err := os.Create(fn)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create output file: %s\n", err)
			os.Exit(1)
		}

		defer of.Close()

		if opts.Verbose {
			fmt.Println(fn)
		}

		if err := format.Node(of, fs, ff); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write code: %s\n", err)
			os.Exit(1)
		}
	}
}
