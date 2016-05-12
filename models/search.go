package models

import (
  elastic "gopkg.in/olivere/elastic.v3"
  "log"
  "os"
  "errors"
  "fmt"
)

const (
  // IndexGeoFences specifies the name of the geofences index
  IndexGeoFences = "geoFences"
  // TypeGeoFence specifies the name of the geofences type
  TypeGeoFence = "geoFence"
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

func MigrateGeofencesToElasticSearch() {
  log.Print("Migrating to ElasticSearch")
  var geoFences []Fence
  DB.Find(&geoFences)

  for _, fence := range geoFences {
    log.Printf("Indexing Fence %v %v", fence.Key(), fence.Name)

    err := IndexGeoFence(&fence)
    if err != nil {
      log.Fatal(err)
    }
  }
}

func IndexGeoFence(fence *Fence) error {
  data := fmt.Sprintf("{'name': '%s', 'center': {'location': {'lat': %f, 'lon': %f}}, 'radius': %d, 'ownerId': %d}", fence.Name, fence.Lat, fence.Lon, fence.Radius, fence.UserID);
  _, err := ElasticInstance.Index().
    Index(IndexGeoFences).
    Type(TypeGeoFence).
    Id(fence.Key()).
    BodyString(data).
    Do()

  return err
}

func FindGeoFences(centerLat float64, centerLon float64, radius int) ([]string, error) {
  query := elastic.NewGeoDistanceQuery("center").Distance(fmt.Sprintf("%dm", radius)).Lat(centerLat).Lon(centerLon)

  searchResult, err := ElasticInstance.Search().
    Index(IndexGeoFences).
    Query(query).
    Do()

  if err != nil {
    return nil, err
  }

  if searchResult.Hits != nil {
    fences := make([]string, searchResult.TotalHits(), searchResult.TotalHits())
    fmt.Printf("Found a total of %d GeoFences\n", searchResult.Hits.TotalHits)

    // Iterate through results
    for i, hit := range searchResult.Hits.Hits {
      fences[i] = hit.Id
    }
    return fences, nil
  }

  fmt.Print("Found no fences\n")
  return make([]string, 0), nil
}

func DeleteGeoFence(id string) error {
  _, err := ElasticInstance.Delete().Index(IndexGeoFences).Id(id).Do()
  return err
}
