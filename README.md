# ntc-gsuperwss
**ntc-gsuperwss** is an example golang websocket server handle more than 1 million websockets connections, lightweight, high performance, large scale.

## Install dependencies
```shell
# install library dependencies
#make deps
go mod download

# update go.mod file
go mod tidy
```

## Build
```shell
export GO111MODULE=on
make build
```

## Start with mode: development | production | staging | test
```shell
# development
make run

# production
make run-prod

# staging
make run-stag

# test
make run-test
```

## Clean build
```shell
make clean
```


## License
This code is under the [Apache License v2](https://www.apache.org/licenses/LICENSE-2.0).  
