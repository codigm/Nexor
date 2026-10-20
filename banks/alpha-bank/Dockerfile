FROM eclipse-temurin:17-jdk-alpine AS build
WORKDIR /workspace/app

# Since we use a parent pom, we'll copy the whole banks folder during build in the real context
# For this dockerfile to work seamlessly, it should be built from the /banks directory
COPY pom.xml .
COPY nexor-common/pom.xml nexor-common/
COPY nexor-common/src nexor-common/src
COPY helios-bank/pom.xml helios-bank/
COPY helios-bank/src helios-bank/src
COPY orion-bank/pom.xml orion-bank/
COPY vesper-bank/pom.xml vesper-bank/

# Create the maven wrapper structure manually to allow maven build (or assume it exists)
RUN apk add --no-cache maven
RUN mvn clean package -DskipTests -pl helios-bank -am

FROM eclipse-temurin:17-jre-alpine
VOLUME /tmp
COPY --from=build /workspace/app/helios-bank/target/*.jar app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
