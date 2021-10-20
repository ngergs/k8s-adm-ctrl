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
This interface is supposed to be directly called after the IO part of the HTTP admission review request has been handled. The library provides a function that wraps a Reviewer interface into a HTTP handler.
```go
func ToHandler(reviewer Reviewer) func(w http.ResponseWriter, r *http.Request)
```

### Reference implementations
The library provides two reference implementations of the Reviewer interface:
```go
type MutatingReviewer struct {
	Mutater ResourceMutater
}
type ValidatingReviewer struct {
	Validator ResourceValidator
}
```
These wrap again some generalized logic around the core Mutation/Validation business logic which is encased in the ResourceMutater and ResourceValidator interface. Especially the MutatingReviewer handles the construction of the JSON patch from the ResourceMutater response.

## Example application
In the namespace_webhook.go an example implementation of the ResourceMutater and ResourceValidator interface which hold the core business logic.

The wrapping in the corresponding Review interface implementations and furthermore wrapping to HTTP handles can then be easily setup as demonstrated in the main.go.
```go
http.HandleFunc("/mutate", admissionreview.ToHandler(
	&admissionreview.MutatingReviewer{
		Mutater: &NamespaceLabelMutater{},
	}))
http.HandleFunc("/validate", admissionreview.ToHandler(
	&admissionreview.ValidatingReviewer{
		Validator: &NamespaceLabelMutater{},
	}))
```

## Tests
The tests use generated code from mockgen, to run them execute:
```bash
go generate ./... &&\
go test ./...
```

## Dockerfile
The provided Dockerfile builds the application using the official golang alpine image and then wraps the application into an alpine Linux server image which ends up at around ~20MB image size.
