package lolgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/jerrodrurik/go-lol/go-lol-generator/lolregi"
)

type Generator struct {
	*ast.File
	*bytes.Buffer

	reg *lolregi.Registry
}

func New(reg *lolregi.Registry) *Generator {
	return &Generator{Buffer: &bytes.Buffer{}, reg: reg}
}

func (g *Generator) Generate() []byte {
	g.generatePackage()

	g.generateRegions(g.reg.Regions)

	for _, c := range g.reg.Classes {
		g.GenerateResponseClass(c)
	}

	for _, v := range g.reg.Resources {
		g.generateResource(v)
	}

	src := g.Bytes()
	return src
}

func (g *Generator) generatePackage() {
	g.P(`// Generated by go-lol-generator. DO NOT EDIT.`)
	g.P()
	g.P(`package `, g.reg.Pkg.Name(), `;`)

	g.P(`import "encoding/json"`)
	g.P(`import "io"`)
	g.P(`import "strconv"`)
	g.P(`import "net/http"`)
	g.P(`import "net/url"`)
	g.P()

	g.P(`import "golang.org/x/net/context"`)
	g.P(`import `, strconv.Quote(g.reg.Pkg.Path()+"/uritemplates"))
	g.P()

	g.P(`var _ = json.Marshal`)
	g.P(`var _ = io.EOF`)
}

func (g *Generator) generateResource(res *lolregi.Resource) {
	for _, endpoint := range res.Endpoints {
		g.generateEndpoint(res, endpoint)
	}
}

func (g *Generator) generateEndpoint(res *lolregi.Resource, e *lolregi.Endpoint) {
	for _, op := range e.Operations {
		g.generateOperation(res, e, op)
	}
}

func (g *Generator) generateOperation(res *lolregi.Resource, e *lolregi.Endpoint, op *lolregi.Operation) {
	info, ret := op.Info(), op.ReturnValue
	if info.MapKey != 0 {
		if m, ok := ret.(*types.Map); ok {
			ret = types.NewMap(types.Typ[info.MapKey], m.Elem())
		}
	}

	g.generateOpType(op)
	g.generateOpCreatorFunc(res, e, op)
	g.generateOpDoRequestFunc(op)

	g.P(`// Do executes api request.`)
	g.P(`//`)
	g.P(`// API Errors: `)
	for _, e := range op.Errors {
		g.P(`//  `, e.Code, ` - `, e.Desc)
	}

	g.P(`func (c *`, op.GoType(), `) Do() (`, ret, `, error) {`)

	g.P(`res, err := c.doRequest()
	if err != nil { return nil, err }
	defer closeBody(res)`)
	g.P(`if err := verifyAPIResponse(res); err!=nil { return nil, err }`)

	g.DeclareVar(`ret`, op.ReturnValue)
	g.P(`if err := json.NewDecoder(res.Body).Decode(&ret); err != nil { return nil, err }`)
	if info.MapKey != 0 {
		if m, ok := ret.(*types.Map); ok {
			ret = types.NewMap(types.Typ[info.MapKey], m.Elem())
		}

		//TODO
		g.DeclareVar(`data`, ret)
		g.P(`for k, v := range ret {`)
		g.P(`i, err := strconv.ParseInt(k, 10, 64)`)
		g.P(`if err != nil { return nil, err }`)
		g.P(`data[i] = v`)
		g.P(`}`)

		g.P(`return data, nil`)
	} else {
		g.P(`return ret, nil`)
	}
	g.P(`}`)
	g.P()
}

