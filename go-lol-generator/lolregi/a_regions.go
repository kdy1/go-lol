package lolregi

import (
	"go/constant"
	"go/types"
	"sort"
	"time"
)

func (reg *Registry) initRegions() {
	const name = "Region"

	obj := types.NewTypeName(0, nil, name, types.Typ[types.String])
	reg.Insert(obj)
	typ := types.NewNamed(obj, types.Typ[types.String], nil)

	for _, r := range reg.Regions {
		val := constant.MakeInt64(int64(r.Number))
		con := types.NewConst(0, nil, r.Name, typ, val)
		reg.Insert(con)
	}
}

var allRegions = Regions{
	{Name: "Global", PlatformID: "", Number: 1},
	{Name: "PBE", PlatformID: "PBE", Number: 2},

	// Non-special servers

	{
		Name: "BR", PlatformID: "BR1",
		LaunchedAt: time.Date(2012, 9, 13, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "EUNE", PlatformID: "EUN1",
		LaunchedAt: time.Date(2010, 7, 13, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "EUW", PlatformID: "EUW1",
		LaunchedAt: time.Date(2010, 7, 13, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "KR", PlatformID: "KR",
		LaunchedAt: time.Date(2011, 12, 12, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "LAN", PlatformID: "LA1",
		LaunchedAt: time.Date(2013, 6, 5, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "LAS", PlatformID: "LA2",
		LaunchedAt: time.Date(2013, 6, 5, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "NA", PlatformID: "NA1",
		LaunchedAt: time.Date(2009, 10, 27, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "OCE", PlatformID: "OC1",
		LaunchedAt: time.Date(2013, 7, 28, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "TR", PlatformID: "TR1",
		LaunchedAt: time.Date(2012, 9, 27, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "RU", PlatformID: "RU",
		LaunchedAt: time.Date(2013, 5, 17, 0, 0, 0, 0, time.UTC),
	},
}

func init() {
	sort.Sort(allRegions)

	var lastSpecial int
	for i, r := range allRegions {
		if r.Number != 0 {
			if !r.IsSpecial() {
				panic(`1 ~ 9 is reserved for special servers.`)
			}
			lastSpecial = i
			continue
		}
		r.Number = int32(i - lastSpecial + 9)
	}
}
