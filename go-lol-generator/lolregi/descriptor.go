package lolregi

import (
	"bytes"
	"fmt"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"time"

	"github.com/go-lol/lol/uritemplates"
)

type (
	Region struct {
		LaunchedAt             time.Time
		Name, PlatformID, host string
		Number                 int32
	}

	Regions []*Region

	ResponseError struct {
		Code int
		Desc string
	}

	Resource struct {
		reg *Registry

		ID      string
		Version string
		Regions Regions

		Num int // numeric id

		Endpoints Endpoints
	}

	Endpoint struct {
		Resource *Resource // Ref to parent

		Operations Operations
	}

	Endpoints []*Endpoint

	Operation struct {
		Endpoint *Endpoint // Ref to parent

		Name, Desc string
		Method     string // HTTP method to use.
		Num        int    // numeric id

		Path        Path
		QueryParams []Parameter
		ReturnValue types.Type

		ImplementationNotes string
		RateLimitNotes      string

		Errors []ResponseError
	}

	Operations []*Operation

	Path struct {
		tpl    *uritemplates.URITemplate
		Params []Parameter
	}
)

func (r *Region) IsSpecial() bool {
	return r.Name == "Global" || r.Name == "PBE"
}

func (r *Region) Host() string {
	if len(r.host) != 0 {
		return r.host
	}

	return strings.ToLower(r.Name) + ".api.pvp.net"
}

// String implements fmt.Stringer
func (r *Region) String() string {
	return r.Name
}

// Len is part of sort.Interface.
func (regions Regions) Len() int {
	return len(regions)
}

// Swap is part of sort.Interface.
func (regions Regions) Swap(i, j int) {
	regions[i], regions[j] = regions[j], regions[i]
}

// Less is part of sort.Interface.
func (regions Regions) Less(i, j int) bool {
	if regions[i].Name == "Global" {
		return true
	} else if regions[i].Name == "PBE" {
		return true
	}
	if regions[j].Name == "Global" {
		return false
	} else if regions[j].Name == "PBE" {
		return false
	}

	return regions[i].LaunchedAt.Before(regions[j].LaunchedAt) //TODO:Alphabet sort.
}

func (regions Regions) Join(sep string) string {
	var buf bytes.Buffer
	for i, region := range regions {
		buf.WriteString(region.Name)
		if i != len(regions)-1 {
			buf.WriteString(sep)
		}
	}
	return buf.String()
}

func regionByName(name string) *Region {
	for _, r := range allRegions {
		if r.Name == name {
			return r
		}
	}
	panic(fmt.Sprintf(`Unknown region "%s"`, name))
}

func AllRegions() Regions {
	return allRegions
}

func AllValidRegions() Regions {
	var rs Regions
	for _, r := range allRegions {
		if r.Name != "Global" {
			rs = append(rs, r)
		}
	}
	return rs
}

// APIBase returns empty string if it's not a special operation.
func (res *Resource) APIBase() string {
	switch res.ID {
	case "lol-static-data":
		return "https://global.api.pvp.net"
	case "lol-status":
		return "http://status.leagueoflegends.com"
	default:
		return ""
	}
}

func (res *Resource) APIKeyRequired() bool {
	switch res.ID {
	case "lol-status":
		return false
	default:
		return true
	}
}

// DocURL returns url of riot method reference page.
func (op *Operation) DocURL() string {
	return fmt.Sprintf("https://developer.riotgames.com/api/methods#!/%d/%d", op.Endpoint.Resource.Num, op.Num)
}

// APIKeyRequired returns true if api key is required for this operation.
func (op *Operation) APIKeyRequired() bool {
	return op.Endpoint.Resource.APIKeyRequired()
}

// HasRegionParameter returns true if this operation has a path parameter named 'region' or 'platformId'.
func (op *Operation) HasRegionParameter() bool {
	if op.Endpoint.Resource.APIBase() == "" { // region is required to determine hostname.
		return true
	}

	for _, p := range op.Path.Params {
		if p.IsRegion() {
			return true
		}
	}
	return false
}

