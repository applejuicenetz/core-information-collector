FROM eclipse-temurin:11-jre-focal

ARG BUILD_DATE
ARG VCS_REF

RUN mkdir /app

ADD https://github.com/applejuicenetz/core-information-collector/releases/latest/download/AJCollector.jar /app/AJCollector.jar

CMD ["java", "-Duser.home=/app", "-jar", "/app/AJCollector.jar"]

LABEL org.opencontainers.image.vendor="appleJuiceNET" \
      org.opencontainers.image.url="https://applejuicenet.cc" \
      org.opencontainers.image.created=${BUILD_DATE} \
      org.opencontainers.image.revision=${VCS_REF} \
      org.opencontainers.image.source="https://github.com/applejuicenetz/core-information-collector"
