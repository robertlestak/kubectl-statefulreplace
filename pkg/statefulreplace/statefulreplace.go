package statefulreplace

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Replacement struct {
	Container string `yaml:"container" json:"container"`
	Image     string `yaml:"image" json:"image"`
}

type StatefulReplace struct {
	Namespace    string        `yaml:"namespace" json:"namespace"`
	Kind         string        `yaml:"kind" json:"kind"`
	Name         string        `yaml:"name" json:"name"`
	Replacements []Replacement `yaml:"replacements" json:"replacements"`
}

func LoadConfig(f string) (*StatefulReplace, error) {
	l := log.WithFields(
		log.Fields{
			"action": "LoadConfig",
		},
	)
	l.Debug("LoadConfig")
	if f == "" {
		return nil, fmt.Errorf("file is empty")
	}
	var bd []byte
	var err error
	if f == "-" {
		l.Debug("reading from stdin")
		bd, err = os.ReadFile("/dev/stdin")
		if err != nil {
			l.Errorf("os.ReadFile error=%v", err)
			return nil, err
		}
	} else {
		l.Debugf("reading from file=%s", f)
		bd, err = os.ReadFile(f)
		if err != nil {
			l.Errorf("os.ReadFile error=%v", err)
			return nil, err
		}
	}
	sr := &StatefulReplace{}
	if err := yaml.Unmarshal(bd, sr); err != nil {
		// try json
		if err := json.Unmarshal(bd, sr); err != nil {
			l.Errorf("yaml.Unmarshal error=%v", err)
			return nil, err
		}
	}
	sr.Kind, err = KindName(sr.Kind)
	if err != nil {
		l.Errorf("KindName error=%v", err)
		return nil, err
	}
	return sr, nil
}

func int32Ptr(i int32) *int32 { return &i }

func (sr *StatefulReplace) Run() error {
	l := log.WithFields(
		log.Fields{
			"action": "Run",
		},
	)
	l.Debug("Run")
	switch sr.Kind {
	case "Deployment":
		return sr.ReplaceDeployment()
	case "StatefulSet":
		return sr.ReplaceStatefulSet()
	case "DaemonSet":
		return sr.ReplaceDaemonSet()
	default:
		return fmt.Errorf("unsupported kind: %s", sr.Kind)
	}
}
