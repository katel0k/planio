FROM postgres

ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DATABASE=planbook
ENV POSTGRES_PORT=5432

WORKDIR /

EXPOSE 5432

COPY ddl.sql /docker-entrypoint-initdb.d/
