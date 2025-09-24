package miqkafka

import (
	"bytes"
	"context"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	miqtool "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/helpers/miq-components"
	miqutilsv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/miqutils"
	olmv1 "github.com/operator-framework/api/pkg/operators/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
)

func KafkaCASecret(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme, secretType string) (*corev1.Secret, controllerutil.MutateFn) {
	caSecret := miqtool.InternalCertificatesSecret(cr, client)
	secret := &corev1.Secret{}

	if secretType == "cert" {
		secret.ObjectMeta = metav1.ObjectMeta{Name: cr.Spec.AppName + "-cluster-ca-cert", Namespace: cr.Namespace}
	} else {
		secret.ObjectMeta = metav1.ObjectMeta{Name: cr.Spec.AppName + "-cluster-ca", Namespace: cr.Namespace}
	}

	if secret.ObjectMeta.Annotations == nil {
		secret.ObjectMeta.Annotations = make(map[string]string)
	}

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}

	secretLabels := map[string]string{
		"app.kubernetes.io/instance":   "manageiq",
		"app.kubernetes.io/managed-by": "strimzi-cluster-operator",
		"strimzi.io/cluster":           "manageiq",
		"strimzi.io/component-type":    "certificate-authority",
		"strimzi.io/kind":              "Kafka",
		"strimzi.io/name":              "strimzi",
	}

	mutateFunc := func() error {
		if err := controllerutil.SetControllerReference(cr, secret, scheme); err != nil {
			return err
		}

		miqtool.AddLabels(secretLabels, &secret.ObjectMeta)

		caGen := 0
		lastKnownRevisionStr := secret.ObjectMeta.Annotations["last-known-revision"]
		if lastKnownRevisionStr == "null" || lastKnownRevisionStr == "" {
			lastKnownRevisionStr = "0"
			secret.ObjectMeta.Annotations["last-known-revision"] = lastKnownRevisionStr
		}
		caGen, err := strconv.Atoi(lastKnownRevisionStr)
		if err != nil {
			return err
		}

		if secretType == "cert" {
			secret.ObjectMeta.Annotations["strimzi.io/ca-cert-generation"] = strconv.Itoa(caGen)
			secret.Data["ca.crt"] = caSecret.Data["root_crt"]
		} else {
			secret.ObjectMeta.Annotations["strimzi.io/ca-key-generation"] = strconv.Itoa(caGen)
			secret.Data["ca.key"] = caSecret.Data["root_key"]
		}

		return nil
	}

	return secret, mutateFunc
}

func renewKafkaCASecretCheck(cr *miqv1alpha1.ManageIQ, client client.Client) bool {
	certSecret := miqtool.InternalCertificatesSecret(cr, client)
	kafkaCASecret := miqutilsv1alpha1.FindSecretByName(client, cr.Namespace, cr.Spec.AppName+"-cluster-ca-cert")
	kafkaCAKeySecret := miqutilsv1alpha1.FindSecretByName(client, cr.Namespace, cr.Spec.AppName+"-cluster-ca")
	if certSecret.Data["root_crt"] == nil || certSecret.Data["root_key"] == nil || kafkaCASecret.Data["ca.crt"] == nil || kafkaCAKeySecret.Data["ca.key"] == nil {
		return false
	}

	if bytes.Equal(certSecret.Data["root_crt"], kafkaCASecret.Data["ca.crt"]) || bytes.Equal(certSecret.Data["root_key"], kafkaCAKeySecret.Data["ca.key"]) {
		return false
	}

	return true
}

func updateKafka(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme, update func(*unstructured.Unstructured) *unstructured.Unstructured) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		kafka := miqutilsv1alpha1.FindKafka(client, scheme, cr.Namespace, cr.Spec.AppName)
		kafka = update(kafka)

		err := client.Update(context.TODO(), kafka)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// If manual certificates are introduced after Strimzi has generated its own certificates,
