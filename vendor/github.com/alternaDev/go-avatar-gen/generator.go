package avatarGen

import (
  "crypto/md5"
  "hash/fnv"
  "fmt"
  "io"
  "strings"
  "math/rand"
  "math"
  "image"
	"image/color"
	"image/png"
  "os"
  "net/http"
  "bytes"
  "strconv"
)

const (
  low = "01234567"
  size = 5
)

func main() {
  saveImage(GenerateAvatar("jhhg", 64, 32))
}

func saveImage(avatar *image.RGBA) {
  file, err := os.Create("1.png")
	err = png.Encode(file, avatar)

	if err != nil {
		panic(err)
	}
}

// GenerateAvatar generates an avatar (image.RGBA) using the given string as
// a random seed.
func GenerateAvatar(input string, blockSize int, borderSize int) (*image.RGBA) {
  hash := hashMd5(input)

  pic := [size][size]bool{}

  for i := size - 4; i >= 0; i-- {
    for j := size - 1; j >= 0; j-- {
      if strings.Contains(low, string(hash[size - 1 * i + j])) {
        pic[j][i] = true
        pic[j][size - 1 - i] = true
      }
    }
  }
  for i := size - 1; i >= 0; i-- {
    if strings.Contains(low, string(hash[i])) {

      pic[i][int(math.Ceil(size / 2))] = true
    }
  }

  avatar := image.NewRGBA(image.Rect(0, 0, blockSize * size + borderSize * 2,
    blockSize * size + borderSize * 2))
  bgColor := calcBgColor()

  for x := 0; x < avatar.Bounds().Dx(); x++ {
    for y := 0 ; y < avatar.Bounds().Dy(); y++ {
      avatar.SetRGBA(x, y, bgColor)
    }
  }

  color := calcPixelColor(input)

  for x := 0; x < len(pic); x++ {
    for y := 0; y < len(pic[x]); y++ {
      if(pic[x][y]) {
        for i := 0; i < blockSize; i++ {
          for j := 0; j < blockSize; j++ {
            avatar.SetRGBA(borderSize + x * blockSize + i,
              borderSize + y * blockSize + j, color)
          }
        }
      }
    }
  }

  return avatar
}

// WriteImageToHTTP Sends an image via http.
func WriteImageToHTTP(w http.ResponseWriter, img image.Image) (error) {
  buffer := new(bytes.Buffer)
  if err := png.Encode(buffer, img); err != nil {
    return err
  }

  w.Header().Set("Content-Type", "image/png")
  w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
  if _, err := w.Write(buffer.Bytes()); err != nil {
    return err
  }
  return nil
}

func calcPixelColor(input string) (pixelColor color.RGBA) {
  inputBytes := []byte(input)
  random := rand.New(rand.NewSource(int64(hash(input))))

  value := random.Float64()

	pixelColor.A = 255

	var mutator = byte(value * 4)

	pixelColor.R = inputBytes[0] * mutator
	pixelColor.G = inputBytes[1] * mutator
	pixelColor.B = inputBytes[2] * mutator

	return pixelColor
}

func calcBgColor() (pixelColor color.RGBA) {
  pixelColor.A = 255

	pixelColor.R = 240
	pixelColor.G = 240
	pixelColor.B = 240

	return
}

func hashMd5(input string) (string) {
  h := md5.New()
  io.WriteString(h, input)
  return fmt.Sprintf("%x", h.Sum(nil))
}

func hash(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}
