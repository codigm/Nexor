FROM eclipse-temurin:17-jdk-alpine AS build
WORKDIR /workspace/app

COPY pom.xml .
COPY nexor-common/pom.xml nexor-common/
COPY nexor-common/src nexor-common/src
COPY orion-bank/pom.xml orion-bank/
COPY orion-bank/src orion-bank/src
COPY helios-bank/pom.xml helios-bank/
COPY vesper-bank/pom.xml vesper-bank/

RUN apk add --no-cache maven
RUN mvn clean package -DskipTests -pl orion-bank -am

FROM eclipse-temurin:17-jre-alpine
VOLUME /tmp
COPY --from=build /workspace/app/orion-bank/target/*.jar app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