// then the certificates must be "renewed" to replace the old. This process is outlined here:
// https://strimzi.io/docs/operators/in-development/deploying#proc-replacing-your-own-private-keys-str
func renewKafkaCASecret(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) error {
	if renewKafkaCASecretCheck(cr, client) {
		kafkaCASecret := miqutilsv1alpha1.FindSecretByName(client, cr.Namespace, cr.Spec.AppName+"-cluster-ca-cert")
		kafkaCAKeySecret := miqutilsv1alpha1.FindSecretByName(client, cr.Namespace, cr.Spec.AppName+"-cluster-ca")
		certSecret := miqtool.InternalCertificatesSecret(cr, client)

		pauseReconcile := func(kafka *unstructured.Unstructured) *unstructured.Unstructured {
			kafka.SetAnnotations(map[string]string{"strimzi.io/pause-reconciliation": "true"})
			return kafka
		}
		updateKafka(cr, client, scheme, pauseReconcile)

		useMiqCA := func(kafka *unstructured.Unstructured) *unstructured.Unstructured {
			kafkaCR := kafka.UnstructuredContent()["spec"].(map[string]interface{})
			kafkaCR["clusterCa"] = map[string]interface{}{
				"generateCertificateAuthority": false,
			}

			return kafka
		}
		updateKafka(cr, client, scheme, useMiqCA)

		lastKnownRevisionStr := kafkaCASecret.ObjectMeta.Annotations["last-known-revision"]
		if lastKnownRevisionStr == "null" || lastKnownRevisionStr == "" {
			lastKnownRevisionStr = "0"
			kafkaCASecret.ObjectMeta.Annotations["last-known-revision"] = lastKnownRevisionStr
			kafkaCAKeySecret.ObjectMeta.Annotations["last-known-revision"] = lastKnownRevisionStr
		}

		caGen, err := strconv.Atoi(lastKnownRevisionStr)
		if err != nil {
			return err
		}
		caGen++

		caGenStr := strconv.Itoa(caGen)
		kafkaCASecret.ObjectMeta.Annotations["last-known-revision"] = caGenStr
		kafkaCAKeySecret.ObjectMeta.Annotations["last-known-revision"] = caGenStr

		kafkaCASecret.ObjectMeta.Annotations["strimzi.io/ca-cert-generation"] = caGenStr
		oldCA := kafkaCASecret.Data["ca.crt"]
		kafkaCASecret.Data = map[string][]byte{"ca.crt": certSecret.Data["root_crt"], "ca-old.crt": oldCA}
		if err := client.Update(context.TODO(), kafkaCASecret); err != nil {
			return err
		}

		kafkaCAKeySecret.ObjectMeta.Annotations["strimzi.io/ca-key-generation"] = caGenStr
		kafkaCAKeySecret.Data = map[string][]byte{"ca.key": certSecret.Data["root_key"]}
		if err := client.Update(context.TODO(), kafkaCAKeySecret); err != nil {
			return err
		}

		pauseReconcile = func(kafka *unstructured.Unstructured) *unstructured.Unstructured {
			kafka.SetAnnotations(map[string]string{"strimzi.io/pause-reconciliation": "false"})
			return kafka
		}
		updateKafka(cr, client, scheme, pauseReconcile)
	}

	return nil
}

func MessagingEnvSecret(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*corev1.Secret, controllerutil.MutateFn) {
	secretData := make(map[string]string)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "messaging-env-secret",
			Namespace: cr.Namespace,
		},
		StringData: secretData,
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, secret, scheme); err != nil {
			return err
		}

		miqtool.AddLabels(map[string]string{"app": cr.Spec.AppName}, &secret.ObjectMeta)

		secretData["hostname"] = cr.Spec.AppName + "-kafka-bootstrap"
		secretData["username"] = cr.Spec.AppName + "-kafka-admin"
		secretData["port"] = "9093"
		secretData["sasl_mechanism"] = "SCRAM-SHA-512"

		secret.StringData = secretData

		return nil
	}

	return secret, f
}

