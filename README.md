
# lugh

This project loads MP3 recusively in a specified folder to extract ID3 v2.4 metadata and to generate a YAML file describing the music collection.
It can also apply the ID3 description of the collection to the MP3 files (including the cover image).

## Usage

- Scan files
```
tagger scan --folder ./ --config collection.yaml
```

- Apply description
```
tagger apply --folder ./ --config collection.yaml
```

## Dev

```
export GOROOT=/usr/local/go; export GOPATH=/go; export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

```go
go run cmd/tagger/main.go scan
```
