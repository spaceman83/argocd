#!/usr/bin/env bash
set -eEuox pipefail

: "${HELM:=helm}"

# Whether to include Pod Security Policy (PSP) related resources in the RBAC output.
# Empty string means DO include PSP resources. Any other value means do NOT include PSP resources.
: "${DO_NOT_INCLUDE_POD_SECURITY_POLICY_RESOURCES:=""}"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
pushd "$SCRIPT_DIR" &>/dev/stderr

options=(
  --namespace rook-ceph
  --set crds.enabled=false
  --set csi.csiAddons.enabled=true
)
if [[ -z "${DO_NOT_INCLUDE_POD_SECURITY_POLICY_RESOURCES}" ]]; then
  options+=(--set pspEnable=true)
else
  options+=(--set pspEnable=false)
fi

echo "generating Helm template with options: ${options[*]}" &>/dev/stderr

${HELM} template ../../deploy/charts/rook-ceph "${options[@]}" | ./keep-rbac-yaml.sh

popd &>/dev/stderr
