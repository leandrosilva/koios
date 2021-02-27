GET /_cat/indices

GET /movies/_search

GET /movies/_search
{
    "query": {
        "wildcard": {
            "title": {
                "value": "A Bo*"
            }
        }
    } 
}

GET /movies/_search
{
    "_source": [],
    "size": 0,
    "query": {
        "wildcard": {
            "title": {
                "value": "A B*"
            }
        }
    },
    "aggregations": {
        "autocomplete": {
            "terms": {
                "field": "title",
                "order": {
                    "_count": "desc"
                },
                "size": 25
            }
        }
    }
}

GET /movies/_search
{
    "_source": [],
    "size": 0,
    "min_score": 0.5,
    "query": {
        "bool": {
            "must": [
                {
                    "wildcard": {
                        "title": {
                            "value": "A Bo*"
                        }
                    }
                }
            ],
            "filter": [],
            "should": [],
            "must_not": []
        }
    },
    "aggs": {
        "autocomplete": {
            "terms": {
                "field": "title.keyword",
                "order": {
                    "_count": "desc"
                },
                "size": 25
            }
        }
    }
}

GET /movies/_mapping

DELETE /movies

PUT /movies
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
