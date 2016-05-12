package models

import (
  elastic "gopkg.in/olivere/elastic.v3"
  "os"
)

var ElasticInstance = initElastic(os.Getenv("ELASTICSEARCH_URL"))

func initElastic(url string) *elastic.Client {
  client, err := elastic.NewClient(elastic.SetURL(url))
  if err != nil {
      return nil
  }

  return client
}
