FROM ubuntu:22.04

WORKDIR /work

RUN apt-get update && apt-get install -y openssh-server openssh-client docker.io

RUN mkdir /root/.ssh && ln -s /shared/ssh/id_rsa.pub /root/.ssh/authorized_keys

RUN mkdir -p /etc/docker/certs.d/registry:4443 && ln -s /shared/certs/domain.crt /etc/docker/certs.d/registry:4443/ca.crt

COPY boot.sh .

RUN chmod +x ./boot.sh

HEALTHCHECK --interval=10s CMD pgrep dockerd

CMD [ "./boot.sh" ]
