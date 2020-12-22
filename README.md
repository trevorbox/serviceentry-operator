# serviceentry-operator

```sh
export USERNAME=trevorbox

make docker-build IMG=quay.io/$USERNAME/serviceentry-operator:v0.0.1

docker logout quay.io
docker login quay.io

make docker-push IMG=quay.io/$USERNAME/serviceentry-operator:v0.0.1

make deploy IMG=quay.io/$USERNAME/serviceentry-operator:v0.0.1
```
