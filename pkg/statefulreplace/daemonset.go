package statefulreplace

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (sr *StatefulReplace) ReplaceDaemonSet() error {
	l := log.WithFields(
		log.Fields{
			"action": "ReplaceDaemonSet",
		},
	)
	l.Debug("ReplaceDaemonSet")
	if sr.Kind != "DaemonSet" {
		return fmt.Errorf("kind is not DaemonSet")
	}
	daemonset, err := K8sClient.AppsV1().DaemonSets(sr.Namespace).Get(context.Background(), sr.Name, metav1.GetOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().DaemonSets(sr.Namespace).Get error=%v", err)
		return err
	}
	l.Debug("got daemonset")
	daemonset, err = K8sClient.AppsV1().DaemonSets(sr.Namespace).Get(context.Background(), sr.Name, metav1.GetOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().DaemonSets(sr.Namespace).Get error=%v", err)
		return err
	}
	// patch the DaemonSet with the new image
	for _, replacement := range sr.Replacements {
		for i, container := range daemonset.Spec.Template.Spec.Containers {
			if container.Name == replacement.Container {
				daemonset.Spec.Template.Spec.Containers[i].Image = replacement.Image
			}
		}
	}
	l.Debug("patched daemonset")
	_, err = K8sClient.AppsV1().DaemonSets(sr.Namespace).Update(context.Background(), daemonset, metav1.UpdateOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().DaemonSets(sr.Namespace).Update error=%v", err)
		return err
	}
	l.Debug("updated daemonset, waiting for new pods to be ready")
	// wait for all new pods to be ready
	time.Sleep(5 * time.Second)
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		selector := labels.Set(daemonset.Spec.Selector.MatchLabels)
		pods, err := K8sClient.CoreV1().Pods(sr.Namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: selector.AsSelector().String(),
		})
		if err != nil {
			return false, err
		}
		for _, pod := range pods.Items {
			if pod.Status.Phase != "Running" {
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		l.Errorf("wait.PollImmediate error=%v", err)
		return err
	}
	l.Debug("new pods are ready")
	println("daemonset.apps/" + sr.Name + " replaced")
	return nil
}
