package config

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"kikitoru/util"

	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"
)

// C 全局配置
var C StructConfig
var PathConfig string

// 修改配置结构体时需要同时修改 StructPointerConfig 和 StructConfig
type StructPointerConfig struct {
	Version        *string `json:"version"`
	LogLevel       *string `json:"logLevel"`
	Production     *bool   `json:"production"`
	DatabaseURL    *string `json:"databaseURL"`
	MaxParallelism *int    `json:"maxParallelism"`
	RootFolders    *[]struct {
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"rootFolders"`
	CoverFolderDir           *string        `json:"coverFolderDir"`
	VoiceWorkDefaultPath     *string        `json:"voiceWorkDefaultPath"`
	MD5Secret                *string        `json:"md5secret"`
	JWTSecret                *string        `json:"jwtsecret"`
	ExpiresIn                *time.Duration `json:"expiresIn"`
	ScannerMaxRecursionDepth *int           `json:"scannerMaxRecursionDepth"`
	PageSize                 *int           `json:"pageSize"`
	TagLanguage              *string        `json:"tagLanguage"`
	Retry                    *int           `json:"retry"`
	DLsiteTimeout            *time.Duration `json:"dlsiteTimeout"`
	RetryDelay               *time.Duration `json:"retryDelay"`
	HTTPProxyHost            *string        `json:"httpProxyHost"`
	HTTPProxyPort            *int           `json:"httpProxyPort"`
	ListenPort               *int           `json:"listenPort"`
	BlockRemoteConnection    *bool          `json:"blockRemoteConnection"`
	HTTPSEnabled             *bool          `json:"httpsEnabled"`
	HTTPSPrivateKey          *string        `json:"httpsPrivateKey"`
	HTTPSCert                *string        `json:"httpsCert"`
	HTTPSPort                *int           `json:"httpsPort"`
	SkipCleanup              *bool          `json:"skipCleanup"`
	EnableGzip               *bool          `json:"enableGzip"`
	RewindSeekTime           *time.Duration `json:"rewindSeekTime"`
	ForwardSeekTime          *time.Duration `json:"forwardSeekTime"`
	OffloadMedia             *bool          `json:"offloadMedia"`
	OffloadStreamPath        *string        `json:"offloadStreamPath"`
	OffloadDownloadPath      *string        `json:"offloadDownloadPath"`
}

type StructConfig struct {
	Version        string `json:"version"`
	LogLevel       string `json:"logLevel"`
	Production     bool   `json:"production"`
	DatabaseURL    string `json:"databaseURL"`
	MaxParallelism int    `json:"maxParallelism"`
	RootFolders    []struct {
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"rootFolders"`
	CoverFolderDir           string        `json:"coverFolderDir"`
	VoiceWorkDefaultPath     string        `json:"voiceWorkDefaultPath"`
	MD5Secret                string        `json:"md5secret"`
	JWTSecret                string        `json:"jwtsecret"`
	ExpiresIn                time.Duration `json:"expiresIn"`
	ScannerMaxRecursionDepth int           `json:"scannerMaxRecursionDepth"`
	PageSize                 int           `json:"pageSize"`
	TagLanguage              string        `json:"tagLanguage"`
	Retry                    int           `json:"retry"`
	DLsiteTimeout            time.Duration `json:"dlsiteTimeout"`
	RetryDelay               time.Duration `json:"retryDelay"`
	HTTPProxyHost            string        `json:"httpProxyHost"`
	HTTPProxyPort            int           `json:"httpProxyPort"`
	ListenPort               int           `json:"listenPort"`
	BlockRemoteConnection    bool          `json:"blockRemoteConnection"`
	HTTPSEnabled             bool          `json:"httpsEnabled"`
	HTTPSPrivateKey          string        `json:"httpsPrivateKey"`
	HTTPSCert                string        `json:"httpsCert"`
	HTTPSPort                int           `json:"httpsPort"`
	SkipCleanup              bool          `json:"skipCleanup"`
	EnableGzip               bool          `json:"enableGzip"`
	RewindSeekTime           time.Duration `json:"rewindSeekTime"`
	ForwardSeekTime          time.Duration `json:"forwardSeekTime"`
	OffloadMedia             bool          `json:"offloadMedia"`
	OffloadStreamPath        string        `json:"offloadStreamPath"`
	OffloadDownloadPath      string        `json:"offloadDownloadPath"`
}

var workingDir string

func InitConfig() {
	ex, err := os.Executable()
	if err != nil {
		log.Error(err)
	}
	workingDir = filepath.Dir(ex)
	PathConfig = path.Join(workingDir, "config.json")

	// 如果配置文件存在
	isExist, err := fileExists(PathConfig)
	if err != nil {
		log.Info(err)
	}

	if isExist {
		// 打开配置文件
		jsonFile, err := os.Open(PathConfig)
		if err != nil {
			log.Info(err)
		}
		defer jsonFile.Close()

		// 从打开的配置文件中读取结构体
		var pointerConfig StructPointerConfig
		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &pointerConfig)
		if err != nil {
			log.Error(err)
		}
		// 将配置文件中的字段合并进默认配置
		C = updateConfig(pointerConfig)

		// 当从配置文件中读取的版本号小于定义的版本号时更新
		if C.Version < VERSION {
			C.Version = VERSION
			err = WriteConfig(PathConfig, C)
			if err != nil {
				log.Error(err)
			}
		}

	} else {
		// 写默认配置
		C = getDefaultNewConfig()

		err = WriteConfig(PathConfig, C)
		if err != nil {
			log.Error(err)
		}
	}
}

