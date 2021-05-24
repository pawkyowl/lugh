package tagger

import (
	"bytes"
	"errors"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v2"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	fileFormat    = ".mp3"
	pictureFormat = "image/png"
)

func scanFolder(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			if filepath.Ext(path) == fileFormat {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func getPicture(arr []byte) (image.Image, error) {
	values := strings.SplitN(getString(arr), "\x00", 3)
	if len(values) != 3 || len(values[1]) == 0 {
		return nil, errors.New("incorrect tag")
	}
	if values[0] != pictureFormat {
		log.Println("Warning: picture format must be " + pictureFormat)
	}
	data := []byte(values[2])
	return png.Decode(bytes.NewReader(data))
}

func getString(arr []byte) string {
	if len(arr) >= 2 && (arr[0] == 0 || arr[0] == 3) {
		if arr[len(arr)-1] == 0 {
			return string(arr[1 : len(arr)-1])
		} else {
			return string(arr[1:])
		}
	} else {
		return ""
	}
}

func getBytes(value string) []byte {
	result := []byte{0}
	return append(result, []byte(value)...)
}

func getPictureString(picture image.Image) string {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, picture)
	if err != nil {
		return ""
	}
	result := []byte(pictureFormat)
	result = append(result, 0x00)
	result = append(result, 2)
	result = append(result, []byte("")...)
	result = append(result, 0x00)
	result = append(result, buf.Bytes()...)
	return string(result)
}

func formatString(str string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	res, _, _ := transform.String(t, str)
	return strings.ReplaceAll(strings.Title(strings.ToLower(res)), "\"", "")
}

func loadConfig(path string) (map[string]*Album, error) {
	albums := make(map[string]*Album)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &albums)
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func saveConfig(path string, albums map[string]*Album) error {
	data, err := yaml.Marshal(&albums)
	if err != nil {
		return err
	}
	log.Printf("--- dump:\n%s", string(data))
	err = ioutil.WriteFile(path, []byte(data), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func loadPicture(path string) (image.Image, error) {
	log.Println("Load picture " + path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}

func savePicture(path string, picture image.Image) error {
	if picture == nil {
		return errors.New("Save picture error")
	}
	log.Println("Save picture " + path)
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(out, picture); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return nil
}
