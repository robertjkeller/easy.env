package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"slices"
)

type Collection struct {
	Name     string
	Filename string
	VarIds   []int
}

type Collections []Collection

func NewCollection(name, description string) Collection {
	return Collection{
		Name:     name,
		Filename: description,
	}
}
func (c *Collection) AddVar(varID int) {
	c.VarIds = append(c.VarIds, varID)
}
func (c *Collection) RemoveVar(varID int) {
	for i, id := range c.VarIds {
		if id == varID {
			c.VarIds = slices.Delete(c.VarIds, i, i+1)
			break
		}
	}
}

func (c *Collection) Vars() Vars {
	userVars := NewVars()
	userVars.Load()
	vars := Vars{}
	for _, id := range c.VarIds {
		for _, v := range userVars {
			if v.Id == id {
				vars = append(vars, v)
			}
		}
	}
	return vars
}

func (c *Collection) GetVarIds() []int {
	return c.VarIds
}

func (c *Collection) GetName() string {
	return c.Name
}

func (c *Collection) GetDescription() string {
	return c.Filename
}

func (c *Collection) SetName(name string) {
	c.Name = name
}

func (c *Collection) SetFilename(filename string) {
	c.Filename = filename
}

func (c *Collection) GetVarID(index int) int {
	if index < 0 || index >= len(c.VarIds) {
		return -1
	}
	return c.VarIds[index]
}

func (c *Collection) GetVarCount() int {
	return len(c.VarIds)
}

// Save Collections to user's config folder
func (c *Collections) Save() error {
	err := os.MkdirAll(ConfigDir, os.ModePerm)
	if err != nil {
		return &UserConfigError{Err: err}
	}
	file, err := os.Create(ConfigDir + "/collections.json")
	if err != nil {
		return &UserConfigError{Err: err}
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(c)
	if err != nil {
		return &UserConfigError{Err: err}
	}
	return nil
}

// Load Collections from user's config folder
func (c *Collections) Load() error {
	file, err := os.Open(ConfigDir + "/collections.json")
	if err != nil {
		slog.Debug("User config file not found, creating new one")
		c.Save()
		return nil
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		return &UserConfigError{Err: err}
	}
	return nil
}

func (c *Collection) WriteToEnvFile(dir, filename string) error {
	collectionFile, err := os.Create(dir + "/" + filename)
	if err != nil {
		return &UserConfigError{Err: err}
	}
	defer collectionFile.Close()
	for _, v := range c.Vars() {
		_, err := collectionFile.WriteString(v.Key + "=" + v.Val + "\n")
		if err != nil {
			return &UserConfigError{Err: err}
		}
	}
	return nil
}

func (c *Collection) WriteToSymlink(dir string) error {
	err := os.MkdirAll(ConfigDir+"/env_files", os.ModePerm)
	if err != nil {
		return &UserConfigError{Err: err}
	}

	err = c.WriteToEnvFile(ConfigDir+"/env_files", c.Name+".env")
	if err != nil {
		return &UserConfigError{Err: err}
	}

	symlinkTargetPath := dir + "/" + c.Filename
	if _, statErr := os.Stat(symlinkTargetPath); statErr == nil {
		err = os.Remove(symlinkTargetPath)
		if err != nil {
			return &UserConfigError{Err: err}
		}
	} else if !os.IsNotExist(statErr) {
		return &UserConfigError{Err: statErr}
	}

	err = os.Symlink(ConfigDir+"/env_files/"+c.Name+".env", symlinkTargetPath)
	if err != nil {
		return &UserConfigError{Err: err}
	}

	return nil
}
