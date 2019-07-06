package lolregi

import (
	"go/types"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// Skip is used to exclude field "-" from struct.
// See Config.GetFieldName()
const Skip = "-"

type Config struct {
	// Default: "github.com/jerrodrurik/go-lol", "lol"
	Package *types.Package

	// Dont fix inconsistent id type.
	//
	// int32: map id, summoner spell id, champion id, rune id, mastery id, item id
	DontFixIDType bool

	// return "-" to exclude class.
	// TODO: Add helper methods to replace types from other classes.
	// return empty string to use default.
	GetClassName func(resID, rawName string) string

	// return "-" to exclude field.
	// return empty string to use default.
	GetFieldName func(resID, rawClassName, rawName string) string
}

func (reg *Registry) parseField(resID, rawClassName string,
	s *goquery.Selection) (rawName, name, rawType, desc string) {
	assert(s, "tr")

	s = s.ChildrenFiltered("td")
	rawName = strings.TrimSpace(s.Eq(0).Text())
	if reg.Config.GetFieldName != nil {
		name = reg.Config.GetFieldName(resID, rawClassName, rawName)
	}

	if name == "" {
		name = inflect.Camelize(rawName)
	}
	if name != "" && name != Skip {
		runes := []rune(name)
		runes[0] = unicode.ToUpper(runes[0])
		name = inflect.Camelize(name)
		name = lintName(string(name))
	}

	rawType = s.Eq(1).Text()
	desc = s.Eq(2).Text()
	return
}

func (reg *Registry) className(resID, rawCls string) string {
	rawCls = strings.TrimSpace(rawCls)
	var name string

	if reg.Config.GetClassName != nil {
		name = reg.Config.GetClassName(resID, rawCls)
		if name != "" {
			return lintName(name)
		}
	}

	name = strings.TrimSuffix(rawCls, "Dto")
	name = lintName(name)

	switch resID {
	case "lol-static-data":
		switch name {
		case "Champion":
			return "ChampionData"
		case "ChampionList":
			return "ChampionDataList"
		}
	case "team":
		switch name {
		case "Team", "TeamMemberInfo", "TeamStatDetail":
			return "Rank" + name
		case "MatchHistorySummary", "Roster":
			return name
		}

	case "match":
		switch name {
		case "Player", "Participant", "ParticipantIdentity":
			return "Match" + name
		case "Rune", "Mastery":
			return "Used" + name
		}

	case "current-game":
		switch name {
		case "Rune", "Mastery":
			return "Current" + name
		case "Observer", "BannedChampion":
			return "CurrentGame" + name
		}

	case "featured-games":
		switch name {
		case "Participant", "Observer", "BannedChampion":
			return "FeaturedGame" + name
		}

	case "summoner":
		switch name {
		case "Mastery":
			return "Summoner" + name
		}
	}

	return name
}

func (reg *Registry) fieldType(resID, rawCls, rawField, rawType, fieldDesc string) types.Type {
	switch rawCls {
	case "SummonerSpellDto", "ChampionSpellDto":
		switch rawField {
		case "effect":
			return types.NewSlice(types.NewSlice(types.Typ[types.Float64]))

		case "range":
			return types.NewPointer(reg.Pkg.Scope().Lookup("SpellRange").Type())
		}
	}

	t, err := reg.parseType(resID, rawType)
	if err != nil {
		panic(err)
	}

	// Fix wrong id types.
	if reg.Config.DontFixIDType == false && t == types.Typ[types.Int64] {
		switch rawField {
		case "mapId", "championId", "profileIconId", "runeId", "masteryId",
			"spell1Id", "spell2Id":
			return types.Typ[types.Int32]
		}

		switch rawCls {
		case "MatchReference":
			switch rawField {
			case "champion":
				return types.Typ[types.Int32]
			}
		}

		if resID == "champion" && rawCls == "ChampionDto" {
			switch rawField {
			case "id":
				return types.Typ[types.Int32]
			}
		}
	}

	return t
}
