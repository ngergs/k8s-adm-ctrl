# Admission controller toolkit
 **This library toolkit uses generics. Therefore, go version 1.18+ is required.**

Library function to build an [admission controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).
The library functions provide some interfaces and helper structures to handle the IO part and e.g. the construction of the JSON patch for mutating controllers.
The target main use case are admission controllers that target a single resource.

The compiled binary is around 14MB and the example docker image is around 16MB.

## Usage
### Reviewer implementations
The library provides two core functions to produce a `ReviewerHandler` which implements `http.Handler`.
```go
func MutatingReviewer[T any](mutater ResourceMutater[T], compatibleGroupVersionKinds ...*metav1.GroupVersionKind) ReviewerHandler
func ValidatingReviewer[T any](validator ResourceValidator[T], compatibleGroupVersionKinds ...*metav1.GroupVersionKind) ReviewerHandler
```
To use them the user has to provide the implementation of the `ResourceMutater[T]` or `ResourceValidator[T]` function, respectively.
These two functions hold the core mutation/validation logic and are defined as:
```go
type ResourceMutater[T any] func(request *T) (*ValidateResult, *Patch[T])
type ResourceValidator[T any] func(request *T) *ValidateResult
```


### Example application
The [namespace admission controller](examples/namespace) is an example implementation of the ResourceMutater and ResourceValidator functions.

As the wrapping in the corresponding Review interface implementation also implements the `http.Handler` interface usage together with the http package is simple:
```go
mutater := &NamespaceLabelMutater{}
http.Handle("/mutate", admissionreview.MutatingReviewer(mutater.Patch, compatibleGroupVersionKind))
http.Handle("/validate", admissionreview.ValidatingReviewer(mutater.Validate, compatibleGroupVersionKind))
```

### Reviewer
The internal core interface. It is supposed to be called after the IO part of the HTTP admission review request (including unmarshalling)
has been handled. You might want to use this interface in special cases where the HTTP handling of the given `ValidatingReviewer`
and `MutatingReviewer` implementations do not suffice.
```go
type Reviewer interface {
	Review(*admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse
}
```

### Helm chart
To actually deploy the admission controller [helm example](examples/helm) provides a Helm chart with a more detailed Readme regarding the deployment.

### Dockerfile
The [example Dockerfile](Dockerfile_namespace_example) builds the application using the official golang alpine image and then copies the statically linked application binary into an distroless image which ends up at around ~16MB image size.
Building works via
```bash
docker build -f Dockerfile_namespace_example -t namespace-adm-ctrl .
```
If you want to test the example locally, you can then do so e.g. via (using [httpie](https://httpie.io/) for HTTP requests):
```bash
docker container run --rm -p 10250:10250 namespace-adm-ctrl
http POST localhost:10250/mutate < examples/namespace/testdata/request_invalid.jsom
```