package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/allyraza/souqr"
	"github.com/garyburd/redigo/redis"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Config @todo
type Config struct {
	RedisURL     string
	Dir          string
	Poll         bool
	Sync         bool
	Delay        int
	Debug        bool
	Offset       string
	AccessKey    string
	AccessSecret string
	Bucket       string
	Region       string
	Conn         redis.Conn
}

func main() {
	config := &Config{}

	flag.StringVar(&config.Dir, "dir", "", "name of the file to write data to")
	flag.StringVar(&config.RedisURL, "redis-url", ":6379", "redis server url")
	flag.IntVar(&config.Delay, "delay", 5, "timeout delay")
	flag.BoolVar(&config.Poll, "poll", false, "poll redis for products")
	flag.BoolVar(&config.Sync, "sync", false, "sync local files with aws s3 bucket")
	flag.BoolVar(&config.Debug, "debug", false, "enable debugging")
	flag.StringVar(&config.Offset, "offset", "0", "redis scan offset")
	flag.StringVar(&config.AccessKey, "access-key", "", "S3 access key")
	flag.StringVar(&config.AccessSecret, "access-secret", "", "S3 access secret")
	flag.StringVar(&config.Bucket, "bucket", "souq-feed", "S3 bucket")
	flag.StringVar(&config.Region, "region", "eu-west-1", "S3 bucket")
	flag.Parse()

	if config.Sync {
		syncFileSystem(config)
	} else {
		syncRedis(config)
	}
}

func syncFileSystem(config *Config) {
	files, err := ioutil.ReadDir(config.Dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		Upload(f.Name(), config)
	}
}

func syncRedis(config *Config) {
	conn, err := redis.Dial("tcp", config.RedisURL)
	if err != nil {
		log.Fatalf("Could not connect to redis server: %v", err)
	}
	config.Conn = conn

	current, err := redis.String(conn.Do("GET", "current"))
	if err != nil {
		log.Fatal(err)
	}

	for {
		offset, keys := Fetch(config)
		if offset == "0" && len(keys) <= 3 {
			break
		}
		config.Offset = offset

		for _, k := range keys {
			if k == "categories" || k == "current" || k == current {
				continue
			}

			Dump(k, config)
			Upload(k+".xml", config)
		}
	}
}

// Drop delete the completed bucket
func Drop(key string, config *Config) {
	_, err := config.Conn.Do("DEL", key)
	if err != nil {
		log.Println(err)
	}
}

// Upload uploads the file to s3
func Upload(filename string, config *Config) {
	path := filepath.Join(config.Dir, filename)

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	creds := credentials.NewEnvCredentials()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: creds,
	})

	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.Bucket),
		Key:    aws.String(filename),
		Body:   file,
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("DUMP: %v\n", filename)
}

// Fetch retrieves data from redis
func Fetch(config *Config) (string, []string) {
	var Offset string
	var Keys []string

	values, err := redis.Values(config.Conn.Do("SCAN", config.Offset))
	if err != nil {
		log.Fatal(err)
	}

	_, err = redis.Scan(values, &Offset, &Keys)
	if err != nil {
		log.Fatal(err)
	}

	return Offset, Keys
}

// Dump writes a given bucket to a xml file
func Dump(bucket string, config *Config) {
	filename := filepath.Join(config.Dir, bucket+".xml")

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(file, xml.Header+"<products>")

	for {
		res, err := redis.Bytes(config.Conn.Do("LPOP", bucket))
		if err != nil && res != nil && config.Debug {
			log.Printf("%v", err)
		}

		if res == nil && config.Poll {
			time.Sleep(time.Duration(config.Delay) * time.Second)
			continue
		}

		if res == nil && !config.Poll {
			break
		}

		Write(file, res)
	}

	fmt.Fprintf(file, "</products>")
}

// Write writes bytes to given file
func Write(file *os.File, p []byte) error {
	s := &souqr.Souq{}

	err := json.Unmarshal(p, &s)
	if err != nil {
		fmt.Println(err)
	}

	var InStock int
	if s.PageData.Product.Quantity > 0 {
		InStock = 1
	} else {
		InStock = 0
	}

	categoryPath := fmt.Sprintf("%s > %s",
		s.PageData.Product.Category,
		s.PageData.Product.ParentCategory)

	prd := &souqr.Product{
		ID:           s.PageData.ItemID,
		Name:         s.Name,
		CategoryPath: categoryPath,
		Brand:        s.PageData.Product.Brand,
		Deeplink:     s.URL,
		Description:  s.Description,
		Image:        s.Image,
		Color:        s.Color,
		Gender:       s.PageData.Product.Attributes.TargetedGroup,
		Size:         s.PageData.Product.Attributes.Size,
		EAN:          s.PageData.EAN,
		Quantity:     s.PageData.Product.Quantity,
		InStock:      InStock,
		Price:        s.PageData.Product.Price,
		PriceFrom:    s.PageData.Product.Price,
	}

	prd.Write(file)

	return nil
}
