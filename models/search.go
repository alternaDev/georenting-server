package models

import (
  elastic "gopkg.in/olivere/elastic.v3"
  "log"
  "os"
  "errors"
)

const (
  IndexGeoFences = "geoFences"
)

// ElasticInstance is a usable ElasticSearch instance.
var ElasticInstance = initElastic(os.Getenv("ELASTICSEARCH_URL"))

func initElastic(url string) *elastic.Client {
  client, err := elastic.NewClient(elastic.SetURL(url))
  if err != nil {
      log.Fatalf("Error while connecting to ElasticSearch: %s", err)
      return nil
  }

  initIndices(client)

  return client
}

func initIndices(client *elastic.Client) error {
  exists, err := client.IndexExists(IndexGeoFences).Do()
  if err != nil {
  	return err
  }
  if !exists {
    log.Println("Creating Index for GeoFences.")
    createIndex, err := client.CreateIndex(IndexGeoFences).Do()
    if err != nil {
      // Handle error
      return err
    }
    if !createIndex.Acknowledged {
      return errors.New("Could not create Index for GeoFences")
    }
  }
  return nil
}
