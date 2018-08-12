# `mgo` gitops repo structure

```
GITOPS="~/mygitops-repo/"

$GITOPS/installations/<cluster_name>/<helm_install_name>-values.yaml

- <cluster_name> is the name of your cluster as shown in your `kubectl config get-contexts`
- <helm_install_name> is the name you used when you installed the chart via helm
  Eg: `helm install stable/redis --name myredis --values ./installations/minikube/myredis-values.yaml`
```

### Value files structure

`*-values.yaml` files are basically your usual values file you used in helm.

Additionally, you need to add a header to your value files that will provide
hints to `mygitops` on how it should update things.

#### Value file headers

```yaml
# This is the header, everything goes under the `__mygitops` key
__mygitops:
  # Chart name like repo/chart
  chart: stable/redis

  # Chart version
  version: 0.1.0

  # name used by helm when installing this chart
  name: redis-cache

  # namespace used by helm when installing this chart
  namespace: app

  # images used in your chart. mygitops will -only- update those images when
  # there's an incoming deploy trigger
  images:
    github.com/myself/my-cool-service: &my-cool-service
      repository: "myself/my-cool-service"
      tag: "latest"

# Your normal values.yaml file content after this

........ snip ......

my-cool-service:
  image:
    # We do a YAML merge (http://yaml.org/type/merge.html)
    # It's the non-invasive way of keeping all automatic updates constrained to
    # the __mygitops section, not have to magically figure out what goes where
    <<: *my-cool-service
    pullPolicy: Always

........ snip ......
```

