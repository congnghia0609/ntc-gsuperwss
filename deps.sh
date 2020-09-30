#!/bin/bash
# Author:       nghiatc
# Email:        congnghia0609@gmail.com

# source /etc/profile

echo "Install library dependencies..."
go get -u github.com/congnghia0609/ntc-gconf
go get -u github.com/gorilla/websocket
go get -u github.com/spf13/viper
go get -u github.com/tools/godep
go get -u github.com/gorilla/mux
go get -u github.com/sirupsen/logrus
go get -u github.com/natefinch/lumberjack
go get -u github.com/satori/go.uuid
go get -u github.com/gobwas/ws

echo "Install dependencies complete..."
