FROM docker.io/eclipse-temurin:25-jre-alpine

ARG VERSION

RUN mkdir /app

ADD https://github.com/applejuicenetz/collector/releases/download/${VERSION}/AJCollector.jar /app/AJCollector.jar

WORKDIR /app

CMD ["java", "-Duser.home=/app", "-jar", "/app/AJCollector.jar"]
