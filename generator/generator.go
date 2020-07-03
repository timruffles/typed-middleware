package generator

import (
	"bytes"
	"fmt"
	"go/types"
	"strings"

	"github.com/dave/jennifer/jen"
)

const thisPackageName = "github.plaid.com/plaid/typedmiddleware"

func Generate(packagePath string, parsed *targetStackParsed) (*bytes.Buffer, error) {
	suffixedTargetName := func(s string) string {
		return parsed.obj.Name() + s
	}

	f := jen.NewFilePath(packagePath)

	stackInterfaceName := suffixedTargetName("Stack")

	/*
		type Stack interface {
			Run(http.Request) (< target interface >, error)
		}
	*/
	runSignature := jen.Id("Run").Params(
		jen.Id("req").Op("*").
			Qual("net/http", "Request"),
	).Params(
		jen.Id(parsed.obj.Name()),
		jen.Op("*").Qual(thisPackageName, "MiddlewareResponse"),
	)
	f.Type().Id(stackInterfaceName).Interface(
		runSignature,
	)

	// constructor for implementation struct
	/*
		func NewStack(
			<initializer>
		) *< result struct > {
			return &< struct > {
				// < fields >
			}
		}
	*/
	implementationStructName := suffixedTargetName("StackImpl")
	implementationParams, embeddedMiddleware, structInitialisers := generateImplementationComponents(parsed)

	f.Func().Id("New" + suffixedTargetName("Stack")).
		Params(implementationParams...).
		Add(
			jen.Op("*").Id(implementationStructName),
		).
		Block(
			jen.Return(
				jen.Op("&").Id(implementationStructName).
					Values(structInitialisers),
			),
		)

	// implementation struct
	/*
		type <struct> struct {
			<fields>
		}
	*/
	f.Type().Id(implementationStructName).
		Struct(embeddedMiddleware...)

	// Run(...) method on implementation struct
	implStatements := generateRunBody(parsed)

	f.Func().Params(
		jen.Id("s").Op("*").Id(implementationStructName),
	).Add(runSignature).Block(
		implStatements...
	)

	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func generateImplementationComponents(parsed *targetStackParsed) ([]jen.Code, []jen.Code, jen.Dict) {
	var implementationParams []jen.Code
	var embeddedMiddleware []jen.Code
	structInitialisers := make(jen.Dict)
	for _, m := range parsed.byId {
		name := m.implementation.Name()

		// for generated struct
		embeddedMiddleware = append(embeddedMiddleware,
			jen.Qual(m.implementation.Pkg().Path(), name),
		)

		// for constructor
		implementationParams = append(implementationParams,
			jen.Id(toParamName(name)).
				Qual(m.implementation.Pkg().Path(), name),
		)
		structInitialisers[jen.Id(name)] = jen.Id(toParamName(name))
	}
	return implementationParams, embeddedMiddleware, structInitialisers
}

func generateRunBody(parsed *targetStackParsed) []jen.Code {
	var body []jen.Code
	for _, id := range parsed.middlewareOrder {
		mw := parsed.byId[id]

		runParams := []jen.Code{
			jen.Id("req"),
		}
		if mw.runHasDependencies() {
			runParams = []jen.Code{
				jen.Id("req"),
				jen.Id("s"),
			}
		}

		stanza := []jen.Code{
			// result, err := s.xxMiddleware.Run(r)
			jen.List(
				jen.Id("result"),
				jen.Id("err"),
			).Op(":=").
				Id("s").
				Dot(mw.implementation.Name()).
				Dot("Run").
				Call(runParams...),
			// if result != nil: result
			jen.If(
				jen.Id("result").
					Op("!=").
					Nil(),
			).Block(
				jen.Return(jen.List(
					jen.Nil(),
					jen.Id("result"),
				)),
			),
			// if result != nil: err
			jen.If(
				jen.Id("err").
					Op("!=").
					Nil(),
			).Block(
				jen.Return(jen.List(
					jen.Nil(),
					jen.Qual(thisPackageName, "NewErrorResult").
						Call(jen.Id("err")),
				)),
			),
		}

		body = append(body, stanza...)
	}
	body = append(body,
		jen.Return(
			jen.Id("s"),
			jen.Nil(),
		),
	)
	return body
}

func toParamName(name string) string {
	if (len(name) < 2) {
		return name
	}
	return fmt.Sprintf("%s%s", strings.ToLower(name[:1]), name[1:])
}


func toTargetName(basename string) string {
	name := strings.TrimSuffix(basename, ".go")
	return fmt.Sprintf("%s_middleware.go", name)
}

func objToQual(obj types.Object) *jen.Statement {
	return jen.Qual(obj.Pkg().Path(), obj.Name())
}
