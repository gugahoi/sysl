package swagger

import (
	"github.com/sirupsen/logrus"

	"github.com/go-openapi/spec"
)

type Param struct {
	Field
	In string
}

type Parameters struct {
	items       map[string]Param
	insertOrder []string
}

func (p *Parameters) Add(param Param) {
	if p.items == nil {
		p.items = map[string]Param{}
	}
	if _, found := p.items[param.Name]; !found {
		p.insertOrder = append(p.insertOrder, param.Name)

	}
	p.items[param.Name] = param
}

func (p Parameters) findParams(where string) []Param {
	res := []Param{}
	for _, name := range p.insertOrder {
		item := p.items[name]
		if item.In == where {
			res = append(res, item)
		}
	}
	return res
}
func (p Parameters) QueryParams() []Param {
	return p.findParams("query")
}
func (p Parameters) HeaderParams() []Param {
	return p.findParams("header")
}
func (p Parameters) BodyParams() []Param {
	return p.findParams("body")
}
func (p Parameters) PathParams() []Param {
	return p.findParams("path")
}
func (p Parameters) CookieParams() []Param {
	return p.findParams("cookie")
}

func buildParam(param spec.Parameter, types TypeList, globals Parameters, logger *logrus.Logger) Param {
	paramTypeName := param.Type
	if paramTypeName == "" {
		if param.Schema != nil {
			paramTypeName = findReferencedType(*param.Schema, logger)
		} else if refURL := param.Ref.GetURL(); refURL != nil {
			refParamName := getReferenceFragment(refURL)
			if p, ok := globals.items[refParamName]; ok {
				return p
			}
			logger.Panicf("referenced parameter %s unknown", refParamName)
		}
	} else if paramTypeName != "string" {
		paramTypeName = mapSwaggerTypeAndFormatToType(param.Type, param.Format, logger)
	}
	ptype, found := types.Find(paramTypeName)
	if !found {
		logger.Warnf("Type '%s' unknown in param %s\n", paramTypeName, param.Name)
	}
	p := Param{Field: Field{
		Name:     param.Name,
		Optional: !param.Required,
		Type:     ptype,
	},
		In: param.In,
	}

	return p
}

func buildParameters(params []spec.Parameter, types TypeList,
	globals Parameters, baseParams Parameters, logger *logrus.Logger) Parameters {

	res := baseParams
	for _, param := range params {

		p := buildParam(param, types, globals, logger)
		res.Add(p)
	}

	return res
}

func buildGlobalParams(params map[string]spec.Parameter, types TypeList,
	logger *logrus.Logger) Parameters {

	res := Parameters{}
	for key, param := range params {

		p := buildParam(param, types, res, logger)
		p.Name = key
		res.Add(p)
	}
	return res
}
