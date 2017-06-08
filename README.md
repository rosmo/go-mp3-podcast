# go-mp3-podcast

This is an utility to take a directory containing MP3s and turn them into an
RSS feed. The utility is configured using a simple YAML file (see config.yml
for example).

## How to run

### RSS mode

```
go-mp3-podcast -config config.yml -dir /path/to/mp3s -mode rss
```

### Directory index mode

```
go-mp3-podcast -config config.yml -dir /path/to/mp3s -mode index
```

## License

Licensed under MIT. Copyright 2017 Taneli Leppä <rosmo@rosmo.fi>

## Acknowledgements

Uses great libraries from:
- [eduncan911/podcast](https://github.com/eduncan911/podcast)
- [mikkyang/id3-go](https://github.com/mikkyang/id3-go)
