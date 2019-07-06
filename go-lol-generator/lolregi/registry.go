package lolregi

import (
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/Sirupsen/logrus"
	"github.com/jerrodrurik/go-lol/uritemplates"
)

// go-lol pakcage path.
const lolPackagePath = "github.com/jerrodrurik/go-lol"

var lolPkg = types.NewPackage(lolPackagePath, "lol")

//
//
//TODO; Lazy validation.
type Registry struct {
	importer types.Importer

	Pkg       *types.Package
	Regions   Regions
	Resources []*Resource
	Classes   map[string]*ResponseClass

	Config Config
}

// New creates a new registry.
func New(conf Config) *Registry {
	if conf.Package == nil {
		conf.Package = lolPkg
	}

	reg := &Registry{
		Pkg:    conf.Package,
		Config: conf,

		importer:  importer.Default(),
		Regions:   allRegions,
		Resources: make([]*Resource, 0),
		Classes:   make(map[string]*ResponseClass, 0),
	}
	reg.initRegions()

	{
		it := types.NewInterface(nil, nil)
		srn := types.NewTypeName(token.NoPos, lolPkg, "SpellRange", it)
		reg.Pkg.Scope().Insert(srn)
	}
	return reg
}

func NewDefault() *Registry {
	reg := New(Config{})
	reg.InitDocument()
	return reg
}

func (reg *Registry) InitDocument() {
	doc := NewDocument()

	sels := doc.Find(".resource")
	ids := reg.sortResourceIDs(sels)
	log.Infoln("Resources:", ids)

	// instead of sorting, this just use simple map.
	selectionsByID := make(map[string]*goquery.Selection)

	sels.Each(func(i int, s *goquery.Selection) {
		id, _, _ := reg.parseResourceInfo(s)
		selectionsByID[id] = s
	})

	for _, id := range ids {
		s := selectionsByID[id]
		reg.Resources = append(reg.Resources, reg.newResource(s))
	}
}

func (reg *Registry) Insert(obj types.Object) {
	if obj.Name() == "Image" {
		return
	}
	if alredy := reg.Pkg.Scope().Insert(obj); alredy != nil {
		panic(fmt.Sprintf("Object '%s' already exists in scope", obj.Name()))
	}
}

func (reg *Registry) sortResourceIDs(s *goquery.Selection) (ids []string) {
	ids = make([]string, 1)
	ids[0] = "lol-static-data" // dirty hack.

	for i := range s.Nodes {
		id, _, _ := reg.parseResourceInfo(s.Eq(i))
		//TODO: Tournament API
		if id == "tournament-provider" {
			// I hate rito.. I can't write a post on riot api forums.
			continue
		}

		if id != "lol-static-data" {
			ids = append(ids, id)
		}

	}

	return
}

func (reg *Registry) newResource(s *goquery.Selection) *Resource {
	res := &Resource{
		Endpoints: make(Endpoints, 0),
	}

	res.ID, res.Version, res.Num = reg.parseResourceInfo(s)
	res.Regions = reg.parseResourceRegions(s)

	reg.parseResponseClasses(res, s)

	endpoints := s.ChildrenFiltered(`.endpoints`)
	for i := range endpoints.Nodes {
		es := endpoints.Eq(i)

		es.ChildrenFiltered(`.endpoint`).Each(func(i int, s *goquery.Selection) {
			res.Endpoints = append(res.Endpoints, reg.newEndpoint(res, i, s))
		})
	}

	return res
}

func (reg *Registry) newEndpoint(res *Resource, i int, s *goquery.Selection) *Endpoint {
	e := &Endpoint{Resource: res, Operations: make(Operations, 0)}

	ops := s.ChildrenFiltered(`.operations`)
	for i := range ops.Nodes {
		s := ops.Eq(i)

		s.ChildrenFiltered(`.operation`).Each(func(i int, s *goquery.Selection) {
			e.Operations = append(e.Operations, reg.newOperation(e, i, s))
		})
	}

	return e
}

