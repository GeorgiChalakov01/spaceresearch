source ../config.sh

docker exec -it $DB_CONTAINER_NAME psql -h localhost -p 5432 -d $DB_SCHEMA -U $DB_USER
