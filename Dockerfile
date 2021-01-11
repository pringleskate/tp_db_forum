FROM golang AS builder

WORKDIR /usr/src/app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN ls
RUN go build -v -work -o tp_db_forum cmd/server.go

FROM ubuntu:20.04
ENV PGVER 12
ENV PORT 5000
ENV POSTGRES_HOST localhost
ENV POSTGRES_PORT 5432
ENV POSTGRES_DB tp_forum
ENV POSTGRES_USER forum_user
ENV POSTGRES_PASSWORD 1221
EXPOSE $PORT

RUN apt-get -y update && apt-get install -y tzdata

ENV TZ=Russia/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get -y update && apt-get install -y postgresql-$PGVER

#RUN apt-get -y update && apt-get install -y --no-install-recommends apt-utils
#RUN apt-get install -y postgresql-$PGVER

USER postgres
COPY /init.sql .

RUN service postgresql start &&\
    psql --command "CREATE USER forum_user WITH SUPERUSER PASSWORD '1221';" &&\
    createdb -O forum_user tp_forum &&\
    psql tp_forum -a -f init.sql &&\
    service postgresql stop

USER root
RUN echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf &&\
    echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "shared_buffers=256MB" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "full_page_writes=off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "unix_socket_directories = '/var/run/postgresql'" >> /etc/postgresql/$PGVER/main/postgresql.conf

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

COPY --from=builder /usr/src/app/tp_db_forum .
CMD service postgresql start && ./tp_db_forum
