FROM openjdk:8-jre-alpine

ADD https://github.com/applejuicenet/core-information-collector/releases/latest/download/core-information-collector.jar /core-information-collector.jar

CMD ["java", "-jar", "/core-information-collector.jar"]
