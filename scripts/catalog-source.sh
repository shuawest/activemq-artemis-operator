#!/bin/bash

if [[ "$OSTYPE" == "darwin"* ]]; then
  PATH="/usr/local/opt/gnu-sed/libexec/gnubin:$PATH"
  WHICHSED=$(which sed)
  echo "Using '$WHICHSED' sed on OSX"
fi

CATALOG_NS="openshift-marketplace"

# if [[ -z ${1} ]]; then
#     CATALOG_NS="operator-lifecycle-manager"
# else
#     CATALOG_NS=${1}
# fi

CSV=`cat deploy/olm-catalog/activemq-artemis-operator/0.13.0-continuity/activemq-artemis-operator.v0.13.0-continuity.clusterserviceversion.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`
CRD=`cat deploy/crds/broker_v2alpha2_activemqartemis_crd.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`
CRDActivemqartemisaddress=`cat deploy/crds/broker_v2alpha1_activemqartemisaddress_crd.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`
CRDActivemqartemisscaledown=`cat deploy/crds/broker_v2alpha1_activemqartemisscaledown_crd.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`
CRDActivemqartemiscontinuity=`cat deploy/crds/continuity_v2alpha2_activemqartemiscontinuity_crd.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`
PKG=`cat deploy/olm-catalog/activemq-artemis-operator/activemq-artemis.package.yaml | sed -e 's/^/          /' | sed '0,/ /{s/          /        - /}'`

cat << EOF > deploy/catalog_resources/catalog-source.yaml
apiVersion: v1
kind: List
items:
  - apiVersion: v1
    kind: ConfigMap
    metadata:
      name: activemq-artemis-resources
      namespace: ${CATALOG_NS}
    data:
      clusterServiceVersions: |
${CSV}
      customResourceDefinitions: |
${CRD}
${CRDActivemqartemisaddress}
${CRDActivemqartemisscaledown}
${CRDActivemqartemiscontinuity}
      packages: >
${PKG}

  - apiVersion: operators.coreos.com/v1alpha1
    kind: CatalogSource
    metadata:
      name: activemq-artemis-resources
      namespace: ${CATALOG_NS}
    spec:
      configMap: activemq-artemis-resources
      displayName: ActiveMQ Artemis Operator with Continuity
      publisher: Red Hat
      sourceType: internal
    status:
      configMapReference:
        name: activemq-artemis-resources
        namespace: ${CATALOG_NS}
EOF
