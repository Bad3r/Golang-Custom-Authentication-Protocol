FROM golang:1.15

ENV APP_USER www-data
ENV APP_HOME /go/src/web_application

ARG USER_ID
ARG GROUP_ID

RUN userdel -f $APP_USER &&\
    if getent group $APP_USER ; then groupdel $APP_USER; fi &&\
    groupadd -g ${GROUP_ID} $APP_USER &&\
    useradd -l -u ${USER_ID} -g $APP_USER $APP_USER &&\
    install -d -m 0755 -o $APP_USER -g $APP_USER /home/$APP_USER &&\
    chown --changes --silent --no-dereference --recursive \
          --from=33:33 ${USER_ID}:${GROUP_ID} \
        /home/$APP_USER


USER $APP_USER

RUN mkdir -p $APP_HOME

COPY . $APP_HOME

WORKDIR $APP_HOME

RUN go build -o main main.go