FROM eclipse-temurin:17-jdk-alpine AS build
WORKDIR /workspace/app

COPY pom.xml .
COPY olympus-common/pom.xml olympus-common/
COPY olympus-common/src olympus-common/src
COPY gamma-bank/pom.xml gamma-bank/
COPY gamma-bank/src gamma-bank/src
COPY alpha-bank/pom.xml alpha-bank/
COPY beta-bank/pom.xml beta-bank/

RUN apk add --no-cache maven
RUN mvn clean package -DskipTests -pl gamma-bank -am

FROM eclipse-temurin:17-jre-alpine
VOLUME /tmp
COPY --from=build /workspace/app/gamma-bank/target/*.jar app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
