# injecto
[![CircleCI](https://circleci.com/gh/hightouchio/injecto/tree/master.svg?style=svg)](https://circleci.com/gh/hightouchio/injecto/tree/master)
![Go Version](https://shields.io/github/go-mod/go-version/hightouchio/injecto)
[![Docker Hub](https://shields.io/docker/v/hightouchio/injecto?label=docker&sort=semver)](https://hub.docker.com/r/hightouchio/injecto)

injecto dynamically transfers the contents of an image into a running container, regardless of the base image of the running container.

Imagine a scenario where you're running a `FROM scratch` image but want to exec into the container for debugging purposes. Using injecto you could inject the contents of Alpine or any other base image into the container and then have access to its shell and package manager.

## Usage

injecto can be run as either a standalone binary or via Docker.

```
injecto <image> <container>
```

```
docker run -it -v /var/run/docker.sock:/var/run/docker.sock -v /tmp:/tmp hightouchio/injecto <image> <container>
```

## Example

First, let's run a container built from a `FROM scratch` image. `joshwget/sleeper` only contains a simple Go binary that spins forever.

```
> docker run -d joshwget/sleeper /sleeper
79f7232e5f30
```

Trying to exec into this container will result in an error because the base image contains no shell.

```
> docker exec -it 79f7232e5f30 sh
exec: \"sh\": executable file not found in $PATH
```

Now let's use injecto to inject Alpine.

```
> docker run -it -v /var/run/docker.sock:/var/run/docker.sock -v /tmp:/tmp hightouchio/injecto alpine 79f7232e5f30
```

If you try to exec again, you'll be in a shell!

```
> docker exec -it 79f7232e5f30 sh
/ #
```