func (reg *Registry) newOperation(endpoint *Endpoint, i int, s *goquery.Selection) *Operation {
	op := &Operation{Endpoint: endpoint}

	id, ok := s.Attr("id")
	if !ok {
		panic(`Operation id required`)
	}
	if num, err := strconv.Atoi(strings.Split(id, "_")[1]); err != nil {
		panic(err)
	} else {
		op.Num = num
	}

	// REST method (always get)
	op.Method = strings.TrimSpace(s.Find(".http_method").Text())

	path := s.Find(".heading > .path").Text()
	if tpl, err := uritemplates.Parse(strings.TrimSpace(path)); err != nil {
		panic(err)
	} else {
		op.Path.tpl = tpl
	}

	op.Desc = strings.TrimSpace(s.Find(`.heading > .options > li`).Text())
	op.Desc = strings.TrimSpace(strings.TrimSuffix(op.Desc, `(REST)`))

	info := op.Info()
	op.Name = op.Info().Name

	s.ChildrenFiltered(`.heading`).Remove() // useless

	// Blocks
	blocks := s.Find(".api_block")
	for i := range blocks.Nodes {
		bs := blocks.Eq(i)
		assert(bs, "div.api_block")

		if IsEmpty(bs) {
			bs.Remove()
			continue
		}

		bh := bs.Children().First()
		assert(bh, "h4")
		t := bh.Remove().Text()

		switch t {
		case "Implementation Notes":
			op.ImplementationNotes = bs.Text()
		case "Rate Limit Notes":
			op.RateLimitNotes = strings.TrimSpace(bs.Text())
		case "Response Classes":
			op.ReturnValue = reg.parseReturnValue(endpoint.Resource.ID, info, bs)
		case "Response Errors":
			op.Errors = parseResponseErrors(bs)
		case "Path Parameters":
			op.Path.Params = reg.parseParams(endpoint.Resource.ID, bs.Children().First())
			if bs.ChildrenFiltered(`h4`) != nil {
				t := bs.ChildrenFiltered(`h4`).Remove().Text()
				switch t {
				case "Query Parameters": // Has query parameters.
					op.QueryParams = reg.parseParams(endpoint.Resource.ID, bs.Contents())
				case "":
				default:
					h, _ := bs.Html()
					panic(t + "\n" + h)
				}
			}

		case "Select Region to Execute Against": //featured-games
		default:
			panic(`Unknown api block: ` + t)
		}
	}

	return op
}

func (reg *Registry) parseParams(resID string, table *goquery.Selection) []Parameter {
	assert(table, "table")

	var params []Parameter

	table.ChildrenFiltered(`thead`).Remove()
	body := table.Children().First()
	assert(body, "tbody.operation-params")

	trs := body.ChildrenFiltered(`tr`)

	for i := range trs.Nodes {
		param, tr := Parameter{}, trs.Eq(i)

		param.required = tr.Find(`.required`).Remove().Text() == "true"
		param.Raw = tr.ChildrenFiltered(`td.code`).Text() // raw name
		param.Name = parameterName(param.Raw)
		rawType := tr.Find(`span.model-signature`).Text()

		param.Desc = tr.Children().Last().Text()
		//TODO: Check for 'Comma-separated list' prefix and 'Maximum allowed at once is '

		if typ, err := reg.parseType(resID, rawType); err != nil {
			panic(err)
		} else {
			param.typ = typ
		}

		//TODO: Handle enums?
		params = append(params, param)
	}

	table.Remove() // important

	return params
}

func (reg *Registry) parseReturnValue(resID string, info OpInfo, s *goquery.Selection) types.Type {
	sels := s.Find(`.response_body`)
	s = sels.Eq(0)

	t := s.Children().First().Remove().Text() // <b>Return Value:</b>
	if t != "Return Value:" {
		panic("Expected return value block. Got: " + t)
	}
	retVal, err := reg.parseType(resID, s.Text())
	if err != nil {
		log.Infoln(err, retVal, s.Text())
		panic(err)
	}

	return retVal
}

//  Args:
// 		- resID:	Resource ID
//		- s:		Resource selector
// parse response classes, register classes to registry.
func (reg *Registry) parseResponseClasses(res *Resource, rs *goquery.Selection) {
	sels := rs.Find(`.response_body`)
	for i := len(sels.Nodes) - 1; i >= 0; i-- {
		s := sels.Eq(i)

		rawClsName := s.Children().First().Text()
		if rawClsName == "Return Value:" {
			continue
		}

		cls := &ResponseClass{
			res:     res,
			rawName: rawClsName,
			name:    reg.className(res.ID, rawClsName),
			fields:  make([]*Field, 0),
			methods: make([]*types.Func, 0),
		}

		if conflict, ok := reg.Classes[cls.Name()]; ok {
			if conflict.ResID() == res.ID && conflict.RawName() == rawClsName {
				log.Debugf(`%s is already declared. Raw: %s`, cls.Name(), rawClsName)
				continue
			}

			log.Fatalf("Conflict class %s:%s", conflict.DebugString(), cls.DebugString())
			panic('!')
		}

		log.Debugln("Parsing response class", strconv.Quote(cls.Name()))
		table := s.ChildrenFiltered(`table`).Remove()
		cls.Desc = strings.TrimSpace(s.Text())
		cls.Desc = strings.TrimSpace(strings.TrimPrefix(cls.Desc, "-"))

		table.ChildrenFiltered(`thead`).Remove() // <thead>,,,</thead>

		table.Find("tbody > tr").Each(func(i int, s *goquery.Selection) {
			rawName, name, rawType, desc := reg.parseField(res.ID, cls.rawName, s)

			if name == Skip {
				return
			}

			typ := reg.fieldType(res.ID, cls.RawName(), rawName, rawType, desc)

			field := NewField(reg.Pkg, rawName, name, typ, desc)
			cls.fields = append(cls.fields, field)
		})

		log.WithField("resource", res.ID).Debugln("Registered class", strconv.Quote(cls.Name()))
		reg.Classes[cls.Name()] = cls
		reg.Insert(types.NewTypeName(0, reg.Pkg, cls.Name(), cls))
	}

	return
}

