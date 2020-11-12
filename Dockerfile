FROM openjdk:8-jre-alpine

ARG BUILD_DATE
ARG VCS_REF

ADD https://github.com/applejuicenetz/core-information-collector/releases/latest/download/core-information-collector.jar /core-information-collector.jar

CMD ["java", "-jar", "/core-information-collector.jar"]

LABEL org.opencontainers.image.vendor="appleJuiceNET" \
      org.opencontainers.image.url="https://applejuicenet.cc" \
      org.opencontainers.image.created=${BUILD_DATE} \
      org.opencontainers.image.revision=${VCS_REF} \
      org.opencontainers.image.source="https://github.com/applejuicenetz/core-information-collector"
