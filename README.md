# Wildfly Operator

This Wildfly Operator is not intended for production use. It is a self-training
lab meant to explore the capabilities of operators and the 
![operator-sdk](http://github.com/operator-framework/operator-sdk).

## Building
To build the operator use the `operator-sdk` CLI tool.
```
$ operator-sdk build quay.io/gbsalinetti/wildfly-operator
```

The resulting image must be pushed to public registry (change registry name 
and organization name accordigly):
```
$ docker push quay.io/gbsalinetti/wildfly-operator
```

**IMPORTANT**: Don't forget to update the *image* field in the 
`deploy/operator.yaml` file accordingly with the correct registry and
organization names.
```
sed -i 's|quay.io/gbsalinetti/wildfly-operator|quay.io/example/wildfly-operator|g' deploy/operator.yaml
```

## Deploying
First, create a new namespace to keep the resources organized:
```
$ kubectl create ns wildfly
```

Then, deploy the manifest related to roles, role bindings, and service 
accounts:
```
$ kubectl create -f deploy/service_account.yaml -n wildfly
$ kubectl create -f deploy/role.yaml -n wildfly
$ kubectl create -f deploy/role_binding.yaml -n wildfly
```

Create the CRD for the Wildfly resource:
```
$ kubectl create -f deploy/crds/wildfly_v1alpha1_wildfly_crd.yaml
```

Finally, deploy the operator:
```
$ kubectl create -f deploy/operator.yaml -n wildfly
```

Monitor the operator pod bootstrap (it will take more time at first run to
pull the operator image):
```
$ kubectl get pods -n wildfly
NAME                               READY     STATUS    RESTARTS   AGE
wildfly-operator-65b784bbb-zckg6   1/1       Running   0          2m
```

The deployment will be managed by the operator accordingly to the fields 
passed in the custom resource. The following example shows the default
fields. 
Change to **image** and **version** fields accordingly to the correct image
(most probably something with an application deployed in it). 
The **size** field is the number of desired replicas.
The **cmd** field is an array of commands and parameters to be executed. If
missing, the operator will fill it up with a default command to run a 
default instance listening on port 0.0.0.0.
```
apiVersion: wildfly.extraordy.com/v1alpha1
kind: Wildfly
metadata:
  name: example-wildfly
spec:
  size: 1
  image: "docker.io/jboss/wildfly"
  version: "14.0.1.Final"
  cmd:
    - "/opt/jboss/wildfly/bin/standalone.sh"
    - "-b"
    - "0.0.0.0"
```

After successful deployment of the operator the custom resource can be 
deployed:
```
$ kubectl create -f deploy/crds/wildfly_v1alpha1_wildfly_cr.yaml -n wildfly
```

This will create a new Wildfly resource:
```
$ kubectl get Wildfly -n wildfly
NAME              AGE
example-wildfly   18h
```

Deployments and services can be monitored as usual:
```
$ kubectl get pods -n wildfly
NAME                               READY     STATUS    RESTARTS   AGE
example-wildfly-5c49b588f7-n6rqn   1/1       Running   0          2m
wildfly-operator-65b784bbb-zckg6   1/1       Running   0          10m

$ kubectl get svc -n wildfly
NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
example-wildfly   ClusterIP   10.102.146.54   <none>        8080/TCP   2m
```
## TODO
- Add service port management in the custom resource
- Add configmaps for the Wildfly config files.
- Use Go template to change config files contents (datasources could be the 
  first try).

## Contributing
Please feel free to file Issues to improve usability and PRs to add new features.
### License
Wildfly-Operator is under the Apache 2.0 license.
