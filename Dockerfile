#FROM ubuntu:22.04
FROM actlab.azurecr.io/repro_base

WORKDIR /app

ADD actlabs-managed-server ./

EXPOSE 8883/tcp

ENTRYPOINT [ "/bin/bash", "-c", "./actlabs-managed-server" ]