package amazons3

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containerops/dockyard/backend/driver"
	"github.com/containerops/wrench/setting"
)

func init() {
	driver.Register("amazons3", InitFunc)
}

func InitFunc() {
	driver.InjectReflect.Bind("amazons3save", amazons3save)
}

func amazons3save(file string) (url string, err error) {

	var key string

	for _, key = range strings.Split(file, "/") {

	}

	fin, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fin.Close()

	url = "http://" + setting.Endpoint + "/" + setting.Bucket + "/" + key

	newSession := session.New()
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(setting.AccessKeyID, setting.AccessKeysecret, ""),
		Endpoint:         aws.String(setting.Endpoint),
		Region:           aws.String(setting.Region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	// Create an S3 service object in the default region.
	s3Client := s3.New(newSession, s3Config)

	Iparams := &s3.PutObjectInput{
		Bucket: aws.String(setting.Bucket),
		Body:   fin,
		Key:    &key,
	}
	_, err = s3Client.PutObject(Iparams)
	if err != nil {
		return "", err
	}

	return url, nil

}
