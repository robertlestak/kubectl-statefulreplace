package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/robertlestak/kubectl-statefulreplace/pkg/statefulreplace"
	log "github.com/sirupsen/logrus"
)

var (
	Version = "dev"
	flagset *flag.FlagSet
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func usage() {
	cmdName := os.Args[0]
	if strings.HasPrefix(filepath.Base(cmdName), "kubectl-") {
		cmdName = "kubectl " + strings.TrimPrefix(filepath.Base(cmdName), "kubectl-")
	}
	usageCmd := cmdName + " -n <namespace> [kind]/[name] [container]/[image] [container]/[image] ..."
	usageFile := cmdName + " -f <config-file>"
	println("Usage:")
	println(usageCmd)
	println(usageFile)
	flagset.PrintDefaults()
}

func version() {
	println("Version: " + Version)
}

func main() {
	// usage:
	// stateful-replace -n <namespace> [kind]/[name] [container]/[image] [container]/[image] ...
	// stateful-replace -f <config-file>
	var namespace string
	var configFile string
	var logLevel string
	var showVersion bool
	flagset = flag.NewFlagSet("statefulreplace", flag.ExitOnError)
	flagset.Usage = usage
	flagset.StringVar(&logLevel, "log-level", log.GetLevel().String(), "Log level")
	flagset.StringVar(&namespace, "n", "", "Namespace")
	flagset.StringVar(&configFile, "f", "", "Config file")
	flagset.BoolVar(&showVersion, "version", false, "Version")
	flagset.Parse(os.Args[1:])
	ll, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %s", logLevel)
	}
	log.SetLevel(ll)
	if showVersion {
		version()
		os.Exit(0)
	}
	var sr *statefulreplace.StatefulReplace
	if namespace == "" {
		namespace = "default"
	}
	if configFile != "" {
		sr, err = statefulreplace.LoadConfig(configFile)
		if err != nil {
			log.Fatalf("LoadConfig error=%v", err)
		}
	} else {
		sr = &statefulreplace.StatefulReplace{
			Namespace: namespace,
		}
		for _, arg := range flagset.Args() {
			// kind/name
			if sr.Kind == "" {
				if strings.Contains(arg, "/") {
					ss := strings.Split(arg, "/")
					kind := ss[0]
					name := strings.Join(ss[1:], "/")
					kind, err := statefulreplace.KindName(kind)
					if err != nil {
						log.Fatalf("Invalid kind: %s", kind)
					}
					sr.Kind = kind
					sr.Name = name
					continue
				}
			}
			// container/image
			if strings.Contains(arg, "/") {
				ss := strings.Split(arg, "/")
				sr.Replacements = append(sr.Replacements, statefulreplace.Replacement{
					Container: ss[0],
					Image:     strings.Join(ss[1:], "/"),
				})
				continue
			}
			log.Fatalf("Invalid argument: %s", arg)
		}
	}
	if sr.Kind == "" {
		usage()
		os.Exit(1)
	}
	if sr.Namespace == "" {
		sr.Namespace = "default"
	}
	if err := sr.Run(); err != nil {
		log.Fatalf("Run error=%v", err)
	}
}
