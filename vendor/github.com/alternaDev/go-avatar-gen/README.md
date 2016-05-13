# go-avatar-gen
Generate Avatars for Users.

# Usage
`GenerateAvatar(input string, blockSize int, borderSize int`

Use like this:
```go

import (
  avatarGen "github.com/alternaDev/go-avatar-gen"
)

avatar := avatarGen.GenerateAvatar("HeinzPeda", 64, 32) // => *image.RGBA
```

You can then write this avatar to a HTTP Response like this:

```go
import (
  avatarGen "github.com/alternaDev/go-avatar-gen"
)

avatar := avatarGen.GenerateAvatar("HeinzPeda", 64, 32) // => *image.RGBA
err := avatarGen.WriteImageToHTTP(respWriter, avatar)
```