func (op *Operation) Info() OpInfo {
	knownOps := knownOperations[op.Endpoint.Resource.ID]

	for suffix, o := range knownOps {
		if strings.HasSuffix(op.Path.String(), suffix) {
			return o
		}
	}

	panic(`Unknown operation: ` + op.DocURL())
}

// GoType returns a name for operation builder struct.
func (op *Operation) GoType() string { return op.Name + "Call" }

// Has returns true if this path has a parameter with a such name.
func (p Path) Has(name string) bool {
	for _, p := range p.Params {
		if p.Name == name {
			return true
		}
	}
	return false
}

// String returns raw string.
func (p Path) String() string {
	return p.tpl.String()
}

// IsRegion returns true if this parameter is related to region.
func (p Parameter) IsRegion() bool {
	switch p.Name {
	case "region", "platformID":
		return true
	default:
		return false
	}
}

// String returns parameter name in golang style.
func (p Parameter) String() string { return p.Name }

// Type returns golang type.
func (p Parameter) Type() types.Type {
	if p.typ == types.Typ[types.String] {
		if p.Name == "summonerIDs" { // summonerIDs ...
			return types.NewSlice(types.Typ[types.Int64])
		}
		if p.Name == "summonerNames" || p.Name == "teamIDs" {
			return types.NewSlice(types.Typ[types.String])
		}
	}

	return p.typ
}

// IsRequired returns true if this parameter is always required.
func (p Parameter) IsRequired() bool { return p.required }

type ResponseClass struct {
	*types.Named
	res *Resource

	rawName string
	name    string
	Desc    string
	// Key: rawName
	fields  []*Field
	methods []*types.Func
}

func (c *ResponseClass) Fields() []*Field {
	return c.fields
}

func (c *ResponseClass) AddField(f *Field) {
	c.fields = append(c.fields, f)
}
func (c *ResponseClass) Methods() []*types.Func {
	return c.methods
}

func (c *ResponseClass) AddMethod(m *types.Func) {
	c.methods = append(c.methods, m)
}

func (c *ResponseClass) Name() string {
	return c.name
}

func (c *ResponseClass) String() string {
	return c.Name()
}

func (c *ResponseClass) DebugString() string {
	return fmt.Sprintf("'%s'(raw: '%s', res: '%s')", c.Name(), c.rawName, c.ResID())
}

func (c *ResponseClass) Underlying() types.Type {
	return c
}

func (e ResponseError) String() string {
	return strconv.Itoa(e.Code) + "(" + e.Desc + ")"
}

// ResID returns a resource id.
func (c *ResponseClass) ResID() string {
	return c.res.ID
}

func (c *ResponseClass) Field(name string) *Field {
	for _, f := range c.Fields() {
		if name == f.Name() {
			return f
		}
	}
	return nil
}

func (c *ResponseClass) FieldByRawName(name string) *Field {
	for _, f := range c.Fields() {
		if name == f.RawName() {
			return f
		}
	}
	return nil
}

func (c *ResponseClass) RawName() string {
	return c.rawName
}

type Field struct {
	*types.Var
	rawName string
	Tag     string
	Desc    string
}

func NewField(pkg *types.Package, rawName, name string, typ types.Type, desc string) *Field {
	return &Field{
		Var:     types.NewField(token.NoPos, pkg, name, typ, false),
		rawName: rawName,
		Tag:     fmt.Sprintf(`json:"%s,omitempty"`, rawName),
		Desc:    desc,
	}
}

func (f *Field) RawName() string {
	return f.rawName
}

func (f *Field) SetType(typ types.Type) {
	f.Var = types.NewField(f.Pos(), f.Pkg(), f.Name(), typ, f.Anonymous())
}

type Parameter struct {
	Raw, Name, Desc string
	typ             types.Type
	required        bool
}
