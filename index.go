package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/olivere/elastic/v7"
)

var (
	indexName = "movies"
	indices   = []string{indexName}
)

var indexMapping = `
{
	"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
	},
	"mappings":{
		"properties":{
			"show_id":{
				"type":"integer"
			},
			"type":{
				"type":"keyword"
			},
			"title":{
				"type":"wildcard"
			},
			"director":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"cast":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"country":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"date_added":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"release_year":{
				"type":"integer"
			},
			"rating":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"duration":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"listed_in":{
				"type":"text",
				"store": true,
				"fielddata": true
			},
			"description":{
				"type":"text",
				"store": true,
				"fielddata": true
			}
		}
	}
}
`

func createMovieIndex(ctx context.Context, esURL string, csvPath string) (string, error) {
	client, err := connectElasticsearch(esURL)
	if err != nil {
		return "", err
	}

	// Check whether index already exists
	indexExistsService := elastic.NewIndicesExistsService(client)
	exist, err := indexExistsService.Index(indices).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("failed while checking index existence due to: %s", err)
	}
	if exist {
		return "index already exists. (aborting)", nil
	}

	// Create movie index
	createIndexService := elastic.NewIndicesCreateService(client)
	created, err := createIndexService.Index(indexName).Body(indexMapping).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("failed while creating index due to: %s", err)
	}
	if created == nil {
		return "", errors.New("expected create index result to be != nil")
	}
	log.Printf("Index %s is created.", indexName)

	// Load the movie data
	movies, err := loadMovies(csvPath)
	if err != nil {
		return "", fmt.Errorf("failed to load movies from '%s' due to: %s", csvPath, err)
	}
	moviesCount := len(movies)
	log.Printf("There is %d movies to be indexed in Elasticsearch.", moviesCount)

	// Bulk insert
	bulk := client.Bulk()
	for i, movie := range movies {
		bulkIndexRequest := elastic.NewBulkIndexRequest()
		bulkIndexRequest.OpType("index")
		bulkIndexRequest.Index(indexName)
		bulkIndexRequest.Id(fmt.Sprintf("%d", movie.ShowID))
		bulkIndexRequest.Doc(movie)

		bulk.Add(bulkIndexRequest)
		log.Printf("Added movie #%d to bulk: ID '%d', Title '%s'", i, movie.ShowID, movie.Title)
	}
	bulkCount := bulk.NumberOfActions()
	log.Printf("Bulk is ready with %d movies.", bulkCount)

	bulkResponse, err := bulk.Do(ctx)
	if err != nil {
		return "", fmt.Errorf("failed while bulk indexing %d movies due to: %s", bulkCount, err)
	}
	if bulkResponse == nil {
		return "", errors.New("expected bulk indexing result to be != nil")
	}

	// Bulk indexing report
	indexed := len(bulkResponse.Indexed())
	log.Printf("Bulk indexed %d/%d movies.", indexed, bulkCount)

	return fmt.Sprintf("movies indexed in Elasticsearch: %d/%d", indexed, bulkCount), nil
}

func destroyMovieIndex(ctx context.Context, esURL string) (string, error) {
	client, err := connectElasticsearch(esURL)
	if err != nil {
		return "", err
	}

	// Check whether index already exists
	indicesExistsService := elastic.NewIndicesExistsService(client)
	exist, err := indicesExistsService.Index(indices).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("failed while checking index existence due to: %s", err)
	}
	if !exist {
		return "index doesn't exist. (aborting)", nil
	}

	// Delete movie index
	deleteIndexService := elastic.NewIndicesDeleteService(client)
	deleted, err := deleteIndexService.Index(indices).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("failed while deleting index due to: %s", err)
	}
	if deleted == nil {
		return "", errors.New("expected delete index result to be != nil")
	}
	log.Printf("Index %s is deleted.", indexName)

	return "movies indexed is deleted in Elasticsearch", nil
}

func searchMovieIndex(ctx context.Context, esURL string, q string) (interface{}, error) {
	client, err := connectElasticsearch(esURL)
	if err != nil {
		return "", err
	}

	// Check whether index already exists
	indicesExistsService := elastic.NewIndicesExistsService(client)
	exist, err := indicesExistsService.Index(indices).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("failed while checking index existence due to: %s", err)
	}
	if !exist {
		return "index doesn't exist. (aborting)", nil
	}

	// Search
	searchService := elastic.NewSearchService(client)
	result, err := searchService.Index(indexName).
		Query(elastic.NewWildcardQuery("title", fmt.Sprintf("%s*", q))).
		Aggregation("autocomplete", elastic.NewTermsAggregation().Field("title").OrderByCount(false).Size(10)).
		Do(ctx)
	if err != nil {
		return "", fmt.Errorf("failed while searching index with '%s' due to: %s", q, err)
	}
	if result == nil {
		return "", errors.New("expected search result to be != nil")
	}
	if result.Aggregations == nil {
		return "", errors.New("expected search aggregation to be != nil")
	}

	aggs, found := result.Aggregations.Terms("autocomplete")
	if !found {
		return "", errors.New("expected search aggregation to have autocomplete term")
	}

	titles := []string{}
	for _, bucket := range aggs.Buckets {
		titles = append(titles, fmt.Sprintf("%v", bucket.Key))
	}

	content := map[string]interface{}{}
	content["count"] = result.TotalHits()
	content["titles"] = titles

	return content, nil
}

func connectElasticsearch(esURL string) (*elastic.Client, error) {
	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(esURL),
		elastic.SetHealthcheckInterval(5*time.Second), // quit trying after 5 seconds
	)
	if err != nil {
		return nil, fmt.Errorf("failed while connecting to Elasticsearch at '%s' due to: %s", esURL, err)
	}
	log.Printf("Connected to Elasticsearch at '%s'", esURL)

	return client, nil
}

func loadMovies(csvPath string) ([]Movie, error) {
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open '%s' file due to: %s", csvPath, err)
	}

	r := csv.NewReader(csvFile)

	_, err = r.Read()
	if err != nil {
		return nil, fmt.Errorf("Couldn't read header (a.k.a. first line) of CVS file due to: %s", err)
	}

	movies := []Movie{}
	i := 0
	for {
		i++
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Couldn't read CSV record #%d due to: %s", i, err)
		}

		showID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse show_id due to: %s", err)
		}

		releaseYear, err := strconv.Atoi(record[7])
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse release_year due to: %s", err)
		}

		movie := Movie{
			ShowID:      showID,
			Type:        record[1],
			Title:       record[2],
			Director:    record[3],
			Cast:        record[4],
			Country:     record[5],
			DateAdded:   record[6],
			ReleaseYear: releaseYear,
			Rating:      record[8],
			Duration:    record[9],
			ListedIn:    record[10],
			Description: record[11],
		}

		movies = append(movies, movie)
	}

	return movies, nil
}
