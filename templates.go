package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"strings"
	"text/template"
)

// dealing with os layouts and templates

// get a listing of the operating systems
func (c *Config) OSListGet() (os map[string]*operatingSystem) {
	os = make(map[string]*operatingSystem)
	list, _ := c.fs.List("boot")
	logger.Info("OS listing ", list)
	for _, i := range list {
		logger.Info(" ----- " + i + "-------")
		tmpOS := &operatingSystem{Name: i, Description: i}
		if tmpOS.CheckAndLoad(c) == true {
			os[i] = tmpOS
		}
	}
	return
}

// check and load the operating system
func (os *operatingSystem) CheckAndLoad(c *Config) (pass bool) {
	subList, err := c.fs.List("boot/" + os.Name)
	if err != nil {
		logger.Error("SubList Error %s", err)
		return false
	}
	for _, i := range subList {
		logger.Debug("found folder %s", i)
		if i == "template" {
			templatePass := os.LoadTemplates(c)
			if templatePass == false {
				return false
			}

		}
	}
	return true
}

// load the templates for the operating system
func (os *operatingSystem) LoadTemplates(c *Config) (pass bool) {
	path := "boot/" + os.Name + "/template"
	templateList, err := c.fs.List(path)
	if err != nil {
		logger.Critical("Template Load fail %s", err)
		return false
	}
	//load all the templates for the operating sytem
	newTemplates := template.New("")
	for i, j := range templateList {
		template, _ := c.fs.Get(path + "/" + j)
		defer template.Close()
		if strings.HasSuffix(j, ".tmpl") {
			data, err := ioutil.ReadAll(template)
			name := strings.TrimSuffix(j, ".tmpl")
			_, err = newTemplates.New(name).Parse(string(data))
			if err != nil {
				fmt.Println("template ", name, " -> ", string(data), err)
			}
			logger.Critical("#%d : %s", i, j)
		}
	}
	os.templates = newTemplates
	// check for operating system classes
	classPath := "boot/" + os.Name + "/" + "classes.toml"
	classFile, err := c.fs.Get(classPath)
	defer classFile.Close()
	if err != nil {
		logger.Critical("Class List fail, %s", err)
		// still returns true , so the OS is added
		// TODO class loader
		return true
	}
	// load the clases file
	classString, _ := ioutil.ReadAll(classFile)
	cl := &classes{}
	_, err = toml.Decode(string(classString), &cl)
	fmt.Println(cl, err)
	// attach to the os list
	os.Classes = cl.Classes
	logger.Debug("Class File : %s", cl)
	os.HasClasses = true
	return true
}
