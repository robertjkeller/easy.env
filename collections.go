package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"slices"
)

type Collection struct {
	Name        string
	Description string
	VarIds      []int
}

type Collections []Collection

func NewCollection(name, description string) Collection {
	return Collection{
		Name:        name,
		Description: description,
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
func (c *Collection) GetVarIds() []int {
	return c.VarIds
}
func (c *Collection) GetName() string {
	return c.Name
}
func (c *Collection) GetDescription() string {
	return c.Description
}
func (c *Collection) SetName(name string) {
	c.Name = name
}
func (c *Collection) SetDescription(description string) {
	c.Description = description
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
