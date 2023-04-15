# Admission controller toolkit
 **This library toolkit uses generics. Therefore, go version 1.18+ is required.**

Library function to build an [admission controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).
The library functions provide some interfaces and helper structures to handle the IO part and e.g. the construction of the JSON patch for mutating controllers.
The target main use case are admission controllers that target a single resource.

The compiled binary is around 14MB and the example docker image is around 16MB.

## Interfaces
### Reviewer
The core interface will usually not be used explicitly is supposed to be directly called after the IO part of the HTTP admission review request has been handled.
```go
type Reviewer interface {
	Review(*admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse
}
```

### Reviewer implementations
Two implementations of the `Reviewer` (and `http.Handler` interface) corresponding to the two admission controller types are provided via the generic functions:
```go
func MutatingReviewer[T any](mutater ResourceMutater[T]) ReviewerHandler
func ValidatingReviewer[T any](validator ResourceValidator[T]) ReviewerHandler
```
To use them the library user has to provide the implementation of the `ResourceMutater[T]` or `ResourceValidator[T]`, respectively.
These then hold the core mutation/validation logic:
```go
type ResourceMutater[T any] func(request *T) (*ValidateResult, *Patch[T])
type ResourceValidator[T any] func(request *T) *ValidateResult
```

## Example application
The [namespace admission controller](examples/namespace) is an example implementation of the ResourceMutater and ResourceValidator functions.

As the wrapping in the corresponding Review interface implementation also implements the `http.Handler` interface usage together with the http package is simple:
```go
mutater := &NamespaceLabelMutater{}
http.Handle("/mutate", admissionreview.MutatingReviewer(mutater.Patch, compatibleGroupVersionKind))
http.Handle("/validate", admissionreview.ValidatingReviewer(mutater.Validate, compatibleGroupVersionKind))
```

## Helm chart
To actually deploy the admission controller [helm example](examples/helm) provides a Helm chart with a more detailed Readme regarding the deployment.

## Dockerfile
The [example Dockerfile](examples/build/Dockerfile) builds the application using the official golang alpine image and then copies the statically linked application binary into an distroless image which ends up at around ~16MB image size.