func KafkaClusterSpec() map[string]interface{} {
	return map[string]interface{}{
		"kafka": map[string]interface{}{
			"replicas": 1,
			"listeners": []map[string]interface{}{
				map[string]interface{}{
					"name": "kafka",
					"port": 9093,
					"type": "internal",
					"tls":  true,
					"authentication": map[string]interface{}{
						"type": "scram-sha-512",
					},
				},
			},
			"config": map[string]interface{}{
				"offsets.topic.replication.factor":         1,
				"transaction.state.log.replication.factor": 1,
				"transaction.state.log.min.isr":            1,
				"default.replication.factor":               1,
				"min.insync.replicas":                      1,
			},
			"template": map[string]interface{}{
				"pod": map[string]interface{}{
					"securityContext": map[string]interface{}{
						"runAsNonRoot": true,
					},
				},
				"kafkaContainer": map[string]interface{}{
					"securityContext": map[string]interface{}{
						"allowPrivilegeEscalation": false,
						"capabilities": map[string]interface{}{
							"drop": []string{"ALL"},
						},
						"privileged":             false,
						"readOnlyRootFilesystem": false,
						"runAsNonRoot":           true,
					},
				},
			},
			"storage": map[string]interface{}{
				"type":        "persistent-claim",
				"deleteClaim": true,
			},
			"authorization": map[string]interface{}{
				"type": "simple",
			},
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{},
				"limits":   map[string]interface{}{},
			},
		},
		"zookeeper": map[string]interface{}{
			"replicas": 1,
			"template": map[string]interface{}{
				"pod": map[string]interface{}{
					"securityContext": map[string]interface{}{
						"runAsNonRoot": true,
					},
				},
				"zookeeperContainer": map[string]interface{}{
					"securityContext": map[string]interface{}{
						"allowPrivilegeEscalation": false,
						"capabilities": map[string]interface{}{
							"drop": []string{"ALL"},
						},
						"privileged":             false,
						"readOnlyRootFilesystem": false,
						"runAsNonRoot":           true,
					},
				},
			},
			"storage": map[string]interface{}{
				"type":        "persistent-claim",
				"deleteClaim": true,
			},
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{},
				"limits":   map[string]interface{}{},
			},
		},
		"entityOperator": map[string]interface{}{
			"template": map[string]interface{}{
				"pod": map[string]interface{}{
					"securityContext": map[string]interface{}{
						"runAsNonRoot": true,
					},
				},
				"topicOperatorContainer": map[string]interface{}{
					"securityContext": map[string]interface{}{
						"allowPrivilegeEscalation": false,
						"capabilities": map[string]interface{}{
							"drop": []string{"ALL"},
						},
						"privileged":             false,
						"readOnlyRootFilesystem": false,
						"runAsNonRoot":           true,
					},
				},
				"userOperatorContainer": map[string]interface{}{
					"securityContext": map[string]interface{}{
						"allowPrivilegeEscalation": false,
						"capabilities": map[string]interface{}{
							"drop": []string{"ALL"},
						},
						"privileged":             false,
						"readOnlyRootFilesystem": false,
						"runAsNonRoot":           true,
					},
				},
				"tlsSidecarContainer": map[string]interface{}{
					"securityContext": map[string]interface{}{
						"allowPrivilegeEscalation": false,
						"capabilities": map[string]interface{}{
							"drop": []string{"ALL"},
						},
						"privileged":             false,
						"readOnlyRootFilesystem": false,
						"runAsNonRoot":           true,
					},
				},
			},
			"tlsSidecar": map[string]interface{}{
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    "500m",
						"memory": "128Mi",
					},
					"limits": map[string]interface{}{
						"cpu":    "500m",
						"memory": "128Mi",
					},
				},
			},
			"userOperator": map[string]interface{}{
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    1,
						"memory": "1Gi",
					},
					"limits": map[string]interface{}{
						"cpu":    1,
						"memory": "1Gi",
					},
				},
			},
			"topicOperator": map[string]interface{}{
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    1,
						"memory": "1Gi",
					},
					"limits": map[string]interface{}{
						"cpu":    1,
						"memory": "1Gi",
					},
				},
			},
		},
	}
}

