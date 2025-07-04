source ./config.sh

docker stop $DB_CONTAINER_NAME
docker rm $DB_CONTAINER_NAME

cd db
./rebuild.sh
cd ..

docker run -d \
	--name $DB_CONTAINER_NAME \
	-p "$DB_PORT:$DB_PORT" \
	-e POSTGRES_PASSWORD=$DB_PWD \
	-e POSTGRES_USER=$DB_USER \
	-e POSTGRES_DB=$DB_SCHEMA \
	--network net \
	-v /home/gchalakov/services/spaceresearch/db/pgdata:/var/lib/postgresql/ \
	spaceresearch-postgres
