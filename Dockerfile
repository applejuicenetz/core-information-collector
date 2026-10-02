FROM docker.io/eclipse-temurin:25-jre-alpine

RUN mkdir /app

COPY target/AJCollector.jar /app/AJCollector.jar

WORKDIR /app

CMD ["java", "-Duser.home=/app", "-jar", "/app/AJCollector.jar"]