func getDefaultNewConfig() StructConfig {
	md5S, err := util.GenerateRandomSecret()
	if err != nil {
		log.Fatal(err)
	}
	jwtS, err := util.GenerateRandomSecret()
	if err != nil {
		log.Fatal(err)
	}

	var defaultConfig = StructConfig{
		Version:        VERSION,
		LogLevel:       "info",
		Production:     false,
		DatabaseURL:    "postgres://username:password@localhost/kikitoru?sslmode=disable",
		MaxParallelism: 16,
		RootFolders: []struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}{},
		CoverFolderDir:           path.Join(workingDir, "covers"),
		VoiceWorkDefaultPath:     path.Join(workingDir, "VoiceWork"),
		MD5Secret:                md5S,
		JWTSecret:                jwtS,
		ExpiresIn:                2592000 * time.Second,
		ScannerMaxRecursionDepth: 2,
		PageSize:                 12,
		TagLanguage:              "zh-cn",
		Retry:                    5,
		DLsiteTimeout:            10000 * time.Millisecond,
		RetryDelay:               2000 * time.Millisecond,
		HTTPProxyHost:            "",
		HTTPProxyPort:            0,
		ListenPort:               8080,
		BlockRemoteConnection:    false,
		HTTPSEnabled:             false,
		HTTPSPrivateKey:          "kikitoru.key",
		HTTPSCert:                "kikitoru.crt",
		HTTPSPort:                8443,
		SkipCleanup:              false,
		EnableGzip:               true,
		RewindSeekTime:           5 * time.Second,
		ForwardSeekTime:          30 * time.Second,
		OffloadMedia:             false,
		OffloadStreamPath:        "/media/stream/",
		OffloadDownloadPath:      "/media/download/",
	}
	return defaultConfig
}

func WriteConfig(configPath string, c StructConfig) error {
	file, _ := json.MarshalIndent(c, "", "    ")
	err := os.WriteFile(configPath, file, 0644)
	return err
}

func updateConfig(p StructPointerConfig) StructConfig {
	cDef := getDefaultNewConfig()
	mergeConfig(p, &cDef)
	return cDef
}

// mergeConfig(t1,&t2) t2传地址，将t1中非空的值写到t2中
func mergeConfig(t1 interface{}, t2 interface{}) {
	// 获取 t1 和 t2 的反射值
	v1 := reflect.ValueOf(t1)
	// t2 传地址 解指针
	v2 := reflect.ValueOf(t2).Elem()

	// 获取 test1 的类型
	t1Type := reflect.TypeOf(t1)

	// 遍历 test1 的字段
	for i := 0; i < t1Type.NumField(); i++ {
		// 获取 test1 字段的 reflect.StructField
		fieldType := t1Type.Field(i)

		// 获取 test1 字段的值
		fieldValue := v1.Field(i)

		// 检查字段值是否为空
		if fieldValue.IsNil() {
			continue
		}

		// 获取 test2 字段的反射值
		field := v2.FieldByName(fieldType.Name)

		// 检查字段是否存在
		if field.IsValid() && field.CanSet() {
			// 将 test1 的字段值赋给 test2
			field.Set(fieldValue.Elem())
		}
	}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ConvertConfigMatchFrontend(config StructConfig) StructConfig {
	config.DLsiteTimeout = config.DLsiteTimeout / time.Millisecond
	config.RetryDelay = config.RetryDelay / time.Millisecond
	config.RewindSeekTime = config.RewindSeekTime / time.Second
	config.ForwardSeekTime = config.ForwardSeekTime / time.Second
	config.ExpiresIn = config.ExpiresIn / time.Second
	return config
}

func ConvertFrontendMatchConfig(config StructConfig) StructConfig {
	config.DLsiteTimeout = config.DLsiteTimeout * time.Millisecond
	config.RetryDelay = config.RetryDelay * time.Millisecond
	config.RewindSeekTime = config.RewindSeekTime * time.Second
	config.ForwardSeekTime = config.ForwardSeekTime * time.Second
	config.ExpiresIn = config.ExpiresIn * time.Second
	return config
}
