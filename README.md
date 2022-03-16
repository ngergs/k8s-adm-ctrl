# Admission controller toolkit
 **This library toolkit uses generics. Therefore, go version 1.18+ is required.**

Library function to build an [admission controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).
 The library functions provide some interfaces and helper structures to handle the IO part and e.g. the construction of the JSON patch for mutating controllers.

The compiled binary is around 5MB and the docker image ngergs/webserver is around 7.5MB.

## Helm chart
To actually deploy the admission controller the subfolder helm provides a Helm chart with a more detailed Readme regarding the deployment.

## Interfaces
### Reviewer
```go
type Reviewer interface {
	Review(*admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse
}
```
This interface is supposed to be directly called after the IO part of the HTTP admission review request has been handled.
The library provides two functions (see next section) that already handles this. For normal usage it is not expected to directly implement this interface.

### Reviewer implementations
Two implementations of the reviewer (and http.Handler interface) corresponding to the two admission controller types are provided via the generic functions:
```go
func MutatingReviewer[T any](mutater ResourceMutater[T]) ReviewerHandler
func ValidatingReviewer[T any](validator ResourceValidator[T]) ReviewerHandler
```
To use them the library user has to provide the implementation of the corresponding named function types, these hold the core mutation/validation logic:
```go
type ResourceMutater[T any] func(request *T) (*ValidateResult, *Patch[T])
type ResourceValidator[T any] func(request *T) *ValidateResult
```

## Example application
The [namespace_webhook.go](namespace_webhook.go) is an example implementation of the ResourceMutater and ResourceValidator functions.

As the wrapping in the corresponding Review interface implementation also implements the http.Handler interface usage together with the http package is simple:
```go
mutater := &NamespaceLabelMutater{}
http.Handle("/mutate", admissionreview.MutatingReviewer(mutater.Patch, compatibleGroupVersionKind))
http.Handle("/validate", admissionreview.ValidatingReviewer(mutater.Validate, compatibleGroupVersionKind))
```

## Dockerfile
The provided Dockerfile builds the application using the official golang alpine image and then wraps the application into an alpine Linux server image which ends up at around ~20MB image size.
