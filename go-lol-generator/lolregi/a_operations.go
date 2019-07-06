package lolregi

import "go/types"

// OpInfo represets a predeclared operation info.
type OpInfo struct {
	Name string
	// Override map key in return value.
	// Loader will panic if it's not map.
	MapKey types.BasicKind
}

// map[resource name]map[path suffix]Operation
//
// Note: Operation ids may change without api change.
var knownOperations = map[string]map[string]OpInfo{
	"lol-static-data": { // v1.2
		"/champion":            {Name: "ChampionDatas"},
		"/champion/{id}":       {Name: "ChampionData"},
		"/item":                {Name: "Items"},
		"/item/{id}":           {Name: "Item"},
		"/language-strings":    {Name: "LanguageStrings"},
		"/languages":           {Name: "Languages"},
		"/map":                 {Name: "Maps"},
		"/mastery":             {Name: "Masteries"},
		"/mastery/{id}":        {Name: "Mastery"},
		"/realm":               {Name: "Realm"},
		"/rune":                {Name: "Runes"},
		"/rune/{id}":           {Name: "Rune"},
		"/summoner-spell":      {Name: "SummonerSpells"},
		"/summoner-spell/{id}": {Name: "SummonerSpell"},
		"versions":             {Name: "Versions"},
	},

	"champion": { // v1.2
		"/champion":      {Name: "Champions"},
		"/champion/{id}": {Name: "Champion"},
	},

	"current-game": { // v1.0
		"/getSpectatorGameInfo/{platformId}/{summonerId}": {Name: "SpectatorGameInfo"},
	},

	"featured-games": { //v1.0
		"/featured": {Name: "FeaturedGames"},
	},

	"game": { // v1.3
		"/game/by-summoner/{summonerId}/recent": {Name: "RecentGames"},
	},

	"league": { // v2.5
		"/league/by-summoner/{summonerIds}":       {Name: "LeaguesBySummonerID"},
		"/league/by-summoner/{summonerIds}/entry": {Name: "LeagueEntriesBySummonerID"},
		"/league/by-team/{teamIds}":               {Name: "LeaguesByTeamID"},
		"/league/by-team/{teamIds}/entry":         {Name: "LeagueEntriesByTeamID"},
		"/league/challenger":                      {Name: "Challenger"},
		"/league/master":                          {Name: "Master"},
	},

	"lol-status": { // v1.0
		"/shards":          {Name: "Shards"},
		"/shards/{region}": {Name: "ShardsInRegion"},
	},

	"match": { // v2.2
		"/match/{matchId}":                          {Name: "Match"},
		"/match/by-tournament/{tournamentCode}/ids": {Name: "MatchesByTournement"},
		"/match/for-tournament/{matchId}":           {Name: "MatchForTournement"},
	},

	"matchlist": { // v2.2
		"/matchlist/by-summoner/{summonerId}": {Name: "MatchesBySummonerID"},
	},

	"stats": { // v1.3
		"/stats/by-summoner/{summonerId}/ranked":  {Name: "RankedStats"},
		"/stats/by-summoner/{summonerId}/summary": {Name: "StatsSummary"},
	},

	"summoner": { // v1.4
		"/summoner/by-name/{summonerNames}": {Name: "SummonersByName"},
		"/summoner/{summonerIds}":           {Name: "Summoners", MapKey: types.Int64},
		"/summoner/{summonerIds}/masteries": {Name: "SummonerMasteries", MapKey: types.Int64},
		"/summoner/{summonerIds}/name":      {Name: "SummonerNames", MapKey: types.Int64},
		"/summoner/{summonerIds}/runes":     {Name: "SummonerRunes", MapKey: types.Int64},
	},

	"team": { // v2.4
		"/team/by-summoner/{summonerIds}": {Name: "TeamsBySummonerID", MapKey: types.Int64},
		"/team/{teamIds}":                 {Name: "Teams"},
	},

	//TODO: Better naming
	"tournament-provider": {},
}
