#!/bin/sh
minio server /data --console-address :9001 &
sleep 5
/bin/mc --insecure alias set myminio http://localhost:9000 "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}"
/bin/mc mb myminio/documents
tail -f /dev/null
