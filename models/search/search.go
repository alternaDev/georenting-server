package search

import (
  elastic "gopkg.in/olivere/elastic.v3"
  "regexp"
  "strings"
  "log"
  "os"
  "errors"
  "fmt"
  "strconv"

  models "github.com/alternaDev/georenting-server/models"
)

const (
  // IndexGeoFences specifies the name of the geofences index
  IndexGeoFences = "geofences"
  // TypeGeoFence specifies the name of the geofences type
  TypeGeoFence = "geofence"
)

// ElasticInstance is a usable ElasticSearch instance.
var ElasticInstance *elastic.Client

func parseBonsaiURL(url string) (string, string, string){
	rex, _ := regexp.Compile(".*?://([a-z0-9]{1,}):([a-z0-9]{1,})@.*$")
	user := rex.ReplaceAllString(url, "$1")
	pass := rex.ReplaceAllString(url, "$2")
	host := strings.Replace(url, user+":"+pass+"@", "", -1)
	return user,pass,host
}

func init() {
  log.Println("Initializing ElasticSearch.")

  elastic, err := initElastic(os.Getenv("ELASTICSEARCH_URL"))
  if err != nil {
		panic(err)
	}
  ElasticInstance = elastic
}

func initElastic(www string) (*elastic.Client, error) {
  username, password, host := parseBonsaiURL(www)

  log.Printf("Initializing ES: %v.", host)

  client, err := elastic.NewClient(elastic.SetURL(host), elastic.SetMaxRetries(10), elastic.SetBasicAuth(username, password), elastic.SetSniff(false))
  if err != nil {
      log.Fatalf("Error while connecting to ElasticSearch: %s", err)
      return nil, err
  }

  log.Println("Initializing Indices.")

  err = initIndices(client)
  if err != nil {
      log.Fatalf("Error while creating ElasticSearch Indices: %s", err)
      return nil, err
  }

  return client, err
}

func initIndices(client *elastic.Client) error {
  exists, err := client.IndexExists(IndexGeoFences).Do()
  if err != nil {
  	return err
  }
  if !exists {
    log.Println("Creating Index for GeoFences.")
    mapping := `{
        "settings":{
            "number_of_shards":1,
            "number_of_replicas":0
        },
        "mappings":{
            "geofence":{
                "properties":{
                    "name":{
                        "type":"string"
                    },
                    "radius":{
                        "type":"double"
                    },
                    "center":{
                        "type":"geo_point"
                    },
                    "owner": {
                        "type": "string"
                    }
                }
            }
        }
    }`
    createIndex, err := client.CreateIndex(IndexGeoFences).BodyString(mapping).Do()
    if err != nil {

      return err
    }
    if !createIndex.Acknowledged {
      return errors.New("Could not create Index for GeoFences")
    }
  }
  return nil
}

/*func MigrateGeofencesToElasticSearch() {
  log.Print("Migrating to ElasticSearch")
  var geoFences []Fence
  DB.Find(&geoFences)

  for _, fence := range geoFences {
    log.Printf("Indexing Fence %v %v", fence.ID, fence.Name)

    err := IndexGeoFence(&fence)
    if err != nil {
      log.Fatal(err)
    }
  }
}*/

func IndexGeoFence(fence *models.Fence) error {
  data := fmt.Sprintf(`{"name": "%s", "center": {"lat": %f, "lon": %f}, "radius": %d, "owner": %d}`, fence.Name, fence.Lat, fence.Lon, fence.Radius, fence.UserID);
  log.Println("Indexing: " + data)
  _, err := ElasticInstance.Index().
    Index(IndexGeoFences).
    Type(TypeGeoFence).
    Id(strconv.Itoa(int(fence.ID))).
    BodyString(data).
    Do()

  return err
}

func FindGeoFences(centerLat float64, centerLon float64, radius int64) ([]int64, error) {
  query := elastic.NewGeoDistanceQuery("center").Distance(fmt.Sprintf("%dm", radius)).Lat(centerLat).Lon(centerLon)

  searchResult, err := ElasticInstance.Search().
    Index(IndexGeoFences).
    Query(query).
    Do()

  if err != nil {
    return nil, err
  }

  if searchResult.Hits != nil {
    fences := make([]int64, searchResult.TotalHits(), searchResult.TotalHits())
    fmt.Printf("Found a total of %d GeoFences\n", searchResult.Hits.TotalHits)

    // Iterate through results
    for i, hit := range searchResult.Hits.Hits {
      stringID, _ := strconv.ParseInt(hit.Id, 10, 64)
      fences[i] = stringID
    }
    return fences, nil
  }

  fmt.Print("Found no fences\n")
  return make([]int64, 0), nil
}

func DeleteGeoFence(fence *models.Fence) error {
  _, err := ElasticInstance.Delete().Index(IndexGeoFences).Id(strconv.Itoa(int(fence.ID))).Do()
  return err
}