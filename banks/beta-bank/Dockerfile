FROM eclipse-temurin:17-jdk-alpine AS build
WORKDIR /workspace/app

COPY pom.xml .
COPY olympus-common/pom.xml olympus-common/
COPY olympus-common/src olympus-common/src
COPY beta-bank/pom.xml beta-bank/
COPY beta-bank/src beta-bank/src
COPY alpha-bank/pom.xml alpha-bank/
COPY gamma-bank/pom.xml gamma-bank/

RUN apk add --no-cache maven
RUN mvn clean package -DskipTests -pl beta-bank -am

FROM eclipse-temurin:17-jre-alpine
VOLUME /tmp
COPY --from=build /workspace/app/beta-bank/target/*.jar app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
