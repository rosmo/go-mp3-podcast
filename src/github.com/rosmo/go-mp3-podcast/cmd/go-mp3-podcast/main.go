package main

import (
	"flag"
	"fmt"
	"github.com/eduncan911/podcast"
	"github.com/rosmo/go-mp3-podcast/config"
	"github.com/rosmo/go-mp3-podcast/process"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Index struct {
	Config   config.Configuration
	Podcasts []*process.AudioFile
}

type ByPublishDate []*process.AudioFile

func (s ByPublishDate) Len() int {
	return len(s)
}
func (s ByPublishDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByPublishDate) Less(i, j int) bool {
	return s[i].PublishDate.Unix() > s[j].PublishDate.Unix()
}

func generateIndex(cfg *config.Configuration, files []*process.AudioFile) {
	t := template.New("index")

	funcMap := template.FuncMap{
		"formatDate": func(date time.Time) string {
			dateFormat := "01.02.2006 15:04"
			switch cfg.Index.DateFormat {
			case "dd.mm.yyyy":
				dateFormat = "01.02.2006"
			case "yyyy.mm.dd":
				dateFormat = "2006.01.02"
			case "yyyy-mm-dd":
				dateFormat = "2006-01-02"
			case "dd-mm-yyyy":
				dateFormat = "02-01-2006"
			case "dd.mm.yyyy hh:ii":
				dateFormat = "01.02.2006 15:04"
			case "yyyy.mm.dd hh:ii":
				dateFormat = "2006.01.02 15:04"
			case "yyyy-mm-dd hh:ii":
				dateFormat = "2006-01-02 15:04"
			case "dd-mm-yyyy hh:ii":
				dateFormat = "02-01-2006 15:04"
			}
			return date.Format(dateFormat)
		},
	}
	t = t.Funcs(funcMap)

	t.Parse(cfg.Index.Template)
	index := Index{
		Config:   *cfg,
		Podcasts: files,
	}

	err := t.Execute(os.Stdout, index)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to render template: %v\n", err.Error())
	}
}

func generateFeed(cfg *config.Configuration, files []*process.AudioFile) {
	var extensionToEnclosureType = map[string]podcast.EnclosureType{
		".mp3":  podcast.MP3,
		".m4a":  podcast.M4A,
		".m4v":  podcast.M4V,
		".mp4":  podcast.MP4,
		".mov":  podcast.MOV,
		".pdf":  podcast.PDF,
		".epub": podcast.EPUB}

	now := time.Now()
	feed := podcast.New(
		cfg.Channel.Title,
		cfg.Channel.Link,
		cfg.Channel.Description,
		nil,
		&now)
	if cfg.Image.Url != "" {
		feed.AddImage(cfg.Image.Url)
	}
	if cfg.Channel.Language != "" {
		feed.Language = cfg.Channel.Language
	}
	feed.Generator = ""

	for _, file := range files {
		link := cfg.Items.Link.BaseUrl + file.Filename
		guid := cfg.Items.Guid.BaseUrl + file.Filename
		enclosure := cfg.Items.Enclosure.BaseUrl + file.Filename

		item := podcast.Item{
			Title:       file.Title,
			Description: file.Title,
			Link:        link,
			GUID:        guid,
		}

		enclosureType := extensionToEnclosureType[strings.ToLower(filepath.Ext(file.Filename))]

		item.AddEnclosure(enclosure, enclosureType, file.Length)
		item.AddPubDate(&file.PublishDate)

		_, err := feed.AddItem(item)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add file %v: %v\n", file, err)
		}
	}
	fmt.Print(feed.String())
}

func processFiles(cfg *config.Configuration, directory string) ([]*process.AudioFile, error) {
	var supportedExtensions = [...]string{".aac", ".flac", ".m4a", ".mp3", ".ogg", ".ogm"}

	if directory[len(directory)-1] != '/' {
		directory = directory + "/"
	}
	fmt.Fprintf(os.Stderr, "Scanning directory: %v\n", directory)

	foundFiles, err := filepath.Glob(directory + "*")
	if err != nil {
		return nil, err
	}

	done := make(chan bool, 1)
	var wg sync.WaitGroup

	var podcasts []*process.AudioFile

	numberOfExtensions := len(supportedExtensions)
	go func() {
		results := make(chan *process.AudioFile)
		for _, file := range foundFiles {
			finfo, err := os.Stat(file)
			if err != nil {
				continue
			}
			fileSize := finfo.Size()
			ext := strings.ToLower(filepath.Ext(file))
			i := sort.SearchStrings(supportedExtensions[:], ext)

			if fileSize >= cfg.Items.Filter.MinimumSize &&
				i < numberOfExtensions && supportedExtensions[i] == ext {
				wg.Add(1)
				go func(file string) {
					defer wg.Done()

					audioFile, err := process.ProcessAudioFile(cfg, file)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error processing: %v: %v\n", file, err)
					} else {
						results <- audioFile
					}
				}(file)
			}
		}
		go func() {
			for res := range results {
				podcasts = append(podcasts, res)
			}
		}()
		wg.Wait()

		done <- true
	}()

	<-done
	sort.Sort(ByPublishDate(podcasts))
	fmt.Fprintf(os.Stderr, "Scan complete.\n")
	return podcasts, nil
}

func main() {
	configFile := flag.String("config", "", "configuration file (yaml)")
	directory := flag.String("dir", ".", "directory containing mp3 files")
	mode := flag.String("mode", "rss", "select mode (rss or index)")

	flag.Parse()
	if *configFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := config.Parse(*configFile)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", *configFile, err)
		os.Exit(2)
	}

	files, err := processFiles(cfg, *directory)
	if err != nil {
		fmt.Printf("Error processing files: %v\n", err)
		os.Exit(2)
	}

	if *mode == "index" {
		generateIndex(cfg, files)
	} else {
		generateFeed(cfg, files)
	}

	os.Exit(0)
}
