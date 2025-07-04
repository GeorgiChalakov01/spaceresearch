source ../config.sh

cd app
./compile.sh
cd ..

docker buildx build -t $BE_IMAGE_NAME .
