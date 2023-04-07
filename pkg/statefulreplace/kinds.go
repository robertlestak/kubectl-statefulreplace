package statefulreplace

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	KindShortNames = map[string][]string{
		"Deployment":  {"deploy", "deployments"},
		"StatefulSet": {"sts", "statefulsets"},
		"DaemonSet":   {"ds", "daemonsets"},
	}
)

func KindName(kind string) (string, error) {
	l := log.WithFields(log.Fields{
		"action": "KindName",
		"kind":   kind,
	})
	l.Debug("KindName")
	for k, v := range KindShortNames {
		if strings.EqualFold(kind, k) {
			l.Debugf("found kind %s", k)
			return k, nil
		}
		for _, s := range v {
			if strings.EqualFold(kind, s) {
				l.Debugf("found kind %s", k)
				return k, nil
			}
		}
	}
	return "", fmt.Errorf("kind %s not found", kind)
}
