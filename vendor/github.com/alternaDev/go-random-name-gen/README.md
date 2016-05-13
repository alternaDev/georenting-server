# go-random-name-gen
Generate Random Names for Go.

# Usage

`GenerateName(amountOfAdjectives, amountOfNouns, placesForRandomNumber)`

Use like this:
```go

import (
  nameGen "github.com/alternaDev/go-random-name-gen"
)

name, err := nameGen.GenerateName(1, 1, 3) // => DreadfulJoe091
```

You can also include custom file arguments (but make sure you include them when deploying!):
`GenerateName(amountOfAdjectives, amountOfNouns, placesForRandomNumber, pathToAdjectiveFile, pathToNounFile)`

## Development
After adding new Adjectives or Nouns, please run `go generate` so they will get included.
