package lookup

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"
)

type SearchResult struct {
	Value string `json:"value"`
	Druh  string `json:"druh"`
	Cislo string `json:"cislo"`
	Nazev string `json:"nazev"`
	Trasa string `json:"trasa"`
	Zeme  string `json:"zeme"`
}

type VagonWebLookup struct {
	client   *http.Client
	baseURL  *url.URL
	trainURL *url.URL
}

func NewVagonWebLookup() *VagonWebLookup {
	baseURL, err := url.Parse("https://www.vagonweb.cz")
	if err != nil {
		// Doing this because I don't like not handling errors, but this should not happen anyway
		panic(err)
	}
	trainURL, err := url.Parse("https://www.vagonweb.cz/razeni/vlak.php")
	if err != nil {
		// Doing this because I don't like not handling errors, but this should not happen anyway
		panic(err)
	}
	return &VagonWebLookup{
		client:   &http.Client{Timeout: time.Second * 10},
		baseURL:  baseURL,
		trainURL: trainURL,
	}
}

// GetTrainLink tries to find the Train on vagonweb.cz and returns the url to the corresponding page
func (v *VagonWebLookup) GetTrainLink(searchterm string) (string, error) {
	searchURL := v.baseURL.JoinPath("/razeni/json_vlaky.php")
	query := make(url.Values)
	query.Set("jmeno", searchterm)
	searchURL.RawQuery = query.Encode()
	resp, err := v.client.Get(searchURL.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	searchresults := make([]SearchResult, 0)
	err = dec.Decode(&searchresults)
	if len(searchresults) > 0 {
		m := make(url.Values)
		m.Set("zeme", searchresults[0].Zeme)
		m.Set("kategorie", searchresults[0].Druh)
		m.Set("cislo", searchresults[0].Cislo)
		v.trainURL.RawQuery = m.Encode()
		return v.trainURL.String(), nil
	}

	return "", errors.New("nothing found")

}
