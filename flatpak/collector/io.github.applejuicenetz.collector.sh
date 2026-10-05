#!/bin/sh

if [ -e /tmp/ajcollector.lock ] && kill -0 "$(cat /tmp/ajcollector.lock)" 2>/dev/null; then
  exit 0
fi

JAVA_ARGS="-Xmx512m -XX:+UseSerialGC -Djava.net.preferIPv4Stack=true -Dsun.java2d.xrender=false"

cd /app/share/io.github.applejuicenetz.collector/ || exit 1

exec java --enable-preview --enable-native-access=ALL-UNNAMED $JAVA_ARGS -jar /app/share/io.github.applejuicenetz.collector/AJCollector.jar "$@"
