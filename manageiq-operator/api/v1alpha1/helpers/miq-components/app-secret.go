package miqtools

import (
	"context"
	"fmt"
	"net/url"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const appSecretName = "app-secrets"

func ManageAppSecret(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*corev1.Secret, controllerutil.MutateFn) {
	secretKey := types.NamespacedName{Namespace: cr.ObjectMeta.Namespace, Name: appSecretName}
	secret := &corev1.Secret{}
	secretErr := client.Get(context.TODO(), secretKey, secret)
	if secretErr != nil {
		secret = defaultAppSecret(cr)
	}

	f := func() error {
		if err := controllerutil.SetControllerReference(cr, secret, scheme); err != nil {
			return err
		}

		addAppLabel(cr.Spec.AppName, &secret.ObjectMeta)
		addBackupLabel(cr.Spec.BackupLabelName, &secret.ObjectMeta)

		encryptionKey := string(secret.Data["encryption-key"])
		d := map[string]string{
			"encryption-key": encryptionKey,
			"v2_key":         v2Key(encryptionKey),
		}

		// postgresql://root:98f43f0145d7e283@postgresql:5432/vmdb_production?encoding=utf8&pool=5&wait_timeout=5&sslmode=prefer
		databaseUrl := "postgresql://"
		postgresqlSecretKey := types.NamespacedName{Namespace: cr.ObjectMeta.Namespace, Name: cr.Spec.DatabaseSecret}
		postgresqlSecret := &corev1.Secret{}
		postgresqlSecretErr := client.Get(context.TODO(), postgresqlSecretKey, postgresqlSecret)
		if postgresqlSecretErr == nil {
			sslMode := "prefer"
			if postgresqlSecret.Data["sslmode"] != nil {
				sslMode = string(postgresqlSecret.Data["sslmode"])
			}
			databaseUrl += url.QueryEscape(string(postgresqlSecret.Data["username"])) + ":" + url.QueryEscape(string(postgresqlSecret.Data["password"]))
			databaseUrl += "@" + string(postgresqlSecret.Data["hostname"]) + ":" + string(postgresqlSecret.Data["port"]) + "/" + string(postgresqlSecret.Data["dbname"])
			databaseUrl += "?" + "encoding=utf8&pool=5&wait_timeout=5" + "&sslmode=" + sslMode
			if postgresqlSecret.Data["rootcertificate"] != nil {
				databaseUrl += "&" + "sslrootcert=" + "/.postgresql/root.crt"
			}

			d["database_yml"] = databaseYaml(databaseUrl)
		}

		secret.StringData = d

		return nil
	}

	return secret, f
}

func defaultAppSecret(cr *miqv1alpha1.ManageIQ) *corev1.Secret {
	secretData := map[string]string{
		"encryption-key": generateEncryptionKey(),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appSecretName,
			Namespace: cr.ObjectMeta.Namespace,
		},
		StringData: secretData,
	}

	return secret
}

func databaseYaml(databaseUrl string) string {
	s := `---
production:
  url: %[1]s
`
	return fmt.Sprintf(s, databaseUrl)
}

func v2Key(encryptionKey string) string {
	s := `---
:algorithm: aes-256-cbc
:key: %[1]s
`
	return fmt.Sprintf(s, encryptionKey)
}
