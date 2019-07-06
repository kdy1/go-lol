package lol_test

import (
	"os"
	"strings"
	"testing"

	"github.com/go-lol/go-lol"
	"golang.org/x/net/context"
)

var client *lol.Client

func init() {
	var err error
	apiKey := os.Getenv("RIOT_API_KEY")
	if apiKey == "" {
		panic("Environment variable 'RIOT_API_KEY' is required.")
	}

	client, err = lol.New(nil, apiKey)
	if err != nil {
		panic(err)
	}
}

func TestSummonersAPI(t *testing.T) {
	// credit: https://github.com/kevinohashi/php-riot-api/blob/master/testing.php
	const (
		testID     = 585897
		testName   = "RiotSchmick"
		testRegion = lol.NA
	)

	check := func(id int64, name string, region lol.Region, summoner *lol.Summoner) {
		t.Logf("Checking summoner info: %v", summoner)

		if summoner.ID != id {
			t.Fatalf("Invalid summoner id returned. Expected %d", id)
			return
		} else if summoner.Name != name {
			t.Fatalf("Invalid summoner name returned. Expected %s", name)
			return
		}
	}

	summoners, err := client.Summoners(context.TODO(), testRegion, []int64{testID}).Do()
	t.Log("Summoners() returned: ", summoners)
	if err != nil {
		t.Fatal("Failed to get summoner information", err)
		return
	} else if len(summoners) != 1 {
		t.Fatal("Expected [1]Summoner{...}")
		return
	}
	check(testID, testName, testRegion, summoners[testID])
	t.Log("Summoners() successed.")

	// then..
	summonersByNames, err := client.SummonersByName(context.TODO(), testRegion, []string{testName}).Do()
	t.Log("SummonersByNames() returned: ", summonersByNames)
	if err != nil {
		t.Fatal("Failed to get summoner information", err)
		return
	} else if len(summonersByNames) != 1 {
		t.Fatal("Expected [1]Summoner{...}")
		return
	}
	//TODO: trim keys.
	check(testID, testName, testRegion, summonersByNames[strings.ToLower(testName)])
	t.Log("SummonersByNames() successed.")

	// then..
	names, err := client.SummonerNames(context.TODO(), testRegion, []int64{testID}).Do()
	if err != nil {
		t.Fatal("Failed to get summoner name", err)
		return
	}
	t.Log("SummonerNames() returned: ", names)
	if names[testID] != testName {
		t.Fatalf("Invalid summoner name returned. Expected %s", testName)
		return
	}
}
