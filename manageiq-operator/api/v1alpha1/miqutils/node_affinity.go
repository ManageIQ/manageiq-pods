package miqutils

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func OperatorNodeAffinityArchValues(deployment *appsv1.Deployment, client client.Client) []string {
	podName := os.Getenv("POD_NAME")
	pod := FindPodByName(client, deployment.ObjectMeta.Namespace, podName)
	values := []string{"amd64"}

	if pod.Spec.Affinity == nil {
		// In case we don't find the operator pod (local testing) or it doesn't have affinities
		return values
	}

	nodeSelectorTerms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms

	for _, selector := range nodeSelectorTerms {
		for _, matchExpression := range selector.MatchExpressions {
			if matchExpression.Key == "kubernetes.io/arch" {
				values = matchExpression.Values
			}
		}
	}

	return values
}

func SetDeploymentNodeAffinity(deployment *appsv1.Deployment, client client.Client) {
	operatorNodeAffinityArchValues := OperatorNodeAffinityArchValues(deployment, client)
	if len(operatorNodeAffinityArchValues) == 0 {
		// We're running local, can't find the operator pod, or it doesn't have any affinities to use as a template.  Skip it.
		return
	}

	matchExpression := corev1.NodeSelectorRequirement{
		Key:      "kubernetes.io/arch",
		Operator: corev1.NodeSelectorOpIn,
		Values:   operatorNodeAffinityArchValues,
	}

	matchExpressions := []corev1.NodeSelectorRequirement{matchExpression}

	nodeSelectorTerm := corev1.NodeSelectorTerm{
		MatchExpressions: matchExpressions,
	}

	nodeSelectionTerms := []corev1.NodeSelectorTerm{nodeSelectorTerm}

	deployment.Spec.Template.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: nodeSelectionTerms,
			},
		},
	}
}

func SetKafkaNodeAffinity(kafkaCRSpec map[string]interface{}, archs []string) map[string]interface{} {
	nodeAffinity := map[string]interface{}{
		"nodeAffinity": map[string]interface{}{
			"requiredDuringSchedulingIgnoredDuringExecution": map[string]interface{}{
				"nodeSelectorTerms": []map[string]interface{}{
					map[string]interface{}{
						"matchExpressions": []map[string]interface{}{
							map[string]interface{}{
								"key":      "kubernetes.io/arch",
								"operator": "In",
								"values":   archs,
							},
						},
					},
				},
			},
		},
	}

	kafkaPod := kafkaCRSpec["kafka"].(map[string]interface{})["template"].(map[string]interface{})["pod"].(map[string]interface{})
	kafkaPod["affinity"] = nodeAffinity
	zookeeperPod := kafkaCRSpec["zookeeper"].(map[string]interface{})["template"].(map[string]interface{})["pod"].(map[string]interface{})
	zookeeperPod["affinity"] = nodeAffinity
	operatorEntityPod := kafkaCRSpec["entityOperator"].(map[string]interface{})["template"].(map[string]interface{})["pod"].(map[string]interface{})
	operatorEntityPod["affinity"] = nodeAffinity

	return kafkaCRSpec
}
