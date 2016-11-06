package search

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	elastic "gopkg.in/olivere/elastic.v3"

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

func parseBonsaiURL(url string) (string, string, string) {
	rex, _ := regexp.Compile(".*?://([a-z0-9]{1,}):([a-z0-9]{1,})@.*$")
	user := rex.ReplaceAllString(url, "$1")
	pass := rex.ReplaceAllString(url, "$2")
	host := strings.Replace(url, user+":"+pass+"@", "", -1)
	return user, pass, host
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

// IndexGeoFence indexes a geofence.
func IndexGeoFence(fence models.Fence) error {
	data := fmt.Sprintf(`{"name": "%s", "center": {"lat": %f, "lon": %f}, "radius": %d, "owner": %d}`, fence.Name, fence.Lat, fence.Lon, fence.Radius, fence.User.ID)
	log.Println("Indexing: " + data)
	_, err := ElasticInstance.Index().
		Index(IndexGeoFences).
		Type(TypeGeoFence).
		Id(strconv.Itoa(int(fence.ID))).
		BodyString(data).
		Do()

	return err
}

// FindGeoFences returns all geofences around a lat/lon pair.
func FindGeoFences(centerLat float64, centerLon float64, radius int64) ([]models.Fence, error) {
	query := elastic.NewGeoDistanceQuery("center").Distance(fmt.Sprintf("%d m", radius)).Lat(centerLat).Lon(centerLon)

	searchResult, err := ElasticInstance.Search().
		Index(IndexGeoFences).
		Query(query).
		Do()

	// Check whether an error appeared or not.
	if err != nil {
		return nil, err
	}

	if searchResult.Hits != nil {
		fences := make([]int64, searchResult.TotalHits())
		fmt.Printf("Found a total of %d GeoFences\n", searchResult.Hits.TotalHits)

		// Iterate through results
		for i, hit := range searchResult.Hits.Hits {
			stringID, _ := strconv.ParseInt(hit.Id, 10, 64)
			fences[i] = stringID
		}
		return models.FindFencesByIDs(fences)
	}

	fmt.Print("Found no fences\n")
	var empty []models.Fence
	return empty, nil
}

// FindGeoFencesExceptByUser returns all geofences around a lat/lon pair, excluding ones from the specified user.
func FindGeoFencesExceptByUser(centerLat float64, centerLon float64, radius int64, excludeBy int) ([]models.Fence, error) {
	query := elastic.NewBoolQuery()
	query = query.MustNot(elastic.NewTermQuery("owner", excludeBy))
	query.Filter(elastic.NewGeoDistanceQuery("center").Distance(fmt.Sprintf("%d m", radius)).Lat(centerLat).Lon(centerLon))

	searchResult, err := ElasticInstance.Search().
		Index(IndexGeoFences).
		Query(query).
		Do()

	// Check whether an error appeared or not.
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
		return models.FindFencesByIDs(fences)
	}

	fmt.Print("Found no fences\n")
	var empty []models.Fence
	return empty, nil
}

// DeleteGeoFence deletes a geofence from the search index.
func DeleteGeoFence(fenceId int) error {
	_, err := ElasticInstance.Delete().Index(IndexGeoFences).Type(TypeGeoFence).Id(strconv.Itoa(fenceId)).Do()
	return err
}
