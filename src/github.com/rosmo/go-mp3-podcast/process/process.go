package process

import (
	"errors"
	id3 "github.com/casept/id3-go"
	"github.com/rosmo/go-mp3-podcast/config"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type AudioFile struct {
	Path        string
	Filename    string
	MimeType    string
	Timestamp   time.Time
	PublishDate time.Time
	Length      int64
	Title       string
}

func getPublishDate(cfg *config.Configuration, file *AudioFile) time.Time {
	if cfg.Items.Date.From == "title" {
		dateFormat := "(?P<day>\\d{1,2})\\.(?P<month>\\d{1,2})\\.(?P<year>\\d{4,4})"
		switch cfg.Items.Date.Format {
		case "yyyy.mm.dd":
			dateFormat = "(?P<year>\\d{4,4})\\.(?P<month>\\d{1,2})\\.(?P<day>\\d{1,2})"
		case "yyyy-mm-dd":
			dateFormat = "(?P<year>\\d{4,4})-(?P<month>\\d{1,2})-(?P<day>\\d{1,2})"
		case "dd-mm-yyyy":
			dateFormat = "(?P<day>\\d{1,2})-(?P<month>\\d{1,2})-(?P<year>\\d{4,4})"
		case "dd.mm.yyyy hh:ii":
			dateFormat = "(?P<day>\\d{1,2})\\.(?P<month>\\d{1,2})\\.(?P<year>\\d{4,4}) (?P<hour>\\d{1,2}):(?P<min>\\d{1,2})"
		case "yyyy.mm.dd hh:ii":
			dateFormat = "(?P<year>\\d{4,4})\\.(?P<month>\\d{1,2})\\.(?P<day>\\d{1,2}) (?P<hour>\\d{1,2}):(?P<min>\\d{1,2})"
		case "yyyy-mm-dd hh:ii":
			dateFormat = "(?P<year>\\d{4,4})-(?P<month>\\d{1,2})-(?P<day>\\d{1,2}) (?P<hour>\\d{1,2}):(?P<min>\\d{1,2})"
		case "dd-mm-yyyy hh:ii":
			dateFormat = "(?P<day>\\d{1,2})-(?P<month>\\d{1,2})-(?P<year>\\d{4,4}) (?P<hour>\\d{1,2}):(?P<min>\\d{1,2})"
		}

		var titlere = regexp.MustCompile(dateFormat)
		match := titlere.FindStringSubmatch(file.Title)
		result := make(map[string]string)
		for i, name := range titlere.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		if len(result) > 2 {
			_, yearPresent := result["year"]
			_, monthPresent := result["month"]
			_, dayPresent := result["day"]
			_, hourPresent := result["hour"]
			_, minutePresent := result["min"]

			var year, month, day, hour, minute int
			if yearPresent && monthPresent && dayPresent {
				year, _ = strconv.Atoi(result["year"])
				month, _ = strconv.Atoi(result["month"])
				day, _ = strconv.Atoi(result["day"])
			}
			if hourPresent && minutePresent {
				hour, _ = strconv.Atoi(result["hour"])
				minute, _ = strconv.Atoi(result["min"])
			}

			return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local)
		}
	}
	return file.Timestamp
}

func ProcessAudioFile(cfg *config.Configuration, file string) (*AudioFile, error) {
	mp3, err := id3.Open(file)
	defer mp3.Close()
	if err != nil {
		return nil, errors.New("failed to open file")
	}
	var result AudioFile
	result.Path = file
	result.Filename = filepath.Base(file)
	result.MimeType = mime.TypeByExtension(filepath.Ext(file))

	finfo, err := os.Stat(file)
	if err != nil {
		return nil, errors.New("failed to stat file")
	}

	result.Timestamp = finfo.ModTime()
	result.Length = finfo.Size()
	result.Title = mp3.Title()
	result.PublishDate = getPublishDate(cfg, &result)

	return &result, nil
}
