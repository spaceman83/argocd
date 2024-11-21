Fix for velero helm chart crd.
Bump version depending on version installed/deployed.
```shell
# Download velero binary
wget https://github.com/vmware-tanzu/velero/releases/download/v1.5.2/velero-v1.5.2-linux-amd64.tar.gz

# Unpack
tar -xvf velero-v1.5.2-linux-amd64.tar.gz

# 
./velero install --crds-only --dry-run -o yaml | kubectl apply -f -
```

Manually sync each resource in argoCD.
