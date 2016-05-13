package nameGen

import (
  "bufio"
  "os"
  "math/rand"
  "bytes"
  "time"
  "strconv"
  "math"
)

//go:generate go run scripts/includetxt.go

var (
  fileCache map[string][]string
  random = rand.New(rand.NewSource(time.Now().UnixNano()))
)


// GenerateNameWithSeed generates a Random Name with adjectiveAmount Adjectives, nounAmount Nouns and a random Number with randomNumberPlaces places.
// The Random Generator uses the specified number as a seed.
func GenerateNameWithSeed(adjectiveAmount int, nounAmount int, randomNumberPlaces int, seed int64) (string, error) {
  random = rand.New(rand.NewSource(seed))

  return GenerateName(adjectiveAmount, nounAmount, randomNumberPlaces)
}

// GenerateName generates a Random Name with adjectiveAmount Adjectives, nounAmount Nouns and a random Number with randomNumberPlaces places.
func GenerateName(adjectiveAmount int, nounAmount int, randomNumberPlaces int) (string, error) {
  var nameBuffer bytes.Buffer
  for i := 0; i < adjectiveAmount; i++ {
    adj, err := getRandomAdjective()
    if err != nil {
      return "", err
    }
    nameBuffer.WriteString(adj)
  }

  for i := 0; i < nounAmount; i++ {
    noun, err := getRandomNoun()
    if err != nil {
      return "", err
    }
    nameBuffer.WriteString(noun)
  }

  for i := 0; i < randomNumberPlaces; i++ {
    nameBuffer.WriteString(strconv.Itoa(random.Intn(10)))
  }

  return nameBuffer.String(), nil
}

// GetPossibilities returns the amount of possible Names with the given parameters.
func GetPossibilities(adjectiveAmount int, nounAmount int, randomNumberPlaces int) (float64) {
  return math.Pow(float64(len(adjectives)), float64(adjectiveAmount)) *
       math.Pow(float64(len(nouns)), float64(nounAmount)) *
       math.Pow(10, float64(randomNumberPlaces))
}

// GenerateNameWithFiles generates a Random Name with adjectiveAmount Adjectives, nounAmount Nouns and a random Number with randomNumberPlaces places.
// You can use custom Files with this function.
func GenerateNameWithFiles(adjectiveAmount int, nounAmount int, randomNumberPlaces int, adjectivesFile string, nounsFile string) (string, error) {
  var nameBuffer bytes.Buffer

  for i := 0; i < adjectiveAmount; i++ {
    adj, err := getRandomLineFromFile(adjectivesFile)
    if err != nil {
      return "", err
    }
    nameBuffer.WriteString(adj)
  }

  for i := 0; i < nounAmount; i++ {
    noun, err := getRandomLineFromFile(nounsFile)
    if err != nil {
      return "", err
    }
    nameBuffer.WriteString(noun)
  }

  for i := 0; i < randomNumberPlaces; i++ {
    nameBuffer.WriteString(strconv.Itoa(random.Intn(10)))
  }

  return nameBuffer.String(), nil
}

func getRandomAdjective() (string, error) {
  line := adjectives[random.Intn(len(adjectives))]
  for line == "" {
    line = adjectives[random.Intn(len(adjectives))]
  }

  return line, nil
}

func getRandomNoun() (string, error) {
  line := nouns[random.Intn(len(nouns))]
  for line == "" {
    line = nouns[random.Intn(len(nouns))]
  }

  return line, nil
}

func getRandomLineFromStringArray(lines []string) (string, error) {
  line := lines[random.Intn(len(lines))]
  for line == "" {
    line = lines[random.Intn(len(lines))]
  }

  return line, nil
}

func getRandomLineFromFile(path string) (string, error) {
  lines, err := readFile(path)

  if err != nil {
    return "", err
  }

  return lines[random.Intn(len(lines))], nil
}

func readFile(path string) ([]string, error) {
  if fileCache == nil {
    fileCache = make(map[string][]string)
  }
  if fileCache[path] != nil {
    return fileCache[path], nil
  }

  inFile, _ := os.Open(path)
  defer inFile.Close()
  scanner := bufio.NewScanner(inFile)
  scanner.Split(bufio.ScanLines)

  var lines []string
  for scanner.Scan() {
    lines = append(lines, scanner.Text())
  }

  fileCache[path] = lines

  return lines, nil
}