func parseResponseErrors(s *goquery.Selection) []ResponseError {
	assert(s, "div.api_block")

	var errors []ResponseError

	errSels := s.Find(`tbody > td, tbody> tr`)
	for i := range errSels.Nodes {
		s := errSels.Eq(i)

		codeText := s.Children().First().Remove().Text()
		codeText = strings.TrimSpace(codeText)
		httpCode, err := strconv.ParseInt(codeText, 10, 64)
		if err != nil {
			panic(err)
		}

		errors = append(errors, ResponseError{
			Code: int(httpCode), Desc: strings.TrimSpace(s.Text()),
		})
	}
	return errors
}

func (reg *Registry) parseResourceInfo(s *goquery.Selection) (id, ver string, num int) {
	a, ok := s.Attr("data-version")
	if !ok {
		panic(`Resource info required`)
	}
	ts := strings.Split(a, "-")
	id = strings.TrimSpace(strings.Join(ts[:len(ts)-1], "-"))
	ver = strings.TrimSpace(ts[len(ts)-1])

	idAttr, ok := s.Attr("id")
	if !ok {
		panic(`Resource numeric id required`)
	}
	var err error
	num, err = strconv.Atoi(strings.TrimPrefix(idAttr, "resource_"))
	if err != nil {
		panic(err)
	}
	return
}

func (reg *Registry) parseResourceRegions(s *goquery.Selection) (regions Regions) {
	regions = make(Regions, 0)
	a, ok := s.Attr("data-regions")
	if !ok {
		panic(`Unknown resource regions`)
	}

	a = strings.TrimPrefix(strings.TrimSuffix(a, "]"), "[")

	// Rito pls... "[ALL]" is added for tournament-provider api.
	if a == "ALL" {
		return AllValidRegions()
	}
	rs := strings.Split(a, ",")
	for _, r := range rs {
		regions = append(regions, regionByName(strings.TrimSpace(r)))
	}
	return
}

var javaTypeToGoType = map[string]types.BasicKind{
	"boolean": types.Bool, "int": types.Int32, "long": types.Int64,
	"string": types.String,
	// I'm not sure about this.
	"double": types.Float64, "float": types.Float32,
}

// This does not return nil if err==nil.
func (reg *Registry) parseType(resID, s string) (types.Type, error) {
	s = strings.TrimSpace(s)

	if kind, ok := javaTypeToGoType[s]; ok {
		return types.Typ[kind], nil
	}

	if s == "object" {
		iface := types.NewInterface(nil, nil)
		return types.NewMap(iface, iface), nil
	}

	if strings.HasPrefix(s, "List[") {
		el := strings.TrimSuffix(strings.TrimPrefix(s, "List["), "]")
		elem, err := reg.parseType(resID, el)
		if err != nil {
			return nil, err
		}
		return types.NewSlice(elem), nil
	} else if strings.HasPrefix(s, "Set[") {
		el := strings.TrimSuffix(strings.TrimPrefix(s, "Set["), "]")
		elem, err := reg.parseType(resID, el)
		if err != nil {
			return nil, err
		}
		return types.NewSlice(elem), nil
	} else if strings.HasPrefix(s, "Map[") {
		el := strings.TrimSuffix(strings.TrimPrefix(s, "Map["), "]")
		ss := strings.SplitN(el, ",", 2)
		k, err := reg.parseType(resID, ss[0])
		if err != nil {
			return nil, err
		}
		v, err := reg.parseType(resID, ss[1])
		if err != nil {
			return nil, err
		}
		return types.NewMap(k, v), nil
	}

	clsName := reg.className(resID, s)
	if typ := reg.Classes[clsName]; typ != nil {
		return types.NewPointer(typ), nil
	}

	return nil, fmt.Errorf(`Unknown type "%s" while type for Resource "%s"`, clsName, resID)
}

func (reg *Registry) PrintDebugInfo() {
	for _, res := range reg.Resources {
		log.Infoln(`|+ Resource`, res.ID, res.Version, res.Regions)

		for _, end := range res.Endpoints {
			log.Infoln(`|+-- Endpoint`)

			for _, op := range end.Operations {
				log.Infoln(`|+---- Operation`, op.Method+":", op.Path.String())
				if op.Desc != "" {
					log.Infoln(`        Description:`, op.Desc)
				}

				if op.QueryParams != nil {
					log.Infoln(`       `, `Query parameters:`, op.QueryParams)
				}

				if len(op.RateLimitNotes) != 0 {
					log.Infoln(`       `, `Rate limit notes:`, op.RateLimitNotes)
				}

				if len(op.ImplementationNotes) != 0 {
					log.Infoln(`       `, `Notes:`, op.ImplementationNotes)
				}

				if len(op.Errors) != 0 {
					log.Infoln(`       `, `Errors:`, op.Errors)
				}
			}
		}
	}

	reg.Pkg.Scope().WriteTo(os.Stdout, 1, true)
}

func (reg *Registry) Class(resID, name string) *ResponseClass {
	for _, cls := range reg.Classes {
		if resID == cls.res.ID && name == cls.name {
			return cls
		}
	}
	return nil
}
