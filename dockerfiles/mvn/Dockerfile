FROM anapsix/alpine-java:8u202b08_jdk

MAINTAINER LOC NGO <locngo@fortna.com>

RUN \
  mkdir -p /opt \
  mkdir -p /working

ADD /tmp/maven /opt/maven
RUN mkdir -p /root/.m2
RUN mkdir -p /root/.m2/repository
ADD settings.xml /root/.m2/settings.xml

ENV PATH "$PATH:/opt/maven/bin"
ENV M2_HOME "/opt/maven"
ENV MAVEN_HOME "/opt/maven"

WORKDIR /working
CMD ["mvn"]