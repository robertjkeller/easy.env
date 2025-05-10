package main

import (
	"encoding/json"
	"log/slog"
	"os"
)

var varIDCounter int = 0

type Var struct {
	Id          int
	Key         string
	Val         string
	Description string
}

func NewVar(key, val, description string) Var {
	n := Var{
		Id:          varIDCounter,
		Key:         key,
		Val:         val,
		Description: description,
	}
	varIDCounter++
	return n
}

type Vars []Var

func NewVars() Vars {
	return Vars{}
}

func (v *Vars) All() []Var {
	return *v
}

func (v *Vars) Get(key string) string {
	for _, v := range *v {
		if v.Key == key {
			return v.Val
		}
	}
	return ""
}

func (v *Vars) GetDescription(key string) string {
	for _, v := range *v {
		if v.Key == key {
			return v.Description
		}
	}
	return ""
}

func (v *Vars) Add(key, val, description string) {
	*v = append(*v, NewVar(key, val, description))
}

func (v *Vars) Set(key, val, description string) {
	for i, currentVar := range *v {
		if currentVar.Key == key {
			(*v)[i].Val = val
			(*v)[i].Description = description
			return
		}
	}
	*v = append(*v, NewVar(key, val, description))
}

func (v *Vars) GetIdFromKey(key string) int {
	for _, v := range *v {
		if v.Key == key {
			return v.Id
		}
	}
	return -1
}

// Save Vars to user's config folder
func (v *Vars) Save() error {
	err := os.MkdirAll(ConfigDir, os.ModePerm)
	if err != nil {
		return &UserConfigError{Err: err}
	}
	file, err := os.Create(ConfigDir + "/vars.json")
	if err != nil {
		return &UserConfigError{Err: err}
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(v)
	if err != nil {
		return &UserConfigError{Err: err}
	}
	return nil
}

// Load Vars from user's config folder
func (v *Vars) Load() error {
	file, err := os.Open(ConfigDir + "/vars.json")
	if err != nil {
		slog.Debug("User config file not found, creating new one")
		v.Save()
		return nil
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(v)
	if err != nil {
		return &UserConfigError{Err: err}
	}
	return nil
}
