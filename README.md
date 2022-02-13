# Admission controller toolkit
Some library function to build an [admission controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).
 The library functions provide some interfaces and helper structures to handle the IO part and e.g. the construction of the JSON patch for mutating controllers.

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
Two implementations of the reviewer (and http.Handler interface) corresponding to the two admission controller types are provided via the functions:
```go
func MutatingReviewer(mutater ResourceMutater) ReviewerHandler
func ValidatingReviewer(validator ResourceValidator) ReviewerHandler
```
To use them the library user has to provide the implementation of the corresponding named function types, these hold the core mutation/validation logic:
```go
type ResourceMutater func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) (*ValidateResult, *Patch)
type ResourceValidator func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *ValidateResult
```

## Example application
The [namespace_webhook.go](https://github.com/ngergs/k8s-admission-ctrl/blob/main/namespace_webhook.go) is an example implementation of the ResourceMutater and ResourceValidator functions.

As the wrapping in the corresponding Review interface implementation also implements the http.Handler interface usage together with the http package is simple:
```go
mutater := &NamespaceLabelMutater{}
http.Handle("/mutate", admissionreview.MutatingReviewer(mutater.Patch))
http.Handle("/validate", admissionreview.ValidatingReviewer(mutater.Validate))
```

## Dockerfile
The provided Dockerfile builds the application using the official golang alpine image and then wraps the application into an alpine Linux server image which ends up at around ~20MB image size.
