# Admission controller helm chart
This helm chart setups all the details required to deploy a [admission controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) and especially the corresponding webhook. 

There is no tight-coupling between this helm chart and the go webhook implementation also provided in this repository. An alternative implementation only has to follow the API as outlined by Kubernetes and use the TLS certificates provided by the Helm chart as Kubernetes secret.

As issuer for the self-signed TLS certificates [cert-manager](https://cert-manager.io/) is used. It can be installed via:
```
helm install cert-manager jetstack/cert-manager -f values-cert-manager.yml
```


## What is deployed
* A self-signed TLS certificate as root certificate for the self-signed certificate authority
* A self-signed TLS certificate issuer that uses the aforementioned certificate
* A reference service account for the webhooks.
* Per configured admission controller:
  * Self-signed TLS certificate
  * Deployment that hosts the webhook service and mounts the self-signed-certificate at /etc/certs
  * Service that loadbalances the webhook calls
  * Mutating and validating webhook configurations according to the specified rules. The self-signed certificate issuer is auto-configured as certificate authority for these webhooks.

## Values
See values.yml for an on-hands example.
* admissionControllers: (list)
  * name: reference name for the admission controller
  * imageName: container image name that should be used for the webhook implementation
  * port: container port that receives the webhook calls. defaults to 8080.
  * admissionReviewVersions: which [versions of AdmissionReview](https://pkg.go.dev/k8s.io/api/admission) are supported. MatchPolicy is fixed to Equivalent.
  * mutating: (list)
    * name: reference name for the webhook
    * path: optional HTTP path that should be used when calling the webhook
    * rules: list of [RulesWithOperations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#rulewithoperations-v1-admissionregistration-k8s-io)
  * validating: (list)
    * name: reference name for the webhook
    * path: optional HTTP path that should be used when calling the webhook
    * rules: list of [RulesWithOperations](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#rulewithoperations-v1-admissionregistration-k8s-io)