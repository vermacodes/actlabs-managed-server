#FROM ubuntu:22.04
FROM actlab.azurecr.io/repro_base

WORKDIR /app

ADD actlabs-managed-server ./

EXPOSE 80/tcp
EXPOSE 443/tcp

ENTRYPOINT [ "/bin/bash", "-c", "export PORT='80' && ./actlabs-managed-server" ]