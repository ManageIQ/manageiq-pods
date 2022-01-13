module github.com/ManageIQ/manageiq-pods/manageiq-operator

go 1.14

require (
	github.com/operator-framework/operator-sdk v0.18.2
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	github.com/otiai10/copy => github.com/otiai10/copy v1.0.2
	github.com/otiai10/mint => github.com/otiai10/mint v1.3.0
	k8s.io/api => github.com/ManageIQ/kubernetes-api v0.0.0-20220110152537-707adf1e9ef5 // HACK: Temporary fork of the kubernetes-api to include some 0.19 changes until we can get to 0.19
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
