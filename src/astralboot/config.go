// Config loading and information structures
package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
)

// OperatingSystem : struct ,  loaded in templates.go
type OperatingSystem struct {
	Name        string
	Description string
	HasClasses  bool
	Classes     []string
	templates   *template.Template
}

// c.lass list
// sub type of the operating system type
type classes struct {
	Classes []string `toml:"classes"`
}

//Refs :  references for ipfs loading
type Refs struct {
	Boot   string `toml:"boot"`
	Rocket string `toml:"rocket"`
}

//Config :  base configuration structure
type Config struct {
	Interf string `toml:"interface"`
	// switchable services
	Spawn     bool
	BaseIP    net.IP
	Gateway   net.IP
	Subnet    net.IP
	DNSServer net.IP
	Domain    string `toml:"Domain,omitempty"`
	DBname    string `toml:"DBname,omitempty"`
	IPFS      bool
	Refs      *Refs // ipfs references
	OSList    map[string]*OperatingSystem
	// not exported generated config parts
	fs ROfs
}

//GetConfig :  loads config and settings from ipfs ref or file system
func GetConfig(path string) (c *Config) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		logger.Critical("Config Error %s", err)
		logger.Critical("Config file %s does not exists", path)
		logger.Critical("Starting Configurator")
		c = &Config{}
		c.Setup()
	}
	// loading the refs from seperate file
	re := &Refs{}
	if _, err := toml.DecodeFile("refs.toml", &re); err != nil {
		logger.Critical("Reference file does not exists")
	}
	c.Refs = re
	// bind the cache (not exported)
	// Add items from system not in config file
	if c.Interf == "" {
		c.Interf = "eth1"
	}
	interf, err := net.InterfaceByName(c.Interf)
	if err != nil {
		logger.Error("Interface error ", err)
	}
	//TODO fix interface checks
	addressList, _ := interf.Addrs()
	serverAddress, ipnet, _ := net.ParseCIDR(addressList[0].String())
	logger.Notice("Server Address  : %s", serverAddress)
	c.BaseIP = serverAddress
	b := ipnet.Mask
	c.Subnet = net.IP{b[0], b[1], b[2], b[3]}
	if c.Gateway == nil {
		c.Gateway = serverAddress
	}
	if c.DNSServer == nil {
		c.DNSServer = serverAddress
	}
	// database file name
	if c.DBname == "" {
		c.DBname = "./machines.json"
	}

	if c.Domain == "" {
		c.Domain = "erf"
	}

	var filesystem ROfs
	if c.IPFS == true {
		logger.Critical("Using IPFS for boot files")
		filesystem = &IPfsfs{c.Refs.Boot}
	} else {
		filesystem = &Diskfs{"./data"}
	}
	c.fs = filesystem

	// distributions
	c.OSList = c.OSListGet()

	return
}

//PrintConfig : dumps config to stdout
func (c *Config) PrintConfig() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	if err != nil {
		logger.Fatal("Config Encode error %v", err)
	}
	fmt.Println(buf.String())
}

// Save to file
func (c *Config) Save() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	if err != nil {
		logger.Critical("Config Marshal Fail %v", err)
	}
	f, err := os.Create("config.toml")
	defer f.Close()
	if err != nil {
		logger.Critical("Config Save Fail %v", err)
	}
	f.Write(buf.Bytes())
}
