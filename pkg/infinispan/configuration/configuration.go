package configuration

import (
	"gopkg.in/yaml.v2"
)

// InfinispanConfiguration is the top level configuration type
type InfinispanConfiguration struct {
	Infinispan  Infinispan   `yaml:"infinispan"`
	JGroups     JGroups      `yaml:"jgroups"`
	Keystore    Keystore     `yaml:"keystore,omitempty"`
	Truststore  Truststore   `yaml:"truststore,omitempty"`
	XSite       *XSite       `yaml:"xsite,omitempty"`
	Logging     Logging      `yaml:"logging,omitempty"`
	Endpoints   Endpoints    `yaml:"endpoints"`
	CloudEvents *CloudEvents `yaml:"cloudEvents,omitempty"`
}

type CloudEvents struct {
	BootstrapServers  string `yaml:"bootstrapServers"`
	Acks              string `yaml:"acks"`
	CacheEntriesTopic string `yaml:"cacheEntriesTopic"`
}

type Infinispan struct {
	Authorization    Authorization `yaml:"authorization,omitempty"`
	ClusterName      string        `yaml:"clusterName"`
	ZeroCapacityNode bool          `yaml:"zeroCapacityNode"`
	Locks            Locks         `yaml:"locks"`
}

type Authorization struct {
	Enabled    bool                `yaml:"enabled"`
	RoleMapper string              `yaml:"roleMapper,omitempty"`
	Roles      []AuthorizationRole `yaml:"roles,omitempty"`
}

type AuthorizationRole struct {
	Name        string   `yaml:"name"`
	Permissions []string `yaml:"permissions"`
}

type Endpoint struct {
	Enabled    bool   `yaml:"enabled,omitempty"`
	Qop        string `yaml:"qop,omitempty"`
	ServerName string `yaml:"serverName,omitempty"`
}

type Endpoints struct {
	Enabled        bool     `yaml:"enabled,omitempty"`
	Cors           bool     `yaml:"cors,omitempty"` // TODO: cors not implemented
	Authenticate   bool     `yaml:"auth"`
	DedicatedAdmin bool     `yaml:"dedicatedAdmin"`
	ClientCert     string   `yaml:"clientCert,omitempty"`
	Hotrod         Endpoint `yaml:"hotrod,omitempty"`
	Memcached      Endpoint `yaml:"memcached,omitempty"`
}

type Locks struct {
	Owners      int32  `yaml:"owners,omitempty"`
	Reliability string `yaml:"reliability,omitempty"`
}

// Keystore configuration info for endpoint encryption
type Keystore struct {
	Path         string
	Password     string
	Alias        string
	CrtPath      string `yaml:"crtPath,omitempty"`
	SelfSignCert string `yaml:"selfSignCert,omitempty"`
	Type         string `yaml:"type,omitempty"`
}

// Truststore configuration info for endpoint encryption
type Truststore struct {
	CaFile   string `yaml:"cafile,omitempty"`
	Certs    string `yaml:"certs,omitempty"`
	Path     string `yaml:"path,omitempty"`
	Password string
}

// JGroups configures clustering layer
type JGroups struct {
	Transport   string  `yaml:"transport"`
	DNSPing     DNSPing `yaml:"dnsPing"`
	Diagnostics bool    `yaml:"diagnostics"`
	BindPort    int32   `yaml:"bindPort"`
	Encrypt     bool    `yaml:"encrypt"`
	Relay       Relay   `yaml:"relay"`
}

type Relay struct {
	BindPort int32 `yaml:"bindPort"`
}

// DNSPing configures DNS cluster lookup settings
type DNSPing struct {
	Query      string `yaml:"query"`
	Address    string `yaml:"address"`
	RecordType string `yaml:"recordType"`
}

type XSite struct {
	Address         string       `yaml:"address"`
	Name            string       `yaml:"name"`
	Port            int32        `yaml:"port"`
	Transport       string       `yaml:"transport"`
	MaxSiteMasters  int32        `yaml:"maxSiteMasters"`
	Backups         []BackupSite `yaml:"backups"`
	MasterCandidate bool         `yaml:"masterCandidate"`
	Relay           RelayXSite   `yaml:"relay"`
}

type RelayXSite struct {
	BindPort int32 `yaml:"bindPort"`
}

type BackupSite struct {
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
	Port    int32  `yaml:"port"`
}

type Logging struct {
	Console    Console           `yaml:"console"`
	File       File              `yaml:"file"`
	Categories map[string]string `yaml:"categories,omitempty"`
}

type Console struct {
	Level   string `yaml:"level"`
	Pattern string `yaml:"pattern"`
}

type File struct {
	Level   string `yaml:"level"`
	Pattern string `yaml:"pattern"`
	Path    string `yaml:"path"`
}

func (c *InfinispanConfiguration) Yaml() (string, error) {
	y, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(y), nil
}

func FromYaml(src string) (*InfinispanConfiguration, error) {
	config := &InfinispanConfiguration{}
	if err := yaml.Unmarshal([]byte(src), config); err != nil {
		return nil, err
	}
	return config, nil
}

func DefaultInfinspanConfiguration() InfinispanConfiguration {
	return InfinispanConfiguration{
		Infinispan: Infinispan{
			Authorization: Authorization{Enabled: true, RoleMapper: "cluster"},
			ClusterName:   "infinispan",
			Locks:         Locks{Owners: -1, Reliability: "consistent"}},
		JGroups: JGroups{
			Transport: "tcp",
			BindPort:  7800,
			DNSPing:   DNSPing{RecordType: "A"},
			Relay:     Relay{BindPort: 7900}},
		Keystore: Keystore{
			Password: "password",
			Alias:    "server",
			Type:     "pkcs12"},
		Truststore: Truststore{},
		XSite: &XSite{
			MasterCandidate: true,
			MaxSiteMasters:  1,
			Transport:       "tcp"},
		Logging: Logging{
			Console: Console{Level: "trace", Pattern: "'%d{HH:mm:ss,SSS} %-5p (%t) [%c] %m%throwable%n'"},
			File: File{Level: "trace",
				Pattern: "'%d{yyyy-MM-dd HH:mm:ss,SSS} %-5p (%t) [%c] %m%throwable%n'",
				Path:    "'${sys:infinispan.server.log.path}/server.log'"},
			Categories: map[string]string{
				"com.arjuna":     "warn",
				"org.infinispan": "info",
				"org.jgroups":    "warn",
				"io.netty.handler.ssl.ApplicationProtocolNegotiationHandler": "error"}},
		Endpoints: Endpoints{
			Authenticate: true,
			ClientCert:   "none",
			Hotrod:       Endpoint{Enabled: true, Qop: "auth", ServerName: "infinispan"},
			Enabled:      true,
		},
		CloudEvents: &CloudEvents{},
	}
}
