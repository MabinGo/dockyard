/*
Copyright 2015 The ContainerOps Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package setting

import (
	"fmt"

	"github.com/astaxie/beego/config"
)

const (
	DOCKERAPIV1 = iota
	DOCKERAPIV2
	APPAPIV1
)

var (
	// Global Config
	AppName string
	Usage   string
	Version string
	Author  string
	Email   string

	// Runtime Config
	RunMode        string
	ListenMode     string
	HttpsCertFile  string
	HttpsKeyFile   string
	LogPath        string
	LogLevel       string
	DatabaseDriver string
	DatabaseURI    string
	DatabaseName   string
	DatabaseUser   string
	DatabasePasswd string
	DockyardPath   string

	ExternalAddressEnabled bool
	ExternalScheme         string
	ExternalHost           string
	ExternalPort           int64

	// Key Manager config
	KeyManagerMode string
	KeyManagerURI  string

	// Auth config
	AuthEnable   string
	Authn        string
	AuthnAddress string
	AuthnPort    string
	// IAM config
	IAMCertPath    string
	IAMCaPath      string
	IAMServiceAkey string
	IAMServiceSkey string

	// Web config
	AllowedOrigins    string
	AllowedMethods    string
	AllowedHeaders    string
	MaxUploadFileSize int64

	//check resources interval
	Inspectinterval int
	//recycle resources interval
	RecycleInterval int
)

func SetConfig(path string) error {
	conf, err := config.NewConfig("ini", path)
	if err != nil {

		return fmt.Errorf("Read %s error: %v", path, err.Error())
	}

	//config globals
	if appname := conf.String("appname"); appname != "" {
		AppName = appname
	} else if appname == "" {

		return fmt.Errorf("AppName config value is null")
	}

	if usage := conf.String("usage"); usage != "" {
		Usage = usage
	} else if usage == "" {

		return fmt.Errorf("Usage config value is null")
	}

	if version := conf.String("version"); version != "" {
		Version = version
	} else if version == "" {

		return fmt.Errorf("Version config value is null")
	}

	if author := conf.String("author"); author != "" {
		Author = author
	} else if author == "" {

		return fmt.Errorf("Author config value is null")
	}

	if email := conf.String("email"); email != "" {
		Email = email
	} else if email == "" {

		return fmt.Errorf("Email config value is null")
	}

	//config runtime
	if runmode := conf.String("runmode"); runmode != "" {
		RunMode = runmode
	} else if runmode == "" {

		return fmt.Errorf("RunMode config value is null")
	}

	if listenmode := conf.String("listenmode"); listenmode != "" {
		ListenMode = listenmode
	} else if listenmode == "" {

		return fmt.Errorf("ListenMode config value is null")
	}

	if httpscertfile := conf.String("httpscertfile"); httpscertfile != "" {
		HttpsCertFile = httpscertfile
	} else if httpscertfile == "" {
		return fmt.Errorf("HttpsCertFile config value is null")
	}

	if httpskeyfile := conf.String("httpskeyfile"); httpskeyfile != "" {
		HttpsKeyFile = httpskeyfile
	} else if httpskeyfile == "" {
		return fmt.Errorf("HttpsKeyFile config value is null")
	}

	if logpath := conf.String("log::filepath"); logpath != "" {
		LogPath = logpath
	} else if logpath == "" {

		return fmt.Errorf("LogPath config value is null")
	}

	if loglevel := conf.String("log::level"); loglevel != "" {
		LogLevel = loglevel
	} else if loglevel == "" {

		return fmt.Errorf("LogLevel config value is null")
	}

	if databasedriver := conf.String("database::driver"); databasedriver != "" {
		DatabaseDriver = databasedriver
	} else if databasedriver == "" {

		return fmt.Errorf("Database Driver config value is null")
	}

	if databaseuri := conf.String("database::uri"); databaseuri != "" {
		DatabaseURI = databaseuri
	} else if databaseuri == "" {

		return fmt.Errorf("Database URI config vaule is null")
	}

	if databasename := conf.String("database::name"); databasename != "" {
		DatabaseName = databasename
	} else if databasename == "" {

		return fmt.Errorf("Database Name config vaule is null")
	}

	if databaseuser := conf.String("database::user"); databaseuser != "" {
		DatabaseUser = databaseuser
	} else if databaseuser == "" {

		return fmt.Errorf("Database User config vaule is null")
	}

	if databasepasswd := conf.String("database::passwd"); databasepasswd != "" {
		DatabasePasswd = databasepasswd
	}

	if dockyardpath := conf.String("path"); dockyardpath != "" {
		DockyardPath = dockyardpath
	} else if dockyardpath == "" {

		return fmt.Errorf("Dockyard Path config vaule is null")
	}

	ExternalAddressEnabled, _ = conf.Bool("web::external-enabled")
	if ExternalAddressEnabled == true {
		if externalScheme := conf.String("web::external-scheme"); externalScheme != "" {
			ExternalScheme = externalScheme
		} else {
			return fmt.Errorf("Dockyard ExternalScheme config vaule is null")
		}

		if externalHost := conf.String("web::external-host"); externalHost != "" {
			ExternalHost = externalHost
		} else {
			return fmt.Errorf("Dockyard ExternalHost config vaule is null")
		}

		if externalPort, err := conf.Int64("web::external-port"); err == nil {
			ExternalPort = externalPort
		} else {
			return fmt.Errorf("Dockyard ExternalPort config vaule is null")
		}
	}

	if keymanagermode := conf.String("keymanager::kmmode"); keymanagermode != "" {
		KeyManagerMode = keymanagermode
	} else if keymanagermode == "" {
		// default to share mode
		KeyManagerMode = "share"
	}

	// if keymanageruri is empty, don't signature vm/app/image
	if keymanageruri := conf.String("keymanager::kmuri"); keymanageruri != "" {
		KeyManagerURI = keymanageruri
	}

	if authenable := conf.String("auth::enable"); authenable != "" {
		AuthEnable = authenable
	}
	if authn := conf.String("auth::authn"); authn != "" {
		Authn = authn
	}
	if authnAddress := conf.String("auth::authn_address"); authnAddress != "" {
		AuthnAddress = authnAddress
	}
	if authnPort := conf.String("auth::authn_port"); authnPort != "" {
		AuthnPort = authnPort
	}
	if iamCaPath := conf.String("iam::capath"); iamCaPath != "" {
		IAMCaPath = iamCaPath
	}
	if iamCertPath := conf.String("iam::certpath"); iamCertPath != "" {
		IAMCertPath = iamCertPath
	}
	if iamServiceAkey := conf.String("iam::serviceakey"); iamServiceAkey != "" {
		IAMServiceAkey = iamServiceAkey
	}
	if iamServiceSkey := conf.String("iam::serviceskey"); iamServiceSkey != "" {
		IAMServiceSkey = iamServiceSkey
	}

	if allowedorigins := conf.String("web::allowed-origins"); allowedorigins != "" {
		AllowedOrigins = allowedorigins
	} else {
		AllowedOrigins = "*"
	}

	if allowedmethods := conf.String("web::allowed-methods"); allowedmethods != "" {
		AllowedMethods = allowedmethods
	}

	if allowedheaders := conf.String("web::allowed-headers"); allowedheaders != "" {
		AllowedHeaders = allowedheaders
	}

	if maxUploadFileSize, err := conf.Int64("web::max-upload-filesize"); maxUploadFileSize != 0 && err == nil {
		MaxUploadFileSize = maxUploadFileSize
	} else { // 2Gib by default
		MaxUploadFileSize = 2 << 30
	}

	if recycleinterval, err := conf.Int("recycleinterval"); err != nil {

		return fmt.Errorf("Recycle interval config value is null")
	} else {
		RecycleInterval = recycleinterval
	}

	if inspectinterval, err := conf.Int("inspectinterval"); err != nil {

		return fmt.Errorf("Inspection interval config value is null")
	} else {
		Inspectinterval = inspectinterval
	}

	return nil
}