func KafkaCluster(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*unstructured.Unstructured, controllerutil.MutateFn) {
	kafkaClusterCR := &unstructured.Unstructured{}
	kafkaClusterCR.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Kind:    "Kafka",
		Version: "v1beta2",
	})
	kafkaClusterCR.SetName(cr.Spec.AppName)
	kafkaClusterCR.SetNamespace(cr.Namespace)

	kafkaCRSpec := KafkaClusterSpec()

	if cr.Spec.StorageClassName != "" {
		kafkaStorage := kafkaCRSpec["kafka"].(map[string]interface{})["storage"].(map[string]interface{})
		kafkaStorage["class"] = cr.Spec.StorageClassName
		zookeeperStorage := kafkaCRSpec["zookeeper"].(map[string]interface{})["storage"].(map[string]interface{})
		zookeeperStorage["class"] = cr.Spec.StorageClassName
	}

	kafkaCRSpec = miqutilsv1alpha1.SetKafkaNodeAffinity(kafkaCRSpec, []string{"amd64", "arm64", "ppc64le", "s390x"})

	mutateFunc := func() error {
		if err := controllerutil.SetControllerReference(cr, kafkaClusterCR, scheme); err != nil {
			return err
		}

		if *cr.Spec.EnforceWorkerResourceConstraints == true {
			kafkaResourceRequests := kafkaCRSpec["kafka"].(map[string]interface{})["resources"].(map[string]interface{})["requests"].(map[string]interface{})
			kafkaResourceRequests["memory"] = cr.Spec.KafkaMemoryRequest
			kafkaResourceRequests["cpu"] = cr.Spec.KafkaCpuRequest
			kafkaResourceLimits := kafkaCRSpec["kafka"].(map[string]interface{})["resources"].(map[string]interface{})["limits"].(map[string]interface{})
			kafkaResourceLimits["memory"] = cr.Spec.KafkaMemoryLimit
			kafkaResourceLimits["cpu"] = cr.Spec.KafkaCpuLimit

			zookeeperResourceRequests := kafkaCRSpec["zookeeper"].(map[string]interface{})["resources"].(map[string]interface{})["requests"].(map[string]interface{})
			zookeeperResourceRequests["memory"] = cr.Spec.ZookeeperMemoryRequest
			zookeeperResourceRequests["cpu"] = cr.Spec.ZookeeperCpuRequest
			zookeeperResourceLimits := kafkaCRSpec["zookeeper"].(map[string]interface{})["resources"].(map[string]interface{})["limits"].(map[string]interface{})
			zookeeperResourceLimits["memory"] = cr.Spec.ZookeeperMemoryLimit
			zookeeperResourceLimits["cpu"] = cr.Spec.ZookeeperCpuLimit
		}

		if certSecret := miqtool.InternalCertificatesSecret(cr, client); certSecret.Data["root_crt"] != nil && certSecret.Data["root_key"] != nil {
			if err := renewKafkaCASecret(cr, client, scheme); err != nil {
				return err
			} else {
				kafkaCRSpec["clusterCa"] = map[string]interface{}{
					"generateCertificateAuthority": false,
				}
			}
		}

		kafkaStorage := kafkaCRSpec["kafka"].(map[string]interface{})["storage"].(map[string]interface{})
		kafkaStorage["size"] = cr.Spec.KafkaVolumeCapacity

		zookeeperStorage := kafkaCRSpec["zookeeper"].(map[string]interface{})["storage"].(map[string]interface{})
		zookeeperStorage["size"] = cr.Spec.ZookeeperVolumeCapacity

		kafkaClusterCR.UnstructuredContent()["spec"] = kafkaCRSpec

		return nil
	}

	return kafkaClusterCR, mutateFunc
}

