#!/bin/bash

set -e

if [[ -z $APP_NAME ]]; then
  echo "No APP_NAME env variable specified"
  exit 1
fi
if [[ -z $DOMAIN ]]; then
  echo "No DOMAIN env variable specified"
  exit 1
fi
if [[ -z $CF_ORG ]]; then
  echo "No CF_ORG env variable specified"
  exit 1
fi
if [[ -z $CF_SPACE ]]; then
  echo "No CF_SPACE env variable specified"
  exit 1
fi
if [[ -z $CF_USER ]]; then
  echo "No CF_USER env variable specified"
  exit 1
fi
if [[ -z $CF_PASS ]]; then
  echo "No CF_PASS env variable specified"
  exit 1
fi
if [[ -z $CF_ENDPOINT ]]; then
  echo "No CF_ENDPOINT env variable specified"
  exit 1
fi

cd app

cf login -a "${CF_ENDPOINT}" -u "${CF_USER}" -p "${CF_PASS}" -o "${CF_ORG}" -s "${CF_SPACE}"
cf push "${APP_NAME}" -d "${DOMAIN}"
