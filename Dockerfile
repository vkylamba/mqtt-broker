FROM golang:1.17 as gobase

# create directory for the application user
ENV APP_HOME=/home/application/
RUN mkdir -p $APP_HOME

# create application user/group first, to be consistent throughout docker variants
RUN set -x \
    && addgroup --system --gid 1001 application \
    && adduser --system --ingroup application --home $APP_HOME --gecos "application user" --shell /bin/false --uid 1001 application

RUN chown -R 1001:0 $APP_HOME

WORKDIR $APP_HOME

COPY ./src $APP_HOME

RUN go build -o /bin/app-server main.go

FROM emqx/emqx:5.7.0
USER root

COPY --from=gobase /bin/app-server app-server
COPY ./src/start_services.sh start_services.sh
USER emqx
EXPOSE 1883 8083 8084 8883 18083 4370 5369

ENTRYPOINT [ "/bin/bash", "start_services.sh"]
