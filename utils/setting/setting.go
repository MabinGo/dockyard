package setting

import (
	"fmt"

	"github.com/astaxie/beego/config"
)

const (
	APIVERSION_V1 = iota
	APIVERSION_V2
	APIVERSION_ACI
)

// version info
var (
	AppName string
	Usage   string
	Version string
	Author  string
	Email   string
)

// running mode
var (
	RunMode       string
	ListenMode    string
	HttpsCertFile string
	HttpsKeyFile  string
)

// log
var (
	LogPath string
)

// DB
var (
	DBDriver string
	DBUser   string
	DBPasswd string
	DBName   string
	DBURI    string
	DBDB     int64
)

// Dockyard
var (
	Backend             string
	Cachable            bool
	ImagePath           string
	Domains             string
	RegistryVersion     string
	DistributionVersion string
	Standalone          string
	OssSwitch           string
)

// object storage driver config parameters
var (
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	AccessKeysecret string

	//upyun unique
	Secret string

	//qcloud unique
	QcloudAccessID string

	//googlecloud unique
	Projectid          string
	Scope              string
	PrivateKeyFilePath string
	PrivateKeyFile     string
	Clientemail        string

	//rados unique
	Chunksize string
	Poolname  string
	Username  string

	// OSS backend driver parameters
	APIPort      int
	APIHttpsPort int
	PartSizeMB   int
)

// Clair service config parameters
var (
	//Path of the database. Default:  '/db'
	ClairDBPath string
	//Remove all the data in DB after stop the clair service. Default: false
	ClairKeepDB bool
	//Log level of the clair lib. Default: 'info'
	//All values: ['critical, error, warning, notice, info, debug, trace']
	ClairLogLevel string
	//Update CVE date in every '%dh%dm%ds'. Default: '1h0m0s'
	ClairUpdateDuration string
	//Return CVEs with minimal priority to Dockyard. Default: 'Low'
	//All values: ['Unknown, Negligible, Low, Medium, High, Critical, Defcon1']
	ClairVulnPriority string
)

// Auth server configuration
var (
	Issuer     string
	PrivateKey string
	Expiration int64
	Authn      string
)

// ldap authn
var (
	Addr                  string
	TLS                   bool
	InsecureTLSSkipVerify bool
	Domain                string
)

// login auth of Dockyard
var (
	Auth           string
	Authmode       string
	Realm          string
	Service        string
	Issue          string
	Rootcertbundle string
)

func SetConfig(path string) error {
	conf, err := config.NewConfig("ini", path)
	if err != nil {
		return fmt.Errorf("Read %s error: %v", path, err.Error())
	}

	err = setVersionConfig(conf)
	err = setRuntimeConfig(conf)

	return err
}

func setVersionConfig(conf config.Configer) error {
	var err error = nil

	if appname := conf.String("appname"); appname != "" {
		AppName = appname
	} else if appname == "" {
		err = fmt.Errorf("AppName config value is null")
	}

	if usage := conf.String("usage"); usage != "" {
		Usage = usage
	} else if usage == "" {
		err = fmt.Errorf("Usage config value is null")
	}

	if version := conf.String("version"); version != "" {
		Version = version
	} else if version == "" {
		err = fmt.Errorf("Version config value is null")
	}

	if author := conf.String("author"); author != "" {
		Author = author
	} else if author == "" {
		err = fmt.Errorf("Author config value is null")
	}

	if email := conf.String("email"); email != "" {
		Email = email
	} else if email == "" {
		err = fmt.Errorf("Email config value is null")
	}

	return err
}

func setRuntimeConfig(conf config.Configer) error {
	var err error = nil

	err = setRunModeConfig(conf)
	err = setLogConfig(conf)
	err = setDBConfig(conf)
	err = setDockyardConfig(conf)
	err = setBackendConfig(conf)
	err = setClairConfig(conf)
	err = setAuthServerConfig(conf)
	err = setAuthConfig(conf)
	err = setSynchronConfig(conf)
	err = setDRConfig(conf)

	return err
}

