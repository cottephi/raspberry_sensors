#!/bin/bash

LOG_PATH=${HOME}/logs/raspberry_sensors
LOCAL_LOKI_DIR=${HOME}/loki

# Create the necessary directories for Loki runtime data
mkdir -p ${LOCAL_LOKI_DIR}/boltdb-shipper-compactor
mkdir -p ${LOCAL_LOKI_DIR}/index
mkdir -p ${LOCAL_LOKI_DIR}/boltdb-cache
mkdir -p ${LOCAL_LOKI_DIR}/chunks
mkdir -p ${LOCAL_LOKI_DIR}/wal

# Set proper permissions (optional, but recommended)
chmod -R 777 "${LOCAL_LOKI_DIR}/boltdb-shipper-compactor" &>/dev/null
chmod -R 777 "${LOCAL_LOKI_DIR}/index" &>/dev/null
chmod -R 777 "${LOCAL_LOKI_DIR}/boltdb-cache" &>/dev/null
chmod -R 777 "${LOCAL_LOKI_DIR}/chunks" &>/dev/null
chmod -R 777 "${LOCAL_LOKI_DIR}/wal" &>/dev/null

chown $(whoami):$(whoami) -R ${LOCAL_LOKI_DIR} &>/dev/null

if ! [ -f .env ] ; then
  echo "LOG_PATH=${LOG_PATH}" > .env
  echo "LOCAL_LOKI_DIR=${LOCAL_LOKI_DIR}" >> .env
else
  if grep -q "LOCAL_LOKI_DIR=" ".env" ; then
    sed -i '/LOCAL_LOKI_DIR=/c\LOCAL_LOKI_DIR='${LOCAL_LOKI_DIR} .env
  else
    echo "LOCAL_LOKI_DIR=${HOME}/loki" >> .env
  fi

  if grep -q "LOG_PATH=" ".env" ; then
    sed -i '/LOG_PATH=/c\LOG_PATH='${LOG_PATH} .env
  else
    echo "LOG_PATH=${LOG_PATH}" >> .env
  fi
fi

echo "Loki directories created in ${HOME}/loki. Local .env file created/updated."