func KafkaUserSpec() map[string]interface{} {
	return map[string]interface{}{
		"authentication": map[string]interface{}{
			"type": "scram-sha-512",
		},
		"authorization": map[string]interface{}{
			"type": "simple",
			"acls": []map[string]interface{}{
				map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "topic",
						"name":        "*",
						"patternType": "literal",
					},
					"operations": []string{"All"},
					"host":       "*",
				},
				map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "group",
						"name":        "*",
						"patternType": "literal",
					},
					"operations": []string{"All"},
					"host":       "*",
				},
			},
		},
	}
}

func KafkaUser(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*unstructured.Unstructured, controllerutil.MutateFn) {
	kafkaUserCR := &unstructured.Unstructured{}

	kafkaUserCR.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Kind:    "KafkaUser",
		Version: "v1beta2",
	})
	kafkaUserCR.SetName(cr.Spec.AppName + "-kafka-admin")
	kafkaUserCR.SetNamespace(cr.Namespace)
	kafkaUserCR.SetLabels(map[string]string{"strimzi.io/cluster": cr.Spec.AppName})

	kafkaUserSpec := KafkaUserSpec()

	mutateFunc := func() error {
		if err := controllerutil.SetControllerReference(cr, kafkaUserCR, scheme); err != nil {
			return err
		}

		kafkaUserCR.UnstructuredContent()["spec"] = kafkaUserSpec

		return nil
	}

	return kafkaUserCR, mutateFunc
}

func KafkaTopicSpec() map[string]interface{} {
	return map[string]interface{}{
		"partitions": 1,
		"config": map[string]interface{}{
			"retention.ms":  7200000,
			"segment.bytes": 1073741824,
		},
	}
}

func KafkaTopic(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme, topicName string) (*unstructured.Unstructured, controllerutil.MutateFn) {
	kafkaTopicCR := &unstructured.Unstructured{}

	kafkaTopicCR.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Kind:    "KafkaTopic",
		Version: "v1beta2",
	})
	kafkaTopicCR.SetName(topicName)
	kafkaTopicCR.SetNamespace(cr.Namespace)
	kafkaTopicCR.SetLabels(map[string]string{"strimzi.io/cluster": cr.Spec.AppName})

	kafkaTopicSpec := KafkaTopicSpec()

	mutateFunc := func() error {
		if err := controllerutil.SetControllerReference(cr, kafkaTopicCR, scheme); err != nil {
			return err
		}

		kafkaTopicCR.UnstructuredContent()["spec"] = kafkaTopicSpec

		return nil
	}

	return kafkaTopicCR, mutateFunc
}

func KafkaInstall(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*olmv1alpha1.Subscription, controllerutil.MutateFn) {
	kafkaSubscription := &olmv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "strimzi-kafka-operator",
			Namespace: cr.Namespace,
		},
	}

	mutateFunc := func() error {
		if err := controllerutil.SetControllerReference(cr, kafkaSubscription, scheme); err != nil {
			return err
		}

		kafkaSubscription.Spec = &olmv1alpha1.SubscriptionSpec{
			CatalogSource:          "community-operators",
			CatalogSourceNamespace: "openshift-marketplace",
			Package:                "strimzi-kafka-operator",
			Channel:                "strimzi-0.37.x",
		}

		return nil
	}

	return kafkaSubscription, mutateFunc
}

func KafkaOperatorGroup(cr *miqv1alpha1.ManageIQ, scheme *runtime.Scheme) (*olmv1.OperatorGroup, controllerutil.MutateFn) {
	kafkaOperatorGroup := &olmv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.AppName + "-group",
			Namespace: cr.Namespace,
		},
	}

	mutateFunc := func() error {
		if err := controllerutil.SetControllerReference(cr, kafkaOperatorGroup, scheme); err != nil {
			return err
		}

		kafkaOperatorGroup.Spec = olmv1.OperatorGroupSpec{
			TargetNamespaces: []string{cr.Namespace},
		}

		return nil
	}

	return kafkaOperatorGroup, mutateFunc
}
