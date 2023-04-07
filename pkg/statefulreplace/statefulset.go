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

func (sr *StatefulReplace) ReplaceStatefulSet() error {
	l := log.WithFields(
		log.Fields{
			"action": "ReplaceStatefulSet",
		},
	)
	l.Debug("ReplaceStatefulSet")
	if sr.Kind != "StatefulSet" {
		return fmt.Errorf("kind is not StatefulSet")
	}
	// get the current replica count
	statefulset, err := K8sClient.AppsV1().StatefulSets(sr.Namespace).Get(context.Background(), sr.Name, metav1.GetOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().StatefulSets(sr.Namespace).Get error=%v", err)
		return err
	}
	l.Debug("got statefulset")
	var origReplicas int32
	if statefulset.Spec.Replicas != nil {
		origReplicas = *statefulset.Spec.Replicas
	}
	// scale the StatefulSet to 1
	statefulset.Spec.Replicas = int32Ptr(1)
	_, err = K8sClient.AppsV1().StatefulSets(sr.Namespace).Update(context.Background(), statefulset, metav1.UpdateOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().StatefulSets(sr.Namespace).Update error=%v", err)
		return err
	}
	l.Debug("scaled statefulset to 1 replica, waiting for pod to be ready")
	// wait for the new pod to be ready
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		selector := labels.Set(statefulset.Spec.Selector.MatchLabels)
		pods, err := K8sClient.CoreV1().Pods(sr.Namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: selector.AsSelector().String(),
		})
		if err != nil {
			return false, err
		}
		// ensure there is only one pod and that it is running
		if len(pods.Items) != 1 {
			return false, nil
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
	l.Debug("pod is ready")
	statefulset, err = K8sClient.AppsV1().StatefulSets(sr.Namespace).Get(context.Background(), sr.Name, metav1.GetOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().StatefulSets(sr.Namespace).Get error=%v", err)
		return err
	}
	// patch the StatefulSet with the new image
	for _, replacement := range sr.Replacements {
		for i, container := range statefulset.Spec.Template.Spec.Containers {
			if container.Name == replacement.Container {
				statefulset.Spec.Template.Spec.Containers[i].Image = replacement.Image
			}
		}
	}
	l.Debug("patching statefulset")
	// scale the StatefulSet back to the original replica count
	statefulset.Spec.Replicas = int32Ptr(origReplicas)
	_, err = K8sClient.AppsV1().StatefulSets(sr.Namespace).Update(context.Background(), statefulset, metav1.UpdateOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().StatefulSets(sr.Namespace).Update error=%v", err)
		return err
	}
	l.Debug("scaled statefulset back to original replica count, waiting for pods to be ready")
	// wait for all new pods to be ready
	time.Sleep(5 * time.Second)
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		selector := labels.Set(statefulset.Spec.Selector.MatchLabels)
		pods, err := K8sClient.CoreV1().Pods(sr.Namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: selector.AsSelector().String(),
		})
		if err != nil {
			return false, err
		}
		if len(pods.Items) != int(origReplicas) {
			return false, nil
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
	l.Debug("ReplaceStatefulSet complete")
	println("statefulset.apps/" + sr.Name + " replaced")
	return nil
}
