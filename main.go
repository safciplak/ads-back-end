package main

import (
	"encoding/json"
	"fmt"
	"github.com/dghubble/sling"
	"github.com/jdkato/prose/v2"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type SynonymResponse struct {
	Word     string   `json:"word"`
	Synonyms []string `json:"synonyms"`
}

type SearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Response struct {
	Results []SearchResult `json:"results"`
}

func getSynonyms(word string) []string {
	apiURL := fmt.Sprintf("https://api.datamuse.com/words?rel_syn=%s", word)

	var synonyms []string
	response := new([]SynonymResponse)

	_, err := sling.New().Get(apiURL).ReceiveSuccess(response)
	if err != nil {
		log.Fatalf("Error fetching synonyms: %v", err)
	}

	for _, item := range *response {
		synonyms = append(synonyms, item.Word)
	}

	return synonyms
}

func generateVariations(phrase string) []string {
	doc, err := prose.NewDocument(phrase)
	if err != nil {
		log.Fatal(err)
	}

	var variations []string
	variations = append(variations, phrase)

	for _, token := range doc.Tokens() {
		synonyms := getSynonyms(token.Text)
		for _, syn := range synonyms {
			newPhrase := strings.Replace(phrase, token.Text, syn, 1)
			variations = append(variations, newPhrase)
			if len(variations) >= 7 {
				return variations
			}
		}
	}

	return variations
}

func addQueryParam(baseUrl, query string) string {
	u, err := url.Parse(baseUrl)
	if err != nil {
		fmt.Printf("URL parse error: %v\n", err)
		return ""
	}

	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()
	return u.String()
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	queryParam := r.URL.Query().Get("query")
	if queryParam == "" {
		http.Error(w, "Query is empty.", http.StatusBadRequest)
		return
	}

	urls := []string{
		"https://www.gileq.com/dsr?q=",
		"https://search.searchalike.com/serp?q=",
		"https://uk.questtips.com/dsr?q=",
		"https://www.novafluxa.com/dsr?q=",
		"https://explorewebzone.com/dsr?q=",
		"https://www.astartex.com/dsr/?q=",
		"https://nexizonal.com/dsr?q=",
	}

	phrase := queryParam

	variations := generateVariations(phrase)

	var response Response

	// Ensure each URL receives only one unique variation
	for i, variation := range variations {
		if i < len(urls) {
			fullUrl := addQueryParam(urls[i], variation)
			response.Results = append(response.Results, SearchResult{Title: variation, URL: fullUrl})
		} else {
			// if more variations than URLs, repeat the URL list
			fullUrl := addQueryParam(urls[i%len(urls)], variation)
			response.Results = append(response.Results, SearchResult{Title: variation, URL: fullUrl})
		}
	}

	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling response to JSON: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func main() {
	http.HandleFunc("/search", searchHandler)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
