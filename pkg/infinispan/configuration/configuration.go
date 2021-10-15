package configuration

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	rice "github.com/GeertJohan/go.rice"
	consts "github.com/infinispan/infinispan-operator/controllers/constants"
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

func (serverConf *InfinispanConfiguration) InfinispanConfiguration() (infinispan, relay string, err error) {
	// Setup go template to process infinispan.xml and jgroups-relay.xml
	funcMap := template.FuncMap{
		"UpperCase":    strings.ToUpper,
		"LowerCase":    strings.ToLower,
		"ServerRoot":   func() string { return consts.ServerRoot },
		"ListAsString": func(elems []string) string { return strings.Join(elems, ",") },
		"RemoteSites": func(elems []BackupSite) string {
			var ret string
			for i, bs := range elems {
				ret += fmt.Sprintf("%s[%d]", bs.Address, bs.Port)
				if i < len(elems)-1 {
					ret += ","
				}
			}
			return ret
		},
	}
	var ispnXmlTemplate, jgroupsXmlTemplate string
	if box, err := rice.FindBox("resources"); err != nil {
		return "", "", err
	} else {
		if ispnXmlTemplate, err = box.String("ispnXmlTemplate.xmltmpl"); err != nil {
			return "", "", err
		}
		if jgroupsXmlTemplate, err = box.String("jgroupsXmlTemplate.xmltmpl"); err != nil {
			return "", "", err
		}
	}

	tIspn, err := template.New("infinispan.xml").Funcs(funcMap).Parse(ispnXmlTemplate)
	if err != nil {
		return "", "", err
	}
	buffIspn := new(bytes.Buffer)
	err = tIspn.Execute(buffIspn, serverConf)
	if err != nil {
		return "", "", err
	}

	var buffJGroups *bytes.Buffer
	if serverConf.XSite != nil && len(serverConf.XSite.Backups) > 0 {
		// Generate jgroups-relay.xml
		tJGroups, err := template.New("jgroups-relay.xml").Funcs(funcMap).Parse(jgroupsXmlTemplate)
		if err != nil {
			return "", "", err
		}
		buffJGroups = new(bytes.Buffer)
		err = tJGroups.Execute(buffJGroups, serverConf)
		if err != nil {
			return "", "", err
		}
	}
	return buffIspn.String(), buffJGroups.String(), nil
}

// 	// Create admin and user identity properties from secrets
// 	var adminBash string
// 	adminBash, err = security.IdentitiesCliFileFromSecret(adminPropSecret.Data[consts.ServerIdentitiesFilename], "admin", ServerRoot+"/conf/cli-admin-users.properties", ServerRoot+"/conf/cli-admin-groups.properties")
// 	if err != nil {
// 		return "", "", err
// 	}

// 	var usersBash string
// 	if userPropSecret != nil {
// 		if usersBash, err = security.IdentitiesCliFileFromSecret(userPropSecret.Data[consts.ServerIdentitiesFilename], "default", ServerRoot+"/conf/cli-users.properties", ServerRoot+"/conf/cli-groups.properties"); err != nil {
// 			return "", "", err
// 		}
// 	}
// 	bash := adminBash + usersBash

// 	// PEM certs need to be loaded and merget to be used by Infinispan
// 	var pem []byte
// 	if serverConf.Keystore.Type == "pem" {
// 		keystoreSecret := &corev1.Secret{}
// 		if result, err := kube.LookupResource(r.infinispan.GetKeystoreSecretName(), r.infinispan.Namespace, keystoreSecret, r.Client, reqLogger, r.eventRec, r.ctx); result != nil {
// 			return "", "", err
// 		}
// 		pem = append(keystoreSecret.Data["tls.key"], keystoreSecret.Data["tls.crt"]...)
// 	}

// 	// Create secret with all the objects to be mounted as "ServerRoot/conf/operator/"
// 	result, err := controllerutil.CreateOrUpdate(r.ctx, r.Client, infinispanXmlObject, func() error {
// 		infinispanXmlObject.Labels = LabelsResource(r.infinispan.Name, "infinispan-secret-admin-identities")
// 		infinispanXmlObject.Data = map[string][]byte{"infinispan.xml": buffIspn.Bytes()}
// 		if buffJGroups != nil {
// 			infinispanXmlObject.Data["jgroups-relay.xml"] = buffJGroups.Bytes()
// 		}
// 		infinispanXmlObject.Data[consts.ServerIdentitiesCliFilename] = []byte(bash)
// 		infinispanXmlObject.Data[EncryptPemKeystoreName] = []byte(pem)
// 		err = controllerutil.SetControllerReference(r.infinispan, infinispanXmlObject, r.scheme)
// 		return err
// 	})
// 	if err != nil {
// 		return &reconcile.Result{}, err
// 	}
// 	if result != controllerutil.OperationResultNone {
// 		r.reqLogger.Info(fmt.Sprintf("ConfigMap '%s' %s", name, result))
// 	}
// 	return nil, nil
// }
