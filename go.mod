module github.com/crossplaneio/easy-gcp

go 1.12

replace (
	github.com/crossplaneio/crossplane => github.com/muvaf/crossplane v0.2.1-0.20191206114926-904a6e4162f2
	github.com/crossplaneio/crossplane-runtime => github.com/muvaf/crossplane-runtime v0.0.0-20191206111112-1fe048748dc5
	github.com/crossplaneio/stack-gcp => github.com/muvaf/stack-gcp v0.0.0-20191206115400-ca4f6c982d88
)

require (
	github.com/crossplaneio/crossplane-runtime v0.2.3
	github.com/crossplaneio/stack-gcp v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize/api v0.2.0
)
