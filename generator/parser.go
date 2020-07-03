package generator

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func Process(ps []*packages.Package, target string) (*targetStackParsed, error) {
	if len(ps) > 1 {
		return nil, fmt.Errorf("package specifier loaded more than one package")
	}
	if len(ps) == 0 {
		// this should never happen - load should fail above
		return nil, fmt.Errorf("package specified loaded no packages")
	}
	p := ps[0]

	// ensure package could be loaded - Scope will otherwise be nil
	for _, err := range p.Errors {
		if err.Error() == "cannot find package" {
			return nil, err
		}
	}

	parsed, err := parseMiddlewareStack(p.Types.Scope(), target)
	if err != nil {
		return nil, err
	}

	g, err := createGraph(parsed)
	if err != nil {
		return nil, err
	}
	parsed.middlewareOrder = topographicalSort(g.adjacency)
	parsed.byId = g.byId

	return parsed, nil
}

// this a target type specified by a user
type targetStackParsed struct {
	// the middleware stack interface, e.g type SomeHandlerMiddleware interface {}
	obj types.Object
	// scope the middleware was defined in
	scope *types.Scope
	stack []*middlewareParsed

	middlewareOrder []string
	byId            map[string]*middlewareParsed
}

// this is a parsed middleware, specified by embedding its interface
// in a target type
type middlewareParsed struct {
	// object - used to uniquely identify
	obj types.Object
	// the interface of the middleware
	interfaceT *types.Interface
	// the type implementing the interface
	implementation types.Object
	// the run method
	run *types.Func
	// nil if is a one element run function
	stackInterface *types.Interface
	// middleware's own dependency stack
	stack []*middlewareParsed
}

func (p middlewareParsed) runHasDependencies() bool {
	return len(p.stack) > 0
}

func parseMiddlewareStack(scope *types.Scope, target string) (*targetStackParsed, error) {
	// Lookup target and ensure it's an interface
	o := scope.Lookup(target)
	if o == nil {
		return nil, fmt.Errorf("could not find %s in package", target)
	}
	if !types.IsInterface(o.Type()) {
		return nil, fmt.Errorf("%s is not an interface", target)
	}
	ival, ok := o.Type().Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("%s could not resolve to interface type", target)
	}

	stack, err := parseStack(ival, middlewareCache{
		cache: make(map[string]*middlewareParsed),
	})
	if err != nil {
		return nil, err
	}
	return &targetStackParsed{
		obj:   o,
		scope: scope,
		stack: stack,
	}, nil
}

func parseStack(ival *types.Interface, middlewareByName middlewareCache) ([]*middlewareParsed, error) {
	stack := make([]*middlewareParsed, 0)
	// Look for embedded interfaces
	for i := 0; i < ival.NumEmbeddeds(); i++ {
		// 1. Get target interface  // TODO caching
		embedded := ival.EmbeddedType(i)
		named, ok := embedded.(*types.Named)
		if !ok {
			// anonymous (which we can't do anything with)
			continue
		}

		// if we have seen this middleware before, we're done
		fullName := types.ObjectString(named.Obj(), nil)
		mw, err := middlewareByName.get(fullName)
		if err != nil {
			return nil, err
		}
		if mw != nil {
			stack = append(stack, mw)
			continue
		}
		middlewareByName.mark(fullName)

		embeddedInterface, ok := named.Underlying().(*types.Interface)
		if !ok {
			// embedded struct, not relevant
			// TODO - could check if it's named xxxMiddleware and warn
			continue
		}

		// 2. Find a corresponding ${...}Middleware exported by same package
		embeddedName := named.Obj().Name()
		pkg := named.Obj().Pkg()
		nameOfStructImpl := fmt.Sprintf("%sMiddleware", embeddedName)
		implementingObj := pkg.Scope().Lookup(nameOfStructImpl)
		if implementingObj == nil {
			return nil, fmt.Errorf("Could not find %s to implement %s", nameOfStructImpl, embeddedName)
		}

		// 3. Check it implements the target
		foundTyp := implementingObj.Type()
		if err := ensureImplementsMiddlewareInterface(foundTyp, embeddedInterface, nameOfStructImpl, embeddedName); err != nil {
			return nil, err
		}

		// 4. Find Run() and validate, if it has a 2 method Run method
		methods := types.NewMethodSet(types.NewPointer(foundTyp))
		runMethod := getRunMethod(methods)
		if runMethod == nil {
			return nil, fmt.Errorf("%s had no Run() method", nameOfStructImpl)
		}

		// docs: 'its Type() is always a *Signature'
		sig := runMethod.Type().(*types.Signature)
		params := sig.Params()

		//	4.1. Check signature
		if params.Len() == 0 || params.Len() > 2 {
			return nil, fmt.Errorf("%s's Run() method should have one or two params", nameOfStructImpl)
		}
		hasDependencies := params.Len() == 2

		req := params.At(0)
		if fn, ok := validateIsHttpRequest(req); !ok {
			return nil, fmt.Errorf("%s's Run() should accept a http.Request as its first argument, got %s", nameOfStructImpl, fn)
		}

		parsed := middlewareParsed{
			obj:            named.Obj(),
			interfaceT:     embeddedInterface,
			implementation: implementingObj,
			run:            runMethod,
		}

		// Validate optional second argument, and recurse
		if hasDependencies {
			dep := params.At(1)
			//	4.2. ProcessMiddlewareInterface(deps)
			depInt, ok := dep.Type().Underlying().(*types.Interface)
			if !ok {
				return nil, fmt.Errorf("%s's Run() second argument should be a middleware interface stack", nameOfStructImpl)
			}
			stack, err := parseStack(depInt, middlewareByName)
			if err != nil {
				return nil, fmt.Errorf("%s's Run() second could not be parsed as a middleware stack: %w", nameOfStructImpl, err)
			}
			parsed.stack = stack
			parsed.stackInterface = depInt
		}

		stack = append(stack, &parsed)
		middlewareByName.Set(fullName, &parsed)
	}
	return stack, nil
}