func setRunModeConfig(conf config.Configer) error {
	var err error = nil

	if runmode := conf.String("runmode"); runmode != "" {
		RunMode = runmode
	} else if runmode == "" {
		err = fmt.Errorf("RunMode config value is null")
	}

	if listenmode := conf.String("listenmode"); listenmode != "" {
		ListenMode = listenmode
	} else if listenmode == "" {
		err = fmt.Errorf("ListenMode config value is null")
	}

	if httpscertfile := conf.String("httpscertfile"); httpscertfile != "" {
		HttpsCertFile = httpscertfile
	} else if httpscertfile == "" {
		err = fmt.Errorf("HttpsCertFile config value is null")
	}

	if httpskeyfile := conf.String("httpskeyfile"); httpskeyfile != "" {
		HttpsKeyFile = httpskeyfile
	} else if httpskeyfile == "" {
		err = fmt.Errorf("HttpsKeyFile config value is null")
	}

	return err
}

func setLogConfig(conf config.Configer) error {
	if logpath := conf.String("log::filepath"); logpath != "" {
		LogPath = logpath
	} else if logpath == "" {
		return fmt.Errorf("LogPath config value is null")
	}

	return nil
}

func setDBConfig(conf config.Configer) error {
	if dbdriver := conf.String("db::driver"); dbdriver != "" {
		DBDriver = dbdriver
	} else {
		return fmt.Errorf("DB driver config value is null")
	}
	if dburi := conf.String("db::uri"); dburi != "" {
		DBURI = dburi
	}
	if dbuser := conf.String("db::user"); dbuser != "" {
		DBUser = dbuser
	}
	if dbpass := conf.String("db::passwd"); dbpass != "" {
		DBPasswd = dbpass
	}
	if dbname := conf.String("db::name"); dbname != "" {
		DBName = dbname
	}
	dbpartition, _ := conf.Int64("db::db")
	DBDB = dbpartition

	return nil
}

func setDockyardConfig(conf config.Configer) error {
	var err error = nil

	ImagePath = "data" //store image layer in local fs by default
	if imagepath := conf.String("dockyard::path"); imagepath != "" {
		ImagePath = imagepath
	}

	if domains := conf.String("dockyard::domains"); domains != "" {
		Domains = domains
	} else if domains == "" {
		err = fmt.Errorf("Domains value is null")
	}

	if registryVersion := conf.String("dockyard::registry"); registryVersion != "" {
		RegistryVersion = registryVersion
	} else if registryVersion == "" {
		err = fmt.Errorf("Registry version value is null")
	}

	if distributionVersion := conf.String("dockyard::distribution"); distributionVersion != "" {
		DistributionVersion = distributionVersion
	} else if distributionVersion == "" {
		err = fmt.Errorf("Distribution version value is null")
	}

	if standalone := conf.String("dockyard::standalone"); standalone != "" {
		Standalone = standalone
	} else if standalone == "" {
		err = fmt.Errorf("Standalone version value is null")
	}
	if ossswitch := conf.String("dockyard::ossswitch"); ossswitch != "" {
		OssSwitch = ossswitch
	} else if ossswitch == "" {
		OssSwitch = "disable"
	}

	//Config of object storage service
	if Backend = conf.String("dockyard::backend"); Backend != "" {
		if cachable, err0 := conf.Bool("dockyard::cachable"); err0 != nil {
			Cachable = true
		} else {
			Cachable = cachable
		}
	}

	return err
}

