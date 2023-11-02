FROM golang:1.20.10-bullseye AS build

WORKDIR /app

COPY . /app

RUN apt-get update

RUN apt-get install -y libdevmapper-dev btrfs-progs libgpgme11-dev libglib2.0-dev libostree-dev buildah

RUN GOOS=linux go build -tags "containers_image_openpgp exclude_graphdriver_btrfs" -o interceptor cmd/interceptor/main.go

FROM debian:bullseye-slim

RUN apt-get update

RUN apt-get install -y libdevmapper-dev btrfs-progs libgpgme11-dev libglib2.0-dev libostree-dev buildah fuse-overlayfs

WORKDIR /app

COPY storage.conf /etc/containers/storage.conf

COPY --from=build /app/interceptor .

EXPOSE 8001

CMD ["./interceptor"]
