FROM docker.io/eclipse-temurin:17-jre-jammy

ARG VERSION

RUN mkdir /app

ADD https://github.com/applejuicenetz/core-information-collector/releases/download/${VERSION}/AJCollector.jar /app/AJCollector.jar

CMD ["java", "-Duser.home=/app", "-jar", "/app/AJCollector.jar"]