// prints operation initialization function.
func (g *Generator) generateOpCreatorFunc(res *lolregi.Resource, e *lolregi.Endpoint, op *lolregi.Operation) {
	g.P()
	g.P(`// `, op.Desc)
	g.P(`//`)
	if op.ImplementationNotes != "" {
		g.P(`//`)
		g.P(`// Implementation notes: `, op.ImplementationNotes)
	}
	if op.RateLimitNotes != "" {
		g.P(`//`)
		g.P(`// Rate limit notes: `, op.RateLimitNotes)
	}
	g.P(`//`)
	g.P(`// `, op.Method, `: `, res.APIBase(), op.Path)
	g.P(`// `)
	g.P(`// Reference: `, op.DocURL())

	// required arguments
	var args string
	if op.HasRegionParameter() {
		args += `, region Region`
	}
	for _, p := range op.Path.Params {
		if p.IsRegion() {
			continue
		}

		if p.IsRequired() {
			args += `,  ` + p.String() + ` ` + p.Type().String()
		}
	}

	g.P(`func (c *Client) `, op.Name, `(ctx context.Context`, args, `) *`, op.GoType(), ` {`)
	g.P(`path := make(map[string]string)`)

	for _, p := range op.Path.Params {
		if p.IsRegion() {
			continue
		}

		if p.IsRequired() {
			g.P(`path[`, strconv.Quote(p.Raw), `] = convertToString(`, p.String(), `)`)
		}
	}

	fields := `ctx: ctx, client: c, query: make(url.Values), pathParams: path, `
	if op.HasRegionParameter() {
		fields += `region: region,`
	}

	g.P(`return &`, op.GoType(), `{`, fields, `}`)
	g.P(`}`)
	g.P()

	for _, q := range op.QueryParams {
		g.P(`// `, q.Name, ` configures query parameter `, strconv.Quote(q.Raw), `.`)
		g.P(`func (c *`, op.GoType(), `) `, funcName(q.Name, true), `(v `, q.Type(), `) (*`, op.GoType(), `) {`)
		g.P(`c.query.Set(`, strconv.Quote(q.Name), `, convertToString(v))`)
		g.P(`return c`)
		g.P(`}`)
	}
}

func (g *Generator) generateRegions(regions lolregi.Regions) {
	g.P(`const (`)

	for _, r := range regions {
		g.P(`// `, r.Name, ` is a service area of league of legends.`)
		g.P(r.Name, ` Region = `, r.Number)
	}

	g.P(`)`)
	g.P()

	g.P(`var regions = []Region{`)
	for _, r := range regions {
		if r.Name != "Global" {
			g.P(r.Name, `,`)
		}
	}
	g.P(`}`)
	g.P()

	g.P(`var regionByName = map[string]Region{`)
	for _, r := range regions {
		g.P(strconv.Quote(strings.ToLower(r.Name)), `:`, r.Name, `,`)
	}
	g.P(`}`)
	g.P()

	g.P(`var regionByPlatformID = map[string]Region{`)
	for _, r := range regions {
		if r.Name != "Global" {
			g.P(strconv.Quote(strings.ToLower(r.PlatformID)), `:`, r.Name, `,`)
		}
	}
	g.P(`}`)
	g.P()

	g.P(`var nameByRegion = map[Region]string{`)
	for _, r := range regions {
		g.P(r.Name, `:`, strconv.Quote(strings.ToLower(r.Name)), `,`)
	}
	g.P(`}`)
	g.P()

	g.P(`var platformIDByRegion = map[Region]string{`)
	for _, r := range regions {
		g.P(r.Name, `:`, strconv.Quote(strings.ToLower(r.PlatformID)), `,`)
	}
	g.P(`}`)
	g.P()

	g.P(`var hostByRegion = map[Region]string{`)
	for _, r := range regions {
		g.P(r.Name, `:`, strconv.Quote(strings.ToLower(r.Host())), `,`)
	}
	g.P(`}`)
	g.P()
}

func (g *Generator) generateOpType(op *lolregi.Operation) {
	g.P()
	g.P(`// `, op.GoType(), ` is a builder for `+strconv.Quote(op.Name))
	g.P(`type `, op.GoType(), ` struct {
		ctx context.Context
		client *Client
		query url.Values
		pathParams map[string]string`)
	if op.HasRegionParameter() {
		g.P(`	region Region`)
	}
	g.P(`}`)
	g.P()
}

