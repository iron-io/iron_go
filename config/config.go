package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type Settings struct {
	Token      string `json:"token,omitempty"`
	ProjectId  string `json:"project_id,omitempty"`
	Host       string `json:"host,omitempty"`
	Scheme     string `json:"scheme,omitempty"`
	Port       uint16 `json:"port,omitempty"`
	ApiVersion string `json:"api_version,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
}

var (
	goVersion = runtime.Version()
	presets   = map[string]Settings{
		"worker": Settings{
			Scheme:     "https",
			Port:       443,
			ApiVersion: "2",
			Host:       "worker-aws-us-east-1.iron.io",
			UserAgent:  "go.iron/worker 2.0 (Go " + goVersion + ")",
		},
		"mq": Settings{
			Scheme:     "https",
			Port:       443,
			ApiVersion: "1",
			Host:       "mq-aws-us-east-1.iron.io",
			UserAgent:  "go.iron/mq 1.0 (Go " + goVersion + ")",
		},
		"cache": Settings{
			Scheme:     "https",
			Port:       443,
			ApiVersion: "1",
			Host:       "cache-aws-us-east-1.iron.io",
			UserAgent:  "go.iron/cache 1.0 (Go " + goVersion + ")",
		},
	}
)

// fullProduct is like "iron_worker" and "iron_mq", not "worker" or "mq", to
// keep some flexibility in future.
func Config(fullProduct string) (settings Settings) {
	pair := strings.SplitN(fullProduct, "_", 2)
	if len(pair) != 2 {
		panic("Invalid product name, has to use prefix.")
	}
	family, product := pair[0], pair[1]

	base, found := presets[product]

	if !found {
		base = Settings{
			Scheme:     "https",
			Port:       443,
			ApiVersion: "1",
			Host:       product + "-aws-us-east-1.iron.io",
			UserAgent:  "go.iron",
		}
	}

	base.globalConfig(family, product)
	base.globalEnv(family, product)
	base.productEnv(family, product)
	base.localConfig(family, product)

	return base
}

func (s *Settings) globalConfig(family, product string) {
	if u, err := user.Current(); err == nil {
		path := filepath.Join(u.HomeDir, ".iron.json")
		s.UseConfigFile(family, product, path)
	}
}

// The environment variables the scheme looks for are all of the same formula:
// the camel-cased product name is switched to an underscore (“IronWorker”
// becomes “iron_worker”) and converted to be all capital letters. For the
// global environment variables, “IRON” is used by itself. The value being
// loaded is then joined by an underscore to the name, and again capitalised.
// For example, to retrieve the OAuth token, the client looks for “IRON_TOKEN”.
func (s *Settings) globalEnv(family, product string) {
	eFamily := strings.ToUpper(family) + "_"
	s.commonEnv(eFamily)
}

// In the case of product-specific variables (which override global variables),
// it would be “IRON_WORKER_TOKEN” (for IronWorker).
func (s *Settings) productEnv(family, product string) {
	eProduct := strings.ToUpper(family) + "_" + strings.ToUpper(product) + "_"
	s.commonEnv(eProduct)
}

func (s *Settings) localConfig(family, product string) {
	s.UseConfigFile(family, product, "iron.json")
}

func (s *Settings) commonEnv(prefix string) {
	if token := os.Getenv(prefix + "TOKEN"); token != "" {
		s.Token = token
	}
	if pid := os.Getenv(prefix + "PROJECT_ID"); pid != "" {
		s.ProjectId = pid
	}
	if host := os.Getenv(prefix + "HOST"); host != "" {
		s.Host = host
	}
	if scheme := os.Getenv(prefix + "SCHEME"); scheme != "" {
		s.Scheme = scheme
	}
	if port := os.Getenv(prefix + "PORT"); port != "" {
		n, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			panic(err)
		}
		s.Port = uint16(n)
	}
	if vers := os.Getenv(prefix + "API_VERSION"); vers != "" {
		s.ApiVersion = vers
	}
}

func (s *Settings) UseConfigFile(family, product, path string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	data := map[string]interface{}{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		panic("Invalid JSON in " + path + ": " + err.Error())
	}

	s.UseConfigMap(data)

	ipData, found := data[family+"_"+product]
	if found {
		pData := ipData.(map[string]interface{})
		s.UseConfigMap(pData)
	}
}

func (s *Settings) UseConfigMap(data map[string]interface{}) {
	if token, found := data["token"]; found {
		s.Token = token.(string)
	}
	if projectId, found := data["project_id"]; found {
		s.ProjectId = projectId.(string)
	}
	if host, found := data["host"]; found {
		s.Host = host.(string)
	}
	if prot, found := data["scheme"]; found {
		s.Scheme = prot.(string)
	}
	if port, found := data["port"]; found {
		s.Port = uint16(port.(float64))
	}
	if vers, found := data["api_version"]; found {
		s.ApiVersion = vers.(string)
	}
	if vers, found := data["user_agent"]; found {
		s.ApiVersion = vers.(string)
	}
}
