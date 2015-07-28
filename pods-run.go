package main

import (
	"errors"
        "fmt"
	"io/ioutil"
        "log"
	"os"
	"os/exec"
	"strings"

        "gopkg.in/yaml.v2"
)

type port struct {
        ContainerPort int `yaml:"containerPort"`
	HostPort int `yaml:"hostPort"`
}

type volumeMounts struct {
        Name string
	HostPath string `yaml:"hostPath"`
	MountPath string `yaml:"mountPath"`
}

type env struct {
        Key string
	Value string
}

type container struct {
        Name string
	Image string

	Ports []port
	VolumeMounts []volumeMounts `yaml:"volumeMounts"`
	Env []env
}

type pod struct {
	Id string
	Kind string
	DesiredState struct {
		Manifest struct {
			Containers []container
		}
	} `yaml:"desiredState"`
}

func (p *pod) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, p); err != nil {
		return err
	}
	if p.Kind != "Pod" {
		return errors.New("pods: invalid `kind`")
	}
	return nil
}

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	data, err := ioutil.ReadFile(pwd + "/pods.yaml")
        if err != nil {
                log.Fatalf("error: %v", err)
        }

	var p pod
	if err = p.Parse(data); err != nil {
		log.Fatal(err)
	}

	ct := p.DesiredState.Manifest.Containers[0]

	args := []string{}
	args = append(args, "/usr/bin/docker run -d")

	for _,s := range ct.Ports {
		args = append(args, fmt.Sprintf("-p \"%v:%v\"", s.HostPort, s.ContainerPort))
	}

	for _,s := range ct.Env {
		args = append(args, fmt.Sprintf("-e \"%v=%v\"", s.Key, s.Value))
	}

	for _,s := range ct.VolumeMounts {
		args = append(args, fmt.Sprintf("-v \"%v:%v\"", s.HostPath, s.MountPath))
	}

	args = append(args, ct.Image)

	command := strings.Join(args, " ")

	out, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
}
