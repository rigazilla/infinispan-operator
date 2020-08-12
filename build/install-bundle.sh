#!/usr/bin/env bash

OPERATOR_SOURCE_INSTALL_FILES=(
  "deploy/service_account.yaml"
  "deploy/role.yaml"
  "deploy/clusterrole.yaml"
  "deploy/role_binding.yaml"
  "deploy/clusterrole_binding.yaml"
  "deploy/crds/infinispan.org_infinispans_crd.yaml"
  "deploy/crds/infinispan.org_caches_crd.yaml"
  "deploy/crds/infinispan.org_infinispanbackups_crd.yaml"
  "deploy/operator.yaml"
)

OPERATOR_BUNDLE_FILE="deploy/operator-install.yaml"
FIRST_FILE_FLAG="true"

echo "#####################################################################" > "${OPERATOR_BUNDLE_FILE}"
# shellcheck disable=SC2129
echo "### THIS IS AUTOGENERATED INFINISPAN OPERATOR INSTALL BUNDLE FILE ###" >> "${OPERATOR_BUNDLE_FILE}"
echo "###                  PLEASE DON'T EDIT THIS FILE                  ###" >> "${OPERATOR_BUNDLE_FILE}"
echo "#####################################################################" >> "${OPERATOR_BUNDLE_FILE}"
echo "" >> "${OPERATOR_BUNDLE_FILE}"

for FILE in "${OPERATOR_SOURCE_INSTALL_FILES[@]}"; do
  if [[ "$FIRST_FILE_FLAG" == "true" ]]; then
    FIRST_FILE_FLAG="false"
  else
    echo $'\n---\n' >> "${OPERATOR_BUNDLE_FILE}"
  fi

  cat "$FILE" >> "${OPERATOR_BUNDLE_FILE}"
done
