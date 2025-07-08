# friendly-fiesta
BitTorrent style file distribution using Kubernetes clusters and Go concurrency.

## upload
[x] Route to upload, split, and create file metadata

## serve
[x] Route to serve file metadata
[x] Route to serve all files metadata

## download
[x] Route to download chunks 


## install

docker build -t torrent-service .

docker run -p 8080:8080 \
  -v "$(pwd)/torrent-data:/data" \
  --name torrent1 \
  torrent-service

docker build -t peterjbishop/torrent-service:latest .
docker push peterjbishop/torrent-service:latest

minikube start
kubectl apply -f torrent-service-statefulset.yaml
minikube service torrent-service-nodeport --url
minikube stop

