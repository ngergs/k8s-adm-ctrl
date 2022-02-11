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
This interface is supposed to be directly called after the IO part of the HTTP admission review request has been handled. The library provides a function that wraps a Reviewer interface into a HTTP handlefunc.
```go
func ToHandelFunc(reviewer Reviewer) func(w http.ResponseWriter, r *http.Request) 
```

### Reference implementations
The library provides two reference implementations of the Reviewer interface in the form of the named function types
```go
type ResourceMutater func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) (result *ValidateResult, patches *Patches)
type ResourceValidator func(requestGroupVersionKind *metav1.GroupVersionKind, rawRequest []byte) *ValidateResult
```
and corresponding functions to wrap these into a struct that implements the Reviewer interface:
```go
func MutatingReviewer(mutater ResourceMutater) Reviewer
func ValidatingReviewer(validator ResourceValidator) Reviewer
```
These wrapper functions implement again some generalized logic around the core Mutation/Validation business logic which is encased ihe provided function arguments. Especially the MutatingReviewer wrapper handles the construction of the JSON patch from the ResourceMutater response.

## Example application
In the namespace_webhook.go is an example implementation of the ResourceMutater and ResourceValidator functions which hold the core business logic.

The wrapping in the corresponding Review interface implementations and furthermore wrapping to HTTP handles can then be easily setup as demonstrated in the main.go:
```go
http.HandleFunc("/mutate", admissionreview.ToHandlFunc(admissionreview.MutatingReviewer(mutater.Patch)))
http.HandleFunc("/validate", admissionreview.ToHandlFunc(admissionreview.ValidatingReviewer(mutater.Validate)))
```

## Dockerfile
The provided Dockerfile builds the application using the official golang alpine image and then wraps the application into an alpine Linux server image which ends up at around ~20MB image size.
