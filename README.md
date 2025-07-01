# friendly-fiesta
BitTorrent style file distribution using Kubernetes clusters and Go concurrency.

## the challenge

Create a Docker container to act as seeder, then multiply them within a Kubernetes cluster. 

Each container would have a full copy of the file and serve it to the external peer.  

If a new file is uploaded each of the pods would begin downloading from the seeder pod until they could also start downloading from each other until they were also seeders.

Create an external and persistant Peer Client which would request file chucks from the Pods in the cluster to construct a file.

## structure

Seeder Client (for uploads) -> [ Kubernetes Cluster of Seeders ] -> Peer Client (for downloads)




