source ../config.sh

docker buildx build -t $DB_IMAGE_NAME .
