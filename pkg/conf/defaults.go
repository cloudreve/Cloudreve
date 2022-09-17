package conf

import "github.com/cloudreve/Cloudreve/v3/pkg/util"

// RedisConfig Redis服务器配置
var RedisConfig = &redis{
	Network:  util.EnvStr("REDIS_NETWORK", "tcp"),
	Server:   util.EnvStr("REDIS_SERVER", ""),
	Password: util.EnvStr("REDIS_PASSWORD", ""),
	DB:       util.EnvStr("REDIS_DB", "0"),
}

// DatabaseConfig 数据库配置
var DatabaseConfig = &database{
	Type:        util.EnvStr("DATABASE_TYPE", "UNSET"),
	User:        util.EnvStr("DATABASE_USER", "root"),
	Password:    util.EnvStr("DATABASE_PASSWORD", ""),
	Host:        util.EnvStr("DATABASE_HOST", "localhost"),
	Name:        util.EnvStr("DATABASE_NAME", "cloudreve"),
	TablePrefix: util.EnvStr("DATABASE_TABLE_PREFIX", ""),
	DBFile:      util.EnvStr("DATABASE_DBFILE", "data.db"),
	Port:        util.EnvInt("DATABASE_PORT", 3306),
	Charset:     util.EnvStr("DATABASE_CHARSET", "utf8"),
}

// SystemConfig 系统公用配置
var SystemConfig = &system{
	Debug:  util.EnvStr("DEBUG", "false") == "true",
	Mode:   util.EnvStr("MODE", "master"),
	Listen: util.EnvStr("LISTEN", ":5212"),
}

// CORSConfig 跨域配置
var CORSConfig = &cors{
	AllowOrigins:     util.EnvArr("CORS_ALLOW_ORIGINS", []string{"UNSET"}),
	AllowMethods:     util.EnvArr("CORS_ALLOW_METHODS", []string{"PUT", "POST", "GET", "OPTIONS"}),
	AllowHeaders:     util.EnvArr("CORS_ALLOW_HEADERS", []string{"Cookie", "X-Cr-Policy", "Authorization", "Content-Length", "Content-Type", "X-Cr-Path", "X-Cr-FileName"}),
	AllowCredentials: util.EnvStr("CORS_ALLOW_CREDENTIALS", "false") == "true",
	ExposeHeaders:    util.EnvArr("CORS_EXPOSE_HEADERS", nil),
}

// SlaveConfig 从机配置
var SlaveConfig = &slave{
	Secret:          util.EnvStr("SLAVE_SECRET", ""),
	CallbackTimeout: util.EnvInt("SLAVE_CALLBACK_TIMEOUT", 10),
	SignatureTTL:    util.EnvInt("SLAVE_SIGNATURE_TTL", 10),
}

var SSLConfig = &ssl{
	Listen:   util.EnvStr("SSL_LISTEN", ":443"),
	CertPath: util.EnvStr("SSL_CERT_PATH", ""),
	KeyPath:  util.EnvStr("SSL_KEY_PATH", ""),
}

var UnixConfig = &unix{
	Listen:      util.EnvStr("UNIX_LISTEN", ""),
	ProxyHeader: util.EnvStr("UNIX_PROXY_HEADER", "X-Forwarded-For"),
}

var OptionOverwrite = map[string]interface{}{}
