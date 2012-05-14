package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

type Settings struct {
	Token      string `json:"token"`
	ProjectId  string `json:"project_id"`
	Host       string `json:"host"`
	Protocol   string `json:"protocol"`
	Port       int    `json:"port"`
	ApiVersion string `json:"api_version"`
}

var (
	presets = map[string]Settings{
		"worker": Settings{Protocol: "https", Port: 443, ApiVersion: "1", Host: "worker-aws-us-east-1.iron.io"},
		"mq":     Settings{Protocol: "https", Port: 443, ApiVersion: "1", Host: "mq-aws-us-east-1.iron.io"},
		"cache":  Settings{Protocol: "https", Port: 443, ApiVersion: "1", Host: "cache-aws-us-east-1.iron.io"},
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
			Protocol:   "https",
			Port:       443,
			ApiVersion: "1",
			Host:       product + "-aws-us-east-1.iron.io",
		}
	}

	// The global configuration file sets the defaults according to the file hierarchy.
	(&base).globalConfig(family, product)
	fmt.Println(base)

	// The global environment variables overwrite the global configuration file’s values.
	(&base).globalEnv(family, product)
	fmt.Println(base)

	// The product-specific environment variables overwrite everything before them.
	(&base).productEnv(family, product)
	fmt.Println(base)

	// The local configuration file overwrites everything before it according to the file hierarchy.
	(&base).localConfig(family, product)
	fmt.Println(base)

	// The configuration file specified when instantiating the client library overwrites everything before it according to the file hierarchy.
	(&base).forceConfig(family, product)
	fmt.Println(base)

	// The arguments passed when instantiating the client library overwrite everything before them.
	(&base).passedConfig(family, product)
	fmt.Println(base)

	return base
}

func (s *Settings) globalConfig(family, product string) {
	u, err := user.Current()
	if err != nil {
		panic(err.Error())
	}

	path := filepath.Join(u.HomeDir, ".iron.json")
	s.commonConfigFile(family, product, path)
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
	s.commonConfigFile(family, product, "iron.json")
}
func (s *Settings) forceConfig(family, product string)  {}
func (s *Settings) passedConfig(family, product string) {}

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
	if prot := os.Getenv(prefix + "PROTOCOL"); prot != "" {
		s.Protocol = prot
	}
	if port := os.Getenv(prefix + "PORT"); port != "" {
		n, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			panic(err)
		}
		s.Port = int(n)
	}
	if vers := os.Getenv(prefix + "API_VERSION"); vers != "" {
		s.ApiVersion = vers
	}
}

func (s *Settings) commonConfigFile(family, product, path string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	data := map[string]interface{}{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		panic("Invalid JSON in " + path + ": " + err.Error())
	}

	if host, found := data["host"]; found {
		s.Host = host.(string)
	}
	if prot, found := data["protocol"]; found {
		s.Protocol = prot.(string)
	}
	if port, found := data["port"]; found {
		s.Port = int(port.(float64))
	}
	if vers, found := data["api_version"]; found {
		s.ApiVersion = vers.(string)
	}

	ipData, found := data[family+"_"+product]
	if found {
		pData := ipData.(map[string]interface{})
		if host, found := pData["host"]; found {
			s.Host = host.(string)
		}
		if prot, found := pData["protocol"]; found {
			s.Protocol = prot.(string)
		}
		if port, found := pData["port"]; found {
			s.Port = int(port.(float64))
		}
		if vers, found := pData["api_version"]; found {
			s.ApiVersion = vers.(string)
		}
	}
}
