#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

OS_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${OS_ROOT}/hack/common.sh"

cd "${OS_ROOT}"

echo "===== Verifying API Swagger Spec ====="

SPECROOT_REL="api/swagger-spec"
SPECROOT="${OS_ROOT}/${SPECROOT_REL}"
REL_TMP_PATH="_output/verify-generated-swagger-spec"
TMP_SPECROOT="${OS_ROOT}/${REL_TMP_PATH}/${SPECROOT_REL}"

echo "Generating a fresh spec..."
if ! output=`${OS_ROOT}/hack/update-generated-swagger-spec.sh ${REL_TMP_PATH} 2>&1`
then
	echo "FAILURE: Generation of fresh spec failed:"
	echo "$output"
  exit 1
fi

echo "Diffing current spec against freshly generated spec..."
ret=0
diff -Naupr -I 'Auto generated by' "${SPECROOT}" "${TMP_SPECROOT}" || ret=$?
rm -rf "${TMP_SPECROOT}"
if [[ $ret -eq 0 ]]
then
  echo "SUCCESS: Swagger spec up to date."
else
  echo "FAILURE: Swagger spec is out of date. Please run hack/update-generated-swagger-spec.sh"
  exit 1
fi