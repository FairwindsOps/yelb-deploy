# Yelb Deploy

A GitOps repository for deploying the yelb demo app

## Feature Branches

generate lastDeployed: `date -u +"%Y-%m-%dT%H:%M:%SZ"`

find empty feature namespaces

```
kubectl get ns -l fairwinds.com/environment=yelb-feature --no-headers -o custom-columns=":metadata.name"  | xargs -I{} kubectl  get all -n {} 2>&1 | grep "No" | cut -d " " -f 5
```
