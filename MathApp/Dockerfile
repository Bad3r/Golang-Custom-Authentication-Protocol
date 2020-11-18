FROM golang:1.15

# Install beego & bee
RUN go get github.com/astaxie/beego
RUN go get github.com/beego/bee

ENV APP_USER www-data
ENV APP_HOME /go/src/mathapp

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

# RUN mkdir -p $APP_HOME && chown -R ${USER_ID}:${GROUP_ID} $APP_HOME

USER $APP_USER

RUN mkdir -p $APP_HOME

COPY src $APP_HOME

WORKDIR $APP_HOME

RUN go build -o $APP_HOME/main $APP_HOME/main.go

RUN ./$APP_HOME/main