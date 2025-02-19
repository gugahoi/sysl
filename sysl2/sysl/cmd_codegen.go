package main

import (
	"github.com/anz-bank/sysl/sysl2/sysl/syslutil"
	"gopkg.in/alecthomas/kingpin.v2"
)

type codegenCmd struct {
	CmdContextParamCodegen
	outDir  string
	appName string
}

func (p *codegenCmd) Name() string            { return "codegen" }
func (p *codegenCmd) RequireSyslModule() bool { return true }

func (p *codegenCmd) Configure(app *kingpin.Application) *kingpin.CmdClause {

	cmd := app.Command(p.Name(), "Generate code").Alias("gen")
	cmd.Flag("root-transform",
		"sysl root directory for input transform file (default: .)").
		Default(".").StringVar(&p.rootTransform)
	cmd.Flag("transform", "path to transform file from the root transform directory").Required().StringVar(&p.transform)
	cmd.Flag("grammar", "path to grammar file").Required().StringVar(&p.grammar)
	cmd.Flag("app-name",
		"name of the sysl app defined in sysl model."+
			" if there are multiple apps defined in sysl model,"+
			" code will be generated only for the given app").Required().StringVar(&p.appName)
	cmd.Flag("start", "start rule for the grammar").Default(".").StringVar(&p.start)
	cmd.Flag("outdir", "output directory").Default(".").StringVar(&p.outDir)
	EnsureFlagsNonEmpty(cmd)
	return cmd
}

func (p *codegenCmd) Execute(args ExecuteArgs) error {

	output, err := GenerateCode(&p.CmdContextParamCodegen, args.Module, p.appName, args.Filesystem, args.Logger)
	if err != nil {
		return err
	}
	return outputToFiles(output, syslutil.NewChrootFs(args.Filesystem, p.outDir))
}
