module github.com/crossplaneio/easy-gcp

go 1.12

replace github.com/crossplaneio/crossplane-runtime => github.com/muvaf/crossplane-runtime v0.0.0-20191205164816-48eab080955d

require (
	github.com/crossplaneio/crossplane-runtime v0.2.3
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize/api v0.2.0
)
