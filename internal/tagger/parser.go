package tagger

import (
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Collection struct {
	albums map[string]*Album
}

func LoadCollection(path string) (*Collection, error) {
	albums, err := loadConfig(path)
	if err != nil {
		return nil, err
	}
	return &Collection{albums: albums}, nil
}

func (collection *Collection) Save(path string) error {
	return saveConfig(path, collection.albums)
}

func (collection *Collection) Albums() []string {
	keys := make([]string, 0, len(collection.albums))
	for k := range collection.albums {
		keys = append(keys, k)
	}
	return keys
}

func (collection *Collection) GetAlbum(key string) *Album {
	if album, ok := collection.albums[key]; ok {
		return album
	}
	return nil
}

type Album struct {
	Path    string
	Cover   string
	picture image.Image
	Album   string `json:"album"`
	Artist  string `json:"artist"`
	Year    int    `json:"year"`
	Genre   string `json:"genre"`
	Tracks  []*Track
}

func NewAlbum(track *Track) *Album {
	return &Album{
		Path:    track.dir,
		Cover:   track.album.Cover,
		picture: track.album.picture,
		Album:   track.album.Album,
		Artist:  track.album.Artist,
		Year:    track.album.Year,
		Genre:   track.album.Genre,
	}
}

func (album *Album) Compare(newAlbum *Album) {
	log.Println("---")
	log.Println("Compare " + album.Path + " with " + album.Path)
	if album.Cover != newAlbum.Cover {
		log.Println("diff cover: " + album.Cover + " " + newAlbum.Cover)
	}
	if album.Album != newAlbum.Album {
		log.Println("diff album: " + album.Album + " " + newAlbum.Album)
	}
	if album.Artist != newAlbum.Artist {
		log.Println("diff artist: " + album.Artist + " " + newAlbum.Artist)
	}
	if album.Year != newAlbum.Year {
		log.Printf("diff year: %d %d", album.Year, newAlbum.Year)
	}
	if album.Genre != newAlbum.Genre {
		log.Println("diff genre: " + album.Genre + " " + newAlbum.Genre)
	}
	for i, track := range album.Tracks {
		newTrack := album.Tracks[i]
		if track.Filename != newTrack.Filename {
			log.Println("diff filename: " + track.Filename + " " + newTrack.Filename)
		}
		if track.Track != newTrack.Track {
			log.Printf("diff track: %d %d", track.Track, newTrack.Track)
		}
		if track.Title != newTrack.Title {
			log.Println("diff title: " + track.Title + " " + newTrack.Title)
		}
	}
	log.Println("---")
}

func (album *Album) Copy(newAlbum *Album) {
	album.Cover = newAlbum.Cover
	picture, err := loadPicture(album.Cover)
	if err == nil {
		album.picture = picture
	} else {
		log.Printf("Error reading picture: %s", err)
	}
	album.Album = newAlbum.Album
	album.Artist = newAlbum.Artist
	album.Year = newAlbum.Year
	album.Genre = newAlbum.Genre
	for i, track := range album.Tracks {
		newTrack := album.Tracks[i]
		track.album = album
		track.Filename = newTrack.Filename
		track.Track = newTrack.Track
		track.Title = newTrack.Title
	}
}

func (album *Album) SavePicture() error {
	if album.picture != nil {
		return savePicture(album.Cover, album.picture)
	} else {
		return nil
	}
}

type Track struct {
	album    *Album
	folder   string
	dir      string
	tags     map[string][]byte
	data     []byte
	File     string
	Filename string
	Track    int    `json:"track"`
	Title    string `json:"title"`
}

func NewTrack(filename string, path string) *Track {
	dir := filepath.Dir(path)
	parent := filepath.Base(dir)
	album := formatString(parent)
	cover := filepath.Join(dir, album+".png")
	return &Track{File: filename, folder: album, dir: dir, album: &Album{Path: path, Cover: cover, Album: album}}
}

func (track *Track) setNewFilename() {
	if len(track.Title) == 0 {
		track.Filename = ""
	} else {
		track.Filename = fmt.Sprintf("%02d", track.Track) + " - " + formatString(track.Title) + ".mp3"
	}
}

func (track *Track) setMetadata() {
	for key := range track.tags {
		valueBytes := track.tags[key]
		if key == "TPE1" {
			track.album.Artist = getString(valueBytes)
		} else if key == "TALB" {
			track.album.Album = getString(valueBytes)
		} else if key == "TDOR" {
			track.album.Year = 0
			str := string(valueBytes[1:])
			if len(str) > 0 {
				date, err := time.Parse("2006-01-02T15:04:05", str)
				if err == nil {
					track.album.Year = date.Year()
				} else {
					x, err := strconv.Atoi(str)
					if err == nil {
						track.album.Year = x
					} else {
						log.Println(err)
					}
				}
			}
		} else if key == "TCON" {
			track.album.Genre = getString(valueBytes)
		} else if key == "TIT2" {
			track.Title = getString(valueBytes)
		} else if key == "TRCK" {
			track.Track = 0
			str := getString(valueBytes)
			if len(str) > 0 {
				x, err := strconv.Atoi(str)
				if err == nil {
					track.Track = x
				} else {
					log.Println(err)
				}
			}
		} else if key == "APIC" {
			picture, err := getPicture(valueBytes)
			if err == nil {
				track.album.picture = picture
			} else {
				log.Printf("Error reading picture: %s", err)
			}
		} else {
			log.Println("Unkown tag " + key)			
		}
	}
}

func (track *Track) Save() error {
	if len(track.Filename) == 0 {
		return errors.New("Save track error, empty filename")
	}
	path := filepath.Join(track.album.Path, track.Filename)
	log.Printf("Save file %s", path)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	headerByte := make([]byte, 10)
	copy(headerByte[0:3], "ID3")
	copy(headerByte[3:6], []byte{4, 0, 0})
	length := 0
	for key := range track.tags {
		switch key {
		case "TPE1": track.tags[key] = getBytes(track.album.Artist)
		case "TALB": track.tags[key] = getBytes(track.album.Album)
		case "TDOR": track.tags[key] = getBytes(strconv.Itoa(track.album.Year))
		case "TCON": track.tags[key] = getBytes(track.album.Genre)
		case "TIT2": track.tags[key] = getBytes(track.Title)
		case "TRCK": track.tags[key] = getBytes(strconv.Itoa(track.Track))
		case "APIC": track.tags[key] = getBytes(getPictureString(track.album.picture))
		}
	}
	for _, val := range track.tags {
		length += 10 + len(val)
	}
	lengthByte := []byte{
		byte(length >> 21),
		byte(length >> 14),
		byte(length >> 7),
		byte(length),
	}
	copy(headerByte[6:10], lengthByte)
	n, err := file.Write(headerByte)
	if err != nil {
		return err
	}
	if n != 10 {
		return errors.New("Save track error")
	}
	for key, value := range track.tags {
		header := make([]byte, 10)
		for i, val := range key {
			header[i] = byte(val)
		}
		length := len(value)
		header[4] = byte(length >> 21)
		header[5] = byte(length >> 14)
		header[6] = byte(length >> 7)
		header[7] = byte(length)
		_, err := file.Write(header)
		if err != nil {
			return err
		}
		_, err = file.Write(value)
		if err != nil {
			return err
		}
	}
	_, err = file.Write(track.data)
	if err != nil {
		return err
	}
	if track.File != track.Filename {
		return os.Remove(filepath.Join(track.album.Path, track.File))
	} else {
		return nil
	}
}

func Scan(dir string) *Collection {
	albums := make(map[string]*Album)
	files, _ := scanFolder(dir)
	for _, path := range files {
		log.Printf("Parse %s", path)
		track, err := parseFile(path)
		if err == nil {
			folder := track.folder
			if _, ok := albums[folder]; !ok {
				albums[folder] = NewAlbum(track)
			}
			albums[folder].Tracks = append(albums[folder].Tracks, track)
		}
	}
	return &Collection{albums: albums}
}

func parseFile(path string) (*Track, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if reader == nil {
		return nil, errors.New("error file is null")
	}
	defer reader.Close()
	if err != nil {
		return nil, err
	}
	stat, err := reader.Stat()
	if err != nil {
		return nil, err
	}
	track := NewTrack(stat.Name(), path)
	header := make([]byte, 10)
	_, err = reader.Read(header)
	if err != nil {
		return nil, err
	}
	marker := string(header[0:3])
	if marker != "ID3" {
		return nil, errors.New("error file marker")
	}
	version := header[3]
	if version != 4 {
		return nil, errors.New("unsupported id3 version")
	}
	length := 0
	for _, x := range header[6:10] {
		length = (length << 7) | int(x)
	}
	track.tags = make(map[string][]byte)
	cursor := 0
	for cursor < length {
		val := make([]byte, 10)
		_, err := reader.Read(val)
		if err != nil {
			return nil, err
		}
		if val[0] == 0 || val[1] == 0 || val[2] == 0 || val[3] == 0 {
			break
		}
		key := string(val[0:4])
		size := 0
		for _, x := range val[4:8] {
			size = (size << 7) | int(x)
		}
		valueBytes := make([]byte, size)
		_, err = reader.Read(valueBytes)
		if err != nil {
			return nil, err
		}
		track.tags[key] = valueBytes
		cursor += 10 + size
	}
	track.setMetadata()
	track.setNewFilename()
	track.data, err = ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return track, nil
}
