module github.com/example/test-operator

go 1.16

// require (
// 	github.com/gorilla/mux v1.8.0
// 	github.com/onsi/ginkgo v1.16.4
// 	github.com/onsi/gomega v1.13.0
// 	k8s.io/api v0.22.2
// 	k8s.io/apimachinery v0.22.2
// 	k8s.io/client-go v0.22.2
// 	sigs.k8s.io/controller-runtime v0.9.2
// )
require (
	github.com/gorilla/mux v1.8.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	golang.org/x/tools v0.1.7 // indirect
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.2
)