func (g *Generator) generateOpDoRequestFunc(op *lolregi.Operation) {
	g.P()
	g.P(`func (c *`, op.GoType(), `) doRequest() (*http.Response, error) {`)
	g.P(`var body io.Reader`)

	// Parameter validation
	if op.HasRegionParameter() {
		g.P(`switch c.region {`)
		g.P(`case `, op.Endpoint.Resource.Regions.Join(","), `:`)
		g.P(`default:
			return nil, ErrNotSupportedRegion
		}`)
	}

	if op.APIKeyRequired() {
		g.P(`c.query.Set("api_key", c.client.apiKey)`)
	}
	if op.Path.Has("region") {
		g.P(`c.pathParams["region"] = c.region.Name()`)
	}
	if op.Path.Has("platformId") {
		g.P(`c.pathParams["platformID"] = c.region.PlatformID()`)
	}
	g.P()

	var urlsTpl string

	if op.Endpoint.Resource.APIBase() == "" {
		urlsTpl = `c.region.baseURL()`
	} else {
		urlsTpl = strconv.Quote(op.Endpoint.Resource.APIBase())
	}
	g.P(`path, err := uritemplates.Expand(`, strconv.Quote(op.Path.String()), `, c.pathParams)`)
	g.P(`if err != nil { return nil, err }`)

	g.P(`urls := `, urlsTpl+` + path + "?" + c.query.Encode()`)
	g.P()

	g.P(`return c.client.doRequest(c.ctx, `, strconv.Quote(op.Method), `, urls, body)`)
	g.P(`}`)
	g.P()
}

func (g *Generator) GenerateResponseClass(c *lolregi.ResponseClass) {
	g.P()
	if c.Desc != "" {
		g.P(`// `, c.Desc)
		g.P(`//`)
	}
	g.P(`// resource: "`, c.ResID(), `", original name: "`, c.RawName(), `"`)
	g.P(`type `, c.Name(), ` struct {`)
	for _, f := range c.Fields() {
		if f.Desc != "" {
			for _, line := range strings.Split(f.Desc, "\n") {
				g.P(`// `, line)
			}
		}

		g.P(f.Name(), ` `, f.Type(), "`", f.Tag, "`")

	}
	g.P(`}`)
	g.P()
}

func (g *Generator) P(args ...interface{}) {
	g.WriteRune('\t')
	for _, v := range args {
		switch s := v.(type) {
		case string:
			g.WriteString(s)
		case types.Type:
			types.WriteType(g.Buffer, s, func(pkg *types.Package) string {
				if pkg.Path() == g.reg.Pkg.Path() {
					return ""
				}

				return pkg.Name()
			})
		case ast.Expr:
			types.WriteExpr(g.Buffer, s)
		case fmt.Stringer:
			g.WriteString(s.String())
		case bool:
			g.WriteString(fmt.Sprintf("%t", s))
		case *bool:
			g.WriteString(fmt.Sprintf("%t", *s))
		case int:
			g.WriteString(fmt.Sprintf("%d", s))
		case int32:
			g.WriteString(fmt.Sprintf("%d", s))
		case *int32:
			g.WriteString(fmt.Sprintf("%d", *s))
		case *int64:
			g.WriteString(fmt.Sprintf("%d", *s))
		case float64:
			g.WriteString(fmt.Sprintf("%g", s))
		case *float64:
			g.WriteString(fmt.Sprintf("%g", *s))
		default:
			panic(fmt.Errorf("unknown type in printer: %T", v))
		}
	}
	g.WriteByte('\n')
}

// ----- Printer utilities.

func (g *Generator) DeclareVar(name string, typ types.Type) {
	switch t := typ.(type) {
	case *types.Pointer:
		g.P(name, ` := `, `&`, t.Elem().String()+`{}`)
	case *types.Map:
		g.P(name, ` := `, `make(`, t.String()+`)`)
	case *types.Slice:
		g.P(name, ` := `, `make(`, t.String()+`, 0)`)
	default:
		log.Panicf("Failed to decalre varaible %s with type %v", name, typ)
	}
}

func funcName(name string, exported bool) string {
	if len(name) == 0 {
		panic(`funcName: len == 0, ` + name)
	}
	first := []rune(name)[0]
	if exported {
		first = unicode.ToUpper(first)
	} else {
		first = unicode.ToLower(first)
	}
	return string(first) + name[1:]
}

func typeName(name string, exported bool) string {
	if len(name) == 0 {
		panic(`typeName: len == 0, ` + name)
	}
	first := []rune(name)[0]
	if exported {
		first = unicode.ToUpper(first)
	} else {
		first = unicode.ToLower(first)
	}
	return string(first) + name[1:]
}
