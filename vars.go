package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// TODO: this is for easier dev, will be changed
const configDir = "./config"

var varIDCounter int

type Var struct {
	Id          int
	Key         string
	Val         string
	Description string
}

func NewVar(key, val, description string) Var {
	varIDCounter++
	return Var{
		Id:          varIDCounter,
		Key:         key,
		Val:         val,
		Description: description,
	}
}

type Vars []Var

func NewVars() Vars {
	return Vars{}
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

// Save Vars to user's config folder
func (v *Vars) Save() error {
	err := os.MkdirAll(configDir, os.ModePerm)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	file, err := os.Create(configDir + "/vars.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(v)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return nil
}

// Load Vars from user's config folder
func (v *Vars) Load() error {
	file, err := os.Open(configDir + "/vars.json")
	if err != nil {
		slog.Debug("User config file not found, creating new one")
		v.Save()
		return nil
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(v)
	if err != nil {
		return err
	}
	return nil
}