// Validates the xxMiddleware type exported correctly implements the interface
func ensureImplementsMiddlewareInterface(foundTyp types.Type, embeddedInterface *types.Interface, nameOfStructImpl string, name string) error {
	asPtr := types.NewPointer(foundTyp)
	if m, wt := types.MissingMethod(asPtr, embeddedInterface, true); m != nil {
		mName := m.Name()
		errTyp := "was missing"
		if wt {
			errTyp = "had wrong signature for"
		}
		return fmt.Errorf("%s should implement %s, but %s %s", nameOfStructImpl, name, errTyp, mName)
	}
	return nil
}

// Find a 'Run' method in a set, or nil
func getRunMethod(methods *types.MethodSet) *types.Func {
	for i := 0; i < methods.Len(); i++ {
		sel := methods.At(i)
		if sel.Obj().Name() == "Run" {
			// safe by docs: 'Obj returns the object denoted by x.f; a *Var for a field selection
			// and a *Func in all other cases'
			return sel.Obj().(*types.Func)
		}
	}
	return nil
}

// firstParam of Run should be a http.Request
func validateIsHttpRequest(firstParam *types.Var) (string, bool) {
	obj, ok := firstParam.Type().(*types.Named)
	if !ok {
		return "", false
	}

	pkgPath := obj.Obj().Pkg().Path()
	name := obj.Obj().Name()

	fn := fmt.Sprintf("%s.%s", pkgPath, name)

	return fn, fn == "net/http.Request"
}

type middlewareGraph struct {
	adjacency map[string][]string
	byId      map[string]*middlewareParsed
}

func createGraph(p *targetStackParsed) (middlewareGraph, error) {
	stack := p.stack[:]
	g := middlewareGraph{
		adjacency: make(map[string][]string),
		byId:      make(map[string]*middlewareParsed),
	}
	for _, mw := range stack {
		id := types.ObjectString(mw.obj, nil)
		g.byId[id] = mw
		// needs to be present in adj map
		g.adjacency[id] = nil
		for _, depMw := range mw.stack {
			g.adjacency[id] = append(g.adjacency[id],
				types.ObjectString(depMw.obj, nil))
		}
		stack = append(stack, mw.stack...)
	}
	return g, nil
}

type middlewareCache struct {
	working string
	cache   map[string]*middlewareParsed
}

func (m *middlewareCache) get(n string) (*middlewareParsed, error) {
	if n == m.working {
		return nil, fmt.Errorf("Cycle detected rooted at %s", n)
	}
	return m.cache[n], nil
}

// ensure we don't end up with cycles
func (m *middlewareCache) mark(n string) {
	m.working = n
}

func (m *middlewareCache) Set(name string, m2 *middlewareParsed) {
	m.cache[name] = m2
}

func topographicalSort(g map[string][]string) []string {
	linearOrder := []string{}

	// 1. Let inDegree[1..n] be a new array, and create an empty linear array of
	//    verticies
	inDegree := map[string]int{}

	// 2. Set all values in inDegree to 0
	for n := range g {
		inDegree[n] = 0
	}

	// 3. For each vertex u
	for _, adjacent := range g {
		// A. For each vertex *v* adjacent to *u*:
		for _, v := range adjacent {
			//  i. increment inDegree[v]
			inDegree[v]++
		}
	}

	// 4. Make a list next consisting of all vertices u such that
	//    in-degree[u] = 0
	next := []string{}
	for u, v := range inDegree {
		if v != 0 {
			continue
		}

		next = append(next, u)
	}

	// 5. While next is not empty...
	for len(next) > 0 {
		// A. delete a vertex from next and call it vertex u
		u := next[0]
		next = next[1:]

		// B. Add u to the end of the linear order
		linearOrder = append(linearOrder, u)

		// C. For each vertex v adjacent to u
		for _, v := range g[u] {
			// i. Decrement inDegree[v]
			inDegree[v]--

			// ii. if inDegree[v] = 0, then insert v into next list
			if inDegree[v] == 0 {
				next = append(next, v)
			}
		}
	}

	// 6. Return the linear order
	return linearOrder
}


