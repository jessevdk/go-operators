package main

import (
	"./types"
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

func getSourceDir(args []string) string {
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "Please provide only a single directory to parse\n")
		os.Exit(1)
	}

	if len(args) == 0 {
		dirname, err := os.Getwd()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to obtain current working directory\n")
			os.Exit(1)
		}

		return dirname
	}

	return args[0]
}

func main() {
	var opts struct {
		Output  string `short:"o" long:"output" description:"Output directory (required)" required:"true"`
		Verbose bool   `short:"v" long:"verbose" description:"Enable verbose mode"`
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

	dirname := getSourceDir(args)

	p, err := build.ImportDir(dirname, 0)

	if err != nil {
		panic(err)
	}

	fs := token.NewFileSet()
	afs := make([]*ast.File, 0, len(p.GoFiles))

	for _, f := range p.GoFiles {
		af, err := parser.ParseFile(fs, path.Join(dirname, f), nil, 0)

		if err != nil {
			panic(err)
		}

		afs = append(afs, af)
	}

	pp, err := types.Check(dirname, fs, afs)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while parsing: %s\n", err)
		os.Exit(1)
	}

	overloads := pp.Overloads()

	for i, f := range afs {
		f = replace(func(node ast.Node) ast.Node {
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

			// Create function call expression
			call := &ast.CallExpr{
				Fun: sel,
				Args: []ast.Expr{
					info.Oper,
				},
			}

			return call
		}, f).(*ast.File)

		fn := path.Join(opts.Output, p.GoFiles[i])

		of, err := os.Create(fn)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create output file: %s\n", err)
			os.Exit(1)
		}

		defer of.Close()

		if opts.Verbose {
			fmt.Println(fn)
		}

		format.Node(of, fs, f)
	}
}
