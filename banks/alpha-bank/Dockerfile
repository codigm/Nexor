FROM eclipse-temurin:17-jdk-alpine AS build
WORKDIR /workspace/app

# Since we use a parent pom, we'll copy the whole banks folder during build in the real context
# For this dockerfile to work seamlessly, it should be built from the /banks directory
COPY pom.xml .
COPY olympus-common/pom.xml olympus-common/
COPY olympus-common/src olympus-common/src
COPY alpha-bank/pom.xml alpha-bank/
COPY alpha-bank/src alpha-bank/src
COPY beta-bank/pom.xml beta-bank/
COPY gamma-bank/pom.xml gamma-bank/

# Create the maven wrapper structure manually to allow maven build (or assume it exists)
RUN apk add --no-cache maven
RUN mvn clean package -DskipTests -pl alpha-bank -am

FROM eclipse-temurin:17-jre-alpine
VOLUME /tmp
COPY --from=build /workspace/app/alpha-bank/target/*.jar app.jar
ENTRYPOINT ["java","-jar","/app.jar"]
