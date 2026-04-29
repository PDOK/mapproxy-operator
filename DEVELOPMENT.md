## Local testing

### First time setup
The manager and webhook require valid certificates and keys. These are named `ca.crt`, `tls.crt` and `tls.key` and need to be placed in `config/manager/certs` (which is ignored in git). In `example_certs.zip` you have valid example files you can use.

### Deploying to the cluster
- Start an empty cluster
- Apply the CRD at `config/crd/bases/pdok.nl_wmts.yaml` manually to the cluster. It should show up with version v2.
- Build and push the controller to the cluster using `build-and-push-locally.sh <controller-version>`
- Check that `mapproxy-operator-controller-manager` is added as a deployment and wait until the pods have started
- Apply v2 WMTS CR's to the cluster, these should generate deployments
