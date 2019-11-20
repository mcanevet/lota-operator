LOTA : The Lord Of The APIs
===========================

"One API to rule them all, One API to find them,\
One API to bring them all and in the brightness bind them."

What is this?
-------------

A generic Kubernetes Operator that can create external resources using Terraform.

Why this?
---------

There is more and more Kubernetes Operators that allows to manage objects in external APIs. This is something Terraform is good at. So instead of creating a specific Operator for an API to manage, why not use Terraform and benefit from the thousands resources it can manage?

How does it works?
------------------

A CRD that allows to declare a `LotaProvider` object to configure a Terraform Provider.
An Operator that watches the `LotaProvider` resources to create a `CRD` per resource exposed by the Terraform provider, then watches all resources created using this new `CRDs` (not working yet) and create the resources in the external APIs using Terraform.

Status
------

For now this is just an ugly Proof Of Concept written in Shell script.
`CRD` creation works, but I couldn't get the resources created using this CRDs watched by the operator yet.

Example
-------

Define a provider to manage an external PostgreSQL cluster:

```yaml
apiVersion: lotaprovider.lota-operator.io/v1alpha1
kind: LotaProvider
metadata:
  name: pg-prod
spec:
  name: postgresql # the name of the Terraform Provider to use
  version: 1.3.0   # the version of the Terraform Provider to use
  schema:          # the attributes of the Terraform Provider
    - name: host
      value: postgres
    - name: password
      value: mypassword
```

How to test
-----------

Start the docker composition:

```shell
$ docker-compose up -d
```

This will create a k3s cluster for testing.

Register the CRD with the Kubernetes apiserver:

```shell
$ KUBECONFIG=./kubeconfig.yaml kubectl apply -f deploy/crds/lotaprovider.lota-operator.io_lotaproviders_crd.yaml
```

Run the operator:

```shell
$ export OPERATOR_NAME=lota-operator
$ operator-sdk up local --namespace=default --kubeconfig=./kubeconfig.yaml
```

Create a LotaProvider Custom Resource:

```shell
$ KUBECONFIG=./kubeconfig.yaml kubectl apply -f deploy/crds/lotaprovider.lota-operator.io_v1alpha1_lotaprovider_cr.yaml
```

Tou should see the operator create new CRD for all resources managed by the Terraform Provider defined by the `LotaOperator` resource.
