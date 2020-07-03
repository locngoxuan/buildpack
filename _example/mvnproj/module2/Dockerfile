FROM xuanloc0511/wesapp-openjdk14:1.4.0

RUN mkdir -p /application \
	mkdir -p /application/hotfix
ADD libs /application/libs
ADD dist /application/dist
ADD application.yml /application/application.yml

ENV JVM_OPTS=""
ENV APP_OPTS=""
ENV APP_ARGS=""
ENV WESCTL=""
CMD /usr/bin/wesctl --app=java --config=/application/application.yml $WESCTL $JVM_OPTS $APP_OPTS $APP_ARGS
