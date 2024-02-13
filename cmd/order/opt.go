package main

import (
	"fmt"
)

type opt struct {
	host string `yaml:"host"`
	user string `yaml:"user"`
	pass string `yaml:"pass"`
	port string `yaml:"port"`
	name string `yaml:"name"`
}

func (o *opt) ConnectionString() string {
	if o.host == "" {
		o.host = "@"
	}

	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", o.user, o.pass, o.host, o.port, o.name)
}
