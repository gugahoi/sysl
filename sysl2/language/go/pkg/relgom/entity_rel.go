package relgom

import (
	. "github.com/anz-bank/sysl/sysl2/codegen/golang" //nolint:golint,stylecheck
)

func (g *entityGenerator) goAppendRelationDecls(decls []Decl) []Decl {
	decls = append(decls,
		g.goRelationDecl(),
	)
	if g.haveKeys {
		decls = append(decls,
			g.goRelationInsertMethod(),
			g.goRelationUpdateMethod(),
			g.goRelationDeleteMethod(),
			g.goRelationLookupMethod(),
		)
	}
	return decls
}

// // ${relation} represents a set of ${ename}.
// type ${relation} struct {
//     ${relation}Data
//     model PetShopModel
// }
func (g *entityGenerator) goRelationDecl() Decl {
	relation := g.tname + "Relation"
	return Types(TypeSpec{
		Name: *ExportedID(relation),
		Type: Struct(
			Field{Type: I(g.relationDataName)},
			Field{Names: Idents("model"), Type: I(g.modelName)},
		),
	}).WithDoc(Commentf("// %s represents a set of %s.", relation, g.tname))
}

func (g *entityGenerator) relationMethod(f func(recv string, recvDot dotter) FuncDecl) *FuncDecl {
	return method("r", I(g.tname+"Relation"), f)
}

func (g *entityGenerator) goRelationInsertMethod() Decl {
	return g.relationMethod(func(recv string, recvDot dotter) FuncDecl {
		entity := I("t")
		modelSet := Dot(Call(recvDot("model", "Get"+g.tname)), "set")

		innerStmts := []Stmt{}
		if len(g.autoinc) > 0 {
			for _, nt := range g.namedAttrs {
				if g.autoinc.Contains(nt.Name) {
					t := g.typeInfoForSyslType(nt.Type).param
					innerStmts = append(innerStmts,
						Assign(Dot(I("t"), NonExportedName(nt.Name)))("=")(Call(t, I("id"))),
					)
				}
			}
		}
		innerStmts = append(innerStmts,
			Init("set", "_")(Call(Dot(modelSet, "Set"), Dot(entity, g.pkName), entity)),
			Return(I("set"), Nil()),
		)

		outerStmts := []Stmt{}
		var model Expr
		if len(g.autoinc) > 0 {
			outerStmts = append(outerStmts, Init("model", "id")(Call(recvDot("model", "newID"))))
			model = I("model")
		} else {
			model = recvDot("model")
		}
		outerStmts = append(outerStmts,
			Return(AddrOf(Composite(I(g.tname+"Builder"),
				KV(I("model"), model),
				KV(I("apply"), FuncT(*g.applyFuncType(), innerStmts...)),
			))),
		)

		return FuncDecl{
			Doc:  Comments(Commentf("// Insert creates a builder to insert a new %s.", g.tname)),
			Name: *I("Insert"),
			Type: FuncType{Results: Fields(Field{Type: Star(I(g.tname + "Builder"))})},
			Body: Block(outerStmts...),
		}
	})
}

func (g *entityGenerator) goRelationUpdateMethod() Decl {
	return g.relationMethod(func(recv string, recvDot dotter) FuncDecl {
		entity := I("t")
		modelSet := Dot(Call(recvDot("model", "Get"+g.tname)), "set")

		return FuncDecl{
			Doc:  Comments(Commentf("// Update creates a builder to update t in the model.")),
			Name: *I("Update"),
			Type: FuncType{
				Params:  *Fields(Field{Names: Idents("t"), Type: I(g.tname)}),
				Results: Fields(Field{Type: Star(I(g.tname + "Builder"))}),
			},
			Body: Block(
				Init("b")(
					AddrOf(Composite(I(g.tname+"Builder"),
						KV(I(g.dataName), Star(Dot(I("t"), g.dataName))),
						KV(I("model"), recvDot("model")),
						KV(I("apply"), FuncT(*g.applyFuncType(),
							Init("set", "_")(Call(Dot(modelSet, "Set"), Dot(entity, g.pkName), entity)),
							Return(I("set"), Nil()),
						)),
					)),
				),
				CallStmt(I("copy"),
					Slice(Dot(I("b"), "mask")),
					Dot(NonExportedID(g.tname+"StaticMetadata"), "PKMask"),
				),
				Return(I("b")),
			),
		}
	})
}

// // Delete deletes t from the model.
// func (r ${X(.name)}Relation) Delete(t ${X(.name)}) (${.model.name}, error) {
//     set, _ := r.model.Get${X(.name)}().set.Del(t.${u(.name)}PK)
//     relations, _ := r.model.relations.Set(${u(.name)}Key, ${u(.name)}RelationData{set: set})
//     return ${.model.name}{relations: relations}, nil
// }
func (g *entityGenerator) goRelationDeleteMethod() Decl {
	return g.relationMethod(func(recv string, recvDot dotter) FuncDecl {
		entity := I("t")
		modelSet := Dot(Call(recvDot("model", "Get"+g.tname)), "set")

		return FuncDecl{
			Doc:  Comments(Commentf("// Delete deletes t from the model.")),
			Name: *I("Delete"),
			Type: FuncType{
				Params: *Fields(Field{Names: Idents("t"), Type: I(g.tname)}),
				Results: Fields(
					Field{Type: I(g.modelName)},
					Field{Type: I("error")},
				),
			},
			Body: Block(
				Init("set", "_")(Call(Dot(modelSet, "Del"), Dot(entity, g.pkName))),
				Init("relations", "_")(
					Call(
						recvDot("model", "relations", "Set"),
						NonExportedID(g.tname+"Key"),
						Composite(I(g.relationDataName), KV(I("set"), I("set"))),
					),
				),
				Return(Composite(I(g.modelName), KV(I("relations"), I("relations"))), Nil()),
			),
		}
	})
}

func (g *entityGenerator) goRelationLookupMethod() Decl {
	return g.relationMethod(func(recv string, recvDot dotter) FuncDecl {
		fields := []Field{}
		kvs := []Expr{}
		for _, nt := range g.namedAttrs {
			if g.pkey.Contains(nt.Name) {
				fname := NonExportedName(nt.Name)
				fields = append(fields, Field{
					Names: Idents(fname),
					Type:  g.typeInfoForSyslType(nt.Type).param,
				})
				kvs = append(kvs, KV(I(fname), I(fname)))
			}
		}

		return FuncDecl{
			Doc:  Comments(Commentf("// Lookup searches %s by primary key.", g.tname)),
			Name: *I("Lookup"),
			Type: FuncType{
				Params:  *Fields(fields...),
				Results: Fields(Field{Type: I(g.tname)}, Field{Type: I("bool")}),
			},
			Body: Block(
				If(
					Init("t", "has")(Call(recvDot("set", "Get"),
						Composite(I(g.pkName), kvs...),
					)),
					I("has"),
					Return(
						Composite(I(g.tname),
							KV(I(g.dataName), Assert(I("t"), Star(I(g.dataName)))),
							KV(I("model"), recvDot("model")),
						),
						I("true"),
					),
				),
				Return(Composite(I(g.tname)), I("false")),
			),
		}
	})
}
