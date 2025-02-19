package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type sequenceDiagramCmd struct {
	plantumlmixin
	endpointFormat string
	appFormat      string
	title          string
	output         string
	endpointsFlag  []string
	appsFlag       []string
	blackboxesFlag map[string]string
	group          string
}

func (p *sequenceDiagramCmd) Name() string            { return "sd" }
func (p *sequenceDiagramCmd) RequireSyslModule() bool { return true }

func (p *sequenceDiagramCmd) Configure(app *kingpin.Application) *kingpin.CmdClause {

	p.blackboxesFlag = map[string]string{}
	p.endpointsFlag = []string{}
	p.appsFlag = []string{}

	cmd := app.Command(p.Name(), "Generate Sequence Diagram")

	cmd.Flag("endpoint_format",
		"Specify the format string for sequence diagram endpoints. May include "+
			"%(epname), %(eplongname) and %(@foo) for attribute foo (default: %(epname))",
	).Default("%(epname)").StringVar(&p.endpointFormat)

	cmd.Flag("app_format",
		"Specify the format string for sequence diagram participants. "+
			"May include %%(appname) and %%(@foo) for attribute foo (default: %(appname))",
	).Default("%(appname)").StringVar(&p.appFormat)

	cmd.Flag("title", "diagram title").Short('t').StringVar(&p.title)

	p.plantumlmixin.AddFlag(cmd)

	cmd.Flag("output",
		"output file (default: %(epname).png)",
	).Default("%(epname).png").Short('o').StringVar(&p.output)

	cmd.Flag("endpoint",
		"Include endpoint in sequence diagram",
	).Short('s').StringsVar(&p.endpointsFlag)

	cmd.Flag("app",
		"Include all endpoints for app in sequence diagram (currently "+
			"only works with templated --output). Use SYSL_SD_FILTERS env (a "+
			"comma-list of shell globs) to limit the diagrams generated",
	).Short('a').StringsVar(&p.appsFlag)

	cmd.Flag("blackbox",
		"Input blackboxes in the format App <- Endpoint=Some description, "+
			"repeat '-b App <- Endpoint=Some description' to set multiple blackboxes",
	).Short('b').StringMapVar(&p.blackboxesFlag)

	cmd.Flag("groupby", "Enter the groupby attribute (apps having "+
		"the same attribute value are grouped together in one box").Short('g').StringVar(&p.group)

	EnsureFlagsNonEmpty(cmd)
	return cmd
}

func (p *sequenceDiagramCmd) Execute(args ExecuteArgs) error {

	sequenceParams := &CmdContextParamSeqgen{
		endpointFormat: p.endpointFormat,
		appFormat:      p.appFormat,
		title:          p.title,
		output:         p.output,
		endpointsFlag:  p.endpointsFlag,
		appsFlag:       p.appsFlag,
		blackboxesFlag: p.blackboxesFlag,
		group:          p.group,
	}

	result, err := DoConstructSequenceDiagrams(sequenceParams, args.Module, args.Logger)
	if err != nil {
		return err
	}
	return p.GenerateFromMap(result, args.Filesystem)
}