func setBackendConfig(conf config.Configer) error {
	var err error = nil

	// TODO: It should be considered to refine the universal config parameters
	switch Backend {
	case "":
	case "qiniu", "aliyun", "s3":
		if endpoint := conf.String(Backend + "::" + "endpoint"); endpoint != "" {
			Endpoint = endpoint
		} else {
			err = fmt.Errorf("Endpoint value is null")
		}

		if bucket := conf.String(Backend + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if accessKeyID := conf.String(Backend + "::" + "accessKeyID"); accessKeyID != "" {
			AccessKeyID = accessKeyID
		} else {
			err = fmt.Errorf("AccessKeyID value is null")
		}

		if accessKeysecret := conf.String(Backend + "::" + "accessKeysecret"); accessKeysecret != "" {
			AccessKeysecret = accessKeysecret
		} else {
			err = fmt.Errorf("AccessKeysecret value is null")
		}
	case "upyun":
		if endpoint := conf.String(Backend + "::" + "endpoint"); endpoint != "" {
			Endpoint = endpoint
		} else {
			err = fmt.Errorf("Endpoint value is null")
		}

		if bucket := conf.String(Backend + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if secret := conf.String(Backend + "::" + "secret"); secret != "" {
			Secret = secret
		} else {
			err = fmt.Errorf("Secret value is null")
		}
	case "qcloud":
		if endpoint := conf.String(Backend + "::" + "endpoint"); endpoint != "" {
			Endpoint = endpoint
		} else {
			err = fmt.Errorf("Endpoint value is null")
		}

		if accessID := conf.String(Backend + "::" + "accessID"); accessID != "" {
			QcloudAccessID = accessID
		} else {
			err = fmt.Errorf("accessID value is null")
		}

		if bucket := conf.String(Backend + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if accessKeyID := conf.String(Backend + "::" + "accessKeyID"); accessKeyID != "" {
			AccessKeyID = accessKeyID
		} else {
			err = fmt.Errorf("AccessKeyID value is null")
		}

		if accessKeysecret := conf.String(Backend + "::" + "accessKeysecret"); accessKeysecret != "" {
			AccessKeysecret = accessKeysecret
		} else {
			err = fmt.Errorf("AccessKeysecret value is null")
		}
	case "oss":
		APIPort, err = conf.Int(Backend + "::" + "apiport")
		APIHttpsPort, err = conf.Int(Backend + "::" + "apihttpsport")
		PartSizeMB, err = conf.Int(Backend + "::" + "partsizemb")
	case "gcs":
		if projectid := conf.String(Backend + "::" + "projectid"); projectid != "" {
			Projectid = projectid
		} else {
			err = fmt.Errorf("Projectid value is null")
		}

		if scope := conf.String(Backend + "::" + "scope"); scope != "" {
			Scope = scope
		} else {
			err = fmt.Errorf("Scope value is null")
		}

		if bucket := conf.String(Backend + "::" + "bucket"); bucket != "" {
			Bucket = bucket
		} else {
			err = fmt.Errorf("Bucket value is null")
		}

		if keyfilepath := conf.String(Backend + "::" + "keyfilepath"); keyfilepath != "" {
			PrivateKeyFilePath = keyfilepath
		} else {
			err = fmt.Errorf("Privatekey value is null")
		}

		if privatekey := conf.String(Backend + "::" + "privatekey"); privatekey != "" {
			PrivateKeyFile = privatekey
		} else {
			err = fmt.Errorf("Privatekey value is null")
		}

		if clientemail := conf.String(Backend + "::" + "clientemail"); clientemail != "" {
			Clientemail = clientemail
		} else {
			err = fmt.Errorf("Clientemail value is null")
		}
	case "rados":
		if chunksize := conf.String(Backend + "::" + "chunksize"); chunksize != "" {
			Chunksize = chunksize
		} else {
			err = fmt.Errorf("Chunksize value is null")
		}

		if poolname := conf.String(Backend + "::" + "poolname"); poolname != "" {
			Poolname = poolname
		} else {
			err = fmt.Errorf("Poolname value is null")
		}

		if username := conf.String(Backend + "::" + "username"); username != "" {
			Username = username
		} else {
			err = fmt.Errorf("Username value is null")
		}
	default:
		err = fmt.Errorf("Not support %v", Backend)
	}

	return err
}

func setClairConfig(conf config.Configer) error {
	//Config of image security scanning
	ClairDBPath = conf.String("clair::path")
	ClairLogLevel = conf.String("clair::logLevel")
	ClairKeepDB, _ = conf.Bool("clair::keepDB")
	ClairUpdateDuration = conf.String("clair::updateDuration")
	ClairVulnPriority = conf.String("clair::vulnPriority")

	return nil
}

func setAuthServerConfig(conf config.Configer) error {
	var err error = nil

	// Auth server configuration
	if issuer := conf.String("auth_server::issuer"); issuer != "" {
		Issuer = issuer
	} else {
		err = fmt.Errorf("auth_server issuer config value is null")
	}
	if privateKey := conf.String("auth_server::privateKey"); privateKey != "" {
		PrivateKey = privateKey
	} else {
		err = fmt.Errorf("auth_server privateKey config value is null")
	}
	if expiration, err := conf.Int64("auth_server::expiration"); err != nil {
		err = fmt.Errorf("auth_server expiration config value error")
	} else {
		Expiration = expiration
	}
	if authn := conf.String("auth_server::authn"); authn != "" {
		Authn = authn
	} else {
		err = fmt.Errorf("auth_server anthn config value error")
	}
	if Authn == "authn_ldap" {
		if addr := conf.String("authn_ldap::addr"); addr != "" {
			Addr = addr
		} else {
			err = fmt.Errorf("authn_ldap addr config value error")
		}
		if tls, err := conf.Bool("authn_ldap::tls"); err != nil {
			err = fmt.Errorf("authn_ldap tls config value error")
		} else {
			TLS = tls
		}
		if insecureTLSSkipVerify, err := conf.Bool("authn_ldap::insecuretlsskipverify"); err != nil {
			err = fmt.Errorf("authn_ldap insecuretlsskipverify config value error")
		} else {
			InsecureTLSSkipVerify = insecureTLSSkipVerify
		}
		if domain := conf.String("authn_ldap::domain"); domain != "" {
			Domain = domain
		} else {
			err = fmt.Errorf("authn_ldap domain config value error")
		}
	}

	return err
}

func setAuthConfig(conf config.Configer) error {
	var err error = nil

	//login auth mode
	if authmode := conf.String("dockyard::authmode"); authmode != "" {
		Authmode = authmode
		Auth = "auth"
	}
	switch Authmode {
	case "":
	case "token":
		if realm := conf.String(Authmode + "::" + "realm"); realm != "" {
			Realm = realm
		} else {
			err = fmt.Errorf("Realm value is null")
		}

		if service := conf.String(Authmode + "::" + "service"); service != "" {
			Service = service
		} else {
			err = fmt.Errorf("Service value is null")
		}

		if issuer := conf.String(Authmode + "::" + "issuer"); issuer != "" {
			Issue = issuer
		} else {
			err = fmt.Errorf("Issuer value is null")
		}

		if rootcertbundle := conf.String(Authmode + "::" + "rootcertbundle"); rootcertbundle != "" {
			Rootcertbundle = rootcertbundle
		} else {
			err = fmt.Errorf("Rootcertbundle value is null")
		}
	default:
		err = fmt.Errorf("Not support auth mode %v", Authmode)
	}

	return err
}

//TODO
var (
	SynMode     string
	SynUser     string
	SynPasswd   string
	SynInterval int64
	DRList      string
)

func setSynchronConfig(conf config.Configer) error {
	var err error = nil

	SynMode = conf.String("dockyard::synmode")
	switch SynMode {
	case "":
	case "poll":
		if SynInterval, err = conf.Int64(SynMode + "::" + "syninterval"); err != nil || SynInterval <= 0 {
			err = fmt.Errorf("polling syninterval config value error")
		}

		if SynUser = conf.String(SynMode + "::" + "synuser"); SynUser == "" {
			err = fmt.Errorf("synuser value is null")
		}

		if SynPasswd = conf.String(SynMode + "::" + "synpasswd"); SynPasswd == "" {
			err = fmt.Errorf("synpasswd value is null")
		}
	//case "priority":
	default:
		err = fmt.Errorf("Not support synch mode %v", SynMode)
	}

	return err
}

func setDRConfig(conf config.Configer) error {
	var err error = nil

	DRList = conf.String("dockyard::drlist")

	return err
}
