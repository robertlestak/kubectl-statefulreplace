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

func (sr *StatefulReplace) ReplaceDeployment() error {
	l := log.WithFields(
		log.Fields{
			"action": "ReplaceDeployment",
		},
	)
	l.Debug("ReplaceDeployment")
	if sr.Kind != "Deployment" {
		return fmt.Errorf("kind is not Deployment")
	}
	deployment, err := K8sClient.AppsV1().Deployments(sr.Namespace).Get(context.Background(), sr.Name, metav1.GetOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().Deployments(sr.Namespace).Get error=%v", err)
		return err
	}
	l.Debug("got deployment, scaling to 1 replica")
	var origReplicas int32
	if deployment.Spec.Replicas != nil {
		origReplicas = *deployment.Spec.Replicas
	}
	// scale the Deployment to 1 replica
	deployment.Spec.Replicas = int32Ptr(1)
	_, err = K8sClient.AppsV1().Deployments(sr.Namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().Deployments(sr.Namespace).Update error=%v", err)
		return err
	}
	l.Debug("scaled deployment to 1 replica, waiting for pod to be ready")
	// wait for all old pods to be terminated
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		selector := labels.Set(deployment.Spec.Selector.MatchLabels)
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
		if pods.Items[0].Status.Phase != "Running" {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		l.Errorf("wait.PollImmediate error=%v", err)
		return err
	}
	l.Debug("pod is ready, patching deployment")
	deployment, err = K8sClient.AppsV1().Deployments(sr.Namespace).Get(context.Background(), sr.Name, metav1.GetOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().Deployments(sr.Namespace).Get error=%v", err)
		return err
	}
	// patch the Deployment with the new image
	for _, replacement := range sr.Replacements {
		for i, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == replacement.Container {
				deployment.Spec.Template.Spec.Containers[i].Image = replacement.Image
			}
		}
	}
	// scale the Deployment back to the original replica count
	deployment.Spec.Replicas = int32Ptr(origReplicas)
	_, err = K8sClient.AppsV1().Deployments(sr.Namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
	if err != nil {
		l.Errorf("K8sClient.AppsV1().Deployments(sr.Namespace).Update error=%v", err)
		return err
	}
	l.Debug("patched deployment, waiting for new pods to be ready")
	// wait for all new pods to be ready
	time.Sleep(5 * time.Second)
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		selector := labels.Set(deployment.Spec.Selector.MatchLabels)
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
	l.Debug("new pods are ready, done")
	println("deployment.apps/" + sr.Name + " replaced")
	return nil
}
