
### sailmaker-release-api
This app acts as the glue between sailmaker deployer proxy and kubernetes. It has endpoints to read deployment information as well as to read and create ReleaseRequest CRD instances.

#### Running App

```
go mod vendor
go build
./kube-sailmaker-release serve --port=6264
```

### List Deployments
GET /api/deployments?namespace=default

Output:
```
{
  data: [{
    "namespace": "default",
    "kind": "deployment",
    "name": "nginx-app",
    "image": "nginx:latest",
    "replicas": 2 
  }]
}
```
### List Custom Resource Definitions
GET /api/crds
```
{
    "data": [
        {
            "name": "cities.world.io",
            "group": "world.io",
            "resource-type": "cities",
            "kind": "City",
            "version": "v1alpha1",
            "link": "/api/crd-instances?resource-group=world.io&resource-type=cities&resource-version=v1alpha1"
        }
    ]
}
```

### List Custom Resource Instances
GET /api/crd-instances?resource-group=world.io&resource-type=cities&resource-version=v1alpha1

```
{
    "data": [
        {
            "namespace": "default",
            "name": "s-cities",
            "group": "world.io",
            "version": "v1alpha1",
            "kind": "City",
            "link": "/api/crd-instance?resource-group=world.io&resource-type=cities&resource-version=v1alpha1&namespace=default&resource-name=s-cities"
        }
    ]
}   
```

### Get Custom Resource Instance
GET /api/crds?namespace=default&resource-type=cities&resource-group=world.io&resource-version=v1alpha1&resource-name=s-cities

```
{
    "data": {
        "spec": {
            "apps": [
                {
                    "country": "Australia",
                    "name": "sydney"
                },
                {
                    "country": "USA",
                    "name": "san francisco"
                }
            ]
        },
        "metadata": {
            "annotations": {
                "living-expense": "high",
                "weather": "great"
            },
            "creationTimestamp": "2020-04-19T00:22:21Z",
            "generation": 1,
            "labels": {
                "starts-with": "S"
            },
            "name": "s-cities",
            "namespace": "default",
            "resourceVersion": "176136",
            "selfLink": "/apis/world.io/v1alpha1/namespaces/default/cities/s-cities",
            "uid": "d5a9c026-81d3-11ea-b6f1-02430e0005fc"
        }
    }
}
```
